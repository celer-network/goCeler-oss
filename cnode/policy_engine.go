// Copyright 2018-2019 Celer Network

package cnode

import (
	"math/big"

	"github.com/celer-network/goCeler-oss/chain"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/rtconfig"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const (
	// Enum of policy allowed.
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

func getDepositCapacity(nodeConfig common.GlobalNodeConfig, tokenAddr string) (*big.Int, error) {
	conn := nodeConfig.GetEthConn()
	tokenAddrToCheck := ctype.Hex2Addr(tokenAddr)
	// ETH pool acts as a ERC20 for OSP. ETH capacity is on addr of Eth pool, not on OSP addr
	if tokenAddr == ctype.ZeroAddrHex {
		tokenAddrToCheck = nodeConfig.GetEthPoolAddr()
	}

	erc20Contract, err := chain.NewERC20(tokenAddrToCheck, conn)
	if err != nil {
		return ctype.ZeroBigInt, err
	}
	allowance, err := erc20Contract.Allowance(&bind.CallOpts{}, ctype.Hex2Addr(nodeConfig.GetOnChainAddr()), nodeConfig.GetLedgerContract().GetAddr())
	if err != nil {
		log.Errorln(err, "getting allowance for token", tokenAddr)
		return nil, err
	}
	balance, err := erc20Contract.BalanceOf(&bind.CallOpts{}, ctype.Hex2Addr(nodeConfig.GetOnChainAddr()))
	if err != nil {
		log.Errorln(err, "getting balance for token", tokenAddr)
		return nil, err
	}
	if allowance.Cmp(balance) == 1 {
		return allowance, nil
	}
	return balance, nil
}

func RequestStandardDeposit(currentBlock uint64, myAddr ethcommon.Address, initializer *entity.PaymentChannelInitializer, ospToOsp bool) int {
	noPolicyAllowed := 0
	token := initializer.GetInitDistribution().GetToken()
	distribution := initializer.GetInitDistribution().GetDistribution()
	tokenAddr := utils.GetTokenAddrStr(token)
	// figure out peer and deposit.
	depositMap := getDepositMap(distribution)
	var peerAddr ethcommon.Address
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
	if config == nil {
		// client open channel Whitelisted? No config means allow all.
		if rtconfig.GetStandardConfigs() == nil {
			return AllowStandardOpenChannel
		}
		tokenAllowed := false
		config, tokenAllowed = rtconfig.GetStandardConfigs().GetConfig()[tokenAddr]
		if !tokenAllowed {
			return noPolicyAllowed
		}
	}
	// Deadline not big.
	deadline := initializer.GetOpenDeadline()
	if deadline > config.GetMaxDeadlineDelta()+currentBlock {
		log.Errorln("deadline too big")
		return noPolicyAllowed
	}
	if deadline < config.GetMinDeadlineDelta()+currentBlock {
		log.Errorln("deadline too small")
		return noPolicyAllowed
	}

	// OSP deposit no bigger than peer deposit.
	requiredMatchRatio := config.GetMatchingRatio()
	if requiredMatchRatio != 0.0 {
		if peerDeposit.Cmp(big.NewInt(0)) == 0 {
			log.Errorf(
				"peer deposits zero, peer:0x%x, ospDeposit:%s, peerDeposit:%s, required ratio: %f",
				peerAddr, myDeposit.String(), peerDeposit.String(), requiredMatchRatio)
			return noPolicyAllowed
		}
		peerDepositFloat := big.NewFloat(float64(peerDeposit.Int64()))
		myDepositFloat := big.NewFloat(float64(myDeposit.Int64()))
		ratio := big.NewFloat(0.0).Quo(myDepositFloat, peerDepositFloat)
		if ratio.Cmp(big.NewFloat(float64(requiredMatchRatio))) == 1 {
			log.Errorf(
				"Asking me depositing more than required ratio, peer:0x%x, ospDeposit:%s, peerDeposit:%s, required ratio: %f",
				peerAddr, myDeposit.String(), peerDeposit.String(), requiredMatchRatio)
			return noPolicyAllowed
		}
	} else {
		// require 1:1 by default
		if myDeposit.Cmp(peerDeposit) != 0 {
			log.Errorf(
				"Asking osp depositing unequal to peer, peer:0x%x, ospDeposit:%s, peerDeposit:%s",
				peerAddr, myDeposit.String(), peerDeposit.String())
			return noPolicyAllowed
		}
	}
	minDeposit, setOK := new(big.Int).SetString(config.GetMinDeposit(), 10)
	if !setOK {
		log.Errorln("can't parse mindeposit:", config.GetMinDeposit())
		return noPolicyAllowed
	}
	if myDeposit.Cmp(minDeposit) == -1 {
		log.Errorf(
			"Osp deposit smaller than mindeposit peer:0x%x, ospDeposit:%s, minDeposit:%s",
			peerAddr, myDeposit.String(), minDeposit.String())
		return noPolicyAllowed
	}
	// peer deposit no bigger than maxDeposit
	maxDeposit, setOK := new(big.Int).SetString(config.GetMaxDeposit(), 10)
	if !setOK {
		log.Errorln("can't parse maxdeposit:", config.GetMaxDeposit())
		return noPolicyAllowed
	}
	if peerDeposit.Cmp(maxDeposit) == 1 {
		log.Errorf(
			"peer deposit is more than maxDeposit: %s, peer:0x%x, peerDeposit:%s",
			maxDeposit.String(), peerAddr, peerDeposit.String())
		return noPolicyAllowed
	}
	return AllowStandardOpenChannel
}
