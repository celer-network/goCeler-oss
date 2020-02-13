# Celer State Channel Network

This repo is the go/grpc implementation of the off-chain parts of Celer state channel network.

For more information regarding the Celer state channel architecture, please refer to https://www.celer.network/docs/celercore.

# Instructions

## Run OSP with pre-built docker image (recommended)

Follow instructions in the [binary folder](https://github.com/celer-network/goCeler-oss/tree/master/binary).

## Connect webclient locally

1. Checkout [celer-light-client](https://github.com/celer-network/celer-light-client).
2. Update `ospEthAddress` to the OSP account address and `ospNetworkAddress` to the network address of the web proxy.
3. Follow the README in celer-light-client.

## Run OSP from source code
### Requirements
- go 1.12 or later
- geth for generating keystore file

### Steps
1. update ./profile/ropsten.json `gateway` field to Ropsten RPC, eg. https://ropsten.infura.io/v3/xxxxx
2. run `geth account new --keystore .` to generate new keystore file, rename it to ospks.json
3. fund the address in ospks.json 11 or more ETH
3. setup OSP onchain: `go run setup/main.go -ks ospks.json -profile ./profile/ropsten.json -amt 10`
4. start OSP: `go run server/server.go -profile ./profile/ropsten.json -ks ospks.json -rtc ./profile/rtconfig.json -routeData ./profile/routes/channels-ropsten-2020-02-12.json`
