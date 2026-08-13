Paste this into a fresh Claude Code session opened in ~/dev/sasha-x402-kit/
(NOT in marketing/ — this is dev/infra work and belongs in this workspace's
own session per the multi-session handoff pattern).

---

CONTEXT: sasha-x402-kit is a live agent (Sasha Coin, @SashaCoin95) that proves
a PAY -> ACT -> ATTEST loop on Casper testnet via x402. The chain-agnostic
core/ package (zero chain SDK imports) already supports a "single config
change" to add a new chain adapter — see README.md's roadmap table (EVM proof
adapter / Base Sepolia was slated for Q3 2026, "next milestone").

TASK: build a Base-chain payment-link path so Sasha can accept real USDC
payments on Base for four paid content services she sells in-feed on X. Full
rationale: read docs/payment-rails-decision-2026-07-05.md in this repo first
— it explains why Base (not Casper) is the right rail for these, and what NOT
to build.

THE FOUR SERVICES (full spec: marketing/sasha-coin/_context/services-menu.md
in the sibling marketing/ workspace — read it for pricing/guarantee language,
but do not treat pricing as final until it says APPROVED at the top; check
with Gabriel if it still says DRAFT):
1. Public token dissection — $25 one-off, human buyer
2. Private report — $50 one-off, human buyer
3. Sponsored breakdown — $50-150 one-off tiered, human/project buyer
4. x402 signal feed — $20/mo or $0.05/query, agent/dev buyer (FUTURE — build
   this one last, it's explicitly not launching with the other three)

WHAT TO BUILD, IN ORDER:
1. Base adapter for core/ (the EVM proof adapter already on the roadmap) —
   validate on Base Sepolia first, exactly like the Casper spike was
   validated on casper-test before going further.
2. A payment-link flow for services 1-3 (the $25/$50/$50-150 tiers). This is
   NOT an x402 HTTP 402 resource server — those three are sold to humans who
   read a tweet and pay, not to code calling an API. Simplest correct shape:
   a fixed-amount USDC-on-Base receive address (or a per-order address),
   a way to match an incoming payment to an order (buyer replies/DMs a
   ticker + tx hash, or a short-lived unique memo/amount), and a webhook or
   poller that confirms the payment landed before Sasha proceeds. Keep it
   the simplest thing that could work — no need for a shopping cart.
3. x402 resource-server pattern for service 4 (the $20/mo feed) — this one
   DOES fit the existing riskserver pattern (see cmd/riskserver), just
   pointed at a Base facilitator instead of Casper's. Build this last.

HARD CONSTRAINTS:
- Testnet (Base Sepolia) first for everything, same discipline as the Casper
  build (two live tx hashes before calling anything a spike GO).
- Do not wire this into a live, publicly-announced payment link until BOTH:
  (a) Gabriel has approved services-menu.md (check for "APPROVED" at the top
      of that file, not "DRAFT"), and
  (b) the Sasha cadence exit gate has cleared (1,000+ engaged followers OR
      first inbound sponsor inquiry — check with Gabriel or the marketing/
      session if unsure whether this has happened).
- Secret scan gate before any push, same as the existing repo convention
  (see .githooks/ and the existing CI secret-scan workflow).
- This repo stays public and Apache-2.0 — same attribution discipline as
  THIRD_PARTY_NOTICES.md already documents for the Casper stack.

REPORT BACK: once the Base adapter has two real Base Sepolia tx hashes and
the payment-link flow has one real end-to-end test (testnet USDC in, order
matched, confirmed), stop and flag for Gabriel before doing anything with
mainnet funds or a live public link.
