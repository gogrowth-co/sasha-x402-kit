# Payment Rails Decision — Sasha Services Menu (2026-07-05)

Written from the `marketing/` account-manager session during the Sasha Coin
services-menu deep dive. This is a decision/rationale doc only — no payment
code was written in that session (hard sequencing rule: no new x402/payment
infra until the content cadence was confirmed live; see the session report).
Implementation is scoped to a fresh session in THIS workspace via the kickoff
prompt below.

## What needs payment rails

`marketing/sasha-coin/_context/services-menu.md` (DRAFT, pending Gabriel's
approval — do not build against it until approved) specs four paid services:

| Service | Price | Chain/rail fit |
|---|---|---|
| Public token dissection | $25 one-off | Human buyer, social checkout |
| Private report | $50 one-off | Human buyer, social checkout |
| Sponsored breakdown | $50-150 one-off, tiered | Human/project buyer, social checkout |
| x402 signal feed | $20/mo or $0.05/query | Agent/dev buyer, native x402 API |

## Why x402/Base, and why not just reuse the Casper build as-is

This repo already proves the x402 PAY/ATTEST loop end-to-end — but on Casper
testnet, via `cmd/riskserver` (an x402 resource-server pricing scheme) and the
`make-software/casper-x402` facilitator. The README's own roadmap already
lists an **EVM proof adapter (Base Sepolia), Q3 2026, "next milestone"** — the
chain-agnostic `core/` package (zero chain SDK imports) was built specifically
so a new adapter is "a single config change," not a rewrite.

**Decision: accelerate the EVM/Base adapter from the Q3 roadmap slot to now,**
because:
1. Sasha's own wallet, audience, and content all live on Base already
   (`sasha.base.eth`, Gnosis Safe `0x7833...`) — Base is her home chain, not a
   new integration surface for a second ecosystem.
2. The buyer base for the $25/$50/$150 tiers is human CT users paying from a
   Base wallet after reading a tweet, not agents calling an HTTP 402 endpoint
   from code — that's a materially different UX than the Casper riskserver's
   machine-to-machine pattern, and Base is where those buyers already are.
3. The $20/mo signal feed IS the native x402 HTTP 402 pattern (agent/dev
   buyers, API access) — this one can reuse the Casper-proven
   resource-server/facilitator pattern almost directly, just pointed at a
   Base facilitator instead of the Casper one.
4. Casper stays live and unchanged for the buildathon submission — this is an
   additive adapter, not a migration off Casper.

## What does NOT get built yet

- No code changes in this session. This doc is a decision + rationale only.
- The $25/$50/$150 one-off tiers need a **lightweight payment-link flow**
  (buyer pays a fixed USDC amount on Base, references an order id in a reply
  or DM, service is fulfilled against that payment) — this is NOT the same
  shape as the x402 HTTP 402 challenge-response flow and should not be forced
  into it. Treat it as a separate, simpler adapter: a payment-received
  webhook/poller + an order-matching convention, not a resource server.
- The $20/mo feed reuses the x402 resource-server pattern once the Base
  adapter exists — this one DOES fit x402 natively.
- Actual pricing, refund/guarantee mechanics, and disclosure copy are owned by
  `marketing/sasha-coin/_context/services-menu.md` — do not re-derive them
  here; read that file for the current spec once Gabriel approves it.

## Sequencing dependency

Per `marketing/_ops/gtm-portfolio-2026-07-05.md` §4 (Sasha Coin), no service
goes live until the cadence exit gate clears (1,000+ engaged followers OR
first inbound sponsor inquiry). This build can proceed on its own timeline
(engineering isn't gated on follower count), but nothing here should be
wired into a live, publicly-announced payment link until:
1. Gabriel approves `services-menu.md` (pricing + guarantees), AND
2. The cadence exit gate clears.

Build ahead of both gates is fine. Announcing or accepting real payment
ahead of either gate is not.
