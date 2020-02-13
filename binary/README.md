# Instruction
## Prerequisite
- [docker](https://docs.docker.com/install/#supported-platforms/)
- [docker-compose](https://docs.docker.com/compose/install/)
- Make sure you [don't need `sudo`](https://docs.docker.com/install/linux/linux-postinstall/) to run `docker` and `docker-compose`.
Otherwise you'll need to preface `sudo` to all docker commands used below.

## Preparation -- Only Run Once at Beginning
- download the directory and `cd` to the downloaded directory.
- <pre><code>mkdir nodedata && mkdir logdir</code></pre>
- <pre><code>cp <i>your_node_keystore.json</i> `pwd`/profile/nodeks.json</code></pre>
Make sure the key store **DOES NOT HAVE PASSWORD**.
- <pre><code>docker load -i cmnty-server-&ast;</code></pre>
- change "gateway" field in ./profile/ropsten.json to your **ROPSTEN** `geth` client.
You can use [infura](https://infura.io/) or run your own **ROPSTEN** `geth` client.
- <pre><code>docker run --rm -v `pwd`/profile:/profile -v `pwd`/logdir:/logdir -i --entrypoint "first-time-run-setup" cmnty/server:latest -ks /profile/nodeks.json -profile /profile/ropsten.json -amt 10</code></pre>

## Start Celer Node Server
Bring it up by running
<pre><code>docker-compose up -d</code></pre>

Now the celer node is ready to serve!
The node log will be available under `.logdir/`.

You can shut down the node by running `docker-compose down`.

## Run web proxy from source
From the `goCeler-oss` directory, run
```
go run webproxy/cmd/main.go -port 29981 -server localhost:10000
```
The command will bring up a web proxy for celer webclient.
You can configure celer webclient to connect to the proxy on port **29981**.

### Troubleshooting
- If you see any permission issue or error about not having docker machine
to run, try to add `sudo` in all docker and docker-compose
related commands. Alternatively, follow [linux-postinstall](https://docs.docker.com/install/linux/linux-postinstall/)

## Connect to Other Celer Nodes
Once you have a node running and you want to connect to peer node (e.g. eth address *`0x4bace345c30d9244b71218dc6ca694836138b60e`*) at *`cnode1-ropsten.celer.app:10000`*, do following
1. Base64 encode peer node eth addr. The result will be *`S6zjRcMNkkS3EhjcbKaUg2E4tg4=`*. On Linux and Mac, you can do `echo 0x4bace345c30d9244b71218dc6ca694836138b60e | xxd -r -p | base64`.
2. <a id="registerstream"></a>Run following command to let your node to register stream with peer node. Remember to register stream on every node rebooting.
```
curl -d '{"peer_rpc_address":"cnode1-ropsten.celer.app:10000","peer_eth_address":"S6zjRcMNkkS3EhjcbKaUg2E4tg4="} -X POST http://localhost:8090/admin/peer/registerstream
```

3. <a id="openchannel"></a>To open an ETH channel with each side depositing 0.15 ETH, run
```
curl -d '{"peer_eth_address":"S6zjRcMNkkS3EhjcbKaUg2E4tg4=","token_type":1, "token_address":"AAAAAAAAAAAAAAAAAAAAAAAAAAA=", "self_deposit_amt_wei":"150000000000000000", "peer_deposit_amt_wei":"150000000000000000"}' -X POST http://localhost:8090/admin/peer/openchannel
```
That's it! Celer clients connecting to your node should be able to send ETH to clients connecting to the peer node `4bace345c30d9244b71218dc6ca694836138b60e`.

### Play on Mainnet
If you want to play on Ethereum mainnet, the celer node mainnet ETH address is `33ebbb1d4d21e8626c4144ed737989c7532eb588`. The two commands above become
```
curl -d '{"peer_rpc_address":"cnode1-mainnet.celer.app:10000","peer_eth_address":"M+u7HU0h6GJsQUTtc3mJx1MutYg="} -X POST http://localhost:8090/admin/peer/registerstream
curl -d '{"peer_eth_address":"M+u7HU0h6GJsQUTtc3mJx1MutYg=","token_type":1, "token_address":"AAAAAAAAAAAAAAAAAAAAAAAAAAA=", "self_deposit_amt_wei":"150000000000000000", "peer_deposit_amt_wei":"150000000000000000"}' -X POST http://localhost:8090/admin/peer/openchannel
```
**In preparation section, make sure you switch to `mainnet.json` and modify `ropsten` related fields in `docker-compose.yml` to mainnet**.

Disclaimer: Celer team operates the example community node `4bace345c30d9244b71218dc6ca694836138b60e` at `cnode1-ropsten.celer.app:10000` for ropsten and `33ebbb1d4d21e8626c4144ed737989c7532eb588` at `cnode1-mainnet.celer.app:10000` for mainnet. We allow **ETH channel** depositing from `50000000000000000` wei to `200000000000000000` wei.

### Troubleshooting
- If you see "connection to *4bace345c30d9244b71218dc6ca694836138b60e*
already exist" in [step 2](#registerstream), that means you have already registered stream with the address.
- If you see "distribution breaks policy" in [step 3](#openchannel), double check with peer for allowed channel open deposit of you and your peer respectively and change "self_deposit_amt_wei" or "peer_deposit_amt_wei" accordingly.

## OSP Admin Interface
The OSP exposes several http admin APIs for operators. The API and input data structures are defined in [the osp_admin.proto](https://github.com/celer-network/goCeler-oss/blob/master/proto/osp_admin.proto). For example, if you want to let the OSP send tokens to some other nodes, you can post a json to `localhost:8090/admin/sendtoken`, with the body looks like
```
{
  "dst_addr": "4bace345c30d9244b71218dc6ca694836138b60e",
  "amt_wei": "1",
  "token_addr": "0000000000000000000000000000000000000000"
}
```
