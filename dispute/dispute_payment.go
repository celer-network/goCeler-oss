// Copyright 2018-2019 Celer Network

package dispute

import (
	"math/big"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/payresolver"
	"github.com/celer-network/goCeler-oss/chain"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
)

// SettleConditionalPay resolves a conditonal payment on chain in the PayRegistry
func (p *Processor) SettleConditionalPay(payID ctype.PayIDType) error {
	return p.resolvePaymentByConditions(payID)
}

// OnPaymentUpdatedInRegistry reacts to a payment resolve event
// Currently not implemented since dispute by vouched result is not supported
func (p *Processor) OnPaymentUpdatedInRegistry(event *payregistry.PayRegistryPayInfoUpdate) error {
	return nil
}

func (p *Processor) resolvePaymentByVouchedResult(payID ctype.PayIDType) error {
	return nil
}

func (p *Processor) resolvePaymentByConditions(payID ctype.PayIDType) error {
	log.Infoln("resolve payment by conditions, payID:", payID.Hex())
	pay, payBytes, err := p.dal.GetConditionalPay(payID)
	if err != nil {
		log.Errorln(common.ErrPayNotFound, err, payID.Hex())
		return common.ErrPayNotFound
	}
	amt, _, err := p.GetCondPayInfoFromRegistry(payID)
	if err != nil {
		return err
	}
	maxAmt := utils.BytesToBigInt(pay.TransferFunc.MaxTransfer.Receiver.Amt)
	if amt.Cmp(maxAmt) == 0 {
		// return nli if payment is already resolved to max
		return nil
	}
	if pay.ResolveDeadline < p.monitorService.GetCurrentBlockNumber().Uint64() {
		log.Errorln(common.ErrDeadlinePassed, "pay:", utils.PrintConditionalPay(pay))
		return common.ErrDeadlinePassed
	}

	request := &chain.ResolvePayByConditionsRequest{
		CondPay:       payBytes,
		HashPreimages: [][]byte{},
	}
	for _, cond := range pay.GetConditions() {
		if cond.ConditionType == entity.ConditionType_HASH_LOCK {
			lock := ctype.Bytes2Hex(cond.HashLock)
			secret, err2 := p.dal.GetSecretRegistry(lock)
			if err2 != nil {
				log.Errorln("secret not revealed for hash lock", lock)
				return common.ErrSecretNotRevealed
			}
			preimage := utils.Pad(ctype.Hex2Bytes(secret), 32)
			request.HashPreimages = append(request.HashPreimages, preimage)
		}
	}
	serializedRequest, err := proto.Marshal(request)

	_, err = p.transactorPool.SubmitAndWaitMinedWithGenericHandler(
		"resolve payment by conditions",
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 :=
				payresolver.NewPayResolverTransactor(ctype.Bytes2Addr(pay.GetPayResolver()), transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.ResolvePaymentByConditions(opts, serializedRequest)
		})
	if err != nil {
		// check onchain again to handle cases when client call it multiple time
		amt, _, _ := p.GetCondPayInfoFromRegistry(payID)
		if amt.Cmp(maxAmt) == 0 {
			return nil
		}
		log.Errorln("ResolvePaymentByConditions tx error", err, "pay:", utils.PrintConditionalPay(pay))
		return err
	}
	return nil
}

func (p *Processor) GetCondPayInfoFromRegistry(payID ctype.PayIDType) (*big.Int, uint64, error) {
	contract, err := payregistry.NewPayRegistryCaller(
		p.nodeConfig.GetPayRegistryContract().GetAddr(), p.transactorPool.ContractCaller())
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}
	info, err := contract.PayInfoMap(&bind.CallOpts{}, payID)
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}
	log.Debugln("pay registry info", payID.String(), info.Amount, info.ResolveDeadline.Uint64())
	return info.Amount, info.ResolveDeadline.Uint64(), nil
}
