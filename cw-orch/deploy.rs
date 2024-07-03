// ANCHOR: full_counter_example
use counter_contract::{
    msg::InstantiateMsg, CounterContract, CounterExecuteMsgFns, CounterQueryMsgFns,
};
use cw_orch::{anyhow, prelude::*};
use cw_orch::environment::{ChainKind, NetworkInfo};

const LOCAL_MNEMONIC: &str = "tip yard art tape orchard universe angle flame wave gadget raven coyote crater ethics able evoke luxury predict leopard delay peanut embody blast soap";
const LOCAL_LANDSLIDE: ChainInfo = ChainInfo {
    kind: ChainKind::Local,
    chain_id: "landslide-test",
    gas_denom: "stake",
    gas_price: 1 as f64,
    grpc_urls: &["http://127.0.0.1:9090"],
    network_info: LOCAL_LANDSLIDE_NETWORK,
    lcd_url: None,
    fcd_url: None,
};

const LOCAL_LANDSLIDE_NETWORK: NetworkInfo = NetworkInfo {
    chain_name: "landslide",
    pub_address_prefix: "wasm",
    coin_type: 118u32,
};

// cargo run --example deploy
pub fn main() -> anyhow::Result<()> {
    std::env::set_var("LOCAL_MNEMONIC", LOCAL_MNEMONIC);
    // ANCHOR: chain_construction
    dotenv::dotenv().ok(); // Used to load the `.env` file if any
    pretty_env_logger::init(); // Used to log contract and chain interactions

    let network = LOCAL_LANDSLIDE;
    let chain = DaemonBuilder::default().chain(network).build()?;
    // ANCHOR_END: chain_construction

    // ANCHOR: contract_interaction

    let counter = CounterContract::new(chain);

    // ANCHOR: clean_example
    counter.upload()?;
    counter.instantiate(&InstantiateMsg { count: 0 }, None, None)?;

    counter.increment()?;

    let count = counter.get_count()?;
    assert_eq!(count.count, 1);
    // ANCHOR_END: clean_example
    // ANCHOR_END: contract_interaction

    Ok(())
}
// ANCHOR_END: full_counter_example
