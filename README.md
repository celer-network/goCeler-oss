# Celer State Channel Network
Official go/grpc implementation of the off-chain modules of Celer state channel network.

Celer state channel network is a generic framework of state channels with deeply optimized on-chain contracts and off-chain messaging protocols. Please checkout the [overview of our system architecture and design principles](https://www.celer.network/docs/celercore/channel/overview.html). This repo implements the off-chain CelerNodes.

## Run Local Manual Tests

One who plan to run a full Off-chain Service Provider (OSP) node can start by following the [instructions on local manual tests](./test/manual/README.md) to play with the code and essential tools.

## Run OSP on Ropsten Testnet

Please walk through the [local manual tests](./test/manual/README.md) before moving forward to Ropsten deployment. Steps to run from source code on Ropsten are very similar to local manual tests.

### Use prebuilt binaries
Download prebuit binaries from https://github.com/celer-network/goCeler-oss/releases
Then run
```bash
tar xzf goceler-v0.16-linux-amd64.tar.gz
export PATH=$PATH:$PWD/goceler
```

### Steps to run with prebuilt binaries on Ropsten
1. Run **`geth account new --keystore .`** to generate new keystore file, then move it to `deploy/ospks.json`
2. Fund the newly generated ospks.json address some ropsten ETH.
3. Update ./deploy/ropsten/profile.json `gateway` field to Ropsten RPC (eg. https://ropsten.infura.io/v3/xxxxx), `host` filed to the OSP public RPC hostname:port (default rpc port is 10000), `address` field to the OSP ETH address.
4. Setup OSP: **`osp-setup -profile ./deploy/ropsten/profile.json -ks ./deploy/ospks.json -ethpoolamt [ETH amount]`**. The last arg is the amount of OSP's ETH to put into the EthPool for future channel open and deposit.
5. Create a storage folder at `deploy/ropsten/store`
6. Start OSP: **`server -profile ./deploy/ropsten/profile.json -ks ./deploy/ospks.json -svrname s0 -storedir ./deploy/ropsten/store -rtc ./deploy/ropsten/rt_config.json -routedata ./deploy/ropsten/channels_2020_05_03.json`**. The default rpc port is `10000`, default admin http endpoint is `localhost:8090`, use `-port` and `-adminweb` to change those values ([example](./test/manual/run_osp.sh)). OSP data store is located at `deploy/ropsten/store/[ospAddr]`
7. Connect with another OSP through grpc stream: **`osp-admin -adminhostport localhost:8090 -registerstream -peeraddr [peerOspEthAddr] -peerhostport [peerOspRpcHostPort]`**
8. Open channel with another OSP: **`osp-admin -adminhostport localhost:8090 -openchannel -peeraddr [peerOspEthAddr] -selfdeposit 0.1 -peerdeposit 0.1`**
9. Use operation tools [osp-admin](./tools/osp-admin/README.md), [channel-view](./tools/channel-view/README.md), and [channel-op](./tools/channel-op/README.md) to operate the OSP. See [local manual tests](./test/manual/README.md) for example.


## Run OSP on Ethereum Mainnet

### Steps to run with prebuilt binaries on Mainnet
Follow [steps for ropsten](#steps-to-run-from-source-code-on-ropsten) above starting from step 2, replace all keywords `ropsten` with `mainnet`.
