// Package casper implements core.SettlementAdapter against the Casper Network using the
// casper-go-sdk (the spike-validated headless signing path). Attest calls the AgentAttest
// Odra contract's `attest` entrypoint as a TransactionV1; the x402 settle leg lands in Task 1.4.
package casper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	casperSDK "github.com/make-software/casper-go-sdk/v2/casper"
	"github.com/make-software/casper-go-sdk/v2/types"
	"github.com/make-software/casper-go-sdk/v2/types/clvalue"
	"github.com/make-software/casper-go-sdk/v2/types/keypair"

	"github.com/SashaCoin95/sasha-x402-kit/core"
)

// errNotImplemented marks SPINE methods completed in a later task.
var errNotImplemented = errors.New("not implemented in this task")

// Config wires a CasperAdapter to a network + the deployed AgentAttest contract.
type Config struct {
	KeyPath     string // PEM secret key path (gitignored; never logged)
	KeyAlgo     string // "ed25519" (default) or "secp256k1"
	RPCURL      string // Casper JSON-RPC, e.g. https://node.testnet.casper.network/rpc
	ChainName   string // short chain name for the tx, e.g. "casper-test"
	AttestPkg   string // 64-hex package hash of the deployed AgentAttest contract
	AttestMotes uint64 // gas budget (motes) for an attest call; 0 -> default
}

type CasperAdapter struct {
	key       keypair.PrivateKey
	rpcURL    string
	chainName string
	attestPkg string
	motes     uint64
}

const defaultAttestMotes uint64 = 5_000_000_000

// New loads the signing key and returns a ready adapter.
func New(cfg Config) (*CasperAdapter, error) {
	algo := keypair.ED25519
	if cfg.KeyAlgo == "secp256k1" {
		algo = keypair.SECP256K1
	}
	key, err := keypair.NewPrivateKeyFromFile(cfg.KeyPath, algo)
	if err != nil {
		return nil, fmt.Errorf("load key: %w", err)
	}
	motes := cfg.AttestMotes
	if motes == 0 {
		motes = defaultAttestMotes
	}
	return &CasperAdapter{
		key:       key,
		rpcURL:    cfg.RPCURL,
		chainName: cfg.ChainName,
		attestPkg: cfg.AttestPkg,
		motes:     motes,
	}, nil
}

func (c *CasperAdapter) rpc() casperSDK.RPCClient {
	return casperSDK.NewRPCClient(casperSDK.NewRPCHandler(c.rpcURL, http.DefaultClient))
}

// Attest records (summary, metric) on the AgentAttest contract via a signed TransactionV1.
func (c *CasperAdapter) Attest(ctx context.Context, a core.Attestation) (core.AttestResult, error) {
	pubKey := c.key.PublicKey()

	args := types.Args{}
	args.AddArgument("summary", *clvalue.NewCLString(a.Summary)).
		AddArgument("metric", *clvalue.NewCLUInt64(a.Metric))

	packageHash, err := casperSDK.NewHash(c.attestPkg)
	if err != nil {
		return core.AttestResult{}, fmt.Errorf("decode attest package hash: %w", err)
	}
	entryPoint := "attest"

	payload, err := types.NewTransactionV1Payload(
		types.InitiatorAddr{PublicKey: &pubKey},
		types.Timestamp(time.Now().UTC()),
		900000000000,
		c.chainName,
		types.PricingMode{Limited: &types.LimitedMode{
			GasPriceTolerance: 1,
			StandardPayment:   true,
			PaymentAmount:     c.motes,
		}},
		types.NewNamedArgs(&args),
		types.TransactionTarget{Stored: &types.StoredTarget{
			ID: types.TransactionInvocationTarget{
				ByPackageHash: &types.ByPackageHashInvocationTarget{Addr: packageHash, Version: nil},
			},
			Runtime: types.NewVmCasperV1TransactionRuntime(),
		}},
		types.TransactionEntryPoint{Custom: &entryPoint},
		types.TransactionScheduling{Standard: &struct{}{}},
	)
	if err != nil {
		return core.AttestResult{}, fmt.Errorf("build transaction payload: %w", err)
	}

	tx, err := types.MakeTransactionV1(payload)
	if err != nil {
		return core.AttestResult{}, fmt.Errorf("make transaction: %w", err)
	}
	if err := tx.Sign(c.key); err != nil {
		return core.AttestResult{}, fmt.Errorf("sign transaction: %w", err)
	}

	client := c.rpc()
	res, err := client.PutTransactionV1(ctx, *tx)
	if err != nil {
		return core.AttestResult{}, fmt.Errorf("put transaction: %w", err)
	}
	txHash := res.TransactionHash.String()

	if err := c.waitForExecution(ctx, txHash); err != nil {
		return core.AttestResult{TxHash: txHash}, fmt.Errorf("await execution: %w", err)
	}
	return core.AttestResult{TxHash: txHash}, nil
}

// waitForExecution polls until the transaction is executed (or reverts).
func (c *CasperAdapter) waitForExecution(ctx context.Context, txHash string) error {
	client := c.rpc()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	deadline := time.NewTimer(120 * time.Second)
	defer deadline.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("timeout waiting for %s", txHash)
		case <-ticker.C:
			res, err := client.GetTransactionByTransactionHash(ctx, txHash)
			if err != nil {
				continue
			}
			if res.ExecutionInfo == nil || res.ExecutionInfo.BlockHeight == 0 ||
				res.ExecutionInfo.ExecutionResult == nil {
				continue
			}
			if res.ExecutionInfo.ExecutionResult.ErrorMessage != nil {
				return fmt.Errorf("execution failed: %s", *res.ExecutionInfo.ExecutionResult.ErrorMessage)
			}
			return nil
		}
	}
}

// ReadReceipt — completed in Task 1.4 (query the contract's `get` view / named keys).
func (c *CasperAdapter) ReadReceipt(_ context.Context, _ uint32) (core.Receipt, error) {
	return core.Receipt{}, errNotImplemented
}

// Settle — the x402 settle leg, completed in Task 1.4 (drives the casper-x402 facilitator).
func (c *CasperAdapter) Settle(_ context.Context, _ core.PaymentReq) (core.SettleResult, error) {
	return core.SettleResult{}, errNotImplemented
}
