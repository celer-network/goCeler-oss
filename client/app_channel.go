// Copyright 2018-2020 Celer Network

package client

import (
	"math/big"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
)

// NewAppChannelOnVirtualContract initializes a generalized state channel with a virtual contract
// It returns the virtual address of the channel
func (c *CelerClient) NewAppChannelOnVirtualContract(
	byteCode []byte,
	constructor []byte,
	nonce uint64,
	onchainTimeout uint64,
	sc common.StateCallback) (string, error) {
	return c.cNode.AppClient.NewAppChannelOnVirtualContract(byteCode, constructor, nonce, onchainTimeout, sc)
}

// NewAppChannelOnDeployedContract initializes a generalized state channel with a deployed contract
// It returns the session ID of the channel
func (c *CelerClient) NewAppChannelOnDeployedContract(
	contractAddr ctype.Addr,
	nonce uint64,
	players []ctype.Addr,
	onchainTimeout uint64,
	sc common.StateCallback) (string, error) {
	return c.cNode.AppClient.NewAppChannelOnDeployedContract(contractAddr, nonce, players, onchainTimeout, sc)
}

// DeleteAppChannel removes the app channel info from the in memory map
func (c *CelerClient) DeleteAppChannel(cid string) error {
	c.cNode.AppClient.DeleteAppChannel(cid)
	return nil
}

// SettleAppChannel tries to settle a app channel onchain
func (c *CelerClient) SettleAppChannel(cid string, stateproof []byte) error {
	return c.cNode.AppClient.SettleAppChannel(cid, stateproof)
}

// GetAppChannelDeployedAddr get the depolyed address of a app channel
// returns error if it's an undeployed virtual contract channel
func (c *CelerClient) GetAppChannelDeployedAddr(cid string) (ctype.Addr, error) {
	return c.cNode.AppClient.GetAppChannelDeployedAddr(cid)
}

func (c *CelerClient) GetAppChannel(cid string) *app.AppChannel {
	return c.cNode.AppClient.GetAppChannel(cid)
}

// SignAppState returns 1: proto serialized app state, 2: signature, 3: error
func (c *CelerClient) SignAppState(cid string, seqNum uint64, state []byte) ([]byte, []byte, error) {
	return c.cNode.AppClient.SignAppState(cid, seqNum, state)
}

// ---------- after switch to onchain ---------

// OnChainGetAppChannelBooleanOutcome returns 1: isFinalized(cid), 2: getOutcome(query), 3: error
func (c *CelerClient) OnChainGetAppChannelBooleanOutcome(cid string, query []byte) (bool, bool, error) {
	return c.cNode.AppClient.GetBooleanOutcome(cid, query)
}

// OnChainApplyAppChannelAction applies onchain action to a app channel
func (c *CelerClient) OnChainApplyAppChannelAction(cid string, action []byte) error {
	return c.cNode.AppClient.ApplyAction(cid, action)
}

// OnChainFinalizeAppChannelOnActionTimeout finalizes a app channel on action timeout
func (c *CelerClient) OnChainFinalizeAppChannelOnActionTimeout(cid string) error {
	return c.cNode.AppClient.FinalizeAppChannelOnActionTimeout(cid)
}

// OnChainGetAppChannelSettleFinalizedTime gets the onchain settle finalized time
func (c *CelerClient) OnChainGetAppChannelSettleFinalizedTime(cid string) (uint64, error) {
	return c.cNode.AppClient.GetAppChannelSettleFinalizedTime(cid)
}

// OnChainGetAppChannelActionDeadline gets the onchain action deadline
func (c *CelerClient) OnChainGetAppChannelActionDeadline(cid string) (uint64, error) {
	return c.cNode.AppClient.GetAppChannelActionDeadline(cid)
}

// OnChainGetAppChannelSeqNum gets the onchain sequence number
func (c *CelerClient) OnChainGetAppChannelSeqNum(cid string) (uint64, error) {
	return c.cNode.AppClient.GetAppChannelSeqNum(cid)
}

// OnChainGetAppChannelStatus gets the onchain status (0:IDLE, 1:SETTLE, 2:ACTION, 3:FINALIZED)
func (c *CelerClient) OnChainGetAppChannelStatus(cid string) (uint8, error) {
	return c.cNode.AppClient.GetAppChannelStatus(cid)
}

// OnChainGetAppChannelState gets the onchain app state associated with the given key
func (c *CelerClient) OnChainGetAppChannelState(cid string, key *big.Int) ([]byte, error) {
	return c.cNode.AppClient.GetAppChannelState(cid, key)
}

// SettleAppChannelBySigTimeout settle an app channel due to signature timeout
func (c *CelerClient) SettleAppChannelBySigTimeout(cid string, oracleProof []byte) error {
	return c.cNode.AppClient.SettleBySigTimeout(cid, oracleProof)
}

// SettleAppChannelByMoveTimeout settle an app channel due to movement timeout
func (c *CelerClient) SettleAppChannelByMoveTimeout(cid string, oracleProof []byte) error {
	return c.cNode.AppClient.SettleByMoveTimeout(cid, oracleProof)
}

// SettleAppChannelByInvalidTurn settle an app channel due to invalid turn
func (c *CelerClient) SettleAppChannelByInvalidTurn(cid string, oracleProof []byte, cosignedStateProof []byte) error {
	return c.cNode.AppClient.SettleByInvalidTurn(cid, oracleProof, cosignedStateProof)
}

// SettleAppChannelByInvalidState settle an app channel due to invalid state
func (c *CelerClient) SettleAppChannelByInvalidState(cid string, oracleProof []byte, cosignedStateProof []byte) error {
	return c.cNode.AppClient.SettleByInvalidState(cid, oracleProof, cosignedStateProof)
}
