// Copyright 2018-2020 Celer Network

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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/celer-network/goCeler/cnode"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/delegate"
	"github.com/celer-network/goCeler/entity"
	celerx_fee_interface "github.com/celer-network/goCeler/fee-manager/interface"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	pjson                = flag.String("profile", "config/profile.json", "Path to profile json file")
	dbg                  = flag.Bool("debug", false, "enable reflection and verbos log for debug")
	port                 = flag.Int("port", 10000, "The server listening port")
	selfrpc              = flag.String("selfrpc", "", "Internal server host:port for inter-server communication")
	adminrpc             = flag.String("adminrpc", "localhost:11000", "The server admin endpoint")
	adminweb             = flag.String("adminweb", "localhost:8090", "The server admin http endpoint")
	listenerweb          = flag.String("listenerweb", "", "The event listener admin http endpoint")
	ks                   = flag.String("ks", "", "Path to keystore json file")
	depositks            = flag.String("depositks", "", "Path to depositor keystore json file")
	noPassword           = flag.Bool("nopassword", false, "Assume empty password for keystores")
	transactorks         = flag.String("transactorks", "", "Paths to keystore json files for on-chain transactions, separated by comma")
	passwordDir          = flag.String("passworddir", "", "Path to the directory containing passwords")
	storedir             = flag.String("storedir", "", "Path to the store directory")
	storesql             = flag.String("storesql", "", "sql database URL")
	showver              = flag.Bool("v", false, "Show version and exit")
	isosp                = flag.Bool("isosp", true, "Run as an OSP node")
	listenOnChain        = flag.Bool("loc", true, "Listen to on-chain log events")
	svrname              = flag.String("svrname", "", "unique server name")
	rtcfile              = flag.String("rtc", "rt_config.json", "runtime config json file path")
	receiveDoneNotifyee  = flag.String("fmrecvdone", "localhost:8092/notify/osp/feereceived", "end point to notify for a pay received with note")
	payDoneNotifyee      = flag.String("fmsenddone", "localhost:8092/notify/osp/sendcomplete", "end point to notify for a pay send complete")
	redisAddr            = flag.String("redisaddr", "", "Redis address to publish pay event")
	pubRetryInterval     = flag.Int64("pubretryintervalsec", 10, "retry interval in seconds for pay event publish")
	ospMaxRedial         = flag.Int("ospmaxredial", 10, "max retry for osp to dial to another osp")
	ospRedialDelay       = flag.Int64("ospredialdelay", 5, "retry delay (in sec) for osp to dial to another osp")
	routingData          = flag.String("routedata", "", "Path to routing data json file")
	tlsCert              = flag.String("tlscert", "", "Path to TLS cert file")
	tlsKey               = flag.String("tlskey", "", "Path to TLS private key file")
	tlsClient            = flag.Bool("tlsclient", false, "Require tls client cert by CelerCA")
	allowTsDiffInMinutes = flag.Uint64("allowtsdiff", 120, "Allowed timestamp diff (in minutes) when authenticating peer in pay history request")

	routerBcastInterval = flag.Uint64("routerbcastinterval", 0, "interval (in sec) to broadcast route updates, should only set for test purpose")
	routerBuildInterval = flag.Uint64("routerbuildinterval", 0, "interval (in sec) to build routing table, should only set for test purpose")
	routerAliveTimeout  = flag.Uint64("routeralivetimeout", 0, "timeout (in sec) for router aliveness, should only set for test purpose")
	ospClearPayInterval = flag.Uint64("ospclearpayinterval", 0, "interval (in sec) for osp to clear expired and on-chain resolved pays with its peers, should only set for test purpose")
	ospReportInterval   = flag.Uint64("ospreportinterval", 0, "interval (in sec) to report to OSP explorer, should only set for test purpose")
)

var selfHostPort string

const maxItemsPerPage int32 = 1000
const defaultItemsPerPage int32 = 50

type server struct {
	cNode        *cnode.CNode
	netClient    *http.Client
	lastOcTs     int64      // timestamp in seconds of last processed open channel request
	lastOcTsLock sync.Mutex // lock to protect r/w of lastOcTs
	delegate     *delegate.DelegateManager
	config       *common.CProfile
	redisClient  *redis.Client
	rpc.UnimplementedRpcServer
}
type serverInterOSP struct {
	svr      *server
	adminSvr *adminService
	rpc.UnimplementedMultiServerServer
}
type adminService struct {
	// Note: adminService doesn't own cNode.
	cNode                        *cnode.CNode
	ospEthToRPC                  map[ctype.Addr]string
	ospEthToRPCLock              sync.Mutex
	streamRetryCb                rpc.ErrCallbackFunc
	rpc.UnimplementedAdminServer // so new rpc won't break build due to missing interface func
}

func newAdminService(cNode *cnode.CNode) *adminService {
	adminS := &adminService{
		cNode:       cNode,
		ospEthToRPC: make(map[ctype.Addr]string),
	}
	adminS.streamRetryCb = func(addr ctype.Addr, streamErr error) {
		// Register the callback to handle stream errors and try to reconnect.
		// TODO: note that such a two-step API has a tiny race-condition window
		// in case the stream quickly disconnects after a successful call to
		// RegisterStream() and before RegisterStreamErrCallback() is done.
		// The next design should either allow them both to be done atomically
		// or allow RegisterStreamErrCallback() before RegisterStream().
		log.Infoln("streamRetryCb triggered for", addr.Hex(), streamErr)
		delay := time.Duration(*ospRedialDelay) * time.Second
		adminS.ospEthToRPCLock.Lock()
		ospRPC, ok := adminS.ospEthToRPC[addr]
		delete(adminS.ospEthToRPC, addr)
		adminS.ospEthToRPCLock.Unlock()

		if !ok {
			log.Errorln("Not finding RPC to redial peer", addr)
			return
		}

		for i := 0; i < *ospMaxRedial; i++ {
			log.Debugln("streamRetryCb: try to register again", addr.Hex())
			adminS.ospEthToRPCLock.Lock()
			if peerRPC := adminS.ospEthToRPC[addr]; peerRPC == ospRPC {
				return
			}
			err := adminS.cNode.RegisterStream(addr, ospRPC)
			if err == nil {
				log.Infoln("streamRetry:Cb successful re-register", addr.Hex())
				adminS.ospEthToRPC[addr] = ospRPC
				adminS.ospEthToRPCLock.Unlock()
				return
			}
			adminS.ospEthToRPCLock.Unlock()
			log.Errorln("streamRetryCb: register failed", addr.Hex(), err)
			time.Sleep(delay)
		}
		log.Errorln("streamRetry:Cb retry dial", addr.Hex(), "for", *ospMaxRedial, "times, giving up")
	}
	return adminS
}

func (s *server) RequestDelegation(ctx context.Context, in *rpc.DelegationRequest) (*rpc.DelegationResponse, error) {
	log.Infof("RequestDelegation: %x", in.GetProof().GetSigner())
	dal := s.cNode.GetDAL()
	proof := in.GetProof()
	delegationDesc := &rpc.DelegationDescription{}
	signer := utils.RecoverSigner(proof.GetDelegationDescriptionBytes(), proof.GetSignature())
	err := proto.Unmarshal(proof.GetDelegationDescriptionBytes(), delegationDesc)
	if err != nil {
		return nil, err
	}

	if signer != ctype.Bytes2Addr(delegationDesc.GetDelegatee()) {
		return nil, errors.New("Not signed by delegatee")
	}
	if s.cNode.EthAddress != ctype.Bytes2Addr(delegationDesc.GetDelegator()) {
		return nil, errors.New("Not delegating to " + s.cNode.EthAddress.Hex())
	}
	err = dal.UpdatePeerDelegateProof(signer, proof)
	if err != nil {
		return nil, err
	}
	return &rpc.DelegationResponse{}, nil
}

func (s *server) QueryDelegation(ctx context.Context, in *rpc.QueryDelegationRequest) (*rpc.QueryDelegationResponse, error) {
	dal := s.cNode.GetDAL()
	proof, found, err := dal.GetPeerDelegateProof(ctype.Bytes2Addr(in.GetDelegatee()))
	if err != nil {
		return nil, err
	}
	if !found || proof == nil {
		return nil, common.ErrDelegateProofNotFound
	}
	return &rpc.QueryDelegationResponse{Proof: proof}, nil
}

func (s *server) CelerStream(stream rpc.Rpc_CelerStreamServer) error {
	var ctx context.Context
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if msg.GetAuthReq() != nil {
		req := msg.GetAuthReq()
		reqjs, _ := utils.PbToJSONHexBytes(req)
		log.Infoln("Recv AuthReq:", reqjs)
		ackMsg, err2 := s.cNode.HandleAuthReq(req)
		if err2 != nil {
			log.Warnln("AuthReq err:", err2)
			return status.Error(codes.InvalidArgument, err2.Error())
		}
		if req.GetProtocolVersion() >= 1 {
			// send AuthAck back to requester if protocol version >= 1
			// r0.15 code doesn't expect AuthAck msg
			err = stream.Send(ackMsg)
			if err != nil {
				log.Warnln("Send AuthAck err:", err)
				return status.Error(codes.Canceled, err.Error())
			}
		}

		ctx, err = s.cNode.AddCelerStream(msg, stream)
		if err != nil {
			log.Warnln("AddCelerStream err:", err.Error())
			return status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		log.Warnln("first message not AuthReq:", msg.GetMessage())
		return status.Error(codes.Unauthenticated, "must send AuthReq first")
	}
	<-ctx.Done()
	return nil
}

func (s *server) GetPayHistory(ctx context.Context, in *rpc.GetPayHistoryRequest) (*rpc.GetPayHistoryResponse, error) {
	log.Debugf("GetPayHistory peer: %s, item per page: %d, before ts: %d, smallest payid: %s", in.GetPeer(), in.GetItemsPerPage(), in.GetBeforeTs(), in.GetSmallestPayId())
	// Read input
	peerStr := in.GetPeer()
	peer := ctype.Hex2Addr(peerStr)
	beforeTs := in.GetBeforeTs()
	itemsPerPage := in.GetItemsPerPage()
	smallestPayIDStr := in.GetSmallestPayId()
	smallestPayID := ctype.Hex2PayID(smallestPayIDStr)
	// set default values
	beforeTsTime := time.Unix(beforeTs, 0).UTC()
	if beforeTs == 0 {
		beforeTsTime = time.Now().UTC()
	}
	if itemsPerPage == 0 {
		itemsPerPage = defaultItemsPerPage
	}
	if itemsPerPage > maxItemsPerPage {
		itemsPerPage = maxItemsPerPage
	}
	// verify identity
	tsSig := in.GetTsSig()
	tsFromPeer := in.GetTs()
	tsFromServer := uint64(time.Now().Unix())
	if utils.RecoverSigner(utils.Uint64ToBytes(tsFromPeer), tsSig) != peer {
		// sig is invalid
		return nil, errors.New("Invalid Signature")
	}
	// require ts from peer be within a window.
	if tsFromPeer > tsFromServer+*allowTsDiffInMinutes*60 || tsFromPeer < tsFromServer-*allowTsDiffInMinutes*60 {
		return nil, errors.New("Invalid Timestamp")
	}

	// Query history with parameters
	payIDs, pays, instates, createTses, err := s.cNode.GetDAL().GetPayHistory(peer, beforeTsTime, smallestPayID, itemsPerPage)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	log.Debugln("History result:", payIDs, pays, instates, createTses, beforeTsTime)

	// Generate response
	resp := &rpc.GetPayHistoryResponse{
		Pays: make([]*rpc.OneHistoricalPay, 0, len(payIDs)),
	}
	for i := 0; i < len(payIDs); i++ {
		onePay := &rpc.OneHistoricalPay{}
		amt := big.NewInt(0).SetBytes(pays[i].GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
		src := pays[i].GetSrc()
		dst := pays[i].GetDest()
		token := pays[i].GetTransferFunc().GetMaxTransfer().GetToken().GetTokenAddress()
		onePay.Amt = amt.String()
		onePay.Token = ctype.Addr2Hex(ctype.Bytes2Addr(token))
		onePay.Src = ctype.Addr2Hex(ctype.Bytes2Addr(src))
		onePay.Dst = ctype.Addr2Hex(ctype.Bytes2Addr(dst))
		onePay.State = instates[i]
		onePay.PayId = ctype.PayID2Hex(payIDs[i])
		onePay.CreateTs = createTses[i]
		resp.Pays = append(resp.Pays, onePay)
	}
	return resp, nil
}

func (s *server) CelerGetPeerStatus(ctx context.Context, in *rpc.PeerAddress) (*rpc.PeerStatus, error) {
	peer, err := utils.ValidateAndFormatAddress(in.Address)
	if err != nil {
		return nil, err
	}
	tokenAddr, err := utils.ValidateAndFormatAddress(in.TokenAddr)
	if err != nil {
		return nil, err
	}

	if s.cNode == nil {
		return nil, fmt.Errorf("server error: cNode not initialized")
	}

	resp := new(rpc.PeerStatus)
	resp.FreeBalance = "0"
	// look up local database first to check if we could find the channel id locally
	if cid, err := s.cNode.GetChannelIdForPeer(peer, tokenAddr); err == nil {
		b, err2 := s.cNode.GetBalance(cid)
		if err2 == nil {
			resp.FreeBalance = b.MyFree.String() // myfree is peer's receiving capacity
		} else {
			log.Warn("getbalance err. cid:", cid.Hex(), err2)
		}

		resp.JoinStatus = rpc.JoinCelerStatus_LOCAL
		log.Infof("CelerGetPeerStatus server request: peer %x token %x free %s", peer, tokenAddr, resp.FreeBalance)
	} else { // if could not find locally, further check to see if it's a remote endpoint
		joinStatus := s.cNode.GetJoinStatusForNode(peer, tokenAddr)
		if joinStatus != rpc.JoinCelerStatus_LOCAL {
			if joinStatus == rpc.JoinCelerStatus_NOT_JOIN {
				log.Infof("peer %x not join Celer for token %x, error: %s", peer, tokenAddr, err)
				resp.FreeBalance = ""
			}
		}
		resp.JoinStatus = joinStatus
	}

	return resp, nil
}
func (s *adminService) HandleOpenChannelFinish(cid ctype.CidType) {
	log.Infoln("open channel succeeded, cid:", cid.Hex())
}

func (s *adminService) HandleOpenChannelErr(e *common.E) {
	log.Errorln("open channel error, reason: ", e.Reason)
}

func (s *adminService) OspOpenChannel(ctx context.Context, in *rpc.OspOpenChannelRequest) (*empty.Empty, error) {
	log.Infof("OspOpenChannel: peer: %x, self deposit: %s, peer deposit: %s",
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
	existingCid, _ := s.cNode.GetChannelIdForPeer(ctype.Bytes2Addr(in.GetPeerEthAddress()), ctype.Bytes2Addr(in.GetTokenAddress()))
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

func (s *server) CelerOpenTcbChannel(ctx context.Context, in *rpc.OpenChannelRequest) (*rpc.OpenChannelResponse, error) {
	return s.cNode.ProcessTcbRequest(in)
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

func (s *server) CelerMigrateChannel(ctx context.Context, in *rpc.MigrateChannelRequest) (*rpc.MigrateChannelResponse, error) {
	if in == nil {
		return nil, common.ErrInvalidArg
	}

	return s.cNode.ProcessMigrateChannelRequest(in)
}

func beautifyRT(table map[ctype.Addr]ctype.CidType) map[string]string {
	ret := make(map[string]string)
	for k, v := range table {
		ret[ctype.Addr2Hex(k)] = ctype.Cid2Hex(v)
	}
	return ret
}

func (s *adminService) ConfirmOnChainResolvedPaysWithPeerOsps(ctx context.Context, in *rpc.ConfirmOnChainResolvedPaysRequest) (*empty.Empty, error) {
	log.Infoln("Admin: ConfirmOnChainResolvedPaysWithPeerOsps")
	connectedOsps := s.cNode.GetConnectedOsps()
	err := s.actToOspsOnToken(in.GetTokenAddress(), connectedOsps, s.cNode.ConfirmOnChainResolvedPays)
	return &empty.Empty{}, err
}

func (s *adminService) ClearExpiredPaysWithPeerOsps(ctx context.Context, in *rpc.ClearExpiredPaysRequest) (*empty.Empty, error) {
	log.Infoln("Admin: ClearExpiredPaysWithPeerOsps")
	connectedOsps := s.cNode.GetConnectedOsps()
	err := s.actToOspsOnToken(in.GetTokenAddress(), connectedOsps, s.cNode.SettleExpiredPays)
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
			cid, err := s.cNode.GetChannelIdForPeer(osp, token)
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
	if rtconfig.GetTcbConfigs() != nil && rtconfig.GetTcbConfigs().GetConfig() != nil {
		for token := range rtconfig.GetTcbConfigs().GetConfig() {
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

func (s *adminService) RecvBcastRoutingInfo(ctx context.Context, in *rpc.RoutingRequest) (*empty.Empty, error) {
	log.Debugln("Admin: recv bcast routing info from", in.GetSender())
	err := s.cNode.RecvBcastRoutingInfo(in)
	if err != nil {
		log.Errorln("Admin: recv bcast routing info err:", err)
	}
	return &empty.Empty{}, err
}

func (s *adminService) RegisterStream(ctx context.Context, in *rpc.RegisterStreamRequest) (*empty.Empty, error) {
	log.Infoln("Admin: register stream", in.PeerRpcAddress, ctype.Bytes2Hex(in.PeerEthAddress))
	s.ospEthToRPCLock.Lock()
	defer s.ospEthToRPCLock.Unlock()
	if peerRPC := s.ospEthToRPC[ctype.Bytes2Addr(in.PeerEthAddress)]; peerRPC == in.PeerRpcAddress {
		return &empty.Empty{}, status.Errorf(codes.AlreadyExists, "connection to %s already exist", in.PeerRpcAddress)
	}
	err := s.cNode.RegisterStream(ctype.Bytes2Addr(in.PeerEthAddress), in.PeerRpcAddress)
	if err != nil {
		log.Errorln("RegisterStream failed:", ctype.Bytes2Hex(in.PeerEthAddress), in.PeerRpcAddress, err)
		return &empty.Empty{}, status.Errorf(codes.Unknown, "RegisterStream failed: %s", err)
	}
	s.ospEthToRPC[ctype.Bytes2Addr(in.PeerEthAddress)] = in.PeerRpcAddress
	s.cNode.RegisterStreamErrCallback(ctype.Bytes2Addr(in.PeerEthAddress), s.streamRetryCb)
	return &empty.Empty{}, nil
}

func (s *adminService) GetPeerOsps(ctx context.Context, in *empty.Empty) (*rpc.PeerOspsResponse, error) {
	peerOsps := s.cNode.GetPeerOsps()
	resp := &rpc.PeerOspsResponse{}
	for addr, osp := range peerOsps {
		peerOsp := &rpc.PeerOsp{
			OspAddress: ctype.Addr2Hex(addr),
			UpdateTs:   uint64(osp.UpdateTime.Unix()),
		}
		for tk, cid := range osp.TokenCids {
			tkcid := &rpc.TokenCidPair{
				TokenAddress: ctype.Addr2Hex(tk),
				Cid:          ctype.Cid2Hex(cid),
			}
			peerOsp.TokenCidPairs = append(peerOsp.TokenCidPairs, tkcid)
		}
		resp.PeerOsps = append(resp.PeerOsps, peerOsp)
	}
	return resp, nil
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
		tokenAddr, err2 := utils.ValidateAndFormatAddress(in.TokenAddr)
		if err2 != nil {
			return &rpc.SendTokenResponse{Status: 1, Error: "Can't parse token address."}, status.Error(codes.InvalidArgument, "Can't parse token address")
		}
		if tokenAddr != ctype.EthTokenAddr {
			tokenTransfer.Token.TokenAddress = tokenAddr.Bytes()
			tokenTransfer.Token.TokenType = entity.TokenType_ERC20
		}
	}
	pay := &entity.ConditionalPay{
		Src:  s.cNode.EthAddress.Bytes(),
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
		noteType = "reward" // TODO: upgrade reward svr to also use note
	}

	metrics.IncSvrAdminSendTokenCnt(metrics.SvrAdminSendAttempt, noteType)
	payID, err := s.cNode.AddBooleanPay(pay, in.Note)
	if err != nil {
		log.Errorln(
			err, "sending token from admin error to", ctype.Bytes2Hex(dstAddr),
			"tokenAmt:", utils.BytesToBigInt(tokenTransfer.Receiver.Amt),
			"tokenAddr:", ctype.Bytes2Hex(tokenTransfer.Token.TokenAddress))

		// filter too many retries error and no celer stream error
		if strings.Contains(err.Error(), "too many retries") || strings.Contains(err.Error(), common.ErrNoCelerStream.Error()) {
			metrics.IncSvrAdminSendTokenCnt(metrics.SvrAdminSendSucceed, noteType)
		}
		return &rpc.SendTokenResponse{Status: 1, Error: err.Error()}, status.Error(codes.Unavailable, err.Error())
	}

	metrics.IncSvrAdminSendTokenCnt(metrics.SvrAdminSendSucceed, noteType)
	return &rpc.SendTokenResponse{
		Status: 0,
		PayId:  ctype.PayID2Hex(payID),
	}, nil
}

func (s *adminService) Deposit(ctx context.Context, in *rpc.DepositRequest) (*rpc.DepositResponse, error) {
	peerAddr := ctype.Hex2Addr(in.GetPeerAddr())
	tokenAddr := ctype.Hex2Addr(in.GetTokenAddr())
	amount := utils.Wei2BigInt(in.GetAmtWei())
	if amount == nil {
		return &rpc.DepositResponse{Status: 1, Error: "Can't parse amount."}, status.Error(codes.InvalidArgument, "Can't parse amount")
	}
	depositID, err := s.cNode.RequestDeposit(
		peerAddr, tokenAddr, in.GetToPeer(), amount, time.Duration(in.GetMaxWaitS())*time.Second)
	if err != nil {
		return &rpc.DepositResponse{Status: 1, Error: err.Error()}, status.Error(codes.Unavailable, err.Error())
	}
	return &rpc.DepositResponse{Status: 0, DepositId: depositID}, nil
}

func (s *adminService) QueryDeposit(ctx context.Context, in *rpc.QueryDepositRequest) (*rpc.QueryDepositResponse, error) {
	state, errMsg, err := s.cNode.QueryDeposit(in.GetDepositId())
	if err != nil {
		errCode := codes.Unavailable
		if errors.Is(err, common.ErrDepositNotFound) {
			errCode = codes.NotFound
		}
		return &rpc.QueryDepositResponse{
			DepositState: rpc.DepositState_Deposit_NOT_FOUND, Error: err.Error(),
		}, status.Error(errCode, err.Error())
	}
	var depositState rpc.DepositState
	switch state {
	case structs.DepositState_NULL:
		depositState = rpc.DepositState_Deposit_NOT_FOUND
	case structs.DepositState_QUEUED, structs.DepositState_APPROVING_ERC20, structs.DepositState_TX_SUBMITTING:
		depositState = rpc.DepositState_Deposit_QUEUED
	case structs.DepositState_TX_SUBMITTED:
		depositState = rpc.DepositState_Deposit_SUBMITTED
	case structs.DepositState_SUCCEEDED:
		depositState = rpc.DepositState_Deposit_SUCCEEDED
	case structs.DepositState_FAILED:
		depositState = rpc.DepositState_Deposit_FAILED
	}
	return &rpc.QueryDepositResponse{
		DepositState: depositState,
		Error:        errMsg,
	}, nil
}

func postFeeEvent(endpoint string, event proto.Message, netClient *http.Client) error {
	buf, err := utils.PbToJSONString(event)
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

func (s *server) publishPayEvent(category string, payEvent *celerx_fee_interface.FeeEvent, noteTypeUrl string) {
	if s.redisClient != nil {
		payID := payEvent.GetPayId()
		eventJSON, toJSONErr := utils.PbToJSONString(payEvent)
		subkey := "emptynote"
		if noteTypeUrl != "" {
			subkey = noteTypeUrl
		}
		pubTopic := fmt.Sprintf("%s:%s", category, subkey)
		if toJSONErr != nil {
			log.Errorf("Publishing PayEvent Err: %v, topic %s payid %x", toJSONErr, pubTopic, payID)
		} else {
			var numNotified int64
			var pubErr error
			// retry 3 times at most.
			numRetry := 3
			for i := 0; i < numRetry; i++ {
				log.Debugf("Publishing PayEvent: topic %s, payid %x", pubTopic, payID)
				numNotified, pubErr = s.redisClient.Publish(pubTopic, eventJSON).Result()
				if pubErr != nil {
					log.Errorf("Publishing PayEvent Err: %v, topic %s payid %x", pubErr, pubTopic, payID)
				} else if numNotified != 0 {
					break
				}
				if i != numRetry-1 {
					// skip sleep in last retry ending up with failure, which is wasteful.
					time.Sleep(time.Duration(*pubRetryInterval) * time.Second)
				}
			}
			if pubErr == nil && numNotified == 0 {
				log.Warnf("Publishing PayEvent Err: pub topic %s payid %x multiple times but no subscriber", pubTopic, payID)
			}
		}
	}
}

func (s *server) handlePaySendFinalize(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	note *any.Any,
	reason rpc.PaymentSettleReason) {
	paid := false
	if reason == rpc.PaymentSettleReason_PAY_PAID_MAX || reason == rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN {
		paid = true
	}
	log.Infoln("payID", ctype.Bytes2Hex(payID.Bytes()), "Done. Note:", note, "paid", paid)
	if note != nil {
		event := &celerx_fee_interface.FeeEvent{
			Pay:          pay,
			SendSuccess:  paid,
			Note:         note,
			NotePbString: note.String(),
			PayId:        payID.Bytes(),
		}
		// No need to notify delegate for send finalization. Delegate is built inside osp, use function call below instead.
		if !ptypes.Is(note, &delegate.PayOriginNote{}) {
			go s.publishPayEvent("paysendfinalized", event, note.GetTypeUrl())
		} else {
			delegateEvent := &delegate.DelegateEvent{
				PayID:       payID,
				Pay:         pay,
				Note:        note,
				SendSuccess: paid,
			}
			log.Debugln("Notifying delegate for paysend finalization", payID.Hex())
			err := s.delegate.NotifyPaySendFinalize(delegateEvent)
			if err != nil {
				log.Errorln("post fee (recv agent):", err)
				return
			}
		}
	}
}

func (s *server) HandleSendComplete(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	note *any.Any,
	reason rpc.PaymentSettleReason) {
	s.handlePaySendFinalize(payID, pay, note, reason)
}

func (s *server) HandleDestinationUnreachable(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any) {
	log.Errorln(payID.String(), "unreachable")
	s.handlePaySendFinalize(payID, pay, note, rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE)
}

func (s *server) HandleSendFail(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any, errMsg string) {
	log.Errorln(payID.String(), "failed", errMsg)
	s.handlePaySendFinalize(payID, pay, note, rpc.PaymentSettleReason_PAY_REJECTED)
}

func (s *server) HandleNewCelerStream(addr ctype.Addr) {
	log.Debugln("Notifying delegate for new stream", addr.Hex())
	err := s.delegate.NotifyNewStream(addr)
	if err != nil {
		log.Warnln("post newstream (recv agent) to delegate:", err, "addr:", addr.Hex())
		return
	}
}

func (s *server) Initialize(
	masterTxConfig, depositTxConfig *transactor.TransactorConfig,
	transactorConfigs []*transactor.TransactorConfig, routingBytes []byte) {
	s.config = common.ParseProfile(*pjson)
	overrideConfig(s.config)
	var err error
	s.cNode, err = cnode.NewCNode(
		masterTxConfig,
		depositTxConfig,
		transactorConfigs,
		*s.config,
		route.ServiceProviderPolicy,
		routingBytes)
	if err != nil {
		log.Fatalln("Server init error:", err)
	}
	s.delegate = delegate.NewDelegateManager(s.cNode.EthAddress, s.cNode.GetDAL(), s.cNode)
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
	event := &celerx_fee_interface.FeeEvent{
		PayId:        payID.Bytes(),
		Pay:          pay,
		Note:         note,
		NotePbString: note.String(),
	}
	if note != nil {
		go s.publishPayEvent("receivedone", event, note.GetTypeUrl())
	}
}

func (s *serverInterOSP) FwdMsg(ctx context.Context, in *rpc.FwdReq) (*rpc.FwdReply, error) {
	log.Debugln("FwdMsg to peer:", in.GetDest())
	reply := rpc.FwdReply{Accepted: false}

	// Reject if the destination client is not connected to this OSP.
	dest := in.GetDest()
	if !s.svr.cNode.IsLocalPeer(ctype.Hex2Addr(dest)) {
		return &reply, nil
	}

	err := s.svr.cNode.ForwardMsgToPeer(in)
	if err != nil {
		reply.Accepted = false
		if err != common.ErrNoCelerStream {
			reply.Err = err.Error()
		}
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

func (s *serverInterOSP) BcastRoutingInfo(ctx context.Context, in *rpc.BcastRoutingRequest) (*rpc.BcastRoutingReply, error) {
	log.Debugln("BcastRoutingInfo:", in.String())

	s.svr.cNode.BcastRoutingInfo(in.GetReq(), in.GetOsps())

	reply := rpc.BcastRoutingReply{}
	return &reply, nil
}

func overrideConfig(config *common.CProfile) {
	count := 0
	stores := []string{*storedir, *storesql}
	for _, st := range stores {
		if st != "" {
			count++
		}
	}
	if count > 1 {
		log.Fatalln("specify only one of -storedir, -storesql")
		os.Exit(1)
	}

	if count == 1 {
		config.StoreDir = ""
		config.StoreSql = ""
		if *storedir != "" {
			config.StoreDir = *storedir
		} else if *storesql != "" {
			config.StoreSql = *storesql
		}
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
	config.SvrName = *svrname
	if config.SvrName == "" {
		config.SvrName = fmt.Sprintf("svr:%s", uuid.New().String())
	}
	if *isosp {
		config.IsOSP = true
	}
	config.ListenOnChain = *listenOnChain
}

func getHostPort(svrAddr string) (string, int, error) {
	hostport := strings.Split(svrAddr, ":")
	if len(hostport) != 2 {
		return "", 0, fmt.Errorf("address '%s' not in host:port format", svrAddr)
	}

	port, err := strconv.Atoi(hostport[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %s: %w", hostport[1], err)
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
		passwordBytes, passwordErr := ioutil.ReadFile(filepath.Join(*passwordDir, ksAddress))
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

func setGlobalConfig() {
	config.EventListenerHttp = *listenerweb
	if *routerBcastInterval != 0 {
		log.Infof("set router bcast interval to %d seconds", *routerBcastInterval)
		config.RouterBcastInterval = time.Duration(*routerBcastInterval) * time.Second
	}
	if *routerBuildInterval != 0 {
		log.Infof("set router build interval to %d seconds", *routerBuildInterval)
		config.RouterBuildInterval = time.Duration(*routerBuildInterval) * time.Second
	}
	if *routerAliveTimeout != 0 {
		log.Infof("set router alive timeout to %d seconds", *routerAliveTimeout)
		config.RouterAliveTimeout = time.Duration(*routerAliveTimeout) * time.Second
	}
	if *ospClearPayInterval != 0 {
		log.Infof("set osp clear pay interval to %d seconds", *ospClearPayInterval)
		config.OspClearPaysInterval = time.Duration(*ospClearPayInterval) * time.Second
	}
	if *ospReportInterval != 0 {
		log.Infof("set osp report interval to %d seconds", *ospReportInterval)
		config.OspReportInverval = time.Duration(*ospReportInterval) * time.Second
	}
}

func main() {
	flag.Parse()
	if *showver {
		printver()
		os.Exit(0)
	}
	setGlobalConfig()
	var err error

	var ksBytes []byte
	var ksStr string
	if *ks != "" {
		ksBytes, err = ioutil.ReadFile(*ks)
		if err != nil {
			log.Fatalln(err)
		}
		ksStr = string(ksBytes)
	}

	var dksBytes []byte
	var dksStr string
	if *depositks != "" {
		dksBytes, err = ioutil.ReadFile(*depositks)
		if err != nil {
			log.Fatalln(err)
		}
		dksStr = string(dksBytes)
	}

	var routingBytes []byte
	if *routingData != "" {
		routingBytes, err = ioutil.ReadFile(*routingData)
		if err != nil {
			log.Fatalln(err)
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
	rpcServer.netClient = &http.Client{Timeout: 3 * time.Second}
	if *redisAddr != "" {
		rpcServer.redisClient = redis.NewClient(&redis.Options{Addr: *redisAddr})
	}
	masterTxConfig := transactor.NewTransactorConfig(ksStr, readPassword(ksBytes))
	if dksStr != "" {
		depositTxConfig := transactor.NewTransactorConfig(dksStr, readPassword(dksBytes))
		rpcServer.Initialize(masterTxConfig, depositTxConfig, tConfigs, routingBytes)
	} else {
		rpcServer.Initialize(masterTxConfig, nil, tConfigs, routingBytes)
	}
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
		interOSPServer := serverInterOSP{
			svr:      &rpcServer,
			adminSvr: adminS,
		}
		rpc.RegisterMultiServerServer(s2, &interOSPServer)
		go s2.Serve(lis2)
	}

	// Run the main server.
	s.Serve(lis)
}

func setUpAdminService(osp *server) *adminService {
	log.Infoln("Celer server has admin rpc:", *adminrpc)
	lis, err := net.Listen("tcp", *adminrpc)
	if err != nil {
		log.Fatalf("failed to listen on admin rpc: %v", err)
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
	log.Infoln("Celer server has admin HTTP:", *adminweb)
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

// cat ddns.crt. for localhost, 127.0.0.1 and supported DDNS domain names
var certPem = []byte(`-----BEGIN CERTIFICATE-----
MIIEjjCCAnagAwIBAgIQeEANaPaT86NaO+kgWfH7wzANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQDEwdDZWxlckNBMB4XDTIwMDUwNjAwNTIyM1oXDTIzMDUwNjAwNTIy
M1owDzENMAsGA1UEAxMEZGRuczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBALWYuQe+ryY0NemrbknpXnJVaFSEwIr5Y65HEz9okW1wK5RyQ8uOog9+1yyZ
nxxFTxNonnu7VlU5bGkTOs32VKZKMP/IEZOonQlM+dExmuFaP2VcdEoxJkztXjQ/
Q/on1hmhpZr24iMdtbT0r8+L3WZ6W714VB3y4z2C6T0zocmJl4o4tgUXU9j0tNkL
f5/5xnRu8eFV0YJmfczZL3aXZX8ZPi1uej9BcQlLKhgLwRUViFIBVRMSoXcHnO0c
reOfe85g36X5IMEdz+d9B9IZqnTeYQVWHKHCdOod7GJsmhtfFi/mdi7yo0ky3s2T
cE98ijNTNBXNjpplQdIfD/lC4ZkCAwEAAaOB4jCB3zAOBgNVHQ8BAf8EBAMCA7gw
HQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBSXw/Uv9WAT
dHl4Q6vKIbUfRr32yjAfBgNVHSMEGDAWgBSoQm3esw5att/hencMPoPdqhyvXTBu
BgNVHREEZzBlgglsb2NhbGhvc3SCCiouZHludS5jb22CCioubW9vby5jb22CByou
dXMudG+CCyouaG9wdG8ub3JnggsqLnphcHRvLm9yZ4ILKi5zeXRlcy5uZXSCCiou
ZGRucy5uZXSHBH8AAAEwDQYJKoZIhvcNAQELBQADggIBAJC3IxCfpg39J3QBCccf
tSRiTINNzdBx025tv5HHyAcPkg+Okwlb2vgyoVAzOCxouzdNRIeFwHdAjINdlMCa
C69Z73WTh1xGzJz39qaK1AFS/pIgg3QLfJic/RDo0qANlo/dCjxpsdlUuaIPth0W
17sHoWKHwYZ5SoYinaZWm10ZYJXSP6tjsJ7DVzVbkPM1mERvwLq9XrpYjpt+l6Gl
ZS/nIPW5S/RBgQzu/uUj50wRUWFtTy7YSEsXDJ5MVn6kF/KwlAQ99YykY+8xQZ7p
ngPG8C6q7htvPOYSntZ9qSHK1bGU5oGaua0F8UO+HK6dpLCGH1Bv8OQRMuXBNDXD
9gUSX4jnMQdLMrmmqHsv/ymqSx1vM2acPqtLZFmoaXtlz88w2MgTDhPzQqjj/nio
J1Y5KpYQw5lS9pNNJ6cQcgvL1wxle/BuS7s/6bdl94H87aWMRHArodUW332rH5+T
1fpm4souLPTriTBETLmLtdrzPIWX3JToKPovfK78gKc9xdj1gt0GaS0I5EfiRCq/
+7322V9fnA2vTE6oL9NaPrTDyOMFG46vdCdInWRLQda0/0I2UrhBR4AjihFKFwfN
P1tumv7Gj+u5Xo1yim8LYJYZJd6kGeYee5c8dXQLdxL6CNiA6Y8KbWQlBcw1mEam
kuppyJmLe4FI3nhQSA2Hdo0p
-----END CERTIFICATE-----`)

// cat ddns.key
var privPem = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAtZi5B76vJjQ16atuSeleclVoVITAivljrkcTP2iRbXArlHJD
y46iD37XLJmfHEVPE2iee7tWVTlsaRM6zfZUpkow/8gRk6idCUz50TGa4Vo/ZVx0
SjEmTO1eND9D+ifWGaGlmvbiIx21tPSvz4vdZnpbvXhUHfLjPYLpPTOhyYmXiji2
BRdT2PS02Qt/n/nGdG7x4VXRgmZ9zNkvdpdlfxk+LW56P0FxCUsqGAvBFRWIUgFV
ExKhdwec7Ryt4597zmDfpfkgwR3P530H0hmqdN5hBVYcocJ06h3sYmyaG18WL+Z2
LvKjSTLezZNwT3yKM1M0Fc2OmmVB0h8P+ULhmQIDAQABAoIBAGhaJpmIveVaQ1DP
rkTS0e8pnRZUYsB5pwpS1YCuekNFHgXj04l5oS7+WG/Mp137hgFFMS0w7d9zbBTF
dBk+4E3efGG/ejo7vmHF1hicRMASuDNq+XhY3gV+sY+ce1eY5LM97IgsLq63N9+Q
bCb73HIzBdzO9KdxF5wVSipdLS895T1ARx3jUHE0qHi9iOeUpZ0tlDgPyCBhC8yM
44PQHhIYfjcP45e3PxJHNYZ8pW2GyfKLLAp0fB9yFpa73wFF7hQrErYHpavfUrpb
0iVnUruFXK9x8RkpYp9CLSwHzqkLBhrSeun+rd+xoLPALn3z5l7Lr2JD75KwWq36
GN6xJsECgYEA6XQVBYuaZcpHDGPJxEt+FIZoKkrqLONepZy+jSAlAneSHIr5I+rG
Zl07lQu1pelstwRLT21cTFNzTH5E/bSsXY3z2xK6upPlybvtWTdK18Ly//V1h1/o
3Wu45dgRqN61OfZzY3VhCH/m0AsUlhx0ZG4Xe4XNsb3YTG6tuYqVEQ0CgYEAxyKH
LW2W06bUa6Z1TR5j8xDx3EuzV/FGPu43ZsIY3MYr2an50blPvDd6ohaSjTGBA5ji
tLeEvxn3TTETag0JXVHY5Awy7TxibIEc6Ci0uBVM9wpJC7ulr7Bkj3Ujzv+HCOnN
2da7xzoWWnnx4XjQYJe6trNFNq59ek7Fxb7Zt70CgYBZ4zpc4j3Wx6XHnXyIulCX
cpEVB7dXBUziE5JDIIdfhc6acuAu9iziKS/R2apdcZ1McrjUolURhmg0ocHKthBQ
DWJBjdqcEGo/3L4sxd9paETMTSSttT4iU2HkDefuz6GT6dLQDtgHv8eNf1az3czr
BoDALOrgbyqjc1IsyLFQ2QKBgQCEfoe8hR8Gggjx+1Yu0pqj/lE+Ryvzxnf7pRo9
h8VtF3nToFG05fyUecFhmW02/oodr5qfrT9ou0kdFhsZw8TIT43k5ez4heXi1eE6
4uKjXKDR+gjyAo9SHyY5gTIrJz9cgdvSJhrYiuCWoG58m5rhufKkXxEdOeoI9+KF
GN/ZbQKBgQDLivA6qOmamcuqAl+3i0gY/A+LNbyqP55Kj+lPOKDqSRJJwnTMc25C
f2dKuFGmDdeOrOfx/vXm7QVvsFXrPRIQG9/cduVegQNq6/Dpwn+jDuh+SBYMOFgn
o8Wapa1vIvkgQsywSL06IJw+7VdA0uNUS+Abrp26gZMmTMnRfDi7Hw==
-----END RSA PRIVATE KEY-----`)
