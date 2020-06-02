# Celer State Channel Network
Official go/grpc implementation of the off-chain modules of Celer state channel network.

Celer state channel network is a generic framework of state channels with deeply optimized on-chain contracts and off-chain messaging protocols. Please checkout the [overview of our system architecture and design principles](https://www.celer.network/docs/celercore/channel/overview.html). This repo implements the off-chain CelerNodes.

## Run Local Manual Tests

One who plans to run a full Off-chain Service Provider (OSP) node can start by following the [instructions on local manual tests](./test/manual/README.md) to play with the code and essential tools.

## Run OSP on Ethereum Mainnet

Make sure you have walked through the [local manual tests](./test/manual/README.md) before moving forward to Mainnet deployment. Steps to run from source code on Mainnet are very similar to local manual tests.

### Requirements
- There are two storage options: SQLite3 and CockroachDB. Install CockroachDB if you plan to use it.
- Set environment variable `GOCELER` to your local goCeler repo path

### Get prebuilt binaries
Download prebuit binaries from https://github.com/celer-network/goCeler-oss/releases
Then run
```bash
tar xzf goceler-v0.16.11-linux-amd64.tar.gz
export PATH=$PATH:$PWD/goceler
```

### Prepare OSP account
1. Run **`geth account new --keystore .`** to generate a new keystore file, then rename it to `ospks.json`
2. Fund the newly generated ospks.json address some mainnet ETH.
3. Update [deploy/mainnet/profile.json](./deploy/mainnet/profile.json) `gateway` field to your Mainnet RPC (eg. https://mainnet.infura.io/v3/xxxxx), `host` filed to the OSP public RPC hostname:port (default rpc port is 10000), `address` field to the OSP ETH address.
4. Setup OSP: Run **`osp-cli -profile $GOCELER/deploy/mainnet/profile.json -ks ospks.json -ethpooldeposit -amount [ETH amount] -register -blkdelay 2`** to deposit OSP's ETH into the EthPool contract, and register the OSP as a state channel network router.
   - **note 1**: EthPool balance is used by OSP to make channel deposits and accept open channel request from peers.
   - **note 2**: `-blkdelay` specifies how many blocks to wait to confirm the on-chain transactions.

### Run OSP server
#### Option 1: run OSP using SQLite as storage backend (easier setup)
5. Choose a store path (e.g., `${HOME}/celerdb`), the OSP data will be located at `${HOME}/celerdb/[ospAddr]`.
6. Start OSP: **`server -profile $GOCELER/deploy/mainnet/profile.json -ks ospks.json -svrname s0 -storedir ${HOME}/celerdb -rtc $GOCELER/deploy/mainnet/rt_config.json -routedata $GOCELER/deploy/mainnet/channels_2020_05_08.json`**.
- **note 1**: use `-routedata` only when starting OSP from scracth for the first time.
- **note 2**: the default rpc port is `10000`, default admin http endpoint is `localhost:8090`, use `-port` and `-adminweb` to change those values ([example](./test/manual/run_osp.sh)).
- **note 3**: use [log args](https://github.com/celer-network/goutils/blob/v0.1.2/log/log.go) if needed, e.g., `-logdir ${HOME}/logs -logrotate`

#### Option 2: run OSP using CockroachDB as storage backend (higher performance)
5. First install CockroachDB. Then checkout [tools/scripts/cockroachdb.sh](./tools/scripts/cockroachdb.sh), update `STOREPATH` to your preferred storage location, and run **`./cockroachdb.sh start`** to start the cockroachDB process and create database tables.
6. Start OSP: **`server -profile $GOCELER/deploy/mainnet/profile.json -ks ospks.json -svrname s0 -storesql postgresql://celer@localhost:26257/celer?sslmode=disable -rtc $GOCELER/deploy/mainnet/rt_config.json -routedata $GOCELER/deploy/mainnet/channels_2020_05_08.json`**. Be aware of the **notes 1-3** above.

### Open channel with peer OSP
7. Connect with another OSP through grpc stream: **`osp-cli -adminhostport localhost:8090 -registerstream -peer [peerOspAddr] -peerhostport [peerOspHostPort]`**
8. Open channel with another OSP: **`osp-cli -adminhostport localhost:8090 -openchannel -peer [peerOspAddr] -selfdeposit 0.1 -peerdeposit 0.1`**
9. Query channel from database: **`osp-cli -profile $GOCELER/deploy/mainnet/profile.json -storedir ${HOME}/celerdb/[ospAddr] -dbview channel -peer [peerOspAddr]`**. If using CockroachDB, replace the `-storedir` arg with `-storesql postgresql://celer@localhost:26257/celer?sslmode=disable`.
10. Query channel from blockchain: **`osp-cli -profile $GOCELER/deploy/mainnet/profile.json -onchainview channel -cid [channel ID]`**. You can see the channel ID from the output of step 9 above.

### Apply other OSP operations
11. Use [OSP CLI Commands](./tools/osp-cli/README.md) to operate the OSP. See [local manual tests](./test/manual/README.md) for example.

## Run OSP on Ropsten Testnet

Follow [steps for mainnet](#run-osp-on-ethereum-mainnet) above starting from step 2, replace all keywords `mainnet` with `ropsten`.

## TLS Certificate for serving Internet traffic
OSP needs to have a valid TLS certificate for Celer connections over the Internet. If you already have a domain name, you can get one from [Let's Encrypt](https://letsencrypt.org/). Then run OSP with flags `-tlscert mysvr.crt -tlskey mysvr.key`.

Otherwise, the builtin cert supports DDNS with the following domain names:
```
*.dynu.com
*.mooo.com
*.us.to
*.hopto.org
*.zapto.org
*.sytes.net
*.ddns.net
```
and all domains in 1st and 2nd pages at https://freedns.afraid.org/domain/registry/. You can register free account with the DDNS provider, eg. mycelernode.ddns.net, update host field in profile.json to it, and run OSP, no need to specify tlscert or tlskey flag.

If you prefer using IP address directly, please contact cert@celer.network and we'll email you a unique cert for requested IP address.

## Start Web Proxy for Celer web client
`go run webproxy/cmd/main.go -server localhost:10000` assume OSP runs on default 10000 port

Then clone https://github.com/celer-network/celer-light-client repo, update demo/mainnet_config.json ospEthAddress to OSP eth address and ospNetworkAddress to http://[webproxy DNS or IP]:29980
