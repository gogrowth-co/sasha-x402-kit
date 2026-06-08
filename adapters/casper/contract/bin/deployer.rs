//! Livenet deployer for AgentAttest. Reads the pre-built wasm (wasm/AgentAttest.wasm) and deploys
//! it to the network configured via ODRA_CASPER_LIVENET_* env vars. Writes the deployed address
//! to AGENT_ATTEST_ADDRESS_FILE (if set) for downstream tooling (adapter / agent loop).

use agent_attest::attest::AgentAttest;
use odra::host::{Deployer, NoArgs};
use odra::prelude::Addressable;

fn main() {
    let env = odra_casper_livenet_env::env();
    env.set_gas(300_000_000_000);

    let contract = AgentAttest::deploy(&env, NoArgs);
    let address = contract.address().to_string();
    println!("Deployed AgentAttest at: {}", address);

    if let Ok(path) = std::env::var("AGENT_ATTEST_ADDRESS_FILE") {
        std::fs::write(&path, &address).expect("failed to write address file");
        println!("Wrote address to {}", path);
    }
}
