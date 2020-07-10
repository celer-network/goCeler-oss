# Celer State Channel Network
Official go/grpc implementation of the off-chain modules of Celer state channel network.

Celer state channel network is a generic framework of state channels with deeply optimized on-chain contracts and off-chain messaging protocols. Please checkout the [overview of our system architecture and design principles](https://www.celer.network/docs/celercore/channel/overview.html). This repo implements the off-chain CelerNodes.

## Run Local Manual Tests

One who plans to run a full Off-chain Service Provider (OSP) node should start by following the [instructions on local manual tests](./test/manual/README.md) to play with the code and essential tools, and to get familiar with the operaton process.

## Run OSP on Ethereum Mainnet

Please walk through the [local manual tests](./test/manual/README.md) before moving forward to Mainnet deployment. Steps to operate OSPs on Mainnet are very similar to local manual tests.

Current running OSPs can be found at https://explorer.celer.network.

Here we only show how to operate ETH channels as examples. ERC20 channels are also fully supported by adding `-token` arg in related [commands](./tools/osp-cli/README.md). 

### Requirements
- There are two storage options: SQLite3 and CockroachDB. Install CockroachDB if you plan to use it.
- [Get TLS certificate ready](#tls-certificate-for-serving-internet-traffic) for serving Internet traffic.

### Get prebuilt binaries and config files
Download prebuit binaries from https://github.com/celer-network/goCeler-oss/releases
Then run
```bash
tar xzf goceler-v0.16.14-linux-amd64.tar.gz
export PATH=$PATH:$PWD/goceler
```
Download the [profile.json](./deploy/mainnet/profile.json), [rt_config.json](./deploy/mainnet/rt_config.json) and [channels_xxx.json](./deploy/mainnet/channels_2020_05_08.json) files from [deploy/mainnet](./deploy/mainnet/) to your home folder. Instructions below use home folder as the base location. Replace `$HOME` with your preferred local path if needed.


### Prepare OSP account
1. Run **`geth account new --keystore . --lightkdf`** to generate a new keystore file, then rename it to `ks.json`

2. Fund the newly generated ks.json address some mainnet ETH.

3. Update the [profile.json](./deploy/mainnet/profile.json) `gateway` field to your Mainnet RPC (eg. https://mainnet.infura.io/v3/xxxxx), `host` filed to the OSP public RPC hostname:port (default rpc port is 10000), `address` field to the OSP ETH address.

4. Setup OSP: Run **`osp-cli -profile $HOME/profile.json -ks ks.json -ethpooldeposit -amount [ETH amount] -register -blkdelay 2`** to deposit OSP's ETH into the EthPool contract, and register the OSP as a state channel network router.
   - EthPool is used by OSP to accept ETH open channel requests from peers. For example, when node `A` initiates an ETH open channel request with node `B`, node `A` will make channel deposit from its account balance, while node `B` will make deposit from its EthPool balance.
   - As noted in the [CLI Command Reference](./tools/osp-cli/README.md), `amount` is float assuming 18 token decimals.
   - `-blkdelay` specifies how many blocks to wait to confirm the on-chain transactions.

### Run OSP server
#### Option 1: run OSP using SQLite as storage backend (easier setup)
5. Choose a store path (e.g., `$HOME/celerdb`), your OSP data will be located at `$HOME/celerdb/[ospAddr]`.

6. Start OSP: **`server -profile $HOME/profile.json -ks ks.json -svrname s0 -storedir $HOME/celerdb -rtc $HOME/rt_config.json -routedata $HOME/channels_2020_05_08.json`**.

#### Option 2: run OSP using CockroachDB as storage backend (higher performance)
5. First install CockroachDB. Then checkout [tools/scripts/cockroachdb.sh](./tools/scripts/cockroachdb.sh), update `STOREPATH` to your preferred storage location, and run **`./cockroachdb.sh start`** to start the cockroachDB process and create tables.

6. Start OSP: **`server -profile $HOME/profile.json -ks ks.json -svrname s0 -storesql postgresql://celer@localhost:26257/celer?sslmode=disable -rtc $HOME/rt_config.json -routedata $HOME/channels_2020_05_08.json`**.

**Notes (for both options):**
- Use `-routedata` only when starting OSP from scracth for the first time.
- Use [log args](https://github.com/celer-network/goutils/blob/v0.1.13/log/log.go) as needed, e.g., `-logdir $HOME/logs -logrotate`.
- The default rpc port is `10000`, default admin http endpoint is `localhost:8090`, use `-port` and `-adminweb` to change those values ([example](./test/manual/run_osp.sh)) if needed.
- Your OSP should be shown on the [Explorer](https://explorer.celer.network) within 15 minutes after the server started.

### Open channel with peer OSP
7. Connect with another OSP through grpc stream: **`osp-cli -adminhostport localhost:8090 -registerstream -peer [peerOspAddr] -peerhostport [peerOspHostPort]`**.

   If you want to quickly connect to multiple peer OSPs (e.g., reconnect after restart), you can use the `-file` option. Create a `peerservers` file with lines of `addr host:port` you want to connect, for example:
   ```
   00290a43e5b2b151d530845b2d5a818240bc7c70 a.b.net:10000
   003ea363bccfd7d14285a34a6b1deb862df0bc84 x.y.com:10000
   00495b55a68b5d5d1b0860b2c9eeb839e7d3a362 m.n.network:10000
   ```
   Then run **`osp-cli -adminhostport localhost:8090 -registerstream -file peerservers`**.

8. Open channel with another OSP: **`osp-cli -adminhostport localhost:8090 -openchannel -peer [peerOspAddr] -selfdeposit 0.5 -peerdeposit 0.5`**. This would open a channel with each peer making 0.5 ETH deposit. If you get error for any reason (e.g., peer policy violation), wait for 10 minutes beforing trying to open channel with the same peer.

   **Note:** make sure you have enough balance in your ETH account, and the peer you want to open channel with has enough balance in the [EthPool contract](https://etherscan.io/token/0x44e081cac2406a4efe165178c2a4d77f7a7854d4#balances).

9. Query channel from database: **`osp-cli -profile $HOME/profile.json -storedir $HOME/celerdb/[ospAddr] -dbview channel -peer [peerOspAddr]`**. If using CockroachDB, replace the `-storedir` arg with `-storesql postgresql://celer@localhost:26257/celer?sslmode=disable`.

10. Query channel from blockchain: **`osp-cli -profile $HOME/profile.json -onchainview channel -cid [channel ID]`**. You can see the channel ID from the output of step 9 above.

### Apply other OSP operations
11. Use [OSP CLI Commands](./tools/osp-cli/README.md) to operate the OSP. See [local manual tests](./test/manual/README.md) for example.

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
