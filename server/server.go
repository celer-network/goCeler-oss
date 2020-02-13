// Copyright 2018-2019 Celer Network

// Package main implements rpc server logic defined in rpc/rpc.proto
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/cnode"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/metrics"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/rtconfig"
	"github.com/celer-network/goCeler-oss/transactor"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	pjson          = flag.String("profile", "config/profile.json", "Path to profile json file")
	dbg            = flag.Bool("debug", false, "enable reflection and verbos log for debug")
	port           = flag.Int("port", 10000, "The server listening port")
	selfrpc        = flag.String("selfrpc", "", "Internal server host:port for inter-server communication")
	adminrpc       = flag.String("adminrpc", ":11000", "The server admin endpoint")
	adminweb       = flag.String("adminweb", ":8090", "The server admin http endpoint")
	ks             = flag.String("ks", "", "Path to keystore json file")
	noPassword     = flag.Bool("nopassword", false, "Assume empty password for keystores")
	transactorks   = flag.String("transactorks", "", "Paths to keystore json files for on-chain transactions, separated by comma")
	passwordDir    = flag.String("passworddir", "", "Path to the directory containing passwords")
	storedir       = flag.String("storedir", "", "Path to the store directory")
	showver        = flag.Bool("v", false, "Show version and exit")
	isosp          = flag.Bool("isosp", true, "Run as an OSP node")
	listenOnChain  = flag.Bool("loc", true, "Listen to on-chain log events")
	rtcfile        = flag.String("rtc", "rt_config.json", "runtime config json file path")
	ospMaxRedial   = flag.Int("ospmaxredial", 10, "max retry for osp to dial to another osp")
	ospRedialDelay = flag.Int64("ospredialdelay", 5, "retry delay (in sec) for osp to dial to another osp")
	routingData    = flag.String("routeData", "", "Path to routing data json file")
	tlsCert        = flag.String("tlscert", "", "Path to TLS cert file")
	tlsKey         = flag.String("tlskey", "", "Path to TLS private key file")
	tlsClient      = flag.Bool("tlsclient", false, "Require tls client cert by CelerCA")
)

var selfHostPort string

type server struct {
	cNode         *cnode.CNode
	jsonMarshaler *jsonpb.Marshaler
	netClient     *http.Client
	lastOcTs      int64      // timestamp in seconds of last processed open channel request
	lastOcTsLock  sync.Mutex // lock to protect r/w of lastOcTs
}
type serverInterOSP struct {
	svr      *server
	adminSvr *adminService
}
type adminService struct {
	// Note: adminService doesn't own cNode.
	cNode           *cnode.CNode
	ospEthToRPC     map[ctype.Addr]string
	ospEthToRPCLock sync.Mutex
	streamRetryCb   rpc.ErrCallbackFunc
}

func newAdminService(cNode *cnode.CNode) *adminService {
	adminS := &adminService{
		cNode:       cNode,
		ospEthToRPC: make(map[ctype.Addr]string),
	}
	adminS.streamRetryCb = func(addrStr string, streamErr error) {
		// Register the callback to handle stream errors and try to reconnect.
		log.Infoln("streamRetryCb triggered for", addrStr, streamErr)
		delay := time.Duration(*ospRedialDelay) * time.Second
		addr := ctype.Hex2Addr(addrStr)
		adminS.ospEthToRPCLock.Lock()
		ospRPC, ok := adminS.ospEthToRPC[addr]
		delete(adminS.ospEthToRPC, addr)
		adminS.ospEthToRPCLock.Unlock()

		if !ok {
			log.Errorln("Not finding RPC to redial peer", addr)
			return
		}

		for i := 0; i < *ospMaxRedial; i++ {
			log.Debugln("streamRetryCb: try to register again", addrStr)
			adminS.ospEthToRPCLock.Lock()
			if peerRPC := adminS.ospEthToRPC[addr]; peerRPC == ospRPC {
				return
			}
			err := adminS.cNode.RegisterStream(addrStr, ospRPC)
			if err == nil {
				log.Infoln("streamRetry:Cb successful re-register", addrStr)
				adminS.ospEthToRPC[addr] = ospRPC
				adminS.ospEthToRPCLock.Unlock()
				return
			}
			adminS.ospEthToRPCLock.Unlock()
			log.Errorln("streamRetryCb: register failed", addrStr, err)
			time.Sleep(delay)
		}
		log.Errorln("streamRetry:Cb retry dial", addrStr, "for", *ospMaxRedial, "times, giving up")
	}
	return adminS
}

func (s *server) CelerStream(stream rpc.Rpc_CelerStreamServer) error {
	var ctx context.Context
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if msg.GetAuthReq() != nil {
		log.Infof("server got AuthReq from %x", msg.GetAuthReq().GetMyAddr())
		ctx, err = s.cNode.AddCelerStream(msg, stream)
		if err != nil {
			log.Warnln("AddCelerStream err:", err.Error(), "authreq:", msg.GetAuthReq().String())
			return status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		log.Warnln("first message not AuthReq:", msg.GetMessage())
		return status.Error(codes.Unauthenticated, "must send AuthReq first")
	}
	<-ctx.Done()
	return nil
}

func (s *adminService) HandleOpenChannelFinish(cid ctype.CidType) {
	log.Infoln("open channel succeeded, cid:", cid.Hex())
}
func (s *adminService) HandleOpenChannelErr(e *common.E) {
	log.Errorln("open channel error, reason: ", e.Reason)
}
func (s *adminService) OspOpenChannel(ctx context.Context, in *rpc.OspOpenChannelRequest) (*empty.Empty, error) {
	log.Infof("OspOpenChannel: peer: %x, self depo: %s, peer depo: %s",
		in.GetPeerEthAddress(), in.GetSelfDepositAmtWei(), in.GetPeerDepositAmtWei())
	selfDeposit, ok := big.NewInt(0).SetString(in.GetSelfDepositAmtWei(), 10)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "wrong self deposit")
	}
	peerDeposit, ok := big.NewInt(0).SetString(in.GetPeerDepositAmtWei(), 10)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "wrong peer deposit")
	}
	tokenInfo := &entity.TokenInfo{
		TokenType:    in.GetTokenType(),
		TokenAddress: in.GetTokenAddress(),
	}
	existingCid, _ := s.cNode.GetChannelIdForPeer(ctype.Bytes2Hex(in.GetPeerEthAddress()), ctype.Bytes2Hex(in.GetTokenAddress()))
	if bytes.Compare(existingCid.Bytes(), ctype.ZeroCid.Bytes()) != 0 {
		log.Errorf("channel already exist: %x", existingCid)
		return nil, status.Errorf(codes.AlreadyExists, "channel already exist: %x", existingCid)
	}
	err := s.cNode.OpenChannel(ctype.Bytes2Addr(in.PeerEthAddress), selfDeposit, peerDeposit, tokenInfo, true /*ospToOspOpen*/, s)
	if err != nil {
		log.Errorf("failed to open channel: %s", err.Error())
		return nil, status.Errorf(codes.Unknown, "failed to open channel: %s", err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *server) CelerOpenChannel(ctx context.Context, in *rpc.OpenChannelRequest) (*rpc.OpenChannelResponse, error) {
	if in == nil {
		return nil, common.ErrInvalidArg
	}
	ocWait := rtconfig.GetOpenChanWaitSecond()
	if ocWait > 0 {
		now := time.Now().Unix()
		s.lastOcTsLock.Lock()
		if now >= s.lastOcTs+ocWait { // ok to proceed
			s.lastOcTs = now
			s.lastOcTsLock.Unlock()
			return s.cNode.ProcessOpenChannelRequest(in)
		}
		// rate limit, return error
		s.lastOcTsLock.Unlock()
		return nil, common.ErrRateLimited
	}
	// ocWait is 0, proceed directly
	return s.cNode.ProcessOpenChannelRequest(in)
}
func beautifyRT(table map[ctype.Addr]ctype.CidType) map[string]string {
	ret := make(map[string]string)
	for k, v := range table {
		ret[ctype.Addr2Hex(k)] = ctype.Cid2Hex(v)
	}
	return ret
}
func (s *adminService) ConfirmOnChainResolvedPaysWithPeerOsps(ctx context.Context, in *rpc.ConfirmOnChainResolvedPaysRequest) (*empty.Empty, error) {
	log.Infoln("Admin: ConfirmOnChainResolvedPaysWithPeeerOsps")
	activeOsps := s.cNode.GetActiveOsps()
	err := s.actToOspsOnToken(in.GetTokenAddress(), activeOsps, s.cNode.ConfirmOnChainResolvedPays)
	return &empty.Empty{}, err
}

func (s *adminService) ClearExpiredPaysWithPeerOsps(ctx context.Context, in *rpc.ClearExpiredPaysRequest) (*empty.Empty, error) {
	log.Infoln("Admin: ClearExpiredPaysWithPeerOsps")
	activeOsps := s.cNode.GetActiveOsps()
	err := s.actToOspsOnToken(in.GetTokenAddress(), activeOsps, s.cNode.SettleExpiredPays)
	return &empty.Empty{}, err
}

type actionToToken = func(ctype.Addr) error
type actionToCid = func(ctype.CidType) error

// actToOspsOnToken defines an execution flow that apply "action" to "osps" on "tokenAddr" if provided
// or loop over all tokens defined in rtconfig.
func (s *adminService) actToOspsOnToken(tokenAddr []byte, osps map[ctype.Addr]bool, action actionToCid) error {
	return s.actOnToken(tokenAddr, func(token ctype.Addr) error {
		tokenAddrStr := ctype.Addr2Hex(token)
		errs := make(map[string]error)
		for osp := range osps {
			cid, err := s.cNode.GetChannelIdForPeer(ctype.Addr2Hex(osp), tokenAddrStr)
			if err != nil {
				key := ctype.Addr2Hex(osp) + "@" + tokenAddrStr
				errs[key] = err
				continue
			}
			err = action(cid)
			if err != nil {
				key := ctype.Addr2Hex(osp) + "@" + tokenAddrStr
				errs[key] = err
			}
		}
		if len(errs) != 0 {
			errStr := fmt.Sprint(errs)
			return errors.New(errStr)
		}
		return nil
	})
}

func (s *adminService) getSupportedTokensFromRtConfig() map[ctype.Addr]bool {
	tks := make(map[ctype.Addr]bool)
	if rtconfig.GetStandardConfigs() != nil && rtconfig.GetStandardConfigs().GetConfig() != nil {
		for token := range rtconfig.GetStandardConfigs().GetConfig() {
			tks[ctype.Hex2Addr(token)] = true
		}
	}
	return tks
}

// actOnToken defines an execution flow that apply "action" on "tokenAddr" if provided
// or loop over all tokens defined in rtconfig.
func (s *adminService) actOnToken(tokenAddr []byte, action actionToToken) error {
	if action == nil {
		return nil
	}
	if tokenAddr != nil {
		err := action(ctype.Bytes2Addr(tokenAddr))
		if err != nil {
			log.Errorln(ctype.Bytes2Addr(tokenAddr))
		}
		return err
	}
	tks := s.getSupportedTokensFromRtConfig()
	if len(tks) == 0 {
		log.Errorln("no token found to act on")
		return status.Errorf(codes.NotFound, "no token found")
	}
	errs := make(map[string]error) // tokenAddr->error
	for token := range tks {
		log.Infof("Acting on token %x", token.Bytes())
		err := action(token)
		if err != nil {
			log.Errorln(ctype.Addr2Hex(token), err)
			errs[ctype.Addr2Hex(token)] = err
		}
	}
	if len(errs) != 0 {
		errStr := fmt.Sprint(errs)
		return errors.New(errStr)
	}
	return nil
}
func (s *adminService) BuildRoutingTable(ctx context.Context, in *rpc.BuildRoutingTableRequest) (*empty.Empty, error) {
	log.Infoln("Admin: building routing table")
	err := s.actOnToken(in.GetTokenAddress(), func(token ctype.Addr) error {
		table, err := s.cNode.BuildRoutingTable(token)
		if err == nil {
			log.Infoln("New routing table:", beautifyRT(table))
		}
		return err
	})
	return &empty.Empty{}, err
}

func (s *adminService) RegisterStream(ctx context.Context, in *rpc.RegisterStreamRequest) (*empty.Empty, error) {
	log.Infoln("Admin: register stream", in.PeerRpcAddress, ctype.Bytes2Hex(in.PeerEthAddress))
	s.ospEthToRPCLock.Lock()
	defer s.ospEthToRPCLock.Unlock()
	if peerRPC := s.ospEthToRPC[ctype.Bytes2Addr(in.PeerEthAddress)]; peerRPC == in.PeerRpcAddress {
		return &empty.Empty{}, status.Errorf(codes.AlreadyExists, "connection to %s already exist", in.PeerRpcAddress)
	}
	err := s.cNode.RegisterStream(ctype.Bytes2Hex(in.PeerEthAddress), in.PeerRpcAddress)
	if err != nil {
		log.Errorln("RegisterStream failed:", ctype.Bytes2Hex(in.PeerEthAddress), in.PeerRpcAddress, err)
		return &empty.Empty{}, status.Errorf(codes.Unknown, "RegisterStream failed: %s", err)
	}
	s.ospEthToRPC[ctype.Bytes2Addr(in.PeerEthAddress)] = in.PeerRpcAddress
	s.cNode.RegisterStreamErrCallback(ctype.Bytes2Hex(in.PeerEthAddress), s.streamRetryCb)
	return &empty.Empty{}, nil
}

func (s *adminService) GetActivePeerOsps(ctx context.Context, in *empty.Empty) (*rpc.ActivePeerOspsResponse, error) {
	activeOsps := s.cNode.GetActiveOsps()
	ospList := make([]string, 0, len(activeOsps))
	for osp := range activeOsps {
		ospList = append(ospList, ctype.Addr2Hex(osp))
	}
	return &rpc.ActivePeerOspsResponse{Osps: ospList}, nil
}

func (s *adminService) SendToken(ctx context.Context, in *rpc.SendTokenRequest) (*rpc.SendTokenResponse, error) {
	amt := utils.Wei2BigInt(in.AmtWei)
	if amt == nil {
		return &rpc.SendTokenResponse{Status: 1, Error: "Can't parse amount."}, status.Error(codes.InvalidArgument, "Can't parse amount")
	}
	dstAddr, err := hex.DecodeString(strings.TrimPrefix(in.DstAddr, "0x"))
	if err != nil {
		log.Errorln("Error parsing dst:", in.DstAddr)
		return &rpc.SendTokenResponse{Status: 1, Error: "Can't parse dst."}, status.Error(codes.InvalidArgument, "Can't parse dst")
	}
	tokenTransfer := &entity.TokenTransfer{
		Token: &entity.TokenInfo{
			TokenType: entity.TokenType_ETH,
		},
		Receiver: &entity.AccountAmtPair{
			Account: dstAddr,
			Amt:     amt.Bytes(),
		},
	}
	if in.TokenAddr != "" {
		tokenAddr, err2 := hex.DecodeString(strings.TrimPrefix(in.TokenAddr, "0x"))
		if err2 != nil {
			return &rpc.SendTokenResponse{Status: 1, Error: "Can't parse token address."}, status.Error(codes.InvalidArgument, "Can't parse token address")
		}
		tokenTransfer.Token.TokenAddress = tokenAddr
		tokenTransfer.Token.TokenType = entity.TokenType_ERC20
	}
	pay := &entity.ConditionalPay{
		Src:  ctype.Hex2Bytes(s.cNode.EthAddress),
		Dest: dstAddr,
		TransferFunc: &entity.TransferFunction{
			LogicType:   entity.TransferFunctionType_BOOLEAN_AND,
			MaxTransfer: tokenTransfer,
		},
		ResolveDeadline: s.cNode.GetCurrentBlockNumber().Uint64() + config.AdminSendTokenTimeout,
		ResolveTimeout:  config.PayResolveTimeout,
	}

	noteType, _ := ptypes.AnyMessageName(in.Note)
	if noteType == "" {
		noteType = "reward"
	}

	metrics.IncSvrAdminSendTokenCnt(metrics.SvrAdminSendAttempt, noteType)
	payID, err := s.cNode.AddBooleanPay(pay, in.Note)
	if err != nil {
		log.Errorln(
			err, "sending token from admin error to", ctype.Bytes2Hex(dstAddr),
			"tokenAmt:", utils.BytesToBigInt(tokenTransfer.Receiver.Amt),
			"tokenAddr:", ctype.Bytes2Hex(tokenTransfer.Token.TokenAddress))
		return &rpc.SendTokenResponse{Status: 1, Error: err.Error()}, status.Error(codes.Unavailable, err.Error())
	}

	metrics.IncSvrAdminSendTokenCnt(metrics.SvrAdminSendSucceed, noteType)
	return &rpc.SendTokenResponse{
		Status: 0,
		PayId:  payID.Hex(),
	}, nil
}
func postFeeEvent(endpoint string, event proto.Message, jsonMarshaler *jsonpb.Marshaler, netClient *http.Client) error {
	buf, err := jsonMarshaler.MarshalToString(event)
	if err != nil {
		log.Errorln("marshal event:", err)
		return err
	}
	resp, err := netClient.Post(endpoint, "application/json", strings.NewReader(buf))
	if err == nil {
		resp.Body.Close()
	}
	return err
}
func (s *server) HandleSendComplete(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	note *any.Any,
	reason rpc.PaymentSettleReason) {
	paid := false
	if reason == rpc.PaymentSettleReason_PAY_PAID_MAX || reason == rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN {
		paid = true
	}
	log.Infoln("payID", ctype.Bytes2Hex(payID.Bytes()), "Done. Note:", note, "paid", paid)
}
func (s *server) HandleDestinationUnreachable(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any) {
	log.Errorln(payID.String(), "unreachable")
}
func (s *server) HandleSendFail(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any, errMsg string) {
	log.Errorln(payID.String(), "failed", errMsg)
}
func (s *server) HandleNewCelerStream(addr []byte) {
}

func (s *server) Initialize(
	keyStore string, passPhrase string, transactorConfigs []*transactor.TransactorConfig, routingBytes []byte) {
	config := *common.ParseProfile(*pjson)
	overrideConfig(&config)
	var err error
	s.cNode, err =
		cnode.NewCNode(
			keyStore,
			passPhrase,
			transactorConfigs,
			config,
			common.ServiceProviderPolicy,
			routingBytes)
	if err != nil {
		log.Fatalln("Server init error:", err)
	}
	s.cNode.OnReceivingToken(s)
	s.cNode.OnSendToken(s)
	s.cNode.OnNewStream(s)
}
func (s *server) HandleReceivingStart(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any) {
}
func (s *server) HandleReceivingDone(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	note *any.Any,
	reason rpc.PaymentSettleReason) {
}

func (s *serverInterOSP) FwdMsg(ctx context.Context, in *rpc.FwdReq) (*rpc.FwdReply, error) {
	log.Debugln("FwdMsg to peer:", in.GetDest())
	reply := rpc.FwdReply{Accepted: false}

	// Reject if the destination client is not connected to this OSP.
	dest := in.GetDest()
	if !s.svr.cNode.IsLocalPeer(dest) {
		return &reply, nil
	}

	err := s.svr.cNode.ForwardMsgToPeer(in)
	if err != nil {
		log.Error(err)
		reply.Accepted = false
	} else {
		reply.Accepted = true
	}
	return &reply, nil
}

func (s *serverInterOSP) Ping(ctx context.Context, in *rpc.PingReq) (*rpc.PingReply, error) {
	log.Traceln("Ping:", in.String())
	reply := rpc.PingReply{}
	reply.Numclients = uint32(s.svr.cNode.NumClients())
	return &reply, nil
}

func (s *serverInterOSP) PickServer(ctx context.Context, in *rpc.PickReq) (*rpc.PickReply, error) {
	log.Debugln("PickServer:", in.String())
	myAddr := s.svr.cNode.GetRPCAddr()
	reply := rpc.PickReply{Server: myAddr}
	return &reply, nil
}

func overrideConfig(config *common.CProfile) {
	if *storedir != "" {
		config.StoreDir = *storedir
	}

	selfHostPort = *selfrpc
	if selfHostPort != "" {
		host, port2, err := getHostPort(selfHostPort)
		if err != nil {
			log.Fatalf("invalid self-RPC: %s", err)
		}
		if host == "MY_POD_IP" {
			selfHostPort = os.Getenv("MY_POD_IP") + ":" + strconv.Itoa(port2)
			log.Infoln("selfHostPort", selfHostPort)
		}
		config.SelfRPC = selfHostPort
	}
	if *isosp {
		config.IsOSP = true
	}
	config.ListenOnChain = *listenOnChain
	config.WebPort = *adminweb
}

func getHostPort(svrAddr string) (string, int, error) {
	hostport := strings.Split(svrAddr, ":")
	if len(hostport) != 2 {
		return "", 0, fmt.Errorf("address '%s' not in host:port format", svrAddr)
	}

	port, err := strconv.Atoi(hostport[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %s: %s", hostport[1], err)
	}

	return hostport[0], port, nil
}

func readPassword(ksBytes []byte) string {
	if *noPassword {
		return ""
	}
	ksAddress, err := utils.GetAddressFromKeystore(ksBytes)
	if err != nil {
		log.Fatal(err)
	}

	if *passwordDir != "" {
		passwordBytes, passwordErr := ioutil.ReadFile(path.Join(*passwordDir, ksAddress))
		if passwordErr != nil {
			log.Fatal(passwordErr)
		}
		return string(passwordBytes)
	}

	ksPasswordStr := ""
	if terminal.IsTerminal(syscall.Stdin) {
		fmt.Printf("Enter password for %s: ", ksAddress)
		ksPassword, err2 := terminal.ReadPassword(syscall.Stdin)
		if err2 != nil {
			log.Fatalln("Cannot read password from terminal:", err2)
		}
		ksPasswordStr = string(ksPassword)
	} else {
		reader := bufio.NewReader(os.Stdin)
		ksPwd, err2 := reader.ReadString('\n')
		if err2 != nil {
			log.Fatalln("Cannot read password from stdin:", err2)
		}
		ksPasswordStr = strings.TrimSuffix(ksPwd, "\n")
	}

	_, err = keystore.DecryptKey(ksBytes, ksPasswordStr)
	if err != nil {
		log.Fatal(err)
	}
	return ksPasswordStr
}

func main() {
	flag.Parse()
	if *showver {
		printver()
		os.Exit(0)
	}
	var ksBytes []byte
	var ksStr string
	var ksErr error
	if *ks != "" {
		ksBytes, ksErr = ioutil.ReadFile(*ks)
		if ksErr != nil {
			log.Fatalln(ksErr)
		}
		ksStr = string(ksBytes)
	}

	var routingBytes []byte
	var routingErr error
	if *routingData != "" {
		routingBytes, routingErr = ioutil.ReadFile(*routingData)
		if routingErr != nil {
			log.Fatalln(routingErr)
		}
	}

	log.Info("Starting Celer server....")
	if *isosp {
		rterr := rtconfig.Init(*rtcfile)
		if rterr != nil {
			log.Warnln("init runtime config failed:", rterr, "All runtime config values will be default.")
		}
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		os.Exit(2)
	}
	s := grpc.NewServer(getServerTlsOption(),
		grpc.KeepaliveEnforcementPolicy(config.KeepAliveEnforcePolicy),
		grpc.KeepaliveParams(config.KeepAliveServerParams))
	// enable reflection and line number printing for easy debugging via cli
	if *dbg {
		reflection.Register(s)
	}
	var tConfigs []*transactor.TransactorConfig
	tksPaths := *transactorks
	if tksPaths != "" {
		tConfigs = []*transactor.TransactorConfig{}
		tksArr := strings.Split(tksPaths, ",")
		for _, tks := range tksArr {
			tksBytes, err := ioutil.ReadFile(tks)
			if err != nil {
				log.Fatal(err)
			}
			tConfigs =
				append(
					tConfigs,
					transactor.NewTransactorConfig(string(tksBytes), readPassword(tksBytes)))
		}
	}
	var rpcServer server
	rpcServer.jsonMarshaler = &jsonpb.Marshaler{}
	rpcServer.netClient = &http.Client{Timeout: 3 * time.Second}
	rpcServer.Initialize(ksStr, readPassword(ksBytes), tConfigs, routingBytes)
	rpc.RegisterRpcServer(s, &rpcServer)

	adminS := setUpAdminService(&rpcServer)
	// If inter-server communication is needed, start in a goroutine the
	// second server on the second port.
	if selfHostPort != "" {
		_, port2, err := getHostPort(selfHostPort)
		if err != nil {
			log.Fatalf("invalid self-RPC: %s", err)
			os.Exit(2)
		}

		log.Info("Celer server has 2nd port (inter-server)....", port2)
		lis2, err := net.Listen("tcp", fmt.Sprintf(":%d", port2))
		if err != nil {
			log.Fatalf("failed to listen on 2nd port: %v", err)
			os.Exit(2)
		}
		s2 := grpc.NewServer()
		if *dbg {
			reflection.Register(s2)
		}
		interOSPServer := serverInterOSP{&rpcServer, adminS}
		rpc.RegisterMultiServerServer(s2, &interOSPServer)
		go s2.Serve(lis2)
		// Trigger a rt build as router processor post happened before this point will not succeed.
		if *listenOnChain {
			time.AfterFunc(time.Second, func() {
				utils.RequestBuildRoutingTable("localhost" + *adminweb)
			})
		}
	}

	// Run the main server.
	s.Serve(lis)
}

func setUpAdminService(osp *server) *adminService {
	_, port, err := getHostPort(*adminrpc)
	if err != nil {
		log.Fatalf("invalid admin rpc: %s", err)
		os.Exit(2)
	}

	log.Info("Celer server has 3rd port (admin rpc)....", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen on 3rd port: %v", err)
		os.Exit(2)
	}
	s := grpc.NewServer()
	if *dbg {
		reflection.Register(s)
	}
	adminS := newAdminService(osp.cNode)
	rpc.RegisterAdminServer(s, adminS)
	go s.Serve(lis)

	ctx := context.Background()

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = rpc.RegisterAdminHandlerFromEndpoint(ctx, gwmux, *adminrpc, opts)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	http.Handle("/admin/", gwmux)
	http.Handle("/metrics", metrics.GetPromExporter())
	log.Info("Celer server has 4th port (admin HTTP)....", *adminweb)
	go func() {
		err := http.ListenAndServe(*adminweb, http.DefaultServeMux)
		if err != nil {
			log.Errorln(err)
		}
	}()
	return adminS
}

func getServerTlsOption() grpc.ServerOption {
	if *tlsCert != "" && *tlsKey != "" {
		if *tlsClient {
			// require client cert
			cpool := x509.NewCertPool()
			cpool.AppendCertsFromPEM(utils.CelerCA)
			cert, _ := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
			return grpc.Creds(credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientCAs:    cpool,
				ClientAuth:   tls.RequireAndVerifyClientCert,
			}))
		}
		// accept any client
		creds, _ := credentials.NewServerTLSFromFile(*tlsCert, *tlsKey)
		return grpc.Creds(creds)
	}
	// no cert/key file specified, use localhost cert. ignore tlsclient b/c
	// server is only local
	localcert, _ := tls.X509KeyPair(certPem, privPem)
	return grpc.Creds(credentials.NewServerTLSFromCert(&localcert))
}

var (
	version string
	commit  string
)

func printver() {
	fmt.Println("Version:", version)
	fmt.Println("Commit:", commit)
}

// cat localhost.crt. for localhost and 127.0.0.1 only
var certPem = []byte(`-----BEGIN CERTIFICATE-----
MIIEQDCCAiigAwIBAgIRAJ20VoyznH/e/ogsuLsK74EwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEAxMHQ2VsZXJDQTAeFw0xOTA5MTcyMTIyMTNaFw0yMjA5MTcyMTIy
MTNaMBQxEjAQBgNVBAMTCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAJkyDGnACOXvQYCT6xULEfCkZOuRjT8+KuK/toUp48D2s7XTA9o+
G6PZxqJDPWpEUtFRdiCB3NCUeRp5sHKs6d/I99a8yV12fguVu+UnN0mlCC2RNNNQ
VpjylRHq/hkeyIQroBzCvxIzjxc3MJx5V8+AU/igbzcHb30Gn8ZeOkop09ZQMbj7
LO2s8x+anQXnEbKOm9RHedp3R456kszoD61tpTME/Wg5Vva8TE4NOhoBS1j34RlI
OZlwoYv8+yOyVuFea2GIKSix4F3zLHlkPYM1TDHddoOOQEmZUqqAP9PwRhTETn9r
BOUV+iFKzpcCcRNru2GNc66Mcl7QF1AoWu8CAwEAAaOBjjCBizAOBgNVHQ8BAf8E
BAMCA7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBRq
LO+4naYymcYyIwbJJi4EhKECrTAfBgNVHSMEGDAWgBSoQm3esw5att/hencMPoPd
qhyvXTAaBgNVHREEEzARgglsb2NhbGhvc3SHBH8AAAEwDQYJKoZIhvcNAQELBQAD
ggIBAGmIZ/oTDSGFgTn6efGyW+V+QbuD+iC04slfHJ1eRZUmDETyyKipfj+YOBCR
1pzzD0htSUwQ5b1k5spGdTDdkK4sRoWG839NWPYrc7Jx/qoh+W0NV9g6T0lcHhXj
FwpAIixlyE6UbkYh5GRoqHcOolaygmCXYEwQg1IdP3xvSzilcWwMiepc9lOmlK/+
gKCw0uaMjNHuCO7IMfAhvFTEQ69BHyLa5II1PQFNMLOjtVsk4Q36DaIKHzKetAh4
W6laPCOD2Nx/jnYxanwsy8XeQz2LIWvCvm+uxsVfhf3cNSAln3quPvSo/kCB2nGP
/LW/mPCIcm1TPTvbi6MONO8IvJItFKz8JedQQLIRKtcXMvM/PXoWBV/PEDG0jtRS
Q1n79XRLN8Ok7oeTJokaxXWhyr9/5Bu4W5KVI4A67x6S7dYItDTb6K7uj47kDoJL
tlQw2tBB7/qqj7d+BkGg7ljtAq3J9ANRPNO4gqYg+7D8UnYmyNTv6O5nBUwov5rB
R5BIOXOh1hsmSWznrDPmFaNOKZbOkWxvQmxgrrpgkqKaMZP2l1YXznb5r01srfPf
EaxvSjqf67woVThLJ27GiRxMVhBhM7Yt3kT7UrBEf99Igftyt8dADT7iZwpD5law
YAT8LXBz9+i+SecjrmRCBzxoDirS2MZz4UR0Wf+XDWgvWNxW
-----END CERTIFICATE-----`)

// cat localhost.key
var privPem = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAmTIMacAI5e9BgJPrFQsR8KRk65GNPz4q4r+2hSnjwPaztdMD
2j4bo9nGokM9akRS0VF2IIHc0JR5Gnmwcqzp38j31rzJXXZ+C5W75Sc3SaUILZE0
01BWmPKVEer+GR7IhCugHMK/EjOPFzcwnHlXz4BT+KBvNwdvfQafxl46SinT1lAx
uPss7azzH5qdBecRso6b1Ed52ndHjnqSzOgPrW2lMwT9aDlW9rxMTg06GgFLWPfh
GUg5mXChi/z7I7JW4V5rYYgpKLHgXfMseWQ9gzVMMd12g45ASZlSqoA/0/BGFMRO
f2sE5RX6IUrOlwJxE2u7YY1zroxyXtAXUCha7wIDAQABAoIBAQCJSnYfa69NyZ6t
SWL7h+E7BUlAaD/qdp9eeKttKb5n12/0ujiQpOqGbAv8rT/j9Xk3B8dSmK846mah
2H7ONrKeEHA0LRpVPXT2kulCE2QUBueOVry9yBjjlzsLRMsV3iWbdbFXNRyhhj1t
c9OH16NfXcVjYvxol6xNotsbnqSkgva64mlOeOsMcXgmBfqWPFiccwMxN8cciuZJ
FD69MlbGgpFR2UO0FjO1TZwQLiYuVzv9llHy5SJpcmPgrkALdxVfG1plVPxAkcer
sSs9Zv3KMMyRsGKr4ZQN6A0s1wnNvPg06VBGvFKmQK1PQw+ffop2qVqVian/ZeMn
IyedQcIRAoGBAMrwIPoqoiYZcxaJHaAUnjXoIYNLcnO+9RyCWXurdv20Cq6g+02v
kIASQxqSJhrqe+7ZRjKUY3z3ElTbj3pIN+43a+0T5ahX83bCsJ1rWoEfaU54VSrj
N62uEfc39LkfjNv5R4IglKh8IXFyXSJiYf4UXVcU17SIh/owA8oL68dnAoGBAMFA
WjgdTb57SedLh+4b4NEd/RQfWxrhMXnkY3ulFl/Uf1E60mgolTVtIW84BlrQomkh
kQCMj2M9LW2mP+Y9/AHMKg9zGb8FPWyabH+j1BD4QrXVnaJTACKrYGT1Z9R16nya
FXqLUdiKA7wRgfgxX9uRzlXMd4qNPIzxdyACH0M5AoGBAK1k1who/PqIrCkJJuLs
OvHcUSYZhMUY191wEnz0WEsVVjs3GQGbjF+hOuytCxncV+AQjUYSO58+i88tej4F
DqTffbunUIax/zftyXH3k/DXoeaGMl7enWgsXvVYPiUerAAX0d2BcQM0bG6+RI1o
eknZpJcPG+8I6QX/mH0+CkrpAoGAeqQTXV9DcmodqZqmhjbNAwksDjQkBjf5xShq
9hH71A8wSWWyGAYBQymhuUptxf53w45YzmdlrA4sIVULYlvd7WobGzjpku+JXr3V
s19N+wMCmxEY++X+xQHLp+aR4SSADlle3ilCZNCZtCXMPK1g7yBmOM8M4jHlxnCL
MBYIrwkCgYEAugg+eR8E0Alomb1X/XzpC+iXC31EoOXalhFE1PtY3F778kU4omFJ
Ke4lExJlF3X8rSAcWOhWFqyCMaIZgIE7PGpiPzZkSdA6vw4aGfoCOte3zMGSHPeZ
BH0tl2Y4N4mvn7EgpyTklckBOly71tIS3E9dGEBbFL7ruCPQUAAQivE=
-----END RSA PRIVATE KEY-----`)
