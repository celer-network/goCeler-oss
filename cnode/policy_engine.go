// Copyright 2018-2020 Celer Network

package cnode

import (
	"errors"
	"math/big"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	// Enum of policy allowed.
	AllowTcbOpenChannel      = 1 << iota
	AllowStandardOpenChannel = 1 << iota
)

func getDepositMap(dist []*entity.AccountAmtPair) map[ctype.Addr]*big.Int {
	depoMap := make(map[ctype.Addr]*big.Int)
	for _, acntAmtPair := range dist {
		var addr ctype.Addr
		addr = ctype.Bytes2Addr(acntAmtPair.GetAccount())
		depoMap[addr] = new(big.Int).SetBytes(acntAmtPair.GetAmt())
	}
	return depoMap
}
func RequestTcbDeposit(dal *storage.DAL, nodeConfig common.GlobalNodeConfig, initializer *entity.PaymentChannelInitializer) (int, error) {
	noPolicyAllowed := 0
	myAddr := nodeConfig.GetOnChainAddr()
	token := initializer.GetInitDistribution().GetToken()
	distribution := initializer.GetInitDistribution().GetDistribution()
	tokenAddr := utils.GetTokenAddrStr(token)

	depositMap := getDepositMap(distribution)
	myDeposit := depositMap[myAddr]
	// Whitelisted? No config means allowing all.
	if rtconfig.GetTcbConfigs() == nil {
		err := dal.Transactional(increaseTcbCommittedTx, myAddr, ctype.Bytes2Addr(token.TokenAddress), myDeposit)
		if err != nil {
			return noPolicyAllowed, err
		}
		return AllowTcbOpenChannel, nil
	}
	config, tokenAllowed := rtconfig.GetTcbConfigs().GetConfig()[tokenAddr]
	tokenNotAllowed := errors.New("Token not allowed")
	unparsableRtConfig := errors.New("Can't parse rtconfig")
	tcbExceedMaxOspDeposit := errors.New("TCB exceeds max osp deposit")
	tcbZeroOspDeposit := errors.New("TCB asks osp to deposit zero")
	tcbNonZeroPeerDeposit := errors.New("TCB asks peer to deposit non-zero")
	tcbFailOvercommitCheck := errors.New("Osp doesn't have enough money in the pool")
	if !tokenAllowed {
		log.Errorln("No policy allowed")
		return noPolicyAllowed, tokenNotAllowed
	}
	// Deposit Amount
	for addr, deposit := range depositMap {
		if addr == myAddr {
			maxOspDeposit, success := new(big.Int).SetString(config.GetMaxOspDeposit(), 10)
			if !success {
				log.Errorln("Can't parse max osp deposit in decimal", config.GetMaxOspDeposit())
				return noPolicyAllowed, unparsableRtConfig
			}
			if deposit.Cmp(maxOspDeposit) == 1 {
				log.Errorf(
					"TCB exceeds max osp deposit. ask: %s max: %s, tokenAddr: 0x%x", deposit.String(), maxOspDeposit.String(), token.TokenAddress)
				return noPolicyAllowed, tcbExceedMaxOspDeposit
			}
			if deposit.Cmp(ctype.ZeroBigInt) == 0 {
				log.Errorf(
					"TCB asks 0 deposit from osp tokenAddr: 0x%x", token.TokenAddress)
				return noPolicyAllowed, tcbZeroOspDeposit
			}
			myDeposit = deposit
		} else {
			if deposit.Cmp(ctype.ZeroBigInt) != 0 {
				log.Errorf("TCB client deposit not zero %s, tokenAddr: 0x%x", deposit.String(), token.TokenAddress)
				return noPolicyAllowed, tcbNonZeroPeerDeposit
			}
		}
	}

	if config.GetSkipOverCommitCheck() {
		return AllowTcbOpenChannel, nil
	}
	// On-chain deposit capacity left
	depositCapacity, err := getDepositCapacity(nodeConfig, tokenAddr)
	if err != nil {
		log.Errorf("%s, TCB failed on checking deposit capacity for token 0x%s", err, tokenAddr)
		return noPolicyAllowed, err
	}
	err = dal.Transactional(checkAndIncreaseCommittedTx, myAddr, ctype.Bytes2Addr(token.TokenAddress), myDeposit, depositCapacity)
	if err != nil {
		log.Errorf("%s, TCB failed checkAndIncreaseCommittedTx on token %s", err, tokenAddr)
		return noPolicyAllowed, tcbFailOvercommitCheck
	}
	return AllowTcbOpenChannel, nil
}
func RecycleInstantiatedTcbDepositTx(tx *storage.DALTx, args ...interface{}) error {
	descriptor := args[0].(*openedChannelDescriptor)
	myAddr := args[1].(ctype.Addr)

	myDeposit := big.NewInt(0)
	if descriptor.participants[0] == myAddr {
		myDeposit = descriptor.initDeposits[0]
	} else if descriptor.participants[1] == myAddr {
		myDeposit = descriptor.initDeposits[1]
	} else {
		log.Errorf("Not having me. p1: 0x%x p2: 0x%x", descriptor.participants[0], descriptor.participants[1])
		return common.ErrChannelDescriptorNotInclude
	}
	// Add myDeposit back to tcbCommited
	tokenInfo := utils.GetTokenInfoFromAddress(descriptor.tokenAddress)
	tcbCommitted, found, err := tx.GetTcbDeposit(myAddr, tokenInfo)
	if err != nil {
		log.Warnf("%v, p1 0x%x p2 0x%x", err, descriptor.participants[0], descriptor.participants[1])
		return err
	} else if !found {
		log.Warnf("tcb deposit not found: p1 0x%x p2 0x%x", descriptor.participants[0], descriptor.participants[1])
		return common.ErrTcbNotFound
	}
	tcbCommitted.Sub(tcbCommitted, myDeposit)
	err = tx.UpdateTcbDeposit(myAddr, tokenInfo, tcbCommitted)
	if err != nil {
		log.Errorf("%v, p1 0x%x p2 0x%x", err, descriptor.participants[0], descriptor.participants[1])
		return err
	}
	return nil
}

// checkAndIncreaseCommittedTx checks if osp has enough balance based on
// https://docs.google.com/document/d/1ho-FHUkgvWa2Rmr_qFbPRa9rUb6WW8zHQECPKxF-y7Y/edit#bookmark=id.bai7h7eldp16k
// It increases committed TCB balance if balance is sufficient.
func checkAndIncreaseCommittedTx(tx *storage.DALTx, args ...interface{}) error {
	myAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	tokenAddrStr := ctype.Addr2Hex(tokenAddr)
	myDeposit := args[2].(*big.Int)
	depositCapacity := args[3].(*big.Int)
	allowedBalance := big.NewInt(0)

	// allowedBalance = depositCapacity - safeMargin - tcbCommited
	safeMargin, ok := new(big.Int).SetString(rtconfig.GetTcbConfigs().GetConfig()[tokenAddrStr].GetOnchainBalanceSafeMargin(), 10)
	if !ok {
		log.Errorln("can't get safe margin:", rtconfig.GetTcbConfigs().GetConfig()[tokenAddrStr].GetOnchainBalanceSafeMargin())
		return common.ErrUnparsable
	}
	allowedBalance.Sub(depositCapacity, safeMargin)
	tokenInfo := utils.GetTokenInfoFromAddress(tokenAddr)
	tcbCommitted, found, err := tx.GetTcbDeposit(myAddr, tokenInfo)
	if err != nil {
		log.Errorln(err, "get tcb committed. token", tokenAddrStr)
		return err
	} else if !found {
		tcbCommitted = big.NewInt(0)
	}
	allowedBalance.Sub(allowedBalance, tcbCommitted)
	if allowedBalance.Cmp(myDeposit) != 1 {
		log.Errorf("NOT ENOUGH ONCHAIN BALANCE FOR TCB allowed:%s, ask:%s", allowedBalance.String(), myDeposit.String())
		return common.ErrInsufficentDepositCapacity
	}
	tcbCommitted.Add(tcbCommitted, myDeposit)
	if found {
		err = tx.UpdateTcbDeposit(myAddr, tokenInfo, tcbCommitted)
	} else {
		err = tx.InsertTcb(myAddr, tokenInfo, tcbCommitted)
	}
	return err
}

// increaseTcbCommittedTx increases committed tcb balance without any policy checking.
func increaseTcbCommittedTx(tx *storage.DALTx, args ...interface{}) error {
	myAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	tokenAddrStr := ctype.Addr2Hex(tokenAddr)
	myDeposit := args[2].(*big.Int)

	tokenInfo := utils.GetTokenInfoFromAddress(tokenAddr)
	tcbCommitted, found, err := tx.GetTcbDeposit(myAddr, tokenInfo)
	if err != nil {
		log.Errorln(err, "get tcb committed. token", tokenAddrStr)
		return err
	}
	if !found {
		tcbCommitted = big.NewInt(0)
	}
	tcbCommitted.Add(tcbCommitted, myDeposit)
	if found {
		err = tx.UpdateTcbDeposit(myAddr, tokenInfo, tcbCommitted)
	} else {
		err = tx.InsertTcb(myAddr, tokenInfo, tcbCommitted)
	}
	return err
}
func getDepositCapacity(nodeConfig common.GlobalNodeConfig, tokenAddr string) (*big.Int, error) {
	conn := nodeConfig.GetEthConn()
	tokenAddrToCheck := ctype.Hex2Addr(tokenAddr)
	// ETH pool acts as a ERC20 for OSP. ETH capacity is on addr of Eth pool, not on OSP addr
	if tokenAddr == ctype.EthTokenAddrStr {
		tokenAddrToCheck = nodeConfig.GetEthPoolAddr()
	}

	erc20Contract, err := chain.NewERC20(tokenAddrToCheck, conn)
	if err != nil {
		return ctype.ZeroBigInt, err
	}
	allowance, err := erc20Contract.Allowance(&bind.CallOpts{}, nodeConfig.GetOnChainAddr(), nodeConfig.GetLedgerContract().GetAddr())
	if err != nil {
		log.Errorln(err, "getting allowance for token", tokenAddr)
		return nil, err
	}
	balance, err := erc20Contract.BalanceOf(&bind.CallOpts{}, nodeConfig.GetOnChainAddr())
	if err != nil {
		log.Errorln(err, "getting balance for token", tokenAddr)
		return nil, err
	}
	if allowance.Cmp(balance) == 1 {
		return allowance, nil
	}
	return balance, nil
}

func RequestStandardDeposit(
	currentBlock uint64, myAddr ctype.Addr,
	initializer *entity.PaymentChannelInitializer, ospToOsp bool,
	ocem *pem.OpenChannelEventMessage) (int, error) {
	noPolicyAllowed := 0
	token := initializer.GetInitDistribution().GetToken()
	distribution := initializer.GetInitDistribution().GetDistribution()
	tokenAddr := utils.GetTokenAddrStr(token)
	// figure out peer and deposit.
	depositMap := getDepositMap(distribution)
	var peerAddr ctype.Addr
	for k := range depositMap {
		if k != myAddr {
			peerAddr = k
			break
		}
	}
	myDeposit := depositMap[myAddr]
	peerDeposit := depositMap[peerAddr]
	var config *rtconfig.StandardConfig
	config = nil
	// determine to use config of osp-osp policy or osp-client policy
	if ospToOsp {
		configs := rtconfig.GetOspToOspOpenConfigs().GetConfigs()
		// Osp open channel Whitelisted? Nil config means allow all.
		if configs != nil {
			ospToOspConfigs, ok := configs[ctype.Addr2Hex(peerAddr)]
			if ok {
				tokenConfigs := ospToOspConfigs.GetTokensConfig()
				if tokenConfigs != nil {
					config, _ = tokenConfigs[tokenAddr]
				}
			}
		}
	}
	// Two cases to use StandardConfigs.
	// 1. osp-client open channel
	// 2. osp-osp fallback
	tokenNotAllowed := errors.New("Token not allowed")
	deadlineOutOfRange := errors.New("Deadline out of range")
	requesterDespositZero := errors.New("Requester deposit zero")
	depositViolateRatio := errors.New("deposit violates ratio")
	rtUnparsable := errors.New("rt unparsable")
	depositOutOfRange := errors.New("deposit out of range")
	disputeTimeoutOutOfRange := errors.New("Dispute timeout out of range")

	if config == nil {
		// client open channel Whitelisted? No config means allow all.
		if rtconfig.GetStandardConfigs() == nil {
			return AllowStandardOpenChannel, nil
		}
		tokenAllowed := false
		config, tokenAllowed = rtconfig.GetStandardConfigs().GetConfig()[tokenAddr]
		if !tokenAllowed {
			return noPolicyAllowed, tokenNotAllowed
		}
	}
	// Deadline not big.
	deadline := initializer.GetOpenDeadline()
	if deadline > config.GetMaxDeadlineDelta()+currentBlock {
		log.Errorln("deadline too late")
		return noPolicyAllowed, deadlineOutOfRange
	}
	if deadline < config.GetMinDeadlineDelta()+currentBlock {
		log.Errorln("deadline too early")
		return noPolicyAllowed, deadlineOutOfRange
	}

	// OSP deposit no bigger than peer deposit.
	requiredMatchRatio := config.GetMatchingRatio()
	if requiredMatchRatio != 0.0 {
		if peerDeposit.Cmp(big.NewInt(0)) == 0 {
			log.Errorf(
				"peer deposits zero, peer:0x%x, ospDeposit:%s, peerDeposit:%s, required ratio: %f",
				peerAddr, myDeposit.String(), peerDeposit.String(), requiredMatchRatio)
			return noPolicyAllowed, requesterDespositZero
		}
		peerDepositFloat := big.NewFloat(float64(peerDeposit.Int64()))
		myDepositFloat := big.NewFloat(float64(myDeposit.Int64()))
		ratio := big.NewFloat(0.0).Quo(myDepositFloat, peerDepositFloat)
		if ratio.Cmp(big.NewFloat(float64(requiredMatchRatio))) == 1 {
			log.Errorf(
				"Asking me depositing more than required ratio, peer:0x%x, ospDeposit:%s, peerDeposit:%s, required ratio: %f",
				peerAddr, myDeposit.String(), peerDeposit.String(), requiredMatchRatio)
			return noPolicyAllowed, depositViolateRatio
		}
	} else {
		// require 1:1 by default
		if myDeposit.Cmp(peerDeposit) != 0 {
			log.Errorf(
				"Asking osp depositing unequal to peer, peer:0x%x, ospDeposit:%s, peerDeposit:%s",
				peerAddr, myDeposit.String(), peerDeposit.String())
			return noPolicyAllowed, depositViolateRatio
		}
	}
	minDeposit, setOK := new(big.Int).SetString(config.GetMinDeposit(), 10)
	if !setOK {
		log.Errorln("can't parse mindeposit:", config.GetMinDeposit())
		return noPolicyAllowed, rtUnparsable
	}
	// osp needs to have a minimum deposit to prevent immediate auto-refill after open the channel.
	if myDeposit.Cmp(minDeposit) == -1 {
		log.Errorf(
			"Osp deposit smaller than mindeposit peer:0x%x, ospDeposit:%s, minDeposit:%s",
			peerAddr, myDeposit.String(), minDeposit.String())
		return noPolicyAllowed, depositOutOfRange
	}
	// peer deposit no bigger than maxDeposit
	maxDeposit, setOK := new(big.Int).SetString(config.GetMaxDeposit(), 10)
	if !setOK {
		log.Errorln("can't parse maxdeposit:", config.GetMaxDeposit())
		return noPolicyAllowed, rtUnparsable
	}
	if peerDeposit.Cmp(maxDeposit) == 1 {
		log.Errorf(
			"peer deposit is more than maxDeposit: %s, peer:0x%x, peerDeposit:%s",
			maxDeposit.String(), peerAddr, peerDeposit.String())
		return noPolicyAllowed, depositOutOfRange
	}
	if initializer.DisputeTimeout > rtconfig.GetMaxDisputeTimeout() || initializer.DisputeTimeout < rtconfig.GetMinDisputeTimeout() {
		return noPolicyAllowed, disputeTimeoutOutOfRange
	}
	return AllowStandardOpenChannel, nil
}
