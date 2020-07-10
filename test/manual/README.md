# Run Local Manual Tests

Follow instructions below to easily start a local testnet and play with multiple OSPs on your local machine.

## 1. Set up tools and environment

- Make sure Geth is installed
- There are two storage options: SQLite3 and CockroachDB. Install CockroachDB if you plan to use it.
- Set environment variable `GOCELER` to your local goCeler repo path.
- Run **`go build $GOCELER/tools/osp-cli`** to build to OSP CLI tool, read the [CLI command reference](../../tools/osp-cli/README.md).

## 2. Start local Ethereum testnet

Run **`./setup.sh`** to start a local Etherem testnet running on your machine.

Take a look at the constants in [setup.go](./setup.go). In addition to start a testnet, this program would also do the following:

- Deploy Celer state channel contracts, test ERC20 token contract and multi-session CelerApp contracts.
- Create 5 test OSP ETH accounts, fund each account with 1 million test ETH and 1 billion test ERC20 tokens.
- Create profiles for the 5 test OSPs, located at `/tmp/celer_manual_test/profile/`. This [sample_profile.json](./sample_profile.json) shows an example of well formated OSP profile.

## 3. Prepare two OSP accounts

Open a new terminal for CLI commands, run **`./osp-cli -profile /tmp/celer_manual_test/profile/o1_profile.json -ks $GOCELER/testing/env/keystore/osp1.json -ethpooldeposit -amount 10000 -register -nopassword`** to deposit 1000 ETH of OSP1 to the EthPool contract and approve to the CelerLedger contract, then register OSP1 as a state channel network router.

Then do the same for OSP2, run **`./osp-cli -profile /tmp/celer_manual_test/profile/o2_profile.json -ks $GOCELER/testing/env/keystore/osp2.json -ethpooldeposit -amount 10000 -register -nopassword`**

## 4. Run two OSPs

### Option 1: use SQLite as storage backend (easier setup)

Run **`./run_osp.sh 1`** and **`./run_osp.sh 2`** in two new terminals respectively to start OSP1 and OSP2. OSP data store is created at `/tmp/celer_manual_test/store/[ospAddr]`

### Option 2: use CockroachDB as storage backend (higher performance)

First, run **`./cockroachdb.sh start`** to start the CockroachDB, then run **`./cockroachdb.sh 1`** and **`./cockroachdb.sh 2`** to create databases and tables for the two OSPs. Remember to run  **`./cockroachdb.sh stop`** after finishing the manual tests.

Then run **`./run_osp.sh 1_crdb`** and **`./run_osp.sh 2_crdb`** in two new terminals respectively to start OSP1 and OSP2. Take a look at the [run_osp.sh](./run_osp.sh) script to see the difference in server arguments.

## 5. Connect two OSPs through grpc stream

Run **`./osp-cli -adminhostport localhost:8190 -registerstream -peer 00290a43e5b2b151d530845b2d5a818240bc7c70 -peerhostport localhost:10002`** to let OSP1 connect with OSP2 through grpc stream. You can see that OSP1 has new log `Admin: register stream ...`, and OSP2 has new log `Recv AuthReq: ...`

If you want to quickly connect to multiple peer OSPs (e.g., reconnect after restart), you can use the `-file` option. For example, create a `peerservers` file for OSP1 with the following lines to let it connect to OSP2-4:
```
00290a43e5b2b151d530845b2d5a818240bc7c70 localhost:10002
003ea363bccfd7d14285a34a6b1deb862df0bc84 localhost:10003
00495b55a68b5d5d1b0860b2c9eeb839e7d3a362 localhost:10004
```
Then run **`./osp-cli -adminhostport localhost:8190 -registerstream -file peerservers`**.

## 6. Open CelerPay channel between two OSPs

Run **`./osp-cli -adminhostport localhost:8190 -openchannel -peer 00290a43e5b2b151d530845b2d5a818240bc7c70 -selfdeposit 10 -peerdeposit 10`** to let OSP1 open an ETH CelerPay channel with OSP2. You can see new logs for channel opening in both OSP terminals.

## 7. Make an off-chain payment

Run **`./osp-cli -adminhostport localhost:8190 -sendtoken -receiver 00290a43e5b2b151d530845b2d5a818240bc7c70 -amount 0.01`** to let OSP1 make an off-chain payment of 0.01 ETH to OSP2. You can see the returned payment ID from the admin tool log. Payment logs are also shown in OSP terminals.

## 8. View off-chain payment state

Run the following command to view the payment state at the local database of OSP2.

- If using SQLite: **`./osp-cli -profile /tmp/celer_manual_test/profile/o2_profile.json -storedir /tmp/celer_manual_test/store/00290a43e5b2b151d530845b2d5a818240bc7c70 -dbview pay -payid [payment ID]`**
- If using CockroachDB: **`./osp-cli -profile /tmp/celer_manual_test/profile/o2_profile.json -storesql postgresql://celer_test_o2@localhost:26257/celer_test_o2?sslmode=disable -dbview pay -payid [payment ID]`**

## 9. View off-chain channel state

Run the following command to view the channel state at the local database of OSP1.

- If using SQLite: **`./osp-cli -profile /tmp/celer_manual_test/profile/o1_profile.json -storedir /tmp/celer_manual_test/store/0015f5863ddc59ab6610d7b6d73b2eacd43e6b7e -dbview channel -peer 00290a43e5b2b151d530845b2d5a818240bc7c70`** 
- If using CockroachDB: **`./osp-cli -profile /tmp/celer_manual_test/profile/o1_profile.json -storesql postgresql://celer_test_o1@localhost:26257/celer_test_o1?sslmode=disable -dbview channel -peer 00290a43e5b2b151d530845b2d5a818240bc7c70`** 

You can see the channel information from the returned output. The simplex channel sequence number and free balances should reflect the 0.01 ETH payment just made.

## 10. View on-chain channel state

Run **`./osp-cli -profile /tmp/celer_manual_test/profile/o1_profile.json -onchainview channel -cid [channel ID]`** to view the on-chain channel information stored in the smart contract. 

## 11. Try other CLI commands and more OSPs

Read the [CLI reference](../../tools/osp-cli/README.md) and try out those commands. Add more OSPs, connect them with each other by any topology you like, and try more scenarios.


