// Command agent runs one PAY -> ACT -> ATTEST cycle of Sasha's x402 agent. Config via env:
//
//	ODRA_CASPER_LIVENET_SECRET_KEY_PATH  PEM key (required)
//	AGENT_ATTEST_PACKAGE                 AgentAttest package hash, 64-hex (required)
//	SIGNAL_URL                           paywalled x402 endpoint (default local /risk-packet, see cmd/riskserver)
//	ODRA_CASPER_LIVENET_NODE_ADDRESS     RPC (default testnet)
//	ODRA_CASPER_LIVENET_CHAIN_NAME       short chain name (default casper-test)
//	CAIP2_CHAIN_ID                       CAIP-2 network (default casper:casper-test)
//	CLIENT_KEY_ALGO                      ed25519 (default) | secp256k1
package main

import (
	"fmt"
	"os"

	"github.com/SashaCoin95/sasha-x402-kit/agent"
)

func main() {
	cfg := agent.Config{
		KeyPath:   os.Getenv("ODRA_CASPER_LIVENET_SECRET_KEY_PATH"),
		KeyAlgo:   envOr("CLIENT_KEY_ALGO", "ed25519"),
		RPCURL:    envOr("ODRA_CASPER_LIVENET_NODE_ADDRESS", "https://node.testnet.casper.network/rpc"),
		ChainName: envOr("ODRA_CASPER_LIVENET_CHAIN_NAME", "casper-test"),
		AttestPkg: os.Getenv("AGENT_ATTEST_PACKAGE"),
		Network:   envOr("CAIP2_CHAIN_ID", "casper:casper-test"),
		SignalURL: envOr("SIGNAL_URL", "http://localhost:4021/risk-packet"),
	}
	if cfg.KeyPath == "" || cfg.AttestPkg == "" {
		fmt.Fprintln(os.Stderr, "required: ODRA_CASPER_LIVENET_SECRET_KEY_PATH and AGENT_ATTEST_PACKAGE")
		os.Exit(1)
	}
	if err := agent.Run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
