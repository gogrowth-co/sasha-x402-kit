//! odra-cli entrypoint for the agent_attest contract: deploy to a livenet and call `attest`.

use agent_attest::attest::AgentAttest;
use odra::host::{HostEnv, NoArgs};
use odra::schema::casper_contract_schema::NamedCLType;
use odra_cli::{
    deploy::DeployScript,
    scenario::{Args, Error, Scenario, ScenarioMetadata},
    CommandArg, ContractProvider, DeployedContractsContainer, DeployerExt, OdraCli,
};

/// Deploys `AgentAttest` and adds it to the container.
pub struct AgentAttestDeployScript;

impl DeployScript for AgentAttestDeployScript {
    fn deploy(
        &self,
        env: &HostEnv,
        container: &mut DeployedContractsContainer,
    ) -> Result<(), odra_cli::deploy::Error> {
        let _c = AgentAttest::load_or_deploy(env, NoArgs, container, 300_000_000_000);
        Ok(())
    }
}

/// Scenario: record one attestation on the deployed contract.
pub struct AttestScenario;

impl Scenario for AttestScenario {
    fn args(&self) -> Vec<CommandArg> {
        vec![
            CommandArg::new("summary", "Decision summary string", NamedCLType::String),
            CommandArg::new("metric", "Metric value (u64)", NamedCLType::U64),
        ]
    }

    fn run(
        &self,
        env: &HostEnv,
        container: &DeployedContractsContainer,
        args: Args,
    ) -> Result<(), Error> {
        let mut contract = container.contract_ref::<AgentAttest>(env)?;
        let summary = args.get_single::<String>("summary")?;
        let metric = args.get_single::<u64>("metric")?;
        env.set_gas(2_000_000_000);
        contract.try_attest(summary, metric)?;
        Ok(())
    }
}

impl ScenarioMetadata for AttestScenario {
    const NAME: &'static str = "attest";
    const DESCRIPTION: &'static str = "Record an attestation (summary, metric) on the deployed contract";
}

pub fn main() {
    OdraCli::new()
        .about("CLI tool for the agent_attest smart contract")
        .deploy(AgentAttestDeployScript)
        .contract::<AgentAttest>()
        .scenario(AttestScenario)
        .build()
        .run();
}
