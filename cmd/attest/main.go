// Command attest records one attestation on the live AgentAttest contract — the SPINE proof
// that the agent can write its decisions on-chain. Config via env (no secrets in argv):
//
//	ODRA_CASPER_LIVENET_SECRET_KEY_PATH  PEM key path (required)
//	AGENT_ATTEST_PACKAGE                 deployed AgentAttest package hash, 64-hex (required)
//	ODRA_CASPER_LIVENET_NODE_ADDRESS     RPC (default testnet)
//	ODRA_CASPER_LIVENET_CHAIN_NAME       chain name (default casper-test)
//	CLIENT_KEY_ALGO                      ed25519 (default) | secp256k1
//	ATTEST_SUMMARY / ATTEST_METRIC       payload (defaults provided)
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/SashaCoin95/sasha-x402-kit/adapters/casper"
	"github.com/SashaCoin95/sasha-x402-kit/core"
)

func main() {
	cfg := casper.Config{
		KeyPath:   os.Getenv("ODRA_CASPER_LIVENET_SECRET_KEY_PATH"),
		KeyAlgo:   envOr("CLIENT_KEY_ALGO", "ed25519"),
		RPCURL:    envOr("ODRA_CASPER_LIVENET_NODE_ADDRESS", "https://node.testnet.casper.network/rpc"),
		ChainName: envOr("ODRA_CASPER_LIVENET_CHAIN_NAME", "casper-test"),
		AttestPkg: os.Getenv("AGENT_ATTEST_PACKAGE"),
	}
	if cfg.KeyPath == "" || cfg.AttestPkg == "" {
		fmt.Fprintln(os.Stderr, "required: ODRA_CASPER_LIVENET_SECRET_KEY_PATH and AGENT_ATTEST_PACKAGE")
		os.Exit(1)
	}

	adapter, err := casper.New(cfg)
	if err != nil {
		fatal(err)
	}

	summary := envOr("ATTEST_SUMMARY", "WETH/USDC LP open")
	metric := uint64(1850)
	if m := os.Getenv("ATTEST_METRIC"); m != "" {
		v, perr := strconv.ParseUint(m, 10, 64)
		if perr != nil {
			fatal(fmt.Errorf("invalid ATTEST_METRIC %q: %w", m, perr))
		}
		metric = v
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()

	fmt.Printf("Attesting summary=%q metric=%d -> package %s\n", summary, metric, cfg.AttestPkg)
	res, err := adapter.Attest(ctx, core.Attestation{Summary: summary, Metric: metric})
	if err != nil {
		fatal(err)
	}
	fmt.Printf("ATTEST OK tx=%s\n", res.TxHash)
	fmt.Printf("explorer: https://testnet.cspr.live/transaction/%s\n", res.TxHash)
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
