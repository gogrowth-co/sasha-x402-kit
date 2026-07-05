// Command riskserver is the x402-paywalled resource server that sells Sasha's own LP risk
// packet (schema sasha.risk_packet.v1) on Casper. It never computes the packet itself — after
// a payment settles, it proxies to a localhost-only internal endpoint (the same risk-packet.ts
// engine the CROO Risk Desk sells over CAP) and returns whatever that engine currently reports.
// Config via env:
//
//	PAYEE_ADDRESS               Casper address (00|01 + 64-hex) payments settle to (required)
//	X402_ASSET_PACKAGE          CEP-18 token package hash, 64-hex (required)
//	X402_ASSET_NAME             token runtime name; must match the deployed contract's name() (required)
//	X402_ASSET_DECIMALS         token decimals (default 9)
//	X402_PRICE                  price per request, e.g. "$0.10" (default "$0.10")
//	X402_FACILITATOR_URL        facilitator base URL (default http://localhost:4022)
//	CAIP2_CHAIN_ID              CAIP-2 network (default casper:casper-test)
//	RISK_PACKET_INTERNAL_URL    localhost risk-packet source (default http://127.0.0.1:8977/risk-packet)
//	PORT                        listen port (default 4021)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	x402 "github.com/x402-foundation/x402/go"
	x402http "github.com/x402-foundation/x402/go/http"
	"github.com/x402-foundation/x402/go/http/nethttp"

	"github.com/SashaCoin95/sasha-x402-kit/adapters/casper"
)

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	payTo := os.Getenv("PAYEE_ADDRESS")
	assetPkg := os.Getenv("X402_ASSET_PACKAGE")
	assetName := os.Getenv("X402_ASSET_NAME")
	if payTo == "" || assetPkg == "" || assetName == "" {
		fmt.Fprintln(os.Stderr, "required: PAYEE_ADDRESS, X402_ASSET_PACKAGE, X402_ASSET_NAME")
		os.Exit(1)
	}
	decimals := 9
	if _, err := fmt.Sscanf(envOr("X402_ASSET_DECIMALS", "9"), "%d", &decimals); err != nil {
		decimals = 9
	}
	network := x402.Network(envOr("CAIP2_CHAIN_ID", "casper:casper-test"))
	riskPacketURL := envOr("RISK_PACKET_INTERNAL_URL", "http://127.0.0.1:8977/risk-packet")
	port := envOr("PORT", "4021")

	facilitator := x402http.NewHTTPFacilitatorClient(&x402http.FacilitatorConfig{
		URL: envOr("X402_FACILITATOR_URL", "http://localhost:4022"),
	})
	scheme := casper.NewServerScheme(assetPkg, assetName, decimals)

	routes := x402http.RoutesConfig{
		"GET /risk-packet": {
			Accepts: x402http.PaymentOptions{
				{Scheme: "exact", PayTo: payTo, Price: envOr("X402_PRICE", "$0.10"), Network: network},
			},
			Description: "Sasha's current LP risk packet (sasha.risk_packet.v1)",
			MimeType:    "application/json",
		},
	}

	middleware := nethttp.PaymentMiddlewareFromConfig(
		routes,
		nethttp.WithFacilitatorClient(facilitator),
		nethttp.WithScheme(network, scheme),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/risk-packet", func(w http.ResponseWriter, r *http.Request) {
		proxyRiskPacket(w, riskPacketURL)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	fmt.Printf("[riskserver] serving GET /risk-packet on :%s (network=%s payTo=%s)\n", port, network, payTo)
	if err := http.ListenAndServe(":"+port, middleware(mux)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// proxyRiskPacket forwards to the internal (payment-gated by the caller, never public) risk-packet
// engine and streams its response back verbatim. Only reached after the x402 middleware settles payment.
func proxyRiskPacket(w http.ResponseWriter, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		http.Error(w, `{"error":"bad internal request"}`, http.StatusInternalServerError)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"risk-packet engine unreachable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}
