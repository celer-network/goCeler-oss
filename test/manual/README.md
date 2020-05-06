# Run Local Manual Tests

Follow instructions below to easily start a local testnet and play with multiple OSPs on your local machine.

## Requirements

- Make sure Geth and SQLite3 are installed
- Set environment variable `GOCELER` to your local goCeler repo path

## Start local Ethereum testnet

Run **`./setup.sh`** to start a local Etherem testnet runing on your machine.

Take a look at the constants in [setup.go](./setup.go). In addition to start a testnet, this program would also do the following:

- Deploy Celer state channel contracts, test ERC20 token contract and multi-session CelerApp contracts.
- Prepare 5 test OSP ETH accounts, fund each account with 1 million test ETH and 1 billion test ERC20 tokens, all in 18 decimals. 
- Create profiles for the 5 test OSPs, located at `/tmp/celer_manual_test/profile/`. This [sample_profile.json](./sample_profile.json) shows an example of well formated OSP profile.

## Setup and run OSP1

First, open a new terminal for running tools, go to [tools/osp-setup](../../tools/osp-setup/), run **`go run osp_setup.go -profile /tmp/celer_manual_test/profile/o1_profile.json -ks ../../testing/env/keystore/osp1.json -ethpoolamt 10000 -blkDelay 0 -nopassword`** to deposit 1000 ETH into the EthPool contract with approval to the CelerLedger contract, and also register the OSP as a network router.

Then, open a new terminal, go to this manual test folder and run **`./run_osp.sh 1`** to start OSP1.

You can see from the log that the OSP is up and running. OSP data store is created at `/tmp/celer_manual_test/store/[ospAddr]`

## Repeat the above process to setup and run OSP2

First, in the tools terminal, run **`go run osp_setup.go -profile /tmp/celer_manual_test/profile/o2_profile.json -ks ../../testing/env/keystore/osp2.json -ethpoolamt 10000 -blkDelay 0 -nopassword`**

Then, open a new terminal, go to this manual test folder and run **`./run_osp.sh 2`** to start OSP2.

## Connect two OSPs through grpc stream

In the tools terminal, go to [tools/osp-admin](../../tools/osp-admin/) folder and run **`go run osp_admin.go -adminhostport localhost:8190 -registerstream -peeraddr ba756d65a1a03f07d205749f35e2406e4a8522ad -peerhostport localhost:10002`** to let OSP1 connect with OSP2 through grpc stream.

You can see that OSP1 has new log `|INFO | server.go:540: Admin: register stream ...`, and OSP2 has new log `|INFO | server.go:218: Recv AuthReq: ...`

Please take a look at the [instructions of OSP admin tool](../../tools/osp-admin/README.md). The arg values match the constants given in [setup.go](./setup.go).

## Open CelerPay channel between two OSPs

In the tools terminal, run **`go run osp_admin.go -adminhostport localhost:8190 -openchannel -peeraddr ba756d65a1a03f07d205749f35e2406e4a8522ad -selfdeposit 10 -peerdeposit 10`** to let OSP1 open an ETH CelerPay channel with OSP2.

You can see new logs for channel opening in both OSP terminals.

## Make an off-chain payment

In the tools terminal, run **`go run osp_admin.go -adminhostport localhost:8190 -sendtoken -receiver ba756d65a1a03f07d205749f35e2406e4a8522ad -amount 0.01`** to let OSP1 make an off-chain payment of 0.01 ETH to OSP2.

You can see the returned payment ID from the admin tool log. Payment logs are also shown in OSP terminals.

## View off-chain payment state

In the tools terminal, go to [tools/channel-view](../../tools/channel-view/) folder and run **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o2_profile.json -storedir /tmp/celer_manual_test/store/ba756d65a1a03f07d205749f35e2406e4a8522ad -dbview pay -payid [payment ID]`** to view the payment state at local database of OSP2.

You can see the payment information from the returned output.

Please take a look at the [instructions of channel view tool](../../tools/channel-view/README.md).

## View off-chain channel state

In the tools terminal, run **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o1_profile.json -storedir /tmp/celer_manual_test/store/6a6d2a97da1c453a4e099e8054865a0a59728863 -dbview channel -peer ba756d65a1a03f07d205749f35e2406e4a8522ad`** to view the channel state at local database of OSP1.

You can see the channel information from the returned output. The simplex channel sequence number and free balances should reflect the 0.01 ETH payment just made.

## View on-chain channel state

In the tools terminal, run **`go run channel_view.go -profile /tmp/celer_manual_test/profile/o1_profile.json -storedir /tmp/celer_manual_test/store/6a6d2a97da1c453a4e099e8054865a0a59728863 -chainview channel -cid [channel ID]`** to view the channel state on the testnet chain.

You can see the on-chain channel information stored in the smart contract. The channel ID can be found from the output of the channel state above. 

## Try other tooling commands

Read instructions of [osp-admin](../../tools/osp-admin/README.md), [channel-view](../../tools/channel-view/README.md), and [channel-op](../../tools/channel-op/README.md), and get familiar with these commands.

## Try more OSPs

Add more OSPs, connect them with each other by any topology you like, and try more scenarios.
