package casper

// x402 pay-side scheme for Casper. Original implementation of the upstream
// x402 SchemeNetworkClient interface: builds an EIP-712 `TransferWithAuthorization`
// authorization with the public casper-eip-712 library and signs it with the agent's
// Casper key (casper-go-sdk). The upstream x402 HTTP client drives the 402->retry dance;
// this type only produces the signed payment payload.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
	"github.com/make-software/casper-go-sdk/v2/types/keypair"
	x402types "github.com/x402-foundation/x402/go/types"
)

// transferWithAuthorizationTypes is the x402 Casper authorization struct (matches the
// on-chain CEP-18 verifier: address from/to, uint256 amounts, bytes32 nonce).
var transferWithAuthorizationTypes = eip712.TypeDefinitions{
	"TransferWithAuthorization": {
		{Name: "from", Type: "address"},
		{Name: "to", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "validAfter", Type: "uint256"},
		{Name: "validBefore", Type: "uint256"},
		{Name: "nonce", Type: "bytes32"},
	},
}

// PayScheme implements the upstream x402 SchemeNetworkClient for the "exact" Casper scheme.
type PayScheme struct {
	key keypair.PrivateKey
}

// NewPayScheme builds a pay-scheme bound to the agent's signing key.
func NewPayScheme(key keypair.PrivateKey) *PayScheme {
	return &PayScheme{key: key}
}

func (s *PayScheme) Scheme() string { return "exact" }

// CreatePaymentPayload builds + signs a TransferWithAuthorization for the given requirements.
func (s *PayScheme) CreatePaymentPayload(
	_ context.Context,
	req x402types.PaymentRequirements,
) (x402types.PaymentPayload, error) {
	pkgBytes, err := hex.DecodeString(req.Asset)
	if err != nil || len(pkgBytes) != 32 {
		return x402types.PaymentPayload{}, fmt.Errorf("invalid asset package hash: %q", req.Asset)
	}
	var pkg [32]byte
	copy(pkg[:], pkgBytes)

	name, _ := req.Extra["name"].(string)
	version, _ := req.Extra["version"].(string)
	if name == "" || version == "" {
		return x402types.PaymentPayload{}, errors.New("requirements.extra must include name and version")
	}

	domain := eip712.BuildDomain(name, version, req.Network, pkg)

	now := time.Now().Unix()
	validAfter := now - 600
	validBefore := now + int64(req.MaxTimeoutSeconds)

	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf("nonce: %w", err)
	}

	fromStr := "00" + s.key.PublicKey().AccountHash().ToHex()
	fromAddr, err := eip712.NewAddressFromHex("0x" + fromStr)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf("from address: %w", err)
	}
	toAddr, err := eip712.NewAddressFromHex("0x" + req.PayTo)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf("to address: %w", err)
	}
	value, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return x402types.PaymentPayload{}, fmt.Errorf("invalid amount: %q", req.Amount)
	}

	message := map[string]interface{}{
		"from":        fromAddr,
		"to":          toAddr,
		"value":       value,
		"validAfter":  big.NewInt(validAfter),
		"validBefore": big.NewInt(validBefore),
		"nonce":       nonce,
	}

	digest, err := eip712.HashTypedData(
		domain,
		transferWithAuthorizationTypes,
		"TransferWithAuthorization",
		message,
		&eip712.TypedDataOptions{DomainTypes: eip712.CasperDomainTypes},
	)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf("hash typed data: %w", err)
	}

	sig, err := s.key.Sign(digest[:])
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf("sign: %w", err)
	}

	payload := map[string]interface{}{
		"signature": hex.EncodeToString(sig),
		"publicKey": s.key.PublicKey().ToHex(),
		"authorization": map[string]interface{}{
			"from":        fromStr,
			"to":          req.PayTo,
			"value":       req.Amount,
			"validAfter":  fmt.Sprintf("%d", validAfter),
			"validBefore": fmt.Sprintf("%d", validBefore),
			"nonce":       hex.EncodeToString(nonce[:]),
		},
	}

	return x402types.PaymentPayload{X402Version: 2, Payload: payload}, nil
}
