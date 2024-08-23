1. **andromeda_kernel.wasm**:
    - No dependencies.

2. **andromeda_vfs.wasm**:
    - Depends on `andromeda_kernel.wasm`.

3. **andromeda_adodb.wasm**:
    - Depends on `andromeda_kernel.wasm`.

4. **andromeda_economics.wasm**:
    - Depends on `andromeda_kernel.wasm`.

5. **andromeda_cw721.wasm**:
    - Depends on `andromeda_kernel.wasm`.

6. **andromeda_auction.wasm**:
    - Depends on `andromeda_kernel.wasm`.

7. **andromeda_crowdfund.wasm**:
    - Depends on `andromeda_kernel.wasm`.
    - Depends on `andromeda_cw721.wasm`.

8. **andromeda_marketplace.wasm**:
    - Depends on `andromeda_kernel.wasm`.

9. **andromeda_cw20.wasm**:
    - Depends on `andromeda_kernel.wasm`.

10. **andromeda_cw20_exchange.wasm**:
    - Depends on `andromeda_kernel.wasm`.
    - Depends on `andromeda_cw20.wasm`.

11. **andromeda_cw20_staking.wasm**:
    - Depends on `andromeda_kernel.wasm`.
    - Depends on `andromeda_cw20.wasm`.

12. **andromeda_merkle_airdrop.wasm**:
    - Depends on `andromeda_kernel.wasm`.

13. **andromeda_lockdrop.wasm**:
    - Depends on `andromeda_kernel.wasm`.
    - Depends on `andromeda_cw20.wasm`.