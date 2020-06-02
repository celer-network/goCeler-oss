// Copyright 2018-2020 Celer Network

package app

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/virtresolver"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	proto "github.com/golang/protobuf/proto"
)

type AppChannel struct {
	Type           entity.ConditionType
	Nonce          uint64
	ByteCode       []byte       // only for virtual contract
	Constructor    []byte       // only for virtual contract
	Players        []ctype.Addr // only for deployed contract
	Session        [32]byte     // only for deployed contract
	DeployedAddr   ctype.Addr
	OnChainTimeout uint64
	Callback       common.StateCallback
	mu             sync.Mutex
	client         *AppClient
	cid            string
	callbackID     monitor.CallbackID
}

type AppClient struct {
	nodeConfig     common.GlobalNodeConfig
	transactor     *transactor.Transactor
	transactorPool *transactor.Pool
	monitorService intfs.MonitorService
	dal            *storage.DAL
	signer         common.Signer
	appChannels    map[string]*AppChannel
	cLock          sync.RWMutex
}

func NewAppClient(
	nodeConfig common.GlobalNodeConfig,
	transactor *transactor.Transactor,
	transactorPool *transactor.Pool,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	signer common.Signer,
) *AppClient {
	p := &AppClient{
		nodeConfig:     nodeConfig,
		transactor:     transactor,
		transactorPool: transactorPool,
		monitorService: monitorService,
		dal:            dal,
		signer:         signer,
		appChannels:    make(map[string]*AppChannel),
	}
	return p
}

func (a *AppChannel) setDeployedAddr(addr ctype.Addr) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.DeployedAddr = addr
}

func (a *AppChannel) getDeployedAddr() ctype.Addr {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.DeployedAddr
}

// onVirtualContractDeploy triggers OnDispute callback for an app based on a virtual contract
// when the contract is deployed
func (a *AppChannel) onVirtualContractDeploy(eLog *types.Log) (bool, error) {
	e := &virtresolver.VirtContractResolverDeploy{}
	virtResolver := a.client.nodeConfig.GetVirtResolverContract().(*chain.BoundContract)
	err := virtResolver.ParseEvent(event.Deploy, *eLog, e)
	if err != nil {
		log.Error(err)
		return false, err
	}
	if ctype.Bytes2Hex(e.VirtAddr[:]) == a.cid {
		a.Callback.OnDispute(0) // seqNum = 0 implies the virtual contract is deployed
		return true, nil
	}
	return false, nil
}

// onDeployedContractSettle triggers OnDispute callback for an app based on a deployed multisession
// contract when the session is created on chain by IntendSettle
func (a *AppChannel) onDeployedContractSettle(eLog *types.Log) (bool, error) {
	e := &IMultiSessionIntendSettle{}
	deployedAdr := a.getDeployedAddr()
	contract, err := chain.NewBoundContract(
		a.client.nodeConfig.GetEthConn(), deployedAdr, IMultiSessionABI)
	if err != nil {
		log.Error(err)
		return false, err
	}
	err = contract.ParseEvent(event.IntendSettle, *eLog, e)
	if err != nil {
		log.Error(err)
		return false, err
	}
	if bytes.Equal(a.Session[:], e.Session[:]) {
		a.Callback.OnDispute(int(e.Seq.Int64()))
		return true, nil
	}
	return false, nil
}

func (c *AppClient) PutAppChannel(cid string, appChannel *AppChannel) {
	c.cLock.Lock()
	defer c.cLock.Unlock()
	c.appChannels[cid] = appChannel
}

func (c *AppClient) GetAppChannel(cid string) *AppChannel {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	return c.appChannels[cid]
}

func (c *AppClient) DeleteAppChannel(cid string) {
	c.cLock.Lock()
	defer c.cLock.Unlock()
	appChannel := c.appChannels[cid]
	if appChannel != nil {
		c.monitorService.RemoveEvent(appChannel.callbackID)
		delete(c.appChannels, cid)
	}
}

func (c *AppClient) NewAppChannelOnVirtualContract(
	byteCode []byte,
	constructor []byte,
	nonce uint64,
	onchainTimeout uint64,
	sc common.StateCallback) (string, error) {

	cid := ctype.Bytes2Hex(GetVirtualAddress(byteCode, constructor, nonce))
	appChannel := &AppChannel{
		Type:           entity.ConditionType_VIRTUAL_CONTRACT,
		Nonce:          nonce,
		ByteCode:       byteCode,
		Constructor:    constructor,
		DeployedAddr:   ctype.ZeroAddr,
		OnChainTimeout: onchainTimeout,
		Callback:       sc,
		client:         c,
		cid:            cid,
	}
	c.PutAppChannel(cid, appChannel)

	_, err := c.monitorService.Monitor(
		event.Deploy,
		c.nodeConfig.GetVirtResolverContract(),
		c.monitorService.GetCurrentBlockNumber(),
		nil,
		true, /* quickCatch */
		false,
		func(id monitor.CallbackID, eLog types.Log) {
			hit, _ := appChannel.onVirtualContractDeploy(&eLog)
			if hit {
				c.monitorService.RemoveEvent(id)
			}
		})
	if err != nil {
		log.Error(err)
	}
	return cid, err
}

func (c *AppClient) NewAppChannelOnDeployedContract(
	contractAddr ctype.Addr,
	nonce uint64,
	players []ctype.Addr,
	onchainTimeout uint64,
	sc common.StateCallback) (string, error) {

	players = SortPlayers(players)
	session, err := c.getSessionID(contractAddr, nonce, players)
	if err != nil {
		return "", err
	}
	cid := ctype.Bytes2Hex(session[:])
	appChannel := &AppChannel{
		Type:           entity.ConditionType_DEPLOYED_CONTRACT,
		Nonce:          nonce,
		Players:        players,
		Session:        session,
		DeployedAddr:   contractAddr,
		OnChainTimeout: onchainTimeout,
		Callback:       sc,
		client:         c,
		cid:            cid,
	}
	c.PutAppChannel(cid, appChannel)

	contract, err := chain.NewBoundContract(
		c.nodeConfig.GetEthConn(), contractAddr, IMultiSessionABI)
	if err != nil {
		log.Error(err)
		return cid, err
	}
	callbackID, err := c.monitorService.Monitor(
		event.IntendSettle,
		contract,
		c.monitorService.GetCurrentBlockNumber(),
		nil,
		true, /* quickCatch */
		false,
		func(id monitor.CallbackID, eLog types.Log) {
			hit, _ := appChannel.onDeployedContractSettle(&eLog)
			if hit {
				c.monitorService.RemoveEvent(id)
			}
		})
	if err != nil {
		log.Error(err)
	}
	appChannel.callbackID = callbackID
	return cid, err
}

func (c *AppClient) SettleAppChannel(cid string, stateproof []byte) error {
	log.Infoln("Settle app channel", cid)
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return fmt.Errorf("SettleAppChannel error: app channel not found")
	}

	if err := c.deployIfNeeded(appChannel); err != nil {
		return err
	}

	err := c.intendSettle(appChannel, stateproof)
	return err
}

func (c *AppClient) GetAppChannelDeployedAddr(cid string) (ctype.Addr, error) {
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return ctype.ZeroAddr, fmt.Errorf("app channel not found")
	}
	addr := appChannel.getDeployedAddr()
	if addr != (ctype.ZeroAddr) || appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		return addr, nil
	}
	virtAddr := GetVirtualAddress(appChannel.ByteCode, appChannel.Constructor, appChannel.Nonce)
	deployed, addr, err := c.isDeployed(virtAddr)
	if err != nil {
		return addr, err
	}
	if !deployed {
		return ctype.ZeroAddr, fmt.Errorf("virtual contract not deployed")
	}
	appChannel.setDeployedAddr(addr)
	return addr, nil
}

// GetBooleanOutcome returns contract isFinalized and getOutcome
func (c *AppClient) GetBooleanOutcome(cid string, query []byte) (bool, bool, error) {
	finalized := false
	result := false
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return false, false, fmt.Errorf("GetBooleanOutcome error: app channel not found")
	}
	if err := c.deployIfNeeded(appChannel); err != nil {
		return false, false, err
	}
	deployedAddr := appChannel.getDeployedAddr()
	contract, err := NewIBooleanOutcomeCaller(
		deployedAddr, c.transactorPool.ContractCaller())
	if err != nil {
		return false, false, fmt.Errorf("GetBooleanOutcome error: %w", err)
	}
	if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
		finalized, err = contract.IsFinalized(&bind.CallOpts{}, nil)
		if err != nil {
			return false, false, fmt.Errorf("contract IsFinalized error: %w", err)
		}
		result, err = contract.GetOutcome(&bind.CallOpts{}, query)
		if err != nil {
			return false, false, fmt.Errorf("contract GetResult error: %w", err)
		}
	} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		finalized, err = contract.IsFinalized(&bind.CallOpts{}, appChannel.Session[:])
		if err != nil {
			return false, false, fmt.Errorf("contract IsFinalized error: %w", err)
		}
		sessionQuery := &SessionQuery{
			Session: appChannel.Session[:],
			Query:   query,
		}
		seralizedSessionQuery, err2 := proto.Marshal(sessionQuery)
		if err2 != nil {
			return false, false, fmt.Errorf("contract GetResult error: %w", err2)
		}
		result, err = contract.GetOutcome(&bind.CallOpts{}, seralizedSessionQuery)
	}
	return finalized, result, err
}

func (c *AppClient) ApplyAction(cid string, action []byte) error {
	log.Infoln("Apply action to app channel", cid)
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return fmt.Errorf("ApplyAction error: app channel not found")
	}

	receipt, err := c.transactor.TransactWaitMined(
		"ApplyAction",
		&transactor.TxConfig{QuickCatch: true},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			addr := appChannel.getDeployedAddr()
			if addr == (ctype.ZeroAddr) {
				return nil, fmt.Errorf("FinalizeOnActionTimeout error: app channel not deployed")
			}
			if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
				contract, err2 := NewISingleSessionTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.ApplyAction(opts, action)
			} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
				contract, err2 := NewIMultiSessionTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.ApplyAction(opts, appChannel.Session, action)
			}
			return nil, errors.New("ApplyAction failed: invalid app channel type")
		})
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("ApplyAction transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (c *AppClient) FinalizeAppChannelOnActionTimeout(cid string) error {
	log.Infoln("Finalize on action timeout on app channel", cid)
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return fmt.Errorf("FinalizeOnActionTimeout error: app channel not found")
	}

	receipt, err := c.transactor.TransactWaitMined(
		"FinalizeOnActionTimeout",
		&transactor.TxConfig{QuickCatch: true},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			addr := appChannel.getDeployedAddr()
			if addr == (ctype.ZeroAddr) {
				return nil, fmt.Errorf("FinalizeOnActionTimeout error: app channel not deployed")
			}
			if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
				contract, err2 := NewISingleSessionTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.FinalizeOnActionTimeout(opts)
			} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
				contract, err2 := NewIMultiSessionTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.FinalizeOnActionTimeout(opts, appChannel.Session)
			}
			return nil, errors.New("FinalizeOnActionTimeout failed: invalid app channel type")
		})
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("FinalizeOnActionTimeout transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (c *AppClient) GetAppChannelSettleFinalizedTime(cid string) (uint64, error) {
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return 0, fmt.Errorf("GetSettleFinalizedTime error: app channel not found")
	}
	deployedAddr := appChannel.getDeployedAddr()
	if deployedAddr == (ctype.ZeroAddr) {
		return 0, fmt.Errorf("GetSettleFinalizedTime error: app channel not deployed")
	}
	if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
		contract, err := NewISingleSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		blkNum, err := contract.GetSettleFinalizedTime(&bind.CallOpts{})
		if err != nil {
			return 0, err
		}
		if blkNum == nil {
			return 0, fmt.Errorf("GetActionDeadline failed, nil blkNum")
		}
		return blkNum.Uint64(), nil
	} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		contract, err := NewIMultiSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		blkNum, err := contract.GetSettleFinalizedTime(&bind.CallOpts{}, appChannel.Session)
		if err != nil {
			return 0, err
		}
		if blkNum == nil {
			return 0, fmt.Errorf("GetActionDeadline failed, nil blkNum")
		}
		return blkNum.Uint64(), nil
	}
	return 0, errors.New("GetSettleFinalizedTime failed: invalid app channel type")
}

func (c *AppClient) GetAppChannelActionDeadline(cid string) (uint64, error) {
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return 0, fmt.Errorf("GetActionDeadline error: app channel not found")
	}
	deployedAddr := appChannel.getDeployedAddr()
	if deployedAddr == (ctype.ZeroAddr) {
		return 0, fmt.Errorf("GetActionDeadline error: app channel not deployed")
	}
	if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
		contract, err := NewISingleSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		blkNum, err := contract.GetActionDeadline(&bind.CallOpts{})
		if err != nil {
			return 0, err
		}
		if blkNum == nil {
			return 0, fmt.Errorf("GetActionDeadline failed, nil blkNum")
		}
		return blkNum.Uint64(), nil
	} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		contract, err := NewIMultiSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		blkNum, err := contract.GetActionDeadline(&bind.CallOpts{}, appChannel.Session)
		if err != nil {
			return 0, err
		}
		if blkNum == nil {
			return 0, fmt.Errorf("GetActionDeadline failed, nil blkNum")
		}
		return blkNum.Uint64(), nil
	}
	return 0, errors.New("GetActionDeadline failed: invalid app channel type")
}

func (c *AppClient) GetAppChannelSeqNum(cid string) (uint64, error) {
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return 0, fmt.Errorf("GetSeqNum error: app channel not found")
	}
	deployedAddr := appChannel.getDeployedAddr()
	if deployedAddr == (ctype.ZeroAddr) {
		return 0, fmt.Errorf("GetSeqNum error: app channel not deployed")
	}
	if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
		contract, err := NewISingleSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		seq, err := contract.GetSeqNum(&bind.CallOpts{})
		if err != nil {
			return 0, err
		}
		if seq == nil {
			return 0, fmt.Errorf("GetSeqNum failed, nil seq")
		}
		return seq.Uint64(), nil
	} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		contract, err := NewIMultiSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		seq, err := contract.GetSeqNum(&bind.CallOpts{}, appChannel.Session)
		if err != nil {
			return 0, err
		}
		if seq == nil {
			return 0, fmt.Errorf("GetSeqNum failed, nil seq")
		}
		return seq.Uint64(), nil
	}
	return 0, errors.New("GetSeqNum failed: invalid app channel type")
}

func (c *AppClient) GetAppChannelStatus(cid string) (uint8, error) {
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return 0, fmt.Errorf("GetStatus error: app channel not found")
	}
	deployedAddr := appChannel.getDeployedAddr()
	if deployedAddr == (ctype.ZeroAddr) {
		return 0, fmt.Errorf("GetStatus error: app channel not deployed")
	}
	if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
		contract, err := NewISingleSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		return contract.GetStatus(&bind.CallOpts{})
	} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		contract, err := NewIMultiSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return 0, err
		}
		return contract.GetStatus(&bind.CallOpts{}, appChannel.Session)
	}
	return 0, errors.New("GetStatus failed: invalid app channel type")
}

func (c *AppClient) GetAppChannelState(cid string, key *big.Int) ([]byte, error) {
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return nil, fmt.Errorf("GetState error: app channel not found")
	}
	deployedAddr := appChannel.getDeployedAddr()
	if deployedAddr == (ctype.ZeroAddr) {
		return nil, fmt.Errorf("GetState error: app channel not deployed")
	}
	if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
		contract, err := NewISingleSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return nil, err
		}
		return contract.GetState(&bind.CallOpts{}, key)
	} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		contract, err := NewIMultiSessionCaller(deployedAddr, c.transactorPool.ContractCaller())
		if err != nil {
			return nil, err
		}
		return contract.GetState(&bind.CallOpts{}, appChannel.Session, key)
	}
	return nil, errors.New("GetState failed: invalid app channel type")
}

// SignAppState returns 1: proto serialized app state, 2: signature, 3: error
func (c *AppClient) SignAppState(cid string, seqNum uint64, state []byte) ([]byte, []byte, error) {
	appChannel := c.GetAppChannel(cid)
	if appChannel == nil {
		return nil, nil, fmt.Errorf("SignAppState error: app channel not found")
	}
	appStateBytes, err := EncodeAppState(appChannel.Nonce, seqNum, state, appChannel.OnChainTimeout)
	if err != nil {
		return nil, appStateBytes, err
	}
	sig, err := c.signer.SignEthMessage(appStateBytes)
	if err != nil {
		return nil, appStateBytes, err
	}
	return appStateBytes, sig, nil
}

func (c *AppClient) SettleBySigTimeout(gcid string, oracleProof []byte) error {
	log.Infoln("Settle by signature timeout", gcid)
	appChannel := c.GetAppChannel(gcid)
	if appChannel == nil {
		return fmt.Errorf("SettleBySigTimeout error: app channel not found")
	}

	receipt, err := c.transactor.TransactWaitMined(
		"SettleBySigTimeout",
		&transactor.TxConfig{QuickCatch: true},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			addr := appChannel.getDeployedAddr()
			if addr == (ctype.ZeroAddr) {
				return nil, fmt.Errorf("FinalizeOnActionTimeout error: app channel not deployed")
			}
			if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
				contract, err2 := NewISingleSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.SettleBySigTimeout(opts, oracleProof)
			} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
				contract, err2 := NewIMultiSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}

				return contract.SettleBySigTimeout(opts, oracleProof)
			}
			return nil, errors.New("SettleBySigTimeout failed: invalid app channel type")
		})
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("SettleBySigTimeout transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (c *AppClient) SettleByMoveTimeout(gcid string, oracleProof []byte) error {
	log.Infoln("Settle by movement timeout", gcid)
	appChannel := c.GetAppChannel(gcid)
	if appChannel == nil {
		return fmt.Errorf("SettleByMoveTimeout error: app channel not found")
	}

	receipt, err := c.transactor.TransactWaitMined(
		"SettleByMoveTimeout",
		&transactor.TxConfig{QuickCatch: true},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			addr := appChannel.getDeployedAddr()
			if addr == (ctype.ZeroAddr) {
				return nil, fmt.Errorf("FinalizeOnActionTimeout error: app channel not deployed")
			}
			if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
				contract, err2 := NewISingleSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.SettleByMoveTimeout(opts, oracleProof)
			} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
				contract, err2 := NewIMultiSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.SettleByMoveTimeout(opts, oracleProof)
			}
			return nil, errors.New("SettleByMoveTimeout failed: invalid app channel type")
		})
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("SettleByMoveTimeout transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (c *AppClient) SettleByInvalidTurn(gcid string, oracleProof []byte, cosignedStateProof []byte) error {
	log.Infoln("Settle by invalid turn", gcid)
	appChannel := c.GetAppChannel(gcid)
	if appChannel == nil {
		return fmt.Errorf("SettleByInvalidTurn error: app channel not found")
	}

	receipt, err := c.transactor.TransactWaitMined(
		"SettleByInvalidTurn",
		&transactor.TxConfig{QuickCatch: true},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			addr := appChannel.getDeployedAddr()
			if addr == (ctype.ZeroAddr) {
				return nil, fmt.Errorf("FinalizeOnActionTimeout error: app channel not deployed")
			}
			if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
				contract, err2 := NewISingleSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.SettleByInvalidTurn(opts, oracleProof, cosignedStateProof)
			} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
				contract, err2 := NewIMultiSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.SettleByInvalidTurn(opts, oracleProof, cosignedStateProof)
			}
			return nil, errors.New("SettleByInvalidTurn failed: invalid app channel type")
		})
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("SettleByInvalidTurn transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (c *AppClient) SettleByInvalidState(gcid string, oracleProof []byte, cosignedStateProof []byte) error {
	log.Infoln("Settle by invalid state", gcid)
	appChannel := c.GetAppChannel(gcid)
	if appChannel == nil {
		return fmt.Errorf("SettleByInvalidState error: app channel not found")
	}

	receipt, err := c.transactor.TransactWaitMined(
		"SettleByInvalidState",
		&transactor.TxConfig{QuickCatch: true},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			addr := appChannel.getDeployedAddr()
			if addr == (ctype.ZeroAddr) {
				return nil, fmt.Errorf("FinalizeOnActionTimeout error: app channel not deployed")
			}
			if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT {
				contract, err2 := NewISingleSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.SettleByInvalidState(opts, oracleProof, cosignedStateProof)
			} else if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
				contract, err2 := NewIMultiSessionWithOracleTransactor(addr, transactor)
				if err2 != nil {
					return nil, err2
				}
				return contract.SettleByInvalidState(opts, oracleProof, cosignedStateProof)
			}
			return nil, errors.New("SettleByInvalidState failed: invalid app channel type")
		})
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("SettleByInvalidState transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (c *AppClient) deployIfNeeded(appChannel *AppChannel) error {
	deployedAddr := appChannel.getDeployedAddr()
	// Deploy virtual contract
	if appChannel.Type == entity.ConditionType_VIRTUAL_CONTRACT && deployedAddr == (ctype.ZeroAddr) {
		deployedAddr, err :=
			c.deployVirtualContract(appChannel.Nonce, appChannel.ByteCode, appChannel.Constructor)
		if err != nil {
			log.Error("virtual contract not deployed")
			return err
		}
		appChannel.setDeployedAddr(deployedAddr)
	}
	return nil
}

func (c *AppClient) deployVirtualContract(
	nonce uint64, byteCode []byte, constructor []byte) (ctype.Addr, error) {

	virtResolverContract := c.nodeConfig.GetVirtResolverContract()
	virtAddr := GetVirtualAddress(byteCode, constructor, nonce)
	deployed, addr, err := c.isDeployed(virtAddr)
	if err != nil {
		log.Error(err)
		return ctype.ZeroAddr, err
	}
	if deployed {
		return addr, nil
	}
	log.Debugln("deploying virtual contract...")
	codeWithCons := append(byteCode, constructor...)

	receipt, err := c.transactorPool.SubmitWaitMined(
		"deploy virtual contract",
		&transactor.TxConfig{QuickCatch: true, GasLimit: 4000000},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 :=
				virtresolver.NewVirtContractResolverTransactor(virtResolverContract.GetAddr(), transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.Deploy(opts, codeWithCons, new(big.Int).SetUint64(nonce))
		})
	if err != nil {
		log.Errorf("deploy virtual contract tx %x error %s", receipt.TxHash, err)
		return ctype.ZeroAddr, err
	}
	deployed, addr, err = c.isDeployed(virtAddr)
	if err != nil {
		return ctype.ZeroAddr, err
	}
	if deployed {
		log.Debugln("deployed virtual contract at", ctype.Addr2Hex(addr))
		return addr, nil
	}
	return ctype.ZeroAddr, fmt.Errorf("virtual contract not deployed")
}

// isDeployed checks if the given virtual address has been deployed on-chain
// if yes, also returns the deployment address
func (c *AppClient) isDeployed(virtAddr []byte) (bool, ctype.Addr, error) {
	contract, err := virtresolver.NewVirtContractResolverCaller(
		c.nodeConfig.GetVirtResolverContract().GetAddr(), c.transactorPool.ContractCaller())
	if err != nil {
		return false, ctype.ZeroAddr, err
	}
	var virt [32]byte
	copy(virt[:], virtAddr[:])
	deployedAddr, err := contract.Resolve(&bind.CallOpts{}, virt)
	if deployedAddr == (ctype.ZeroAddr) {
		return false, deployedAddr, nil
	}
	return true, deployedAddr, nil
}

func (c *AppClient) intendSettle(appChannel *AppChannel, stateproof []byte) error {
	var err error
	if appChannel.Type == entity.ConditionType_DEPLOYED_CONTRACT {
		stateproof, err = SigSortedAppStateProof(stateproof)
		if err != nil {
			return err
		}
	}
	_, err = c.transactorPool.SubmitWaitMined(
		"intend settle app channel",
		&transactor.TxConfig{QuickCatch: true},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			addr := appChannel.getDeployedAddr()
			// intendSettle API for SingleSession and MultiSession contracts are same
			contract, err2 := NewISingleSessionTransactor(addr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.IntendSettle(opts, stateproof)
		})
	if err != nil {
		log.Errorln("intend settle app channel tx error", err)
	}
	return err
}

func (c *AppClient) getSessionID(
	contractAddr ctype.Addr, nonce uint64, players []ctype.Addr) ([32]byte, error) {
	contract, err := NewIMultiSessionCaller(contractAddr, c.transactorPool.ContractCaller())
	if err != nil {
		var s [32]byte
		return s, err
	}
	session, err := contract.GetSessionID(&bind.CallOpts{}, new(big.Int).SetUint64(nonce), players)
	return session, err
}

func GetVirtualAddress(byteCode []byte, constructor []byte, nonce uint64) []byte {
	codeWithCons := append(byteCode, constructor...)
	return crypto.Keccak256(codeWithCons, utils.Pad(new(big.Int).SetUint64(nonce).Bytes(), 32))
}
