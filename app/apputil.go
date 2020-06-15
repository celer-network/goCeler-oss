// Copyright 2018-2020 Celer Network

package app

import (
	"bytes"
	"sort"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

func EncodeAppState(nonce uint64, seqNum uint64, state []byte, disputeTimeout uint64) ([]byte, error) {
	appState := &AppState{
		Nonce:   nonce,
		SeqNum:  seqNum,
		State:   state,
		Timeout: disputeTimeout,
	}
	appStateBytes, err := proto.Marshal(appState)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return appStateBytes, nil
}

func DecodeAppState(appStateBytes []byte) (uint64, uint64, []byte, uint64, error) {
	var appState AppState
	err := proto.Unmarshal(appStateBytes, &appState)
	if err != nil {
		return 0, 0, nil, 0, err
	}
	return appState.Nonce, appState.SeqNum, appState.State, appState.Timeout, nil
}

func EncodeAppStateProof(appStateBytes []byte, sigs [][]byte) ([]byte, error) {
	stateProof := &StateProof{
		AppState: appStateBytes,
	}
	for _, sig := range sigs {
		stateProof.Sigs = append(stateProof.Sigs, sig)
	}
	stateProofBytes, err := proto.Marshal(stateProof)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return stateProofBytes, nil
}

func DecodeAppStateProof(stateProofBytes []byte) ([]byte, [][]byte, error) {
	var stateProof StateProof
	err := proto.Unmarshal(stateProofBytes, &stateProof)
	if err != nil {
		return nil, nil, err
	}
	return stateProof.AppState, stateProof.Sigs, nil
}

func SigSortedAppStateProof(proof []byte) ([]byte, error) {
	var stateproof StateProof
	err := proto.Unmarshal(proof, &stateproof)
	if err != nil {
		return nil, err
	}
	stateproof.Sigs = SortPlayerSigs(stateproof.AppState, stateproof.Sigs)
	return proto.Marshal(&stateproof)
}

type sortPlayerSigs struct {
	players [][]byte
	sigs    [][]byte
}

func (s sortPlayerSigs) Len() int {
	return len(s.sigs)
}

func (s sortPlayerSigs) Less(i, j int) bool {
	switch bytes.Compare(s.players[i], s.players[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Error("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (s sortPlayerSigs) Swap(i, j int) {
	s.players[j], s.players[i] = s.players[i], s.players[j]
	s.sigs[j], s.sigs[i] = s.sigs[i], s.sigs[j]
}

func SortPlayerSigs(state []byte, sigs [][]byte) [][]byte {
	sorted := sortPlayerSigs{
		sigs: sigs,
	}
	for _, sig := range sigs {
		signer, err := eth.RecoverSigner(state, sig)
		if err != nil {
			log.Error(err)
			continue
		}
		sorted.players = append(sorted.players, signer.Bytes())
	}
	sort.Sort(sorted)
	return sorted.sigs
}

type sortPlayers []ctype.Addr

func (b sortPlayers) Len() int {
	return len(b)
}

func (b sortPlayers) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i].Bytes(), b[j].Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortPlayers) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

func SortPlayers(src []ctype.Addr) []ctype.Addr {
	sorted := sortPlayers(src)
	sort.Sort(sorted)
	return sorted
}
