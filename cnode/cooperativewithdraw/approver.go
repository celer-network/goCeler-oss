// Copyright 2018-2019 Celer Network

package cooperativewithdraw

import (
	"errors"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/golang/protobuf/proto"
)

func (p *Processor) ProcessRequest(
	request *rpc.CooperativeWithdrawRequest) error {
	cid := ctype.Bytes2Cid(request.WithdrawInfo.ChannelId)
	return p.sendResponse(cid, request)
}

func (p *Processor) sendResponse(
	cid ctype.CidType, request *rpc.CooperativeWithdrawRequest) error {
	withdrawInfo := request.WithdrawInfo

	peer, err := p.dal.GetPeer(cid)
	if err != nil {
		return err
	}
	serializedInfo, err := proto.Marshal(withdrawInfo)
	if err != nil {
		return err
	}
	if !p.signer.SigIsValid(peer, serializedInfo, request.RequesterSig) {
		return errors.New("Invalid CooperativeWithdrawRequest signature")
	}

	approverSig, err := p.signer.Sign(serializedInfo)
	if err != nil {
		return err
	}

	err = p.dal.Transactional(p.checkWithdrawBalanceTx, cid, withdrawInfo)
	if err != nil {
		return err
	}

	response := &rpc.CooperativeWithdrawResponse{
		WithdrawInfo: withdrawInfo,
		RequesterSig: request.RequesterSig,
		ApproverSig:  approverSig,
	}
	msg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_WithdrawResponse{
			WithdrawResponse: response,
		},
	}
	log.Infoln("Sending withdraw response of seq", withdrawInfo.SeqNum, "to", peer)
	err = p.streamWriter.WriteCelerMsg(peer, msg)
	if err != nil {
		return err
	}
	return nil
}
