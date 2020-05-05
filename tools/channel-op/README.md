### Example commands

`go run channel_op.go -profile [OSP profile file path] -ks [OSP keystore] -storedir [osp sqlite store path]` followed by:

* `-intendsettle [channel ID]`: intend unilaterally settle a channel
* `-confirmsettle [channel ID]`: confirm unilaterally settle a channel
* `-batch -intendsettle [file with a list of channel IDs]`: intend unilaterally settle a list of channels
* `-batch -confirmsettle [file with a list of channel IDs]`: confirm unilaterally settle a list of channels
* `-intendwithdraw [channel ID] -withdrawamt [float amount in unit of 1e18]`: intend unilaterally withdraw from a channel
* `-confirmwithdraw [channel ID]`: confirm unilaterally withdraw from a channel

**Note**: replace `-storedir` with `-storesql` followed by url if using a separate storage server.
