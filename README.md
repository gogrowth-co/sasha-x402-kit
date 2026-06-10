![sasha-x402-kit](assets/sasha-hero.png)

# sasha-x402-kit

[![ci](https://github.com/gogrowth-co/sasha-x402-kit/actions/workflows/secret-scan.yml/badge.svg)](https://github.com/gogrowth-co/sasha-x402-kit/actions/workflows/secret-scan.yml) ![License](https://img.shields.io/badge/license-Apache--2.0-blue) ![Casper](https://img.shields.io/badge/network-casper--test-red) ![x402](https://img.shields.io/badge/protocol-x402-green)

**An autonomous AI agent that makes its real DeFi book verifiable and payable on Casper over the [x402](https://x402.org) payment protocol.**

Most agents claim to run a book. This one attests every decision on-chain so anyone can verify it. Built on a chain-agnostic core with a Casper flagship adapter, the agent runs a four-verb loop that closes the gap between "agent says it did something" and "agent proves it did something." Built for the **Casper Agentic Buildathon 2026**. Testnet only.

---

## The Loop

| Verb | What it does | Status |
|---|---|---|
| **PAY** | Buys the signals it acts on over x402 (HTTP 402 → sign → settle) | Shipped |
| **ACT** | Manages a real testnet position | Roadmap |
| **ATTEST** | Writes every decision to an on-chain attestation contract | Shipped |
| **EXPOSE** | Serves its verified yield as an x402-payable feed | Roadmap (needs external payer) |

---

## Live on `casper-test` (verifiable now)

Every claim here is a real transaction on the public Casper testnet. Click any hash to verify on testnet.cspr.live.

| What | Transaction |
|---|---|
| Attestation contract deploy (`AgentAttest`) | [`577570f2…dba0bfff`](https://testnet.cspr.live/transaction/577570f2f5f486353b8d2e61f7328fca34cd8446053d643ebc395344dba0bfff) |
| Agent loop — PAY (x402 `402→settle`) | [`b419bbcb…13cc5f2b`](https://testnet.cspr.live/transaction/b419bbcbcbefaa6da97eb4e5251461c691ba436f8f6921a316ea82c213cc5f2b) |
| Agent loop — ATTEST (decision on-chain) | [`1f063cc2…dec62f6893`](https://testnet.cspr.live/transaction/1f063cc2d3567079cfac9075c3120d9b15deddcdec2a71eb75fc6fdec62f6893) |

`AgentAttest` package hash: `7b4bb374af24ee46a067f4d41f5cba61b097ba613825617e81a57d7673132262`

---

## Why this is different

The field is full of "resell-data-over-x402" agents. This one is different on two dimensions.

**A real position.** The agent behind this kit, [Sasha](https://x.com/SashaCoin95), runs a live delta-neutral LP/treasury book on Base and Solana and posts it publicly. That means she can attest a book because she actually runs one.

**Verifiable on-chain identity.** Every decision cycle writes an attestation to `AgentAttest` on Casper. The chain doesn't trust the agent's self-report. It stores a proof. That's a different category than agents that log to a database they control.

---

## Architecture

![Agent loop and architecture](assets/agent-loop.png)

```
core/                         chain-agnostic — imports NO chain SDK
  settlement_adapter.go        SettlementAdapter: Attest (shipped); ReadReceipt/Settle = roadmap
  types.go                     chain-neutral types
adapters/
  casper/                      FLAGSHIP (lands first)
    contract/                  Odra (Rust) AgentAttest contract, clean-room ERC-8004 pattern
    casper_adapter.go          Attest via casper-go-sdk TransactionV1
    x402_scheme.go             original x402 pay-scheme (EIP-712 via casper-eip-712)
  evm/                         PROOF adapter (Base Sepolia), proves the seam  [roadmap]
agent/loop.go                  PAY → ACT → ATTEST → EXPOSE orchestrator
cmd/{attest,agent}/            runnable entrypoints
scripts/secret-scan.sh         pre-commit secret gate (.githooks/pre-commit)
```

**The core rule:** `core/` never imports a chain SDK. All chain specifics live behind `SettlementAdapter`. The Casper adapter is built and proven end-to-end before the EVM adapter starts, so "chain-agnostic" is provable in code, not just claimed in a README.

**Stack:** Rust + [Odra](https://github.com/odradev/odra) v2.7 (CasperVM/WASM contract) · [casper-go-sdk](https://github.com/make-software/casper-go-sdk) (headless `TransactionV1` signing) · [`casper-eip-712`](https://github.com/casper-ecosystem/casper-eip-712) (typed-data) · [`make-software/casper-x402`](https://github.com/make-software/casper-x402) facilitator (x402 infra) · CSPR.cloud public testnet RPC.

The `AgentAttest` contract is **clean-room original**. The CEP-18 used for x402 settlement and the facilitator come from the Apache-2.0 projects [`odradev/casper-x402-poc`](https://github.com/odradev/casper-x402-poc) and [`make-software/casper-x402`](https://github.com/make-software/casper-x402), used as dependencies and infra, not vendored. Full attribution and the originality statement: [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

---

## Quick start

**Prereqs:** Rust + cargo-odra, `wasm-opt`/`wasm-strip`, Go 1.25+, a funded `casper-test` key.

```bash
# 1. Contract: test on OdraVM + the real CasperVM, then deploy to testnet
cd adapters/casper/contract
cargo odra test                 # OdraVM (fast)
cargo odra test -b casper       # CasperVM (real engine)
ODRA_CASPER_LIVENET_NODE_ADDRESS=https://node.testnet.casper.network/rpc \
ODRA_CASPER_LIVENET_EVENTS_URL=https://node.testnet.casper.network/events \
ODRA_CASPER_LIVENET_CHAIN_NAME=casper-test \
ODRA_CASPER_LIVENET_SECRET_KEY_PATH=./keys/secret_key.pem \
  cargo run --bin deployer      # -> deploy tx + AgentAttest package hash

# 2. Agent: one PAY -> ACT -> ATTEST cycle
#    (needs an x402 facilitator + paid endpoint reachable)
ODRA_CASPER_LIVENET_SECRET_KEY_PATH=./keys/secret_key.pem \
AGENT_ATTEST_PACKAGE=<package-hash> \
SIGNAL_URL=<paid x402 endpoint> \
  go run ./cmd/agent            # -> live settle tx + attest tx
```

See `.env.example` for the full variable set. `cmd/attest` runs a standalone attestation without the full agent loop.

---

## Security

- **Testnet only. No production keys.** `.env`, `*.pem`, `keys/`, and `state/` are gitignored.
- `scripts/secret-scan.sh` runs as a pre-commit hook (`.githooks/pre-commit`, wired via `git config core.hooksPath .githooks`). It always blocks PEM/env/private-key material and known credential token shapes by name and content. It runs gitleaks additionally when present, and CI re-runs the full-history scan on every push.

---

## 🔗 Real project / launch plan

This is not a throwaway hackathon entry. It's a new on-chain capability for a live agent.

- **Agent:** Sasha — X [@SashaCoin95](https://x.com/SashaCoin95), YouTube [@SashaCoin](https://youtube.com/@SashaCoin)
- **Token:** $SASHA · **Podcast:** *Token Trends* (co-host Max Ledge) · runs autonomously on OpenCLAW
- **Today:** Sasha runs a real delta-neutral LP/treasury book on Base and Solana and posts it publicly.
- **What this adds:** that book becomes verifiable on Casper (every decision attested) and payable over x402 (a verified-yield feed other agents can buy).

**What runs on Casper after the buildathon:** Sasha keeps the `AgentAttest` contract live and attests each decision cycle, then stands up an x402-payable verified-yield feed using the x402 ecosystem credits to keep at least one real data feed paid. The chain-agnostic core means the same agent extends to any chain. The EVM proof adapter (Base Sepolia) is the next milestone after the flagship.

---

## License

Apache-2.0 (see [`LICENSE`](LICENSE)). Third-party attribution and the originality statement are in [`NOTICE`](NOTICE) and [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).
