# sasha-x402-kit

An autonomous agent that makes its real DeFi book **verifiable and payable on Casper** over the
[x402](https://x402.org) payment protocol — a chain-agnostic commerce core with a **Casper flagship
adapter**. Agent loop: **PAY** (buy signals over x402) → **ACT** (manage a position) → **ATTEST**
(write the decision on-chain) → **EXPOSE** (sell the verified yield over x402).

Built for the **Casper Agentic Buildathon 2026**. Testnet only.

> Status: SPINE in progress. Phase-0 spike proved the live path end-to-end on `casper-test`
> (contract deploy + a real x402 `/settle`). See the design + plan in the project docs.

## Architecture

```
x402 Agent Commerce Kit (chain-agnostic core — imports no chain SDK)
  ├─ x402 pay-client        ├─ paywalled x402 server (EXPOSE)
  ├─ synthesis engine       └─ SettlementAdapter (quote/sign/settle/attest/readReceipt)
        │
        ├─ CasperAdapter (FLAGSHIP) — Odra attestation contract + casper-x402 + EIP-712 → casper:casper-test
        └─ EvmAdapter   (PROOF)     — fresh EVM x402 + EAS-style attestation (Base Sepolia)  [later]
```

## Layout

```
adapters/casper/contract/   Odra (Rust) agent-attestation contract
core/                       chain-agnostic interfaces + types  [SPINE]
agent/                      PAY → ACT → ATTEST → EXPOSE orchestrator  [SPINE]
scripts/secret-scan.sh      pre-commit secret gate
```

## Security

- Testnet only. No production keys. `.env`, `*.pem`, `keys/`, `state/` are gitignored.
- `scripts/secret-scan.sh` runs as a pre-commit hook (gitleaks if present, else heuristic block on
  PEM/env/private-key material + known credential token shapes).

## Real project / launch plan

_(Filled in Task 1.5 — links Sasha's live socials (@SashaCoin95), token, podcast, dashboard, and the
chain-agnostic roadmap.)_

## License

Apache-2.0 _(to be added)_
