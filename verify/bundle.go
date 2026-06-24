// SPDX-License-Identifier: Apache-2.0

package verify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protocommon "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	protorekor "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"

	"github.com/sigstore/sigstore-go/pkg/bundle"
)

const simpleSigningMediaType = "application/vnd.dev.cosign.simplesigning.v1+json"

// bundleFromOCIImage fetches the cosign signature artifact for ref and assembles a
// Sigstore protobuf bundle from its annotations (mirroring the sigstore-go OCI
// example). It also returns the signed simple-signing layer's digest (alg, hex).
func bundleFromOCIImage(ctx context.Context, ref string, craneOpts ...crane.Option) (*bundle.Bundle, string, string, error) {
	layer, err := simpleSigningLayer(ctx, ref, craneOpts...)
	if err != nil {
		return nil, "", "", fmt.Errorf("getting simple signing layer: %w", err)
	}

	verificationMaterial, err := bundleVerificationMaterial(layer)
	if err != nil {
		return nil, "", "", fmt.Errorf("getting verification material: %w", err)
	}

	msgSignature, err := bundleMessageSignature(layer)
	if err != nil {
		return nil, "", "", fmt.Errorf("getting message signature: %w", err)
	}

	mediaType, err := bundle.MediaTypeString("0.1")
	if err != nil {
		return nil, "", "", fmt.Errorf("getting bundle media type: %w", err)
	}

	pb := &protobundle.Bundle{
		MediaType:            mediaType,
		VerificationMaterial: verificationMaterial,
		Content:              msgSignature,
	}
	b, err := bundle.NewBundle(pb)
	if err != nil {
		return nil, "", "", fmt.Errorf("creating bundle: %w", err)
	}
	return b, layer.Digest.Algorithm, layer.Digest.Hex, nil
}

// simpleSigningLayer resolves ref to a digest, fetches the cosign
// sha256-<digest>.sig manifest, and returns its simple-signing layer descriptor.
func simpleSigningLayer(ctx context.Context, ref string, craneOpts ...crane.Option) (*v1.Descriptor, error) {
	parsed, err := name.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("parsing image reference: %w", err)
	}

	craneOpts = append([]crane.Option{crane.WithContext(ctx)}, craneOpts...)

	digestStr, err := crane.Digest(ref, craneOpts...)
	if err != nil {
		return nil, fmt.Errorf("resolving image digest: %w", err)
	}
	h, err := v1.NewHash(digestStr)
	if err != nil {
		return nil, fmt.Errorf("parsing image digest: %w", err)
	}

	sigTag := parsed.Context().Tag(fmt.Sprintf("%s-%s.sig", h.Algorithm, h.Hex))
	mf, err := crane.Manifest(sigTag.Name(), craneOpts...)
	if err != nil {
		return nil, fmt.Errorf("getting signature manifest: %w", err)
	}
	sigManifest, err := v1.ParseManifest(bytes.NewReader(mf))
	if err != nil {
		return nil, fmt.Errorf("parsing signature manifest: %w", err)
	}

	if len(sigManifest.Layers) == 0 || sigManifest.Layers[0].MediaType != simpleSigningMediaType {
		return nil, errors.New("no cosign simple-signing layer found in signature manifest")
	}
	return &sigManifest.Layers[0], nil
}

// bundleVerificationMaterial builds the protobuf verification material (signing
// certificate chain + Rekor transparency-log entry) from cosign annotations.
func bundleVerificationMaterial(layer *v1.Descriptor) (*protobundle.VerificationMaterial, error) {
	signingCert, err := certificateChain(layer)
	if err != nil {
		return nil, fmt.Errorf("getting signing certificate: %w", err)
	}
	entries, err := tlogEntries(layer)
	if err != nil {
		return nil, fmt.Errorf("getting tlog entries: %w", err)
	}
	return &protobundle.VerificationMaterial{
		Content:                   signingCert,
		TlogEntries:               entries,
		TimestampVerificationData: nil,
	}, nil
}

// certificateChain decodes the cosign PEM signing certificate annotation.
func certificateChain(layer *v1.Descriptor) (*protobundle.VerificationMaterial_X509CertificateChain, error) {
	pemCert := layer.Annotations["dev.sigstore.cosign/certificate"]
	block, _ := pem.Decode([]byte(pemCert))
	if block == nil {
		return nil, errors.New("failed to decode certificate PEM block")
	}
	return &protobundle.VerificationMaterial_X509CertificateChain{
		X509CertificateChain: &protocommon.X509CertificateChain{
			Certificates: []*protocommon.X509Certificate{{RawBytes: block.Bytes}},
		},
	}, nil
}

// cosignTlogBundle is the JSON shape of the cosign dev.sigstore.cosign/bundle
// annotation. encoding/json base64-decodes the string fields typed as []byte
// (SignedEntryTimestamp and the Rekor body) automatically.
type cosignTlogBundle struct {
	SignedEntryTimestamp []byte `json:"SignedEntryTimestamp"`
	Payload              struct {
		Body           []byte `json:"body"`
		IntegratedTime int64  `json:"integratedTime"`
		LogIndex       int64  `json:"logIndex"`
		LogID          string `json:"logID"` // hex-encoded, not base64
	} `json:"Payload"`
}

// rekorBody is the minimal envelope of a Rekor entry body needed to populate
// the transparency-log entry's kind/version.
type rekorBody struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// tlogEntries reconstructs the Rekor transparency-log entry from the cosign
// dev.sigstore.cosign/bundle annotation.
func tlogEntries(layer *v1.Descriptor) ([]*protorekor.TransparencyLogEntry, error) {
	raw := layer.Annotations["dev.sigstore.cosign/bundle"]
	var cb cosignTlogBundle
	if err := json.Unmarshal([]byte(raw), &cb); err != nil {
		return nil, fmt.Errorf("unmarshaling cosign bundle: %w", err)
	}
	if len(cb.Payload.Body) == 0 || cb.Payload.LogID == "" || len(cb.SignedEntryTimestamp) == 0 {
		return nil, errors.New("cosign bundle missing required tlog fields")
	}

	logID, err := hex.DecodeString(cb.Payload.LogID)
	if err != nil {
		return nil, fmt.Errorf("decoding logID: %w", err)
	}

	var body rekorBody
	if err := json.Unmarshal(cb.Payload.Body, &body); err != nil {
		return nil, fmt.Errorf("unmarshaling rekor body: %w", err)
	}

	return []*protorekor.TransparencyLogEntry{
		{
			LogIndex:       cb.Payload.LogIndex,
			LogId:          &protocommon.LogId{KeyId: logID},
			KindVersion:    &protorekor.KindVersion{Kind: body.Kind, Version: body.APIVersion},
			IntegratedTime: cb.Payload.IntegratedTime,
			InclusionPromise: &protorekor.InclusionPromise{
				SignedEntryTimestamp: cb.SignedEntryTimestamp,
			},
			InclusionProof:    nil,
			CanonicalizedBody: cb.Payload.Body,
		},
	}, nil
}

// bundleMessageSignature builds the message signature (signed layer digest +
// cosign signature annotation) for the bundle.
func bundleMessageSignature(layer *v1.Descriptor) (*protobundle.Bundle_MessageSignature, error) {
	var alg protocommon.HashAlgorithm
	switch layer.Digest.Algorithm {
	case "sha256":
		alg = protocommon.HashAlgorithm_SHA2_256
	default:
		return nil, fmt.Errorf("unsupported digest algorithm: %s", layer.Digest.Algorithm)
	}
	digest, err := hex.DecodeString(layer.Digest.Hex)
	if err != nil {
		return nil, fmt.Errorf("decoding layer digest: %w", err)
	}
	// cosign puts the signature under a different namespace than the cert/bundle
	// (dev.cosignproject.cosign vs dev.sigstore.cosign) — intentional, don't "unify".
	sig, err := base64.StdEncoding.DecodeString(layer.Annotations["dev.cosignproject.cosign/signature"])
	if err != nil {
		return nil, fmt.Errorf("decoding cosign signature: %w", err)
	}
	return &protobundle.Bundle_MessageSignature{
		MessageSignature: &protocommon.MessageSignature{
			MessageDigest: &protocommon.HashOutput{Algorithm: alg, Digest: digest},
			Signature:     sig,
		},
	}, nil
}
