# Third-Party Notices

`sasha-x402-kit` is licensed under Apache-2.0 (see `LICENSE`). It builds on the following
third-party work. We record here, per Apache-2.0 §4, what is used, under what license, and
which parts of this repo are original vs. derived.

## Runtime dependencies (Go modules — see `go.mod`)

| Project | License | How used |
|---|---|---|
| [`x402-foundation/x402/go`](https://github.com/x402-foundation/x402) | Apache-2.0 | The generic x402 protocol framework. We implement its `SchemeNetworkClient` interface; we do not vendor or modify it. |
| [`casper-ecosystem/casper-eip-712`](https://github.com/casper-ecosystem/casper-eip-712) (Go) | Apache-2.0 | EIP-712 typed-data hashing primitives (`BuildDomain`, `HashTypedData`, `encodeAddress`) used by our pay-scheme. Dependency only. |
| [`make-software/casper-go-sdk`](https://github.com/make-software/casper-go-sdk) | Apache-2.0 | Headless `TransactionV1` signing + RPC. Dependency only. |
| [`odradev/odra`](https://github.com/odradev/odra) | MIT/Apache-2.0 | Smart-contract framework for the `AgentAttest` Casper contract. Dependency only. |

## Reference implementations (NOT vendored — used as protocol references / external infra)

| Project | License | Relationship |
|---|---|---|
| [`make-software/casper-x402`](https://github.com/make-software/casper-x402) | Apache-2.0 | The canonical Casper x402 **facilitator** (an HTTP settlement service). We run it as external infrastructure during demos — like an RPC node — and our agent talks to it over HTTP. **None of its Go code is copied into this repo;** our `adapters/casper/x402_scheme.go` is an original implementation of the upstream `x402-foundation` scheme interface. |
| [`odradev/casper-x402-poc`](https://github.com/odradev/casper-x402-poc) | Apache-2.0 | A Casper x402 proof-of-concept (CEP-18 token + `cep3009` EIP-712 `transfer_with_authorization`). We used its CEP-18 as the **test/settlement token** for the x402 PAY leg, and learned the EIP-712 domain/struct conventions (`{name, version, chain_name, contract_package_hash}`; `TransferWithAuthorization`) from it so our pay-scheme is wire-compatible. The token contract itself is **not part of this repository's submission**. |

## Originality statement (Buildathon eligibility)

All code in this repository was written new for the Casper Agentic Buildathon 2026:

- `adapters/casper/contract/src/attest.rs` — the **`AgentAttest`** contract is **clean-room original**:
  a minimal append-only attestation log (`attest(summary, metric) -> id`, `count`, `get`). It is not a
  port of the `casper-x402-poc` CEP-18; it shares no token/transfer logic.
- `adapters/casper/x402_scheme.go`, `adapters/casper/casper_adapter.go`, `agent/loop.go`, `core/*`,
  `cmd/*` — original Go, written against the public libraries above. Functional similarity to the
  reference x402 scheme is inherent to implementing the same open protocol, not copied code.

If any upstream ships a `NOTICE` file, its attribution text is incorporated here by reference.
