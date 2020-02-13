// Copyright 2018 Celer Network

package cnode

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/celer-network/goCeler-oss/chain"
	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/payresolver"
	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/routerregistry"
	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/virtresolver"
	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/wallet"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/cnode/cooperativewithdraw"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/cobj"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/dispatchers"
	"github.com/celer-network/goCeler-oss/dispute"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/handlers"
	"github.com/celer-network/goCeler-oss/ledgerview"
	"github.com/celer-network/goCeler-oss/messager"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/celer-network/goCeler-oss/route"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/transactor"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/celer-network/goCeler-oss/watcher"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const mutexLocked = 1 << iota

type CNode struct {
	transactorPool     *transactor.Pool
	nodeConfig         common.GlobalNodeConfig
	streamWriter       common.StreamWriter
	crypto             common.Crypto
	channelRouter      common.StateChannelRouter
	celerMsgDispatcher *dispatchers.CelerMsgDispatcher
	monitorService     intfs.MonitorService
	messager           *messager.Messager

	// eth address of the node (hex string)
	EthAddress string
	// private key (hex string)
	PrivKey string

	connManager *rpc.ConnectionManager

	onNewStreamLock sync.RWMutex
	onNewStream     event.OnNewStreamCallback

	// PSP server ETH address (for convenience)
	ServerAddr string

	ethclient                    *ethclient.Client
	kvstore                      storage.KVStore
	dal                          *storage.DAL
	watch                        *watcher.WatchService
	openChannelProcessor         *openChannelProcessor
	cooperativeWithdrawProcessor *cooperativewithdraw.Processor
	depositProcessor             *depositProcessor
	isOSP                        bool
	listenOnChain                bool
	// on-chain dispute related data structrues and functions
	Disputer      *dispute.Processor
	rtEdgeBuilder routingTableEdgeBuilder

	// For the multi-server setup.
	isMultiServer   bool
	serverForwarder handlers.ForwardToServerCallback

	// grpc client streaminterceptor. currently only used by client for test
	msgDropper *clientStreamMsgDropper
}
type routingTableEdgeBuilder interface {
	AddEdge(ctype.Addr, ctype.Addr, ctype.CidType, ctype.Addr) error
	RemoveEdge(ctype.CidType) error
	Build(ctype.Addr) (map[ctype.Addr]ctype.CidType, error)
	GetOspInfo() map[ctype.Addr]uint64
}

func (c *CNode) GetConnManager() *rpc.ConnectionManager {
	return c.connManager
}

func (c *CNode) GetKVStore() storage.KVStore {
	return c.kvstore
}

func (c *CNode) RegisterStream(peerOnChainAddr string, peerHTTPTarget string) error {

	conn, err := grpc.Dial(peerHTTPTarget, utils.GetClientTlsOption(), grpc.WithBlock(),
		grpc.WithTimeout(4*time.Second), grpc.WithKeepaliveParams(config.KeepAliveClientParams),
		grpc.WithStreamInterceptor(c.streamInterceptor))
	if err != nil {
		log.Errorln("RegisterStream:", err, peerHTTPTarget)
		return fmt.Errorf("grpcDial %s failed: %s", peerHTTPTarget, err)
	}
	c.connManager.AddConnection(peerOnChainAddr, conn)
	dialClient, err := c.connManager.GetClient(peerOnChainAddr)
	if err != nil {
		return fmt.Errorf("getClient fail: %s", err)
	}

	celerStream, dialErr := dialClient.CelerStream(context.Background())
	if dialErr != nil {
		log.Errorln("CelerStream failed:", dialErr)
		return fmt.Errorf("CelerStream failed: %s", dialErr)
	}
	if celerStream == nil {
		log.Errorln("nil celer stream: gRPC connection is not established")
		return errors.New("nil celer stream: gRPC connection is not established")
	}
	ts, tsSig := c.getTsAndSig()
	authReq := &rpc.AuthReq{
		MyAddr:     c.nodeConfig.GetOnChainAddrBytes(),
		Timestamp:  ts,
		MySig:      tsSig,
		ExpectPeer: ctype.Hex2Bytes(peerOnChainAddr),
	}
	sendErr := celerStream.Send(&rpc.CelerMsg{
		ToAddr:  ctype.Hex2Bytes(peerOnChainAddr),
		Message: &rpc.CelerMsg_AuthReq{AuthReq: authReq},
	})
	if sendErr != nil {
		log.Errorln(sendErr)
		return sendErr
	}
	msgProcessor := c.celerMsgDispatcher.NewStream(ctype.Hex2Addr(peerOnChainAddr))
	c.connManager.AddCelerStream(peerOnChainAddr, celerStream, msgProcessor)

	// Notify application about new streams and they are ready to SEND/RECEIVE pay.

	peerByte, err := hex.DecodeString(peerOnChainAddr)
	if err != nil {
		log.Errorln("RegisterStream:", err)
		return errors.New("MALFORMED_PEER_ADDR")
	}
	c.onNewStreamLock.RLock()
	if c.onNewStream != nil {
		log.Debugf("Notifying onNewStream %x", peerByte)
		go c.onNewStream.HandleNewCelerStream(peerByte)
	}
	c.onNewStreamLock.RUnlock()
	return nil
}

func (c *CNode) OnReceivingToken(callback event.OnReceivingTokenCallback) {
	// OnReceivingToken is not thread-safe since we'll need to pass lock
	log.Debugln("Changing receiving callback")
	c.celerMsgDispatcher.OnReceivingToken(callback)
	log.Debugln("receiving callback changed.")
}
func (c *CNode) OnNewStream(callback event.OnNewStreamCallback) {
	c.onNewStreamLock.Lock()
	defer c.onNewStreamLock.Unlock()
	c.onNewStream = callback
}
func (c *CNode) RegisterStreamErrCallback(peerOnChainAddr string, callback rpc.ErrCallbackFunc) {
	c.connManager.AddErrorCallback(peerOnChainAddr, callback)
}

// Counter for stream with wrong signatures. During app upgrade, this metric can be used to
// count how much traffic left unupgraded.
var addStreamWrongSig = expvar.NewMap("AddStreamWrongSig").Init()

// AddCelerStream checks the auth req in celerMsg coming. If auth passes, start listener on the stream and
// add the stream to connection manager.
func (c *CNode) AddCelerStream(celerMsg *rpc.CelerMsg, stream rpc.CelerStream) (context.Context, error) {
	authReq := celerMsg.GetAuthReq()
	if authReq == nil {
		return nil, errors.New("no AuthReq msg")
	}
	if !isTimestampValid(authReq.Timestamp) {
		return nil, errors.New("invalid timestamp")
	}
	src := authReq.GetMyAddr()
	tsByte := uint64ToBytes(authReq.Timestamp)
	if !c.crypto.SigIsValid(ctype.Bytes2Hex(src), tsByte, authReq.GetMySig()) {
		addStreamWrongSig.Add("Celer", 1)
		return nil, errors.New("invalid signature")
	}
	msgProcessor := c.celerMsgDispatcher.NewStream(ctype.Bytes2Addr(src))
	ctx := c.connManager.AddCelerStream(ctype.Bytes2Hex(src), stream, msgProcessor)
	c.onNewStreamLock.RLock()
	if c.onNewStream != nil {
		log.Debugln("calling callback onNewStream src", ctype.Bytes2Hex(src))
		go c.onNewStream.HandleNewCelerStream(src)
	}
	c.onNewStreamLock.RUnlock()
	return ctx, nil
}

func (c *CNode) getTsAndSig() (ts uint64, sig []byte) {
	ts = uint64(time.Now().Unix())
	sig, _ = c.crypto.Sign(uint64ToBytes(ts))
	return ts, sig
}

func uint64ToBytes(i uint64) []byte {
	ret := make([]byte, 8) // 8 bytes for uint64
	binary.BigEndian.PutUint64(ret, i)
	return ret
}

func isTimestampValid(ts uint64) bool {
	now := uint64(time.Now().Unix())
	return ts >= now-config.AllowedTimeWindow && ts <= now+config.AllowedTimeWindow
}

func (c *CNode) SetMsgDropper(dropRecv, dropSend bool) {
	c.msgDropper.dropRecv = dropRecv
	c.msgDropper.dropSend = dropSend
	log.Infoln("SetMsgDropper to dropRecv", dropRecv, "dropSend", dropSend)
}

// streamInterceptor is grpc.StreamClientInterceptor to be used by grpc.WithStreamInterceptor
func (c *CNode) streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	if c.msgDropper == nil {
		c.msgDropper = &clientStreamMsgDropper{
			ClientStream: s,
		}
	} else {
		c.msgDropper.ClientStream = s
	}
	return c.msgDropper, nil
}

//------------------------------main logic--------------------------------

func (c *CNode) IntendSettlePaymentChannel(cid ctype.CidType) error {
	if err := c.Disputer.IntendSettlePaymentChannel(cid); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *CNode) ConfirmSettlePaymentChannel(cid ctype.CidType) error {
	if err := c.Disputer.ConfirmSettlePaymentChannel(cid); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *CNode) GetSettleFinalizedTime(cid ctype.CidType) (*big.Int, error) {
	return ledgerview.GetOnChainSettleFinalizedTime(cid, c.nodeConfig)
}

func getAddrPrivKey(keyStore string, passPhrase string) (string, string, error) {
	key, err := keystore.DecryptKey([]byte(keyStore), passPhrase)
	if err != nil {
		return "", "", err
	}
	addr := ctype.Addr2Hex(key.Address)
	privKey := ctype.Bytes2Hex(crypto.FromECDSA(key.PrivateKey))
	return addr, privKey, nil
}

func NewCNode(
	keyStore string,
	passPhrase string,
	transactorConfigs []*transactor.TransactorConfig,
	config common.CProfile,
	routingPolicy common.RoutingPolicy,
	routingData []byte) (*CNode, error) {
	rc := &CNode{}

	log.Infoln("CNode config:", config)

	// ctx, _ := context.WithTimeout(context.Background(), 4*time.Second)
	err := rc.initialize(
		context.Background(),
		keyStore,
		passPhrase,
		transactorConfigs,
		&config,
		routingPolicy,
		routingData)
	if err != nil {
		log.Errorf("cNode init error: %s", err)
		return nil, err
	}
	log.Info("Finishing NewCNode")
	return rc, nil
}

func (c *CNode) initialize(
	ctx context.Context,
	keyStore string,
	passPhrase string,
	transactorConfigs []*transactor.TransactorConfig,
	profile *common.CProfile,
	routingPolicy common.RoutingPolicy,
	routingData []byte) error {
	// Seed the rand
	rand.Seed(time.Now().UTC().UnixNano())

	c.isOSP = profile.IsOSP
	c.initMultiServer(profile)
	if c.isMultiServer {
		return errors.New("no multiserver support")
	}
	c.listenOnChain = true

	if profile.DisputeTimeout != 0 {
		config.ChannelDisputeTimeout = profile.DisputeTimeout
	}

	// Initialize the server storage layer.
	if err := c.setupKVStore(profile, keyStore); err != nil {
		c.Close()
		return err
	}

	// set up node account
	var privKeyErr error
	c.EthAddress, c.PrivKey, privKeyErr = getAddrPrivKey(keyStore, passPhrase)
	if privKeyErr != nil {
		c.Close()
		return privKeyErr
	}
	c.ServerAddr = profile.SvrETHAddr

	// initialize on-chain transactor
	// create an IPC based RPC connection to a remote node and an authorized transactor
	ethCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var wsOrigin string
	if profile.WsOrigin != "" {
		wsOrigin = profile.WsOrigin
	} else {
		wsOrigin = "http://celer.network"
	}
	var rpcClient *ethrpc.Client
	var err error
	ethInstance := profile.ETHInstance
	if strings.HasPrefix(ethInstance, "ws") {
		rpcClient, err = ethrpc.DialWebsocket(ethCtx, ethInstance, wsOrigin)
		if err != nil {
			// Retry once for stability.
			time.Sleep(time.Second)
			rpcClient, err = ethrpc.DialWebsocket(ethCtx, ethInstance, wsOrigin)
			if err != nil {
				log.Errorln("Dial ETHInstance WS failed.")
				c.Close()
				return fmt.Errorf("DialETH failed: %s", err)
			}
		}
	} else {
		rpcClient, err = ethrpc.Dial(ethInstance)
		if err != nil {
			// Retry once for stability.
			time.Sleep(time.Second)
			rpcClient, err = ethrpc.Dial(ethInstance)
			if err != nil {
				log.Errorln("Dial ETHInstance HTTP failed.")
				c.Close()
				return fmt.Errorf("DialETH failed: %s", err)
			}
		}
	}
	c.ethclient = ethclient.NewClient(rpcClient)
	// Create transactor pool. If the list of transactor keys isn't specified, use the signing key
	// as the sole transactor.
	if len(transactorConfigs) == 0 {
		transactorConfigs = []*transactor.TransactorConfig{
			transactor.NewTransactorConfig(keyStore, passPhrase),
		}
	}

	blockDelay := profile.BlockDelayNum

	chainId := big.NewInt(profile.ChainId)
	c.transactorPool, err = transactor.NewPool(c.ethclient, blockDelay, chainId, transactorConfigs)
	if err != nil {
		c.Close()
		return err
	}

	// In the multi-server setup provide the connection manager with
	// a callback to register clients that connect at this OSP.
	var regClient rpc.RegisterClientCallbackFunc
	c.connManager = rpc.NewConnectionManager(regClient)

	// Initialize the watcher service.
	log.Infof("Setting up watch service (block delay %d)", blockDelay)
	var pollingInterval uint64 = 10
	if profile.PollingInterval != 0 {
		pollingInterval = profile.PollingInterval
	}
	c.watch = watcher.NewWatchService(c.ethclient, c.dal, pollingInterval)
	if c.watch == nil {
		log.Error("Cannot setup watch service")
		c.Close()
		return fmt.Errorf("newWatchService failed: %s", err)
	}

	c.nodeConfig = cobj.NewCelerGlobalNodeConfig(
		c.EthAddress,
		c.ethclient,
		profile,
		wallet.CelerWalletABI,
		ledger.CelerLedgerABI,
		virtresolver.VirtContractResolverABI,
		payresolver.PayResolverABI,
		payregistry.PayRegistryABI,
		routerregistry.RouterRegistryABI,
	)
	// Init monitor service
	monitorService :=
		monitor.NewService(c.watch, blockDelay, !c.isOSP || c.listenOnChain, c.GetRPCAddr())
	monitorService.Init()
	c.monitorService = monitorService
	c.channelRouter = cobj.NewCelerChannelRouter(routingPolicy, c.dal, ctype.Hex2Addr(profile.SvrETHAddr))
	c.crypto = cobj.NewCelerCrypto(c.PrivKey, c.EthAddress)
	c.streamWriter = cobj.NewCelerStreamWriter(c.connManager)

	selfAddress := c.nodeConfig.GetOnChainAddr()
	ledger := c.nodeConfig.GetLedgerContract().(*chain.BoundContract)
	c.cooperativeWithdrawProcessor, err = cooperativewithdraw.StartProcessor(
		selfAddress,
		c.crypto,
		c.transactorPool,
		c.connManager,
		c.monitorService,
		c.dal,
		c.streamWriter,
		ledger,
		c,
		c.isOSP,
		!c.isOSP)
	if err != nil {
		c.Close()
		return err
	}
	masterTransactor, err :=
		transactor.NewTransactor(keyStore, passPhrase, chainId, c.ethclient, blockDelay)
	if err != nil {
		c.Close()
		return err
	}
	if routingPolicy == common.ServiceProviderPolicy && c.listenOnChain {
		builder := route.NewRoutingTableBuilder(ctype.Hex2Addr(c.nodeConfig.GetOnChainAddr()), c.dal)
		if builder == nil {
			c.Close()
			return errors.New("Fail to initialize routing table builder")
		}
		c.rtEdgeBuilder = builder

		if err = route.StartRoutingRecoverProcess(c.GetCurrentBlockNumber(), routingData, c.nodeConfig, builder); err != nil {
			c.Close()
			return err
		}

		routerProcessor := route.NewRouterProcessor(
			c.nodeConfig, masterTransactor, c.monitorService,
			builder, 0 /*blkDelay*/, "localhost"+profile.WebPort /*adminWebHostAndPort*/)
		routerProcessor.Start()
	} else {
		c.rtEdgeBuilder = nil
	}
	c.openChannelProcessor, err = startOpenChannelProcessor(
		c.nodeConfig,
		c.crypto,
		masterTransactor,
		c.dal,
		ledger,
		c.connManager,
		c.monitorService,
		c,
		c.rtEdgeBuilder,
		c.isOSP)
	if err != nil {
		c.Close()
		return err
	}

	c.depositProcessor, err = startDepositProcessor(
		c.nodeConfig,
		masterTransactor,
		c.dal,
		ledger,
		c.monitorService,
		c,
		!c.isOSP)
	if err != nil {
		c.Close()
		return err
	}

	c.Disputer = dispute.NewProcessor(
		c.nodeConfig,
		masterTransactor,
		c.transactorPool,
		c.rtEdgeBuilder,
		c.monitorService,
		c.dal,
		c.isOSP)

	c.serverForwarder = c.defServerForwarder

	c.messager = messager.NewMessager(
		c.nodeConfig,
		c.crypto,
		c.streamWriter,
		c.channelRouter,
		c.monitorService,
		c.serverForwarder,
		c.dal)
	c.connManager.SetMsgQueueCallback(c.messager.EnableMsgQueue, c.messager.DisableMsgQueue)

	c.celerMsgDispatcher = dispatchers.NewCelerMsgDispatcher(
		c.nodeConfig,
		c.streamWriter,
		c.crypto,
		c.channelRouter,
		c.monitorService,
		c.dal,
		c.cooperativeWithdrawProcessor,
		c.serverForwarder,
		c.Disputer,
		c.messager)

	return nil
}

func (c *CNode) recoverCallbacks() error {

	return nil
}

//----------------------State Persistence-----------------------
// Initialize the server key/value storage.
func (c *CNode) setupKVStore(profile *common.CProfile, keyStore string) error {
	type addrOnly struct {
		Address string
	}
	var ksjson addrOnly
	err := json.Unmarshal([]byte(keyStore), &ksjson)
	if err != nil {
		log.Error(err)
	}

	dir := profile.StoreDir
	if dir == "" {
		dir = fmt.Sprintf("/tmp/cnodestore/%s", ksjson.Address)
	} else {
		dir = fmt.Sprintf("%s/%s", dir, ksjson.Address)
	}
	log.Infof("Setting up local kvstore at %s", dir)
	c.kvstore, err = storage.NewKVStoreLocal(dir, false)
	if err != nil {
		err = fmt.Errorf("Cannot setup local store: %s: %s", dir, err)
		log.Errorln(err)
		return err
	}

	c.dal = storage.NewDAL(c.kvstore)
	return nil
}

func (c *CNode) Close() {
	if c.kvstore != nil {
		c.kvstore.Close()
		c.kvstore = nil
		c.dal = nil
	}
	if c.ethclient != nil {
		c.ethclient.Close()
	}
}

// waitTimeout emits a signal to the timeoutChan when the given timeout has passed
func (c *CNode) waitTimeout(ctx context.Context, timeoutChan chan bool, timeout *big.Int) {
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()
	currentBlkNum := c.GetCurrentBlockNumber()
	deadline := big.NewInt(0)
	deadline.Add(currentBlkNum, timeout)
	for {
		blockNum := c.GetCurrentBlockNumber()
		if blockNum.Cmp(deadline) > 0 {
			select {
			case timeoutChan <- true:
			default:
			}
			return
		}
		// Wait for the next round.
		select {
		case <-ctx.Done():
			return
		case <-queryTicker.C:
		}
	}
}

func (c *CNode) GetCurrentBlockNumber() *big.Int {
	return c.monitorService.GetCurrentBlockNumber()
}

func (c *CNode) CooperativeWithdraw(
	cid ctype.CidType,
	tokenType entity.TokenType,
	tokenAddress string,
	amount string,
	callback cooperativewithdraw.Callback) (string, error) {
	return c.cooperativeWithdrawProcessor.CooperativeWithdraw(
		cid, tokenType, tokenAddress, amount, callback)
}

func (c *CNode) MonitorCooperativeWithdrawJob(
	withdrawHash string,
	callback cooperativewithdraw.Callback) {
	c.cooperativeWithdrawProcessor.MonitorCooperativeWithdrawJob(withdrawHash, callback)
}

func (c *CNode) RemoveCooperativeWithdrawJob(withdrawHash string) {
	c.cooperativeWithdrawProcessor.RemoveCooperativeWithdrawJob(withdrawHash)
}

func (c *CNode) BuildRoutingTable(token ctype.Addr) (map[ctype.Addr]ctype.CidType, error) {
	return c.rtEdgeBuilder.Build(token)
}
func (c *CNode) GetActiveOsps() map[ctype.Addr]bool {
	osps := c.rtEdgeBuilder.GetOspInfo()
	activeOsps := make(map[ctype.Addr]bool)
	for osp := range osps {
		if c.connManager.GetCelerStream(ctype.Addr2Hex(osp)) != nil {
			activeOsps[osp] = true
		}
	}
	return activeOsps
}

// GetPaymentState returns the ingress and egress state of a payment
func (c *CNode) GetPaymentState(payID ctype.PayIDType) (string, string) {
	_, inState, _, _ := c.dal.GetPayIngressState(payID)
	_, outState, _, _ := c.dal.GetPayEgressState(payID)
	return inState, outState
}
