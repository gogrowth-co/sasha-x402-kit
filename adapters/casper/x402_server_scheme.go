package casper

// x402 resource-server-side scheme for Casper. Original implementation of the upstream
// x402 SchemeNetworkServer interface: it does not verify or settle payments itself (that's
// the facilitator's job, reached over HTTP by the generic x402 Go SDK resource server) — it
// only knows how to price a route in the CEP-18 token's atomic units and fill in the EIP-712
// domain fields (name/version) the client-side PayScheme needs to reconstruct the same digest.

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	x402 "github.com/x402-foundation/x402/go"
	"github.com/x402-foundation/x402/go/types"
)

var (
	addressRegex     = regexp.MustCompile(`^(00|01)[0-9a-fA-F]{64}$`)
	packageHashRegex = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
)

// ServerScheme implements the upstream x402 SchemeNetworkServer for the "exact" Casper scheme.
// One instance prices requests against a single CEP-18 asset (this kit only ever sells its own
// signal for its own token), so asset/name/version/decimals are fixed at construction.
type ServerScheme struct {
	assetPackage string // 64-hex CEP-18 package hash
	assetName    string // token runtime name; must match the deployed contract's name() (EIP-712 domain)
	decimals     int
}

// NewServerScheme builds a pricing scheme bound to one CEP-18 asset.
func NewServerScheme(assetPackage, assetName string, decimals int) *ServerScheme {
	return &ServerScheme{assetPackage: assetPackage, assetName: assetName, decimals: decimals}
}

func (s *ServerScheme) Scheme() string { return "exact" }

// GetAssetDecimals satisfies the optional AssetDecimalsProvider interface.
func (s *ServerScheme) GetAssetDecimals(_ string, _ x402.Network) int { return s.decimals }

// ParsePrice converts a USD price string (e.g. "$0.10") into atomic units of this scheme's asset.
func (s *ServerScheme) ParsePrice(price x402.Price, _ x402.Network) (x402.AssetAmount, error) {
	usd, err := parseUSD(price)
	if err != nil {
		return x402.AssetAmount{}, err
	}
	units := new(big.Float).Mul(big.NewFloat(usd), new(big.Float).SetInt(pow10(s.decimals)))
	amount, _ := units.Int(nil)
	return x402.AssetAmount{
		Amount: amount.String(),
		Asset:  s.assetPackage,
		Extra: map[string]interface{}{
			"name":     s.assetName,
			"version":  "1",
			"decimals": strconv.Itoa(s.decimals),
		},
	}, nil
}

// EnhancePaymentRequirements fills in the EIP-712 domain fields the client needs and validates
// the asset/payTo shapes before the requirements go out in a 402 response.
func (s *ServerScheme) EnhancePaymentRequirements(
	_ context.Context,
	requirements types.PaymentRequirements,
	_ types.SupportedKind,
	_ []string,
) (types.PaymentRequirements, error) {
	if !packageHashRegex.MatchString(requirements.Asset) {
		return requirements, fmt.Errorf("invalid CEP-18 package hash: %q", requirements.Asset)
	}
	if !addressRegex.MatchString(requirements.PayTo) {
		return requirements, fmt.Errorf("invalid Casper address: %q", requirements.PayTo)
	}
	if requirements.Extra == nil {
		requirements.Extra = map[string]interface{}{}
	}
	if _, ok := requirements.Extra["name"]; !ok {
		requirements.Extra["name"] = s.assetName
	}
	if _, ok := requirements.Extra["version"]; !ok {
		requirements.Extra["version"] = "1"
	}
	return requirements, nil
}

func parseUSD(price x402.Price) (float64, error) {
	switch v := price.(type) {
	case string:
		clean := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "$"))
		f, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid price %q: %w", v, err)
		}
		return f, nil
	case float64:
		return v, nil
	default:
		return 0, errors.New("price must be a \"$x.xx\" string")
	}
}

func pow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}
