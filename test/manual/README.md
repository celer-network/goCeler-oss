# Run Local Manual Tests

Follow instructions below to easily start a local testnet and play with multiple OSPs on your local machine.

## Requirements

- Make sure Geth is installed
- There are two storage options: SQLite3 and CockroachDB. Install CockroachDB if you plan to use it.
- Set environment variable `GOCELER` to your local goCeler repo path

## 1. Start local Ethereum testnet

Run **`./setup.sh`** to start a local Etherem testnet running on your machine.

Take a look at the constants in [setup.go](./setup.go). In addition to start a testnet, this program would also do the following:

- Deploy Celer state channel contracts, test ERC20 token contract and multi-session CelerApp contracts.
- Create 5 test OSP ETH accounts, fund each account with 1 million test ETH and 1 billion test ERC20 tokens, all in 18 decimals. 
- Create profiles for the 5 test OSPs, located at `/tmp/celer_manual_test/profile/`. This [sample_profile.json](./sample_profile.json) shows an example of well formated OSP profile.

## 2. Prepare two OSP accounts

Open a new terminal for running tools, go to [tools/osp-setup](../../tools/osp-setup/), run **`go run osp_setup.go -profile /tmp/celer_manual_test/profile/o1_profile.json -ks ../../testing/env/keystore/osp1.json -ethpoolamt 10000 -blkdelay 0 -nopassword`** to deposit 1000 ETH of OSP 1 into the EthPool contract with approval to the CelerLedger contract, and also register the OSP1 as a network router.

Then do the same for OSP2, run **`go run osp_setup.go -profile /tmp/celer_manual_test/profile/o2_profile.json -ks ../../testing/env/keystore/osp2.json -ethpoolamt 10000 -blkdelay 0 -nopassword`**

## 3. Setup and run two OSPs

### Option 1: run OSPs using SQLite as storage backend

Run **`./run_osp.sh 1`** and **`./run_osp.sh 2`** in two terminals respectively to start OSP1 and OSP2. OSP data store is created at `/tmp/celer_manual_test/store/[ospAddr]`

### Option 2: run OSPs using CockroachDB as storage backend

First, run **`./cockroachdb.sh start`** to start the cockroachDB process, then run **`./cockroachdb.sh 1`** and **`./cockroachdb.sh 2`** to create databases and tables for the two OSPs.

Then run **`./run_osp.sh 1_crdb`** and **`./run_osp.sh 2_crdb`** in two terminals respectively to start OSP1 and OSP2. Take a look at the [run_osp.sh](./run_osp.sh) script to see the difference in server arguments.

Remember to run  **`./cockroachdb.sh stop`** after finishing all the manual tests.

## 4. Connect two OSPs through grpc stream

In the tools terminal, go to [tools/osp-admin](../../tools/osp-admin/) folder and run **`go run osp_admin.go -adminhostport localhost:8190 -registerstream -peeraddr 00290a43e5b2b151d530845b2d5a818240bc7c70 -peerhostport localhost:10002`** to let OSP1 connect with OSP2 through grpc stream.

You can see that OSP1 has new log `|INFO | server.go:540: Admin: register stream ...`, and OSP2 has new log `|INFO | server.go:218: Recv AuthReq: ...`

Please take a look at the [instructions of OSP admin tool](../../tools/osp-admin/README.md). The arg values match the constants given in [setup.go](./setup.go).

## 5. Open CelerPay channel between two OSPs

In the tools terminal, run **`go run osp_admin.go -adminhostport localhost:8190 -openchannel -peeraddr 00290a43e5b2b151d530845b2d5a818240bc7c70 -selfdeposit 10 -peerdeposit 10`** to let OSP1 open an ETH CelerPay channel with OSP2.

You can see new logs for channel opening in both OSP terminals.

## 6. Make an off-chain payment

In the tools terminal, run **`go run osp_admin.go -adminhostport localhost:8190 -sendtoken -receiver 00290a43e5b2b151d530845b2d5a818240bc7c70 -amount 0.01`** to let OSP1 make an off-chain payment of 0.01 ETH to OSP2.

You can see the returned payment ID from the admin tool log. Payment logs are also shown in OSP terminals.

## 7. View off-chain payment state

In the tools terminal, go to [tools/channel-view](../../tools/channel-view/) folder and run the following command to view the payment state at the local database of OSP2.

- If using SQLite: **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o2_profile.json -storedir /tmp/celer_manual_test/store/00290a43e5b2b151d530845b2d5a818240bc7c70 -dbview pay -payid [payment ID]`**
- If using CockroachDB: **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o2_profile.json -storesql postgresql://celer_test_o2@localhost:26257/celer_test_o2?sslmode=disable -dbview pay -payid [payment ID]`**

Please take a look at the [instructions of channel view tool](../../tools/channel-view/README.md).

## 8. View off-chain channel state

In the tools terminal, run the following command to view the channel state at the local database of OSP1.

- If using SQLite: **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o1_profile.json -storedir /tmp/celer_manual_test/store/0015f5863ddc59ab6610d7b6d73b2eacd43e6b7e -dbview channel -peer 00290a43e5b2b151d530845b2d5a818240bc7c70`** 
- If using CockroachDB: **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o1_profile.json -storesql postgresql://celer_test_o1@localhost:26257/celer_test_o1?sslmode=disable -dbview channel -peer 00290a43e5b2b151d530845b2d5a818240bc7c70`** 

You can see the channel information from the returned output. The simplex channel sequence number and free balances should reflect the 0.01 ETH payment just made.

## 9. View on-chain channel state

In the tools terminal, run **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o1_profile.json -chainview channel -cid [channel ID]`** to view the on-chain channel information stored in the smart contract. The channel ID can be found from the output of the channel state above. 

## 10. Try other tooling commands

Read instructions of [osp-admin](../../tools/osp-admin/README.md), [channel-view](../../tools/channel-view/README.md), and [channel-op](../../tools/channel-op/README.md), and get familiar with these commands.

## 11. Try more OSPs

Add more OSPs, connect them with each other by any topology you like, and try more scenarios.
