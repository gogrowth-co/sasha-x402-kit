// Package agent is the PAY -> ACT -> ATTEST -> EXPOSE orchestrator. The SPINE runs PAY -> ATTEST:
// buy a signal over x402 (live 402->settle), synthesize a decision, and attest it on-chain.
package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	x402 "github.com/x402-foundation/x402/go"
	x402http "github.com/x402-foundation/x402/go/http"

	"github.com/make-software/casper-go-sdk/v2/types/keypair"

	"github.com/SashaCoin95/sasha-x402-kit/adapters/casper"
	"github.com/SashaCoin95/sasha-x402-kit/core"
)

// Config wires one agent cycle.
type Config struct {
	KeyPath   string
	KeyAlgo   string
	RPCURL    string
	ChainName string // short, e.g. "casper-test"
	AttestPkg string // AgentAttest package hash (64-hex)
	Network   string // CAIP-2, e.g. "casper:casper-test"
	SignalURL string // paywalled x402 endpoint to buy the signal from
}

// Run executes one PAY -> ACT -> ATTEST cycle, printing the live tx hashes.
func Run(cfg Config) error {
	algo := keypair.ED25519
	if cfg.KeyAlgo == "secp256k1" {
		algo = keypair.SECP256K1
	}
	key, err := keypair.NewPrivateKeyFromFile(cfg.KeyPath, algo)
	if err != nil {
		return fmt.Errorf("load key: %w", err)
	}

	// --- PAY: buy the signal over x402 (HTTP 402 -> sign -> retry -> settle) ---
	scheme := casper.NewPayScheme(key)
	client := x402.Newx402Client()
	client.Register(x402.Network(cfg.Network), scheme)
	// Bounded client so the PAY leg (402 -> sign -> retry -> settle) can't hang indefinitely.
	payHTTP := &http.Client{Timeout: 90 * time.Second}
	httpClient := x402http.WrapHTTPClientWithPayment(payHTTP, x402http.Newx402HTTPClient(client))

	fmt.Printf("[PAY] GET %s (x402)\n", cfg.SignalURL)
	resp, err := httpClient.Get(cfg.SignalURL)
	if err != nil {
		return fmt.Errorf("pay request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pay not completed: status %d: %s", resp.StatusCode, string(body))
	}

	settleTx := ""
	if pr := resp.Header.Get("Payment-Response"); pr != "" {
		if dec, derr := base64.StdEncoding.DecodeString(pr); derr == nil {
			var m map[string]interface{}
			if json.Unmarshal(dec, &m) == nil {
				if t, ok := m["transaction"].(string); ok {
					settleTx = t
				}
			}
		}
	}
	fmt.Printf("[PAY] paid; settle tx=%s\n", settleTx)

	// --- ACT/synthesize: derive a decision from the purchased signal (strictly validated) ---
	// The signal is Sasha's own LP risk packet (schema sasha.risk_packet.v1) — she pays her
	// own x402 endpoint for the packet, then attests the verdict it returned.
	var signal map[string]interface{}
	if err := json.Unmarshal(body, &signal); err != nil {
		return fmt.Errorf("act: paid response is not JSON: %w", err)
	}
	schema, _ := signal["schema"].(string)
	verdict, _ := signal["verdict"].(string)
	score, scoreOK := signal["score"].(float64)
	if schema == "" || verdict == "" || !scoreOK {
		return fmt.Errorf("act: signal missing required fields schema/verdict/score: %s", string(body))
	}
	summary := fmt.Sprintf("bought x402 risk packet (%s); verdict=%s score=%.0f", schema, verdict, score)
	metric := uint64(0)
	if score > 0 {
		metric = uint64(score)
	}
	fmt.Printf("[ACT] decision: %q metric=%d\n", summary, metric)

	// --- ATTEST: record the decision on-chain ---
	adapter, err := casper.New(casper.Config{
		KeyPath:   cfg.KeyPath,
		KeyAlgo:   cfg.KeyAlgo,
		RPCURL:    cfg.RPCURL,
		ChainName: cfg.ChainName,
		AttestPkg: cfg.AttestPkg,
	})
	if err != nil {
		return fmt.Errorf("adapter: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	res, err := adapter.Attest(ctx, core.Attestation{Summary: summary, Metric: metric})
	if err != nil {
		return fmt.Errorf("attest: %w", err)
	}
	fmt.Printf("[ATTEST] tx=%s\n", res.TxHash)

	fmt.Println("\n=== AGENT CYCLE COMPLETE (PAY -> ACT -> ATTEST) ===")
	if settleTx != "" {
		fmt.Printf("PAY/settle: https://testnet.cspr.live/transaction/%s\n", settleTx)
	}
	fmt.Printf("ATTEST:     https://testnet.cspr.live/transaction/%s\n", res.TxHash)
	return nil
}
