// Copyright 2018-2020 Celer Network

package e2e

import "github.com/celer-network/goCeler/ctype"

const (
	// outPathPrefix is the path prefix for all output from e2e (incl. chain data, binaries etc)
	// the code will append epoch second to this and create the folder
	// the folder will be deleted after test ends successfully
	outRootDirPrefix = "/tmp/celer_e2e_"

	// etherbase and osp addr/priv key in hex
	etherBaseAddr    = "b5bb8b7f6f1883e0c01ffb8697024532e6f3238c"
	osp1EthAddr      = "6a6d2a97da1c453a4e099e8054865a0a59728863"
	osp2EthAddr      = "ba756d65a1a03f07d205749f35e2406e4a8522ad"
	osp3EthAddr      = "90506865876ac95f642c8214cc3b570121ba4e4e"
	osp4EthAddr      = "59fb4bc14618ee4e9e83ff563c72b7fd1222731c"
	osp5EthAddr      = "cc09df4fe5e09ff9dec5781170f146990dd77ba3"
	depositorEthAddr = "d0c5e4abfadbc0e71bfb6c4955e66b8a6bf4da51"
	ospEthAddr       = osp1EthAddr

	etherBasePriv = "69ef4da8204644e354d759ca93b94361474259f63caac6e12d7d0abcca0063f8"
	osp1Priv      = "a7c9fa8bcd45a86fdb5f30fecf88337f20185b0c526088f2b8e0f726cad12857"
	osp2Priv      = "c2ff7d4ce25f7448de00e21bbbb7b884bb8dc0ca642031642863e78a35cb933d"
	osp3Priv      = "ec88168be62ef61b5d189415254370ac2118fdf42fbd51e2c944f60f6fc437b3"
	osp4Priv      = "a54790ec0e8a172fc37f727a44c356eb957ff6438ba981906c6d8f3d62edbeee"
	osp5Priv      = "16212eaf5b15f468306e81af23c7e6036860ed117745015de6f84bdff5ba5b2c"
	depositorPriv = "c76dc6d854247299174f9582566cd195a57d8236a7fb2d0c085377f274584cec"

	ethGateway = "http://127.0.0.1:8545"

	// try to do some allocation for port: 10xyz are osp,
	// x is osp 0-based index
	// yz are osp ports like grpc, adminweb, selfrpc etc
	localhost = "127.0.0.1:"

	sPort     = "10000"
	sSelfRPC  = "localhost:30000"
	sAdminRPC = "localhost:11000"
	sAdminWeb = "localhost:8090"

	// below are ports/addrs for multi-osp tests
	// o1 is the server for osp1, it is mapped to the default server above
	o1Port     = sPort
	o1SelfRPC  = sSelfRPC
	o1AdminRPC = sAdminRPC
	o1AdminWeb = sAdminWeb
	// o2 is the single server of osp2
	o2Port     = "10002"
	o2AdminRPC = "localhost:11002"
	o2AdminWeb = "localhost:8290"
	// o3 is the single server of osp3
	o3Port     = "10003"
	o3AdminRPC = "localhost:11003"
	o3AdminWeb = "localhost:8390"
	// o4 is the single server of osp4
	o4Port     = "10004"
	o4AdminRPC = "localhost:11004"
	o4AdminWeb = "localhost:8490"
	// o5 is the single server of osp5
	o5Port     = "10005"
	o5AdminRPC = "localhost:11005"
	o5AdminWeb = "localhost:8590"

	sStoreDir               = "/tmp/sStore"
	c1StoreSettleDisputeDir = "/tmp/c1StoreSettleDispute"
	oracleStoreDir          = "/tmp/oracleStore"

	sendAmt      = "1"
	tokenAddrEth = ctype.EthTokenAddrStr

	accountBalance      = "50000000000000000000" // 50 ETH
	initialBalance      = "5000000000000000000"  // 5 ETH
	initOspToOspBalance = "8000000000000000000"  // 8 ETH

	rtConfig         = "../../testing/profile/rt_config.json"
	rtConfigMultiOSP = "../../testing/profile/rt_config_multiosp.json"
	tokensConfig     = "../../testing/profile/tokens.json"

	osp1Keystore    = "../../testing/env/keystore/osp1.json"
	osp2Keystore    = "../../testing/env/keystore/osp2.json"
	osp3Keystore    = "../../testing/env/keystore/osp3.json"
	osp4Keystore    = "../../testing/env/keystore/osp4.json"
	osp5Keystore    = "../../testing/env/keystore/osp5.json"
	depositKeystore = "../../testing/env/keystore/osp1_depositor.json"
	ospKeystore     = osp1Keystore
)
