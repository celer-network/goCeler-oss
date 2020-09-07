// Copyright 2018-2020 Celer Network

package testapp

import (
	"math/big"
	"strings"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/golang/protobuf/proto"
)

var (
	AppCode                = ctype.Hex2Bytes(SimpleSingleSessionAppBin)
	AppWithOracleCode      = ctype.Hex2Bytes(SimpleSingleSessionAppWithOracleBin)
	Nonce                  = big.NewInt(666)
	Timeout                = big.NewInt(2)
	PlayerNum              = big.NewInt(2)
	ContractAddr           = "58712219a4bdbb0e581dcaf6f5c4c2b2d2f42158"
	ContractWithOracleAddr = "283ab9db53f25d84fa30915816ec53f8affaa86e"

	GomokuMinOffChain = uint8(0)
	GomokuMaxOnChain  = uint8(10)
	GomokuAddr        = "4e4a0101cd72258183586a51f8254e871b9c544a"
)

// GetSingleSessionConstructor generates abi-conforming constructors for SingleSession App
func GetSingleSessionConstructor(players []ctype.Addr) []byte {
	abi, err := abi.JSON(strings.NewReader(SimpleSingleSessionAppABI))
	input, err := abi.Pack("", players, Nonce, Timeout)
	if err != nil {
		log.Error(err)
	}
	return input
}

func GetSingleSessionWithOracleConstructor(players []ctype.Addr, oracle ctype.Addr) []byte {
	abi, err := abi.JSON(strings.NewReader(SimpleSingleSessionAppABI))
	input, err := abi.Pack("", players, Nonce, Timeout, Timeout, oracle)
	if err != nil {
		log.Error(err)
	}
	return input
}

func GetAppState(winner uint8, nonce uint64) []byte {
	appState := &app.AppState{
		Nonce:  nonce,
		SeqNum: 10,
		State:  []byte{winner},
	}
	serializedAppState, err := proto.Marshal(appState)
	if err != nil {
		log.Error(err)
	}
	return serializedAppState
}

func GetAppStateWithOracle(turn uint8, winner uint8, nonce uint64) []byte {
	appState := &app.AppState{
		Nonce:  nonce,
		SeqNum: 10,
		State:  []byte{turn, winner},
	}
	serializedAppState, err := proto.Marshal(appState)
	if err != nil {
		log.Error(err)
	}
	return serializedAppState
}

func GetGetGomokuBoardState() []byte {
	var state [228]byte
	state[1] = 2 // turn
	state[2] = 1
	state[3] = 1
	return state[:]
}

func GetGomokuState() []byte {
	appState := &app.AppState{
		Nonce:   Nonce.Uint64(),
		SeqNum:  10,
		State:   GetGetGomokuBoardState(),
		Timeout: 3,
	}
	serializedAppState, err := proto.Marshal(appState)
	if err != nil {
		log.Error(err)
	}
	return serializedAppState
}
