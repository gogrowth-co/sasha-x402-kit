# sasha-x402-kit

An autonomous AI agent that makes its **real DeFi book verifiable and payable on Casper** over the
[x402](https://x402.org) payment protocol. A chain-agnostic commerce core with a **Casper flagship
adapter**; the agent runs a four-verb loop:

**PAY** — buys the signals it acts on over x402 (HTTP 402 → sign → settle).
**ACT** — manages a real testnet position. *(roadmap — see Status)*
**ATTEST** — writes every decision to an on-chain attestation contract, making its book verifiable.
**EXPOSE** — serves its verified yield as an x402-payable feed. *(roadmap — needs an external payer)*

Built for the **Casper Agentic Buildathon 2026**. **Testnet only.**

The agent is **[Sasha](https://x.com/SashaCoin95)** — a live AI persona with a token, a podcast, and a
public track record (see *Real project* below). The differentiator versus the field's generic
"resell-data-over-x402" agents: a **real, persona-owned position** plus a **verifiable on-chain agent
identity** — Sasha can credibly attest a book because she actually runs one.

## Live on `casper-test` (verifiable now)

Every claim here is a real transaction on the public Casper testnet:

| What | Transaction |
|---|---|
| Attestation contract deploy (`AgentAttest`) | [`577570f2…dba0bfff`](https://testnet.cspr.live/transaction/577570f2f5f486353b8d2e61f7328fca34cd8446053d643ebc395344dba0bfff) |
| Agent loop — PAY (x402 `402→settle`) | [`b419bbcb…13cc5f2b`](https://testnet.cspr.live/transaction/b419bbcbcbefaa6da97eb4e5251461c691ba436f8f6921a316ea82c213cc5f2b) |
| Agent loop — ATTEST (decision on-chain) | [`1f063cc2…dec62f6893`](https://testnet.cspr.live/transaction/1f063cc2d3567079cfac9075c3120d9b15deddcdec2a71eb75fc6fdec62f6893) |

`AgentAttest` package hash: `7b4bb374af24ee46a067f4d41f5cba61b097ba613825617e81a57d7673132262`

## Architecture

```
core/                         chain-agnostic — imports NO chain SDK
  settlement_adapter.go        SettlementAdapter: Attest / ReadReceipt / Settle
  types.go                     chain-neutral types
adapters/
  casper/                      FLAGSHIP (lands first)
    contract/                  Odra (Rust) AgentAttest contract — clean-room ERC-8004 pattern
    casper_adapter.go          Attest via casper-go-sdk TransactionV1
    x402_scheme.go             original x402 pay-scheme (EIP-712 via casper-eip-712)
  evm/                         PROOF adapter (Base Sepolia) — proves the seam  [roadmap]
agent/loop.go                  PAY → ACT → ATTEST → EXPOSE orchestrator
cmd/{attest,agent}/            runnable entrypoints
scripts/secret-scan.sh         pre-commit secret gate (.githooks/pre-commit)
```

**Rule:** `core/` never imports a chain SDK; all chain specifics live behind `SettlementAdapter`. The
Casper adapter is built and proven end-to-end before the EVM adapter starts — so "chain-agnostic" is
provable in code, not just claimed.

**Stack:** Rust + [Odra](https://github.com/odradev/odra) v2.7 (CasperVM/WASM contract) ·
[casper-go-sdk](https://github.com/make-software/casper-go-sdk) (headless `TransactionV1` signing) ·
[`casper-eip-712`](https://github.com/casper-ecosystem/casper-eip-712) (typed-data) ·
[`make-software/casper-x402`](https://github.com/make-software/casper-x402) facilitator (x402 infra) ·
CSPR.cloud public testnet RPC. The contract derives from the
[`odradev/casper-x402-poc`](https://github.com/odradev/casper-x402-poc) CEP-18 (CC-attributed); all
agent/adapter code is original to this repo.

## Quick start

Prereqs: Rust + cargo-odra, `wasm-opt`/`wasm-strip`, Go 1.25+, a funded `casper-test` key.

```bash
# 1. contract: test on OdraVM + the real CasperVM, then deploy to testnet
cd adapters/casper/contract
cargo odra test                 # OdraVM (fast)
cargo odra test -b casper       # CasperVM (real engine)
ODRA_CASPER_LIVENET_NODE_ADDRESS=https://node.testnet.casper.network/rpc \
ODRA_CASPER_LIVENET_EVENTS_URL=https://node.testnet.casper.network/events \
ODRA_CASPER_LIVENET_CHAIN_NAME=casper-test \
ODRA_CASPER_LIVENET_SECRET_KEY_PATH=./keys/secret_key.pem \
  cargo run --bin deployer      # -> deploy tx + AgentAttest package hash

# 2. agent: one PAY -> ACT -> ATTEST cycle (needs an x402 facilitator + paid endpoint reachable)
ODRA_CASPER_LIVENET_SECRET_KEY_PATH=./keys/secret_key.pem \
AGENT_ATTEST_PACKAGE=<package-hash> \
SIGNAL_URL=<paid x402 endpoint> \
  go run ./cmd/agent            # -> live settle tx + attest tx
```

See `.env.example` for the full variable set. `cmd/attest` runs a standalone attestation.

## Security

- **Testnet only. No production keys.** `.env`, `*.pem`, `keys/`, `state/` are gitignored.
- `scripts/secret-scan.sh` runs as a pre-commit hook (`.githooks/pre-commit`, wired via
  `git config core.hooksPath .githooks`): always blocks PEM/env/private-key material + known
  credential token shapes by name *and* content; runs gitleaks additionally when present.

## Real project / launch plan

This is not a throwaway hackathon entry — it's a new on-chain capability for a **live agent**:

- **Agent:** Sasha — X [@SashaCoin95](https://x.com/SashaCoin95), YouTube [@SashaCoin](https://youtube.com/@SashaCoin)
- **Token:** $SASHA · **Podcast:** *Token Trends* (co-host Max Ledge) · runs autonomously on OpenCLAW
- **Today:** Sasha runs a real delta-neutral LP/treasury book on Base/Solana and posts it publicly.
- **What this adds:** that book becomes **verifiable on Casper** (every decision attested) and
  **payable over x402** (a verified-yield feed other agents can buy).

**What runs on Casper after the buildathon:** Sasha keeps the `AgentAttest` contract live and attests
each decision cycle, and stands up an x402-payable verified-yield feed — using the x402 ecosystem
credits to keep paying for at least one real data feed. The chain-agnostic core means the same agent
extends to any chain; the EVM proof adapter (Base Sepolia) is the next milestone after the flagship.

## License

Apache-2.0.
