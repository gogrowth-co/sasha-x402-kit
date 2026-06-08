//! AgentAttest — an append-only on-chain log of an autonomous agent's decisions/positions.
//!
//! ERC-8004-style agent attestation, written clean-room in Rust/Odra for Casper (originality
//! rule: no port of existing Solidity/JS). Each cycle the agent calls `attest(summary, metric)`;
//! the record is keyed by an incrementing id and stamped with the caller + block time, making the
//! agent's book independently verifiable on `casper-test`.

use odra::prelude::*;

/// A single attested agent decision/position.
#[odra::odra_type]
pub struct Record {
    /// The account that recorded the attestation (the agent).
    pub author: Address,
    /// Human-readable decision summary, e.g. "WETH/USDC LP open".
    pub summary: String,
    /// A single numeric metric tied to the decision (price, size, bps, ...).
    pub metric: u64,
    /// Block time when the attestation was recorded.
    pub ts: u64,
}

/// Append-only attestation log.
#[odra::module]
pub struct AgentAttest {
    records: Mapping<u32, Record>,
    count: Var<u32>,
}

#[odra::module]
impl AgentAttest {
    /// Initialize an empty log.
    pub fn init(&mut self) {
        self.count.set(0);
    }

    /// Record an attestation; returns the assigned id.
    pub fn attest(&mut self, summary: String, metric: u64) -> u32 {
        let id = self.count.get_or_default();
        self.records.set(
            &id,
            Record {
                author: self.env().caller(),
                summary,
                metric,
                ts: self.env().get_block_time(),
            },
        );
        self.count.set(id + 1);
        id
    }

    /// Total number of attestations recorded.
    pub fn count(&self) -> u32 {
        self.count.get_or_default()
    }

    /// Read an attestation by id (reverts if absent).
    pub fn get(&self, id: u32) -> Record {
        self.records.get(&id).unwrap_or_revert(&self.env())
    }
}
