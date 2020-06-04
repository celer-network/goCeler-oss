// Copyright 2018-2020 Celer Network

package cnode

import (
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/routerregistry"
	"github.com/celer-network/goCeler/chain/channel-eth-go/virtresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/wallet"
	"github.com/celer-network/goCeler/cnode/cooperativewithdraw"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/cobj"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/deposit"
	"github.com/celer-network/goCeler/dispatchers"
	"github.com/celer-network/goCeler/dispute"
	"github.com/celer-network/goCeler/handlers"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/lrucache"
	"github.com/celer-network/goCeler/messager"
	"github.com/celer-network/goCeler/migrate"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/watcher"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const mutexLocked = 1 << iota

var dropMsg = flag.Bool("dropmsg", false, "add grpc interceptor to test drop msg, only use for tests.")

type CNode struct {
	transactorPool     *eth.TransactorPool
	nodeConfig         common.GlobalNodeConfig
	streamWriter       common.StreamWriter
	celerMsgDispatcher *dispatchers.CelerMsgDispatcher
	monitorService     intfs.MonitorService
	messager           *messager.Messager

	EthAddress        ctype.Addr // ETH address of the node
	signer            eth.Signer
	externalSigner    bool // if the signer is external
	masterTransactor  *eth.Transactor
	depositTransactor *eth.Transactor

	connManager *rpc.ConnectionManager

	onNewStreamLock sync.RWMutex
	onNewStream     event.OnNewStreamCallback

	// OSP server ETH address (for convenience)
	ServerAddr ctype.Addr

	ethclient                    *ethclient.Client
	kvstore                      storage.KVStore
	dal                          *storage.DAL
	watch                        *watcher.WatchService
	openChannelProcessor         *openChannelProcessor
	cooperativeWithdrawProcessor *cooperativewithdraw.Processor
	depositProcessor             *deposit.Processor
	isOSP                        bool
	listenOnChain                bool
	Disputer                     *dispute.Processor
	routeForwarder               *route.Forwarder
	routeController              *route.Controller
	migrateChannelProcessor      *migrate.MigrateChannelProcessor

	// For the multi-server setup.
	isMultiServer   bool
	clientCache     *lrucache.LRUCache // dest client -> osp server
	serverCache     *lrucache.LRUCache // osp server -> grpc conn
	serverCacheLock sync.Mutex
	serverForwarder handlers.ForwardToServerCallback

	AppClient *app.AppClient

	// grpc client streaminterceptor. currently only used by client for test
	msgDropper *clientStreamMsgDropper

	// signal for goroutines to exit
	quit chan bool
}

func (c *CNode) GetConnManager() *rpc.ConnectionManager {
	return c.connManager
}

func (c *CNode) GetKVStore() storage.KVStore {
	return c.kvstore
}

func (c *CNode) GetDAL() *storage.DAL {
	return c.dal
}

func (c *CNode) dialOpts(drop bool) []grpc.DialOption {
	opts := []grpc.DialOption{
		utils.GetClientTlsOption(), grpc.WithBlock(),
		grpc.WithTimeout(config.GrpcDialTimeout * time.Second),
		grpc.WithKeepaliveParams(config.KeepAliveClientParams),
	}
	if drop {
		opts = append(opts, grpc.WithStreamInterceptor(c.streamInterceptor))
	}
	return opts
}

// WARNING: msg drop interceptor only supports single stream, and should only be used in testing
func (c *CNode) RegisterStream(peerAddr ctype.Addr, peerHTTPTarget string) error {
	if s := c.connManager.GetCelerStream(peerAddr); s != nil {
		return common.ErrStreamAleadyExists
	}
	conn, err := grpc.Dial(peerHTTPTarget, c.dialOpts(*dropMsg)...)
	if err != nil {
		return fmt.Errorf("grpcDial %s failed: %w", peerHTTPTarget, err)
	}
	c.connManager.AddConnection(peerAddr, conn)
	dialClient := rpc.NewRpcClient(conn)
	celerStream, err := dialClient.CelerStream(context.Background())
	if err != nil {
		return fmt.Errorf("CelerStream failed: %w", err)
	}
	authReq, err := c.getAuthReq(peerAddr)
	if err != nil {
		return fmt.Errorf("getAuthReq failed %w", err)
	}
	err = celerStream.Send(&rpc.CelerMsg{
		ToAddr:  peerAddr.Bytes(),
		Message: &rpc.CelerMsg_AuthReq{AuthReq: authReq},
	})
	if err != nil {
		return fmt.Errorf("celerStream Send failed %w", err)
	}
	// after NewStream, dispatcher spins goroutine blocks on reading from msgChan
	msgChan := c.celerMsgDispatcher.NewStream(peerAddr)
	// Note we block here on celerStream.Recv and ensure AuthAck is right
	// before moving on to commMgr.AddCelerStream which does loop Recv
	// if auth failed, server will return err so err2 isn't nil
	// have timeout to avoid blocking forever (esp. for sdk/mobile)
	celerMsg, err := waitRecvWithTimeout(celerStream, config.AuthAckTimeout)
	if err != nil {
		return fmt.Errorf("waitRecvWithTimeout failed %w", err)
	}
	ackMsg := celerMsg.GetAuthAck()
	if ackMsg == nil {
		// unlikely case that we connect to old version osp, first msg isn't authack
		log.Errorln("nil authack:", celerMsg)
		return common.ErrInvalidMsgType
	}
	// verify ackMsg and update db if needed
	c.HandleAuthAck(peerAddr, ackMsg)
	// AddCelerStream spins goroutine that blocks on celerStream.Recv and write to msgChan
	c.connManager.AddCelerStream(peerAddr, celerStream, msgChan)

	// Notify application about new streams and they are ready to SEND/RECEIVE pay.
	// note for now only server.go implements HandleNewCelerStream for delegate.
	// so code here only happens when one OSP dials to another.
	c.onNewStreamLock.RLock()
	if c.onNewStream != nil {
		log.Debugf("Notifying onNewStream %x", peerAddr)
		go c.onNewStream.HandleNewCelerStream(peerAddr)
	}
	c.onNewStreamLock.RUnlock()
	// When registering stream, start channel migration process
	go c.migrateChannelProcessor.CheckPeerChannelMigration(peerAddr)
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
func (c *CNode) RegisterStreamErrCallback(peerAddr ctype.Addr, callback rpc.ErrCallbackFunc) {
	c.connManager.AddErrorCallback(peerAddr, callback)
}

// AddCelerStream is called on server side after authReq passed
// add the stream to connection manager.
func (c *CNode) AddCelerStream(celerMsg *rpc.CelerMsg, stream rpc.CelerStream) (context.Context, error) {
	authReq := celerMsg.GetAuthReq()
	src := authReq.GetMyAddr()
	msgChan := c.celerMsgDispatcher.NewStream(ctype.Bytes2Addr(src))
	ctx := c.connManager.AddCelerStream(ctype.Bytes2Addr(src), stream, msgChan)
	c.onNewStreamLock.RLock()
	if c.onNewStream != nil {
		log.Debugln("calling callback onNewStream src", ctype.Bytes2Hex(src))
		go c.onNewStream.HandleNewCelerStream(ctype.Bytes2Addr(src))
	}
	c.onNewStreamLock.RUnlock()
	return ctx, nil
}

func (c *CNode) SetMsgDropper(dropRecv, dropSend bool) {
	if c.msgDropper != nil {
		c.msgDropper.dropRecv = dropRecv
		c.msgDropper.dropSend = dropSend
		log.Infoln("SetMsgDropper to dropRecv", dropRecv, "dropSend", dropSend)
	} else {
		log.Info("Ignore drop request due to nil msgDropper")
	}

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
	if err := c.Disputer.IntendSettlePaymentChannel(cid, true); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *CNode) ConfirmSettlePaymentChannel(cid ctype.CidType) error {
	if err := c.Disputer.ConfirmSettlePaymentChannel(cid, true); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *CNode) GetSettleFinalizedTime(cid ctype.CidType) (*big.Int, error) {
	return ledgerview.GetOnChainSettleFinalizedTime(cid, c.nodeConfig)
}

func NewCNode(
	masterTxConfig *eth.TransactorConfig,
	depositTxConfig *eth.TransactorConfig,
	transactorConfigs []*eth.TransactorConfig,
	profile common.CProfile,
	routingPolicy route.RoutingPolicy,
	routingData []byte) (*CNode, error) {
	return newCNode(
		masterTxConfig, depositTxConfig, transactorConfigs,
		ctype.ZeroAddr, nil, false, profile, routingPolicy, routingData)
}

// NewCNodeWithExternalSigner is only used by client
func NewCNodeWithExternalSigner(
	address ctype.Addr,
	signer eth.Signer,
	profile common.CProfile) (*CNode, error) {
	return newCNode(nil, nil, nil, address, signer, true, profile, route.GateWayPolicy, nil)
}

func newCNode(
	masterTxConfig *eth.TransactorConfig,
	depositTxConfig *eth.TransactorConfig,
	transactorConfigs []*eth.TransactorConfig,
	address ctype.Addr,
	signer eth.Signer,
	externalSigner bool,
	profile common.CProfile,
	routingPolicy route.RoutingPolicy,
	routingData []byte) (*CNode, error) {

	c := &CNode{quit: make(chan bool)}

	log.Infoln("CNode config:", profile)
	config.SetGlobalConfigFromProfile(&profile)

	err := c.setupEthClient(&profile)
	if err != nil {
		log.Errorln("cNode setup ethClient error:", err)
		return nil, err
	}

	if externalSigner {
		err = c.setupExternalTransactor(address, signer)
		if err != nil {
			log.Errorln("cNode setupExternalTransactor error:", err)
			return nil, err
		}
	} else {
		err = c.setupTransactor(masterTxConfig, depositTxConfig, transactorConfigs, &profile)
		if err != nil {
			log.Errorln("cNode setupTransactors error:", err)
			return nil, err
		}
	}

	err = c.initialize(&profile, routingPolicy, routingData)
	if err != nil {
		log.Errorln("cNode initialize error:", err)
		return nil, err
	}
	log.Info("Finishing NewCNode")
	return c, nil
}

func (c *CNode) setupEthClient(profile *common.CProfile) error {
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
	var err error
	var rpcClient *ethrpc.Client
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
				return fmt.Errorf("DialETH failed: %w", err)
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
				return fmt.Errorf("DialETH failed: %w", err)
			}
		}
	}
	c.ethclient = ethclient.NewClient(rpcClient)
	return nil
}

func (c *CNode) setupTransactor(
	masterTxConfig *eth.TransactorConfig,
	depositTxConfig *eth.TransactorConfig,
	transactorConfigs []*eth.TransactorConfig,
	profile *common.CProfile) error {

	var err error
	var privKey string
	// set up node account and signer
	c.EthAddress, privKey, err =
		eth.GetAddrPrivKeyFromKeystore(masterTxConfig.Keyjson, masterTxConfig.Passphrase)
	if err != nil {
		c.Close()
		return err
	}
	c.signer, err = eth.NewSigner(privKey)
	if err != nil {
		c.Close()
		return err
	}

	// Create transactor pool. If the list of transactor keys isn't specified, use the signing key
	// as the sole transactor.
	if len(transactorConfigs) == 0 {
		transactorConfigs = []*eth.TransactorConfig{
			eth.NewTransactorConfig(masterTxConfig.Keyjson, masterTxConfig.Passphrase),
		}
	}

	c.transactorPool, err = eth.NewTransactorPoolFromConfig(c.ethclient, transactorConfigs)
	if err != nil {
		c.Close()
		return err
	}
	c.masterTransactor, err =
		eth.NewTransactor(masterTxConfig.Keyjson, masterTxConfig.Passphrase, c.ethclient)
	if err != nil {
		c.Close()
		return err
	}
	if depositTxConfig != nil {
		c.depositTransactor, err =
			eth.NewTransactor(depositTxConfig.Keyjson, depositTxConfig.Passphrase, c.ethclient)
		if err != nil {
			c.Close()
			return err
		}
	} else {
		c.depositTransactor = c.masterTransactor
	}

	return nil
}

func (c *CNode) setupExternalTransactor(address ctype.Addr, signer eth.Signer) error {
	c.EthAddress = address
	c.signer = signer
	c.masterTransactor = eth.NewTransactorByExternalSigner(address, signer, c.ethclient)
	c.depositTransactor = c.masterTransactor
	var err error
	c.transactorPool, err = eth.NewTransactorPool([]*eth.Transactor{c.masterTransactor})
	if err != nil {
		c.Close()
		return err
	}
	return nil
}

func (c *CNode) initialize(
	profile *common.CProfile,
	routingPolicy route.RoutingPolicy,
	routingData []byte) error {
	// Seed the rand
	rand.Seed(time.Now().UnixNano())

	c.isOSP = profile.IsOSP
	c.initMultiServer(profile)
	if c.isMultiServer {
		c.listenOnChain = profile.ListenOnChain
	} else {
		c.listenOnChain = true
	}

	// Initialize the storage layer.
	err := c.setupKVStore(profile, c.EthAddress)
	if err != nil {
		c.Close()
		return err
	}

	c.ServerAddr = ctype.Hex2Addr(profile.SvrETHAddr)

	// In the multi-server setup provide the connection manager with
	// a callback to register clients that connect at this OSP.
	var regClient rpc.RegisterClientCallbackFunc
	if c.isMultiServer {
		regClient = c.registerClientForServer
	}
	c.connManager = rpc.NewConnectionManager(regClient)

	// Initialize the watcher service.
	c.watch = watcher.NewWatchService(c.ethclient, c.dal, config.BlockIntervalSec)
	if c.watch == nil {
		log.Error("Cannot setup watch service")
		c.Close()
		return fmt.Errorf("newWatchService failed: %w", err)
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
		c.dal,
	)

	if c.isOSP && c.listenOnChain {
		err = c.registerEventListener()
		if err != nil {
			c.Close()
			return err
		}
		if c.isMultiServer {
			go c.keepAliveEventListener()
		}
	}

	// Init monitor service
	monitorService := monitor.NewService(c.watch, config.BlockDelay, !c.isOSP || c.listenOnChain, c.GetRPCAddr())
	monitorService.Init()
	c.monitorService = monitorService
	c.streamWriter = cobj.NewCelerStreamWriter(c.connManager)

	c.cooperativeWithdrawProcessor, err = cooperativewithdraw.StartProcessor(
		c.nodeConfig,
		c.nodeConfig.GetOnChainAddr(),
		c.signer,
		c.transactorPool,
		c.connManager,
		c.monitorService,
		c.dal,
		c.streamWriter,
		c.isOSP,
		!c.isOSP)
	if err != nil {
		c.Close()
		return err
	}

	c.routeForwarder = route.NewForwarder(routingPolicy, c.dal, ctype.Hex2Addr(profile.SvrETHAddr))
	if routingPolicy == route.ServiceProviderPolicy && c.listenOnChain {
		c.routeController, err = route.NewController(
			c.nodeConfig,
			c.masterTransactor,
			c.monitorService,
			c.dal,
			c.signer,
			c.bcastSend,
			routingData,
			profile.SvrRPC,
			profile.ExplorerUrl)
		if err != nil {
			c.Close()
			return err
		}
		c.routeController.Start()
	} else {
		c.routeController = nil
	}

	c.depositProcessor, err = deposit.StartProcessor(
		c.nodeConfig,
		c.depositTransactor,
		c.dal,
		c.monitorService,
		c.isOSP,
		c.listenOnChain,
		c.quit)
	if err != nil {
		c.Close()
		return err
	}

	c.openChannelProcessor, err = startOpenChannelProcessor(
		c.nodeConfig,
		c.signer,
		c.masterTransactor,
		c.dal,
		c.connManager,
		c.monitorService,
		c.routeController,
		c.depositProcessor,
		c.isOSP)
	if err != nil {
		c.Close()
		return err
	}

	c.Disputer = dispute.NewProcessor(
		c.nodeConfig,
		c.masterTransactor,
		c.transactorPool,
		c.routeController,
		c.monitorService,
		c.dal,
		c.isOSP)

	c.AppClient = app.NewAppClient(
		c.nodeConfig,
		c.masterTransactor,
		c.transactorPool,
		c.monitorService,
		c.dal,
		c.signer)

	if c.isMultiServer {
		c.serverForwarder = c.multiServerForwarder
	} else {
		c.serverForwarder = c.defServerForwarder
	}

	c.messager = messager.NewMessager(
		c.nodeConfig,
		c.signer,
		c.streamWriter,
		c.routeForwarder,
		c.monitorService,
		c.serverForwarder,
		c.depositProcessor,
		c.dal,
		c.isOSP)
	c.connManager.SetMsgQueueCallback(c.messager.EnableMsgQueue, c.messager.DisableMsgQueue)

	c.celerMsgDispatcher = dispatchers.NewCelerMsgDispatcher(
		c.nodeConfig,
		c.streamWriter,
		c.signer,
		c.monitorService,
		c.dal,
		c.cooperativeWithdrawProcessor,
		c.serverForwarder,
		c.Disputer,
		c.routeForwarder,
		c.routeController,
		c.messager)

	c.migrateChannelProcessor = migrate.NewMigrateChannelProcessor(
		c.nodeConfig,
		c.signer,
		c.dal,
		c.connManager,
		c.monitorService,
		c.isOSP,
	)

	if c.isOSP {
		go c.runOspRoutineJob()
	}

	return nil
}

func (c *CNode) recoverCallbacks() error {

	return nil
}

func (c *CNode) SetDelegation(tokens []ctype.Addr, timeout int64) error {
	client, err := c.connManager.GetClient(c.ServerAddr)
	if err != nil {
		return err
	}
	delegatedTks := make([][]byte, 0, len(tokens))
	for _, tk := range tokens {
		delegatedTks = append(delegatedTks, tk.Bytes())
	}
	desc := &rpc.DelegationDescription{
		Delegator:         c.ServerAddr.Bytes(),
		Delegatee:         c.nodeConfig.GetOnChainAddr().Bytes(),
		ExpiresAfterBlock: c.monitorService.GetCurrentBlockNumber().Int64() + timeout,
		TokenToDelegate:   delegatedTks,
	}
	descBytes, err := proto.Marshal(desc)
	if err != nil {
		return err
	}
	sig, err := c.signer.SignEthMessage(descBytes)
	if err != nil {
		return err
	}
	req := &rpc.DelegationRequest{
		Proof: &rpc.DelegationProof{
			DelegationDescriptionBytes: descBytes,
			Signature:                  sig,
			Signer:                     c.nodeConfig.GetOnChainAddr().Bytes(),
		},
	}
	_, err = client.RequestDelegation(context.Background(), req)
	return err
}

//----------------------State Persistence-----------------------
// Initialize the server-side storage.
func (c *CNode) setupServerStore(db string) error {
	log.Infof("Setting up server store at %s", db)
	st, err := storage.NewKVStoreSQL("postgres", db)
	if err != nil {
		return fmt.Errorf("Cannot setup SQL store: %s: %w", db, err)
	}

	c.kvstore = st
	return nil
}

// Initialize the client-side storage.
func (c *CNode) setupClientStore(dir string) error {
	log.Infof("Setting up local store at %s", dir)
	fpath := filepath.Join(dir, "sqlite", "celer.db")
	st, err := storage.NewKVStoreSQL("sqlite3", fpath)
	if err != nil {
		return fmt.Errorf("Cannot setup local store: %s: %w", fpath, err)
	}
	c.kvstore = st
	return nil
}

func (c *CNode) setupKVStore(profile *common.CProfile, addr ctype.Addr) error {
	var err error
	if profile.StoreSql != "" {
		err = c.setupServerStore(profile.StoreSql)
	} else {
		dir := profile.StoreDir
		if dir == "" {
			dir = filepath.Join(os.TempDir(), "cnodestore")
		}
		dir = filepath.Join(dir, ctype.Addr2Hex(addr))
		err = c.setupClientStore(dir)
	}

	if err != nil {
		log.Errorln(err)
		return err
	}

	c.dal = storage.NewDAL(c.kvstore)
	return nil
}

func (c *CNode) Close() {
	close(c.quit)
	if c.kvstore != nil {
		c.kvstore.Close()
		c.kvstore = nil
		c.dal = nil
	}
	if c.ethclient != nil {
		c.ethclient.Close()
	}
	if c.monitorService != nil {
		// exit monitorEvent should also call watcher.Close which unregisters and inner_close
		c.monitorService.Close()
		time.Sleep(500 * time.Millisecond) // not required just to be nice so when watch.Close is called, map should be empty
		// must call watch.Close after monitorService.Close to avoid monitor re-creating watchers
		// note this should be only close(ws.quit) cause individual watches should have already Close
		c.watch.Close()
	}
	if c.connManager != nil {
		c.connManager.CloseNoRetry(c.ServerAddr)
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

func (c *CNode) DepositWithCallback(amt *big.Int, cid ctype.CidType, cb deposit.DepositCallback) (string, error) {
	return c.depositProcessor.DepositWithCallback(amt, cid, cb)
}

func (c *CNode) MonitorDepositJobWithCallback(jobID string, cb deposit.DepositCallback) {
	c.depositProcessor.MonitorJobWithCallback(jobID, cb)
}

func (c *CNode) RemoveDepositJob(jobID string) error {
	return c.depositProcessor.RemoveJob(jobID)
}

func (c *CNode) RequestDeposit(
	peerAddr ctype.Addr, tokenAddr ctype.Addr, toPeer bool, amount *big.Int, maxWait time.Duration) (string, error) {
	token := utils.GetTokenInfoFromAddress(tokenAddr)
	cid, state, found, err := c.dal.GetCidStateByPeerToken(peerAddr, token)
	if err != nil {
		return "", fmt.Errorf("GetCidByPeerToken err: %w", err)
	}
	if !found {
		return "", common.ErrChannelNotFound
	}
	if state != enums.ChanState_OPENED {
		return "", common.ErrInvalidChannelState
	}
	return c.depositProcessor.RequestDeposit(cid, amount, toPeer, maxWait)
}

func (c *CNode) QueryDeposit(depositID string) (int, string, error) {
	return c.depositProcessor.GetDepositState(depositID)
}

func (c *CNode) CooperativeWithdraw(cid ctype.CidType, amount *big.Int, callback cooperativewithdraw.Callback) (string, error) {
	return c.cooperativeWithdrawProcessor.CooperativeWithdraw(cid, amount, callback)
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
	if c.routeController != nil {
		return c.routeController.BuildTable(token)
	}
	return nil, fmt.Errorf("route controller not initialized")
}

func (c *CNode) RecvBcastRoutingInfo(in *rpc.RoutingRequest) error {
	if c.routeController != nil {
		return c.routeController.RecvBcastRoutingInfo(in)
	}
	return fmt.Errorf("route controller not initialized")
}

func (c *CNode) GetConnectedOsps() map[ctype.Addr]bool {
	connectedOsps := make(map[ctype.Addr]bool)
	if c.routeController != nil {
		peerOsps := c.routeController.GetAllNeighbors()
		for ospAddr := range peerOsps {
			if c.IsLocalPeer(ospAddr) {
				connectedOsps[ospAddr] = true
			}
		}
	} else if c.isMultiServer {
		res, err := utils.QueryPeerOsps(config.EventListenerHttp)
		if err != nil {
			log.Error(err)
			return nil
		}
		for _, osp := range res.PeerOsps {
			ospAddr := ctype.Hex2Addr(osp.GetOspAddress())
			if c.IsLocalPeer(ospAddr) {
				connectedOsps[ospAddr] = true
			}
		}
	}
	return connectedOsps
}

func (c *CNode) GetPeerOsps() map[ctype.Addr]*route.NeighborInfo {
	if c.routeController != nil {
		return c.routeController.GetAllNeighbors()
	}
	log.Warn("route controller not initialized")
	return nil
}

func (c *CNode) getConnectedOspCids() ([]ctype.CidType, error) {
	var connectedCids []ctype.CidType
	if c.routeController != nil {
		peerOsps := c.routeController.GetAllNeighbors()
		for ospAddr, osp := range peerOsps {
			if c.IsLocalPeer(ospAddr) {
				for _, cid := range osp.TokenCids {
					connectedCids = append(connectedCids, cid)
				}
			}
		}
	} else if c.isMultiServer {
		res, err := utils.QueryPeerOsps(config.EventListenerHttp)
		if err != nil {
			return nil, err
		}
		for _, osp := range res.PeerOsps {
			ospAddr := ctype.Hex2Addr(osp.GetOspAddress())
			if c.IsLocalPeer(ospAddr) {
				for _, tkcid := range osp.GetTokenCidPairs() {
					connectedCids = append(connectedCids, ctype.Hex2Cid(tkcid.GetCid()))
				}
			}
		}
	}

	return connectedCids, nil
}

// ProcessMigrateChannelRequest handles channel migration request
func (c *CNode) ProcessMigrateChannelRequest(in *rpc.MigrateChannelRequest) (*rpc.MigrateChannelResponse, error) {
	return c.migrateChannelProcessor.ProcessMigrateChannelRequest(in)
}

func (c *CNode) runOspRoutineJob() {
	clearPayTicker := time.NewTicker(config.OspClearPaysInterval)
	defer func() {
		clearPayTicker.Stop()
	}()

	for {
		select {
		case <-clearPayTicker.C:
			err := c.ClearPaymentsWithPeerOsps()
			if err != nil {
				log.Error(err)
			}
		}
	}
}
