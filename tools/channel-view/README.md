### Example commands

`go run channel_view.go -profile [OSP profile file path] -storedir [OSP sqlite store path]` followed by:

#### Off-chain state fast query
* `-dbview channel -cid [channel ID]`: print channel info for a given channel id
* `-dbview channel -peer [peer address] -token [token address]`: print channel info for given peer and token pair
* `-dbview pay -payid [payment ID]`: print payment infomation for a given payment id
* `-dbview deposit -depositid [deposit job ID]`: print deposit job info for a given deposit id
* `-dbview deposit -cid [channel ID]`: print all deposit jobs info for a given channel id
* `-dbview route -dest [destination address] -token [token address]`: print route info
* `-dbview stats -token [token address] -chanstate [channel state (3:opened,4:settling)]`: print number of channels with given token and channel state

#### Off-chain state slow query (table scan)
* `-dbview channel -cid [channel ID] -allpays`: print channel info including full pay history
* `-dbview allchan -token [token address]`: print all my channels of a given token
* `-dbview allchan -token [token address] -inactiveday [#days]`: print all channels being inactive for given days
* `-dbview balance -token [token address]`: print the total balance of all my channels of a given token
* `-dbview stats -token [token address] -chanstate [channel state] -allids`: print all channel ids with given token and channel state
* `-dbview stats -token [token address] -chanstate [channel state] -inactiveday [#days]`: print number of inactive channels with given token and channel state
* `-dbview stats -token [token address] -chanstate [channel state] -inactiveday [#days] -allids`: print all inactive channel ids with given token and channel state

#### On-chain state query
* `-chainview channel -cid [channel ID]`: print onchain channel info from CelerLedger contract
* `-chainview pay -payid [payment ID]`: print onchain payment info from PayRegistry contract
* `-chainview app -appaddr [app contract address] -outcome [arg for query outcome] -finalize [arg for query finalization] -decode`: print onchain CelerApp session info

#### Note
* no need to have `-token` for ETH token.
* no need to have `-storedir` for on-chain query.
* replace `-storedir` with `-storesql` followed by url if using a separate storage server.
