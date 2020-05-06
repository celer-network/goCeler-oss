// Copyright 2018-2020 Celer Network

// SDK APIs dealing with app channels, ie. app session

package celersdk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/big"
	"sort"
	"strings"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/client"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/atomic"
)

const NUM_PLAYERS = 2

var (
	ErrMissingSigs         = errors.New("missing sigs from received ack msg")
	ErrDiffAckState        = errors.New("ack msg has different state")
	ErrWrongSeqNum         = errors.New("wrong seqnum")
	ErrDiffAppState        = errors.New("appstate is different")
	ErrWrongOpcode         = errors.New("unknown opcode")
	ErrWrongPlayers        = errors.New("players list is wrong")
	ErrWrongNonce          = errors.New("wrong nonce")
	ErrWrongOnChainTimeout = errors.New("wrong onchain timeout")
	ErrInvalidSession      = errors.New("invalid app session ")
)

type AppInfo struct {
	DeployedAddr   string
	ContractBin    string
	OnChainTimeout int64
	Callback       AppCallback
}
type AppCallback interface {
	common.StateCallback
}

type AppSession struct {
	ID string
	// TODO(mzhou): MyIdx should not be enforced
	MyIdx         int64 // MyIdx in this appsession, eg. 0 if I'm black on gomoku
	cc            *client.CelerClient
	seqnum        *atomic.Uint64 // maintain seqnum and update on send/recv
	lastSentState []byte
	lastRecvState []byte
	// expectNewStateFromPeer indicates whether I'm expecting to receive new state from peer
	// default true, set to false after receive new state and set to true after signappdata
	expectNewStateFromPeer *atomic.Bool
}

// newAppSession creates a new CApp session and return the session identifier as string
func (mc *Client) newAppSession(capp *AppInfo, constructor string, nonce int64) (string, error) {
	return mc.c.NewAppChannelOnVirtualContract(
		ctype.Hex2Bytes(capp.ContractBin),
		ctype.Hex2Bytes(constructor),
		uint64(nonce),
		uint64(capp.OnChainTimeout),
		capp.Callback)
}

func (mc *Client) CreateAppSessionOnVirtualContract(
	contractBin string,
	constructor string,
	nonce uint64,
	onChainTimeout uint64,
	callback AppCallback) (*AppSession, error) {
	sessionID, err := mc.c.NewAppChannelOnVirtualContract(
		ctype.Hex2Bytes(contractBin),
		ctype.Hex2Bytes(constructor),
		nonce,
		onChainTimeout,
		callback)
	if err != nil {
		return nil, err
	}
	return &AppSession{
		ID:                     sessionID,
		MyIdx:                  0, // dummy
		cc:                     mc.c,
		seqnum:                 atomic.NewUint64(0),
		expectNewStateFromPeer: atomic.NewBool(true),
	}, nil
}

func (mc *Client) CreateAppSessionOnDeployedContract(
	contractAddress string,
	nonce uint64,
	onChainTimeout uint64,
	participants string,
	callback AppCallback) (*AppSession, error) {
	var participantAddrs []ctype.Addr
	splitted := strings.Split(participants, ",")
	for _, participant := range splitted {
		participantAddrs = append(participantAddrs, ctype.Hex2Addr(participant))
	}
	sessionID, err := mc.c.NewAppChannelOnDeployedContract(
		ctype.Hex2Addr(contractAddress),
		nonce,
		participantAddrs,
		onChainTimeout,
		callback)
	if err != nil {
		return nil, err
	}
	return &AppSession{
		ID:                     sessionID,
		MyIdx:                  0, // dummy
		cc:                     mc.c,
		seqnum:                 atomic.NewUint64(0),
		expectNewStateFromPeer: atomic.NewBool(true),
	}, nil
}

func (mc *Client) EndAppSession(sessionid string) error {
	return mc.c.DeleteAppChannel(sessionid)
}

// NewAppSession creates app session object for deployed contract
// deployedAddr is eth address bytes of deployed app contract
// matchid is the matchid string from nakama server
// players is players ETH addresses seperated by comma, eg: ab...12,bc...23
// due to gobind type limitation.
// return AppSession and AppSession.ID can be used to end session
func (mc *Client) NewAppSessionOnDeployedContract(capp *AppInfo, matchid string, players string) (*AppSession, error) {
	contractAddr := ctype.Hex2Addr(capp.DeployedAddr)
	nonce := binary.BigEndian.Uint64(crypto.Keccak256([]byte(matchid)))
	var plist []ctype.Addr
	p := strings.Split(players, ",")
	for _, i := range p {
		plist = append(plist, ctype.Hex2Addr(i))
	}
	if len(plist) != NUM_PLAYERS {
		return nil, ErrWrongPlayers
	}
	// Fair algo for assigning myidx, assume nonce is sufficiently random, need to ensure two clients see different
	// MyIdx based on same nonce. So we sort player by address(to ensure both clients see same), then pick nonce%2 in the players list
	// as idx 0, the other addr is idx 1. Compare my own address with it to get MyIdx
	sortedPlayers := app.SortPlayers(plist)
	idx1addr := sortedPlayers[nonce%NUM_PLAYERS]
	var myidx int64
	if idx1addr == mc.c.MyAddress() {
		myidx = 0
	} else {
		myidx = 1
	}

	sid, err := mc.c.NewAppChannelOnDeployedContract(
		contractAddr, nonce, plist, uint64(capp.OnChainTimeout), capp.Callback)
	if err != nil {
		return nil, err
	}
	return &AppSession{
		ID:                     sid,
		MyIdx:                  myidx,
		cc:                     mc.c,
		seqnum:                 atomic.NewUint64(0),
		expectNewStateFromPeer: atomic.NewBool(true),
	}, nil
}

// SignAppData takes app data, add proper metadata and return final bytes ready to be sent via nakama
// note this func will incr session seq number and set expect new state from peer to true
func (s *AppSession) SignAppData(in []byte) ([]byte, error) {
	newseq := s.seqnum.Inc()
	s.expectNewStateFromPeer.Store(true)
	serialized, sig, err := s.cc.SignAppState(s.ID, newseq, in)
	if err != nil {
		return nil, err
	}
	s.lastSentState = serialized
	return app.EncodeAppStateProof(serialized, [][]byte{sig})
}

// AppData has 2 fields, Received is to be passed to celerx
// AckMsg is to be sent via nakama with opcode OPCODE_ACK
type AppData struct {
	Received []byte
	AckMsg   []byte
}

const (
	OPCODE_NEWSTATE = 1
	OPCODE_ACK      = 2
)

// HandleMatchData process received matchdata via nakama
// opcode 1: new state with one sig
// opcode 2: ack msg with new sig appended
// data should be bytes of AppStateProof
// for opcode 2, *AppData is empty, just check error
// for opcode 1, if *AppData isn't nil, both fields should be set
func (s *AppSession) HandleMatchData(opcode int, data []byte) (*AppData, error) {
	appstate, sigs, err := app.DecodeAppStateProof(data)
	if err != nil {
		return nil, err
	}
	switch opcode {
	case OPCODE_ACK:
		if len(sigs) != NUM_PLAYERS { // only work for 1v1 now
			return nil, ErrMissingSigs
		}
		if !bytes.Equal(s.lastSentState, appstate) {
			log.Errorf("%s expect:%x recv:%x", ErrDiffAckState, s.lastSentState, appstate)
			return nil, ErrDiffAckState
		}
		return new(AppData), nil
	case OPCODE_NEWSTATE:
		nonce, seqn, recv, timeout, err := app.DecodeAppState(appstate)
		if err != nil {
			return nil, err
		}
		appChannel := s.cc.GetAppChannel(s.ID)
		if appChannel == nil {
			log.Error(ErrInvalidSession)
			return nil, ErrInvalidSession
		}
		if nonce != appChannel.Nonce {
			log.Errorf("%s expect:%d recv:%d", ErrWrongSeqNum, appChannel.Nonce, nonce)
			return nil, ErrWrongNonce
		}
		if timeout != appChannel.OnChainTimeout {
			log.Errorf("%s expect:%d recv:%d", ErrWrongOnChainTimeout, appChannel.OnChainTimeout, timeout)
			return nil, ErrWrongOnChainTimeout
		}
		myseq := s.seqnum.Load()
		isExpect := s.expectNewStateFromPeer.Load()
		if !isExpect {
			// not expecting msg, consider this is due to peer re-send last state
			// myseq already incr in last recv so seqn should be the same
			if seqn != myseq {
				log.Errorf("%s expect:%d recv:%d", ErrWrongSeqNum, myseq, seqn)
				return nil, ErrWrongSeqNum
			}
			if !bytes.Equal(recv, s.lastRecvState) {
				log.Errorf("%s expect:%x recv:%x", ErrDiffAppState, s.lastRecvState, recv)
				return nil, ErrDiffAppState
			}
		} else {
			// expect new state, seqn must be myseq+1
			if seqn != myseq+1 {
				log.Errorf("%s expect:%d recv:%d", ErrWrongSeqNum, myseq+1, seqn)
				return nil, ErrWrongSeqNum
			}
			s.lastRecvState = recv
			s.seqnum.Store(seqn)
		}
		s.expectNewStateFromPeer.Store(false)

		appstate2, mysig, err := s.cc.SignAppState(s.ID, seqn, recv)
		if !bytes.Equal(appstate, appstate2) {
			log.Errorf("%s recv:%x tosend:%x", ErrDiffAppState, appstate, appstate2)
			return nil, ErrDiffAppState
		}
		sigs = append(sigs, mysig)
		ackMsg, err := app.EncodeAppStateProof(appstate, sigs)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return &AppData{
			Received: recv,
			AckMsg:   ackMsg,
		}, nil
	default:
		return nil, ErrWrongOpcode
	}
}

// GetPlayerIdxForMatch returns a player idx (0 based) for myaddr. myaddr must be in players.
// players is ETH addresses seperated by comma, eg: ab...12,bc...23
// it uses a fair algo for assigning myidx, based on a number generated from matchid
// players list is sorted first for consistency
// return -1 on err
func GetPlayerIdxForMatch(matchid, myaddr, players string) int64 {
	nonce := binary.BigEndian.Uint64(crypto.Keccak256([]byte(matchid)))
	var plist []ctype.Addr
	p := strings.Split(players, ",")
	for _, i := range p {
		plist = append(plist, ctype.Hex2Addr(i))
	}
	if len(p) < 2 {
		log.Error("too few players: ", len(p))
		return -1
	}
	// sort plist in place
	sort.Slice(plist, func(i, j int) bool { return bytes.Compare(plist[i].Bytes(), plist[j].Bytes()) < 0 })
	// rearrange plist so cutIdx becomes 0
	cutIdx := nonce % uint64(len(p))
	plist = append(plist[cutIdx:], plist[0:cutIdx]...)
	my := ctype.Hex2Addr(myaddr)
	for i, a := range plist {
		if a == my {
			return int64(i)
		}
	}
	log.Error("not found in players: ", myaddr)
	return -1
}

// -------------------- Switch to Onchain -------------------

// SwitchToOnchain submits offchain stateproof to onchain and starts onchain play
func (s *AppSession) SwitchToOnchain(stateproof []byte) error {
	return s.cc.SettleAppChannel(s.ID, stateproof)
}

// GetDeployedAddress get the depolyed address of the app
// returns error if the app is based on an undeployed virtual contract
func (s *AppSession) GetDeployedAddress() (string, error) {
	addr, err := s.cc.GetAppChannelDeployedAddr(s.ID)
	return ctype.Addr2Hex(addr), err
}

// AppBooleanOutcome has two fields
// Finalized: if the app is finalized
// Outcome: the boolean outcome with given query arg
type AppBooleanOutcome struct {
	Finalized bool
	Outcome   bool
}

// OnChainGetBooleanOutcome returns app boolean outcome
func (s *AppSession) OnChainGetBooleanOutcome(query []byte) (*AppBooleanOutcome, error) {
	finalized, outcome, err := s.cc.OnChainGetAppChannelBooleanOutcome(s.ID, query)
	boolres := AppBooleanOutcome{Finalized: finalized, Outcome: outcome}
	return &boolres, err
}

// OnChainApplyAction applies an action on chain
func (s *AppSession) OnChainApplyAction(action []byte) error {
	return s.cc.OnChainApplyAppChannelAction(s.ID, action)
}

// OnChainFinalizeOnActionTimeout finalizes the app on action timeout
func (s *AppSession) OnChainFinalizeOnActionTimeout() error {
	return s.cc.OnChainFinalizeAppChannelOnActionTimeout(s.ID)
}

// OnChainGetSettleFinalizedTime gets the app onchain settle finalized time
func (s *AppSession) OnChainGetSettleFinalizedTime() (int64, error) {
	blkNum, err := s.cc.OnChainGetAppChannelSettleFinalizedTime(s.ID)
	return int64(blkNum), err
}

// OnChainGetActionDeadline gets the app onchain action deadline
func (s *AppSession) OnChainGetActionDeadline() (int64, error) {
	blkNum, err := s.cc.OnChainGetAppChannelActionDeadline(s.ID)
	return int64(blkNum), err
}

// OnChainGetStatus gets the app onchain status (0:IDLE, 1:SETTLE, 2:ACTION, 3:FINALIZED)
func (s *AppSession) OnChainGetStatus() (int8, error) {
	status, err := s.cc.OnChainGetAppChannelStatus(s.ID)
	return int8(status), err
}

// OnChainGetState gets the app onchain state associated with the given key
func (s *AppSession) OnChainGetState(key int64) ([]byte, error) {
	bigkey := big.NewInt(key)
	return s.cc.OnChainGetAppChannelState(s.ID, bigkey)
}

// OnChainGetSeqNum gets the app onchain sequence number
func (s *AppSession) OnChainGetSeqNum() (int64, error) {
	seq, err := s.cc.OnChainGetAppChannelSeqNum(s.ID)
	return int64(seq), err
}

// SettleBySigTimeout settle an app channel due to signature timeout
func (s *AppSession) SettleBySigTimeout(oracleProof []byte) error {
	return s.cc.SettleAppChannelBySigTimeout(s.ID, oracleProof)
}

// SettleByMoveTimeout settle an app channel due to movement timeout
func (s *AppSession) SettleByMoveTimeout(oracleProof []byte) error {
	return s.cc.SettleAppChannelByMoveTimeout(s.ID, oracleProof)
}

// SettleByInvalidTurn settle an app channel due to invalid turn
func (s *AppSession) SettleByInvalidTurn(oracleProof []byte, cosignedStateProof []byte) error {
	return s.cc.SettleAppChannelByInvalidTurn(s.ID, oracleProof, cosignedStateProof)
}

// SettleByInvalidState settle an app channel due to invalid state
func (s *AppSession) SettleByInvalidState(oracleProof []byte, cosignedStateProof []byte) error {
	return s.cc.SettleAppChannelByInvalidState(s.ID, oracleProof, cosignedStateProof)
}
