use agent_attest::attest::AgentAttest;
use odra::host::{Deployer, NoArgs};

#[test]
fn records_a_decision_and_reads_it_back() {
    let env = odra_test::env();
    let mut c = AgentAttest::deploy(&env, NoArgs);

    // empty to start
    assert_eq!(c.count(), 0u32);

    // first attestation gets id 0
    let id = c.attest("WETH/USDC LP open".to_string(), 1850u64);
    assert_eq!(id, 0u32);
    assert_eq!(c.count(), 1u32);

    // read it back, fields intact, author = caller (deployer = account 0)
    let rec = c.get(id);
    assert_eq!(rec.summary, "WETH/USDC LP open");
    assert_eq!(rec.metric, 1850u64);
    assert_eq!(rec.author, env.get_account(0));

    // second attestation increments the id + count
    let id2 = c.attest("ETH short opened".to_string(), 42u64);
    assert_eq!(id2, 1u32);
    assert_eq!(c.count(), 2u32);
    assert_eq!(c.get(1).metric, 42u64);
}
