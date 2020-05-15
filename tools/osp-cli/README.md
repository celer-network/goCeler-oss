## OSP CLI Command Reference

#### Notes
* ETH token is used by default if `-token` arg is not provided.
* `amount` is float assuming 18 token decimals.
* replace `-storedir` with `-storesql` followed by sql database URL if using CockroachDB.

### Operation through OSP admin HTTP interface

`osp-cli -adminhostport [host:port of admin http endpoint]` followed by:

* `-registerstream -peer [peer addr] -peerhostport [peer grpc host:port]`: register stream with a peer OSP
* `-openchannel -peer [peer addr] -token [token addr] -selfdeposit [amount] -peerdeposit [amount]`: open a state channel with a peer OSP
* `-sendtoken -receiver [receiver addr] -token [token addr] -amount [amount]`: make an off-chain payment
* `-deposit -peer [peer addr] -token [token addr] -amount [amount]`: make an on-chain deposit
* `-querydeposit -depositid [deposit job ID]`: query the status of a deposit job
* `-querypeerosps`: get information of all peer OSPs

### Query information from database

`osp-cli -profile [profile file] -storedir [sqlite store directory]` followed by:

#### Off-chain state fast query
* `-dbview channel -cid [channel ID]`: get channel info for a given channel id
* `-dbview channel -peer [peer addr] -token [token addr]`: get channel info for given peer and token pair
* `-dbview channel -count -token [token addr] -chanstate [int]`: count channels for given token and state
* `-dbview pay -payid [payment ID]`: get payment infomation for a given payment id
* `-dbview deposit -depositid [deposit job ID]`: get deposit job info for a given deposit id
* `-dbview deposit -cid [channel ID]`: get all deposit jobs for a given channel id
* `-dbview route -dest [destination addr] -token [token addr]`: get route info

#### Off-chain state slow query (table scan)
* `-dbview channel -cid [channel ID] -payhistory`: get channel info including full pay history
* `-dbview channel -list -token [token addr] -chanstate [int]`: list channel ids for given token and state
* `-dbview channel -list -detail -token [token addr] -chanstate [int]`: list channel details for given token and state
* `-dbview channel -list -token [token addr] -chanstate [int] -inactiveday [int]`: list inactive channel ids
* `-dbview channel -count [token addr] -chanstate [int] -inactiveday [int]`: count inactive channels
* `-dbview channel -balance -token [token addr] -chanstate [int]`: get the total balance of all my channels

Note: `chanstate` is enum integer, valid states for commands above include 3 for *opened* and 4 for *settling*. Default chanstate is 3 if arg is not provided in command.

### Query information from blockchain
`osp-cli -profile [profile file]` followed by:

* `-onchainview channel -cid [channel ID]`: get onchain channel info from CelerLedger contract
* `-onchainview pay -payid [payment ID]`: get onchain payment info from PayRegistry contract
* `-onchainview tx -txhash [transaction hash]`: get on-chain transaction information
* `-onchainview app -appaddr [app contract addr] -outcome [arg for query outcome] -finalize [arg for query finalization] -decode`: get onchain CelerApp session info

### Unilateral channel settle or withdraw action 

`osp-cli -profile [profile file] -ks [keystore file] -storedir [sqlite store directory]` followed by:

* `-intendsettle -cid [channel ID]`: intend unilaterally settle a channel
* `-confirmsettle -cid [channel ID]`: confirm unilaterally settle a channel
* `-intendsettle -batchfile [file with a list of channel IDs]`: intend unilaterally settle a list of channels
* `-confirmsettle -batchfile [file with a list of channel IDs]`: confirm unilaterally settle a list of channels
* `-intendwithdraw -cid [channel ID] -amount [amount]`: intend unilaterally withdraw from a channel
* `-confirmwithdraw -cid [channel ID]`: confirm unilaterally withdraw from a channel

### OSP account on-chain operation

`osp-cli -profile [profile file] -ks [keystore file]` followed by:

* `-ethpooldeposit -amount [amount]`: deposit ETH into EthPool and approve to CelerLedger
* `-ethpoolwithdraw -amount [amount]`: withdraw ETH from EthPool
* `-register`: register OSP as a state channel router
* `-deregister`: deregister OSP as a state channel router
