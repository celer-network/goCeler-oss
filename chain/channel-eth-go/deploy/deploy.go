package deploy

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/celer-network/goCeler/chain/channel-eth-go/balancelimit"
	"github.com/celer-network/goCeler/chain/channel-eth-go/channel"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ethpool"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledgerstruct"
	"github.com/celer-network/goCeler/chain/channel-eth-go/migrate"
	"github.com/celer-network/goCeler/chain/channel-eth-go/operation"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/routerregistry"
	"github.com/celer-network/goCeler/chain/channel-eth-go/virtresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/wallet"
	"github.com/celer-network/goutils/log"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type CelerChannelAddrBundle struct {
	BalanceLimitAddr  common.Address
	LedgerChannelAddr common.Address
	EthPoolAddr       common.Address
	CelerLedgerAddr   common.Address
	OperationAddr     common.Address
	MigrateAddr       common.Address
	PayRegistryAddr   common.Address
	PayResolverAddr   common.Address
	VirtResolverAddr  common.Address
	LedgerStructAddr  common.Address
	CelerWalletAddr   common.Address
}

// DeployRouterRegistry deploys router registry contract
func DeployRouterRegistry(
	ctx context.Context,
	auth *bind.TransactOpts,
	conn *ethclient.Client,
	blockDelay uint64) common.Address {
	log.Infoln("Deploying RouterRegistry contract...")
	routerRegistryAddr, tx, _, err := routerregistry.DeployRouterRegistry(auth, conn)
	if err != nil {
		log.Fatalf("Failed to deploy RouterRegistry contract: %v", err)
	}
	receipt, err := WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined RouterRegistry: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed RouterRegistry contract at 0x%x\n", routerRegistryAddr)
	return routerRegistryAddr
}

// DeployAll cChannel related contracts.
func DeployAll(
	auth *bind.TransactOpts,
	conn *ethclient.Client,
	ctx context.Context,
	blockDelay uint64) CelerChannelAddrBundle {
	/********** contracts without need of linking **********/
	// Deploy VirtContractResolver contract
	log.Infoln("Deploying VirtContractResolver contract...")
	virtresolverAddr, tx, _, err := virtresolver.DeployVirtContractResolver(auth, conn)
	if err != nil {
		log.Fatalf("Failed to deploy VirtContractResolver contract: %v", err)
	}
	receipt, err := WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined VirtContractResolver: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed VirtContractResolver contract at 0x%x\n", virtresolverAddr)

	// Deploy EthPool contract
	log.Infoln("Deploying EthPool contract...")
	ethPoolAddr, tx, _, err := ethpool.DeployEthPool(auth, conn)
	if err != nil {
		log.Fatalf("Failed to deploy EthPool contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined EthPool: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed EthPool contract at 0x%x\n", ethPoolAddr)

	// Deploy PayRegistry contract
	log.Infoln("Deploying PayRegistry contract...")
	payRegistryAddr, tx, _, err := payregistry.DeployPayRegistry(auth, conn)
	if err != nil {
		log.Fatalf("Failed to deploy PayRegistry contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined PayRegistry: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed PayRegistry contract at 0x%x\n", payRegistryAddr)

	// Deploy PayResolver contract
	log.Infoln("Deploying PayResolver contract...")
	payResolverAddr, tx, _, err := payresolver.DeployPayResolver(auth, conn, payRegistryAddr, virtresolverAddr)
	if err != nil {
		log.Fatalf("Failed to deploy PayResolver contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined PayResolver: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed PayResolver contract at 0x%x\n", payResolverAddr)

	// Deploy CelerWallet contract
	log.Infoln("Deploying CelerWallet contract...")
	walletAddr, tx, _, err := wallet.DeployCelerWallet(auth, conn)
	if err != nil {
		log.Fatalf("Failed to deploy CelerWallet contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined CelerWallet: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed CelerWallet contract at 0x%x\n", walletAddr)

	// Deploy LedgerStruct contract
	log.Infoln("Deploying LedgerStruct contract...")
	ledgerstructAddr, tx, _, err := ledgerstruct.DeployLedgerStruct(auth, conn)
	if err != nil {
		log.Fatalf("Failed to deploy LedgerStruct contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined LedgerStruct: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed LedgerStruct contract at 0x%x\n", ledgerstructAddr)

	/********** contracts with need of linking **********/
	// Deploy LedgerChannel contract
	log.Infoln("Deploying LedgerChannel contract...")
	channelAddr, tx, _, err := DeployContractWithLinks(
		auth,
		conn,
		channel.LedgerChannelABI,
		channel.LedgerChannelBin,
		map[string]common.Address{"LedgerStruct": ledgerstructAddr},
	)
	if err != nil {
		log.Fatalf("Failed to deploy LedgerChannel contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined LedgerChannel: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed LedgerChannel contract at 0x%x\n", channelAddr)

	// Deploy LedgerBalanceLimit contract
	log.Infoln("Deploying LedgerBalanceLimit contract...")
	balancelimitAddr, tx, _, err := DeployContractWithLinks(
		auth,
		conn,
		balancelimit.LedgerBalanceLimitABI,
		balancelimit.LedgerBalanceLimitBin,
		map[string]common.Address{"LedgerStruct": ledgerstructAddr},
	)
	if err != nil {
		log.Fatalf("Failed to deploy LedgerBalanceLimit contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined LedgerBalanceLimit: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed LedgerBalanceLimit contract at 0x%x\n", balancelimitAddr)

	// Deploy LedgerOperation contract
	log.Infoln("Deploying LedgerOperation contract...")
	operationAddr, tx, _, err := DeployContractWithLinks(
		auth,
		conn,
		operation.LedgerOperationABI,
		operation.LedgerOperationBin,
		map[string]common.Address{"LedgerStruct": ledgerstructAddr, "LedgerChannel": channelAddr},
	)
	if err != nil {
		log.Fatalf("Failed to deploy LedgerOperation contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined LedgerOperation: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed LedgerOperation contract at 0x%x\n", operationAddr)

	// Deploy LedgerMigrate contract
	log.Infoln("Deploying LedgerMigrate contract...")
	migrateAddr, tx, _, err := DeployContractWithLinks(
		auth,
		conn,
		migrate.LedgerMigrateABI,
		migrate.LedgerMigrateBin,
		map[string]common.Address{
			"LedgerStruct":    ledgerstructAddr,
			"LedgerOperation": operationAddr,
			"LedgerChannel":   channelAddr,
		},
	)
	if err != nil {
		log.Fatalf("Failed to deploy LedgerMigrate contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined LedgerMigrate: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed LedgerMigrate contract at 0x%x\n", migrateAddr)

	// Deploy CelerLedger contract
	log.Infoln("Deploying CelerLedger contract...")
	ledgerAddr, tx, _, err := DeployContractWithLinks(
		auth,
		conn,
		ledger.CelerLedgerABI,
		ledger.CelerLedgerBin,
		map[string]common.Address{
			"LedgerStruct":       ledgerstructAddr,
			"LedgerOperation":    operationAddr,
			"LedgerChannel":      channelAddr,
			"LedgerBalanceLimit": balancelimitAddr,
			"LedgerMigrate":      migrateAddr,
		},
		ethPoolAddr,
		payRegistryAddr,
		walletAddr,
	)
	if err != nil {
		log.Fatalf("Failed to deploy CelerLedger contract: %v", err)
	}
	receipt, err = WaitMined(ctx, conn, tx, blockDelay)
	if err != nil {
		log.Fatalf("Failed to WaitMined CelerLedger: %v", err)
	}
	log.Infof("Transaction status: %x", receipt.Status)
	log.Infof("Deployed CelerLedger contract at 0x%x\n", ledgerAddr)

	// return addresses of deployed contracts
	return CelerChannelAddrBundle{
		BalanceLimitAddr:  balancelimitAddr,
		LedgerChannelAddr: channelAddr,
		EthPoolAddr:       ethPoolAddr,
		CelerLedgerAddr:   ledgerAddr,
		OperationAddr:     operationAddr,
		MigrateAddr:       migrateAddr,
		PayRegistryAddr:   payRegistryAddr,
		PayResolverAddr:   payResolverAddr,
		VirtResolverAddr:  virtresolverAddr,
		LedgerStructAddr:  ledgerstructAddr,
		CelerWalletAddr:   walletAddr,
	}
}

// The following two functions(ABILinkLibrary and DeployContractWithLinks) are modified based on: https://github.com/joincivil/go-common/blob/master/pkg/eth/utils.go
// ABILinkLibrary replaces references to a library
// with the actual addresses to those library contracts
func ABILinkLibrary(bin string, libraryName string, libraryAddress common.Address) string {
	libstr := fmt.Sprintf("_+%v_+", libraryName)
	libraryRexp := regexp.MustCompile(libstr)

	// Remove the 0x prefix from those addresses, just need the actual hex string
	cleanLibraryAddr := strings.Replace(libraryAddress.Hex(), "0x", "", -1)

	modifiedBin := libraryRexp.ReplaceAllString(bin, cleanLibraryAddr)

	return modifiedBin
}

// DeployContractWithLinks patches a contract bin with provided library addresses
func DeployContractWithLinks(
	auth *bind.TransactOpts,
	backend bind.ContractBackend,
	abiString string,
	bin string,
	libraries map[string]common.Address,
	params ...interface{},
) (common.Address, *types.Transaction, *bind.BoundContract, error) {

	for libraryName, libraryAddress := range libraries {
		bin = ABILinkLibrary(bin, libraryName, libraryAddress)
	}

	parsed, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return bind.DeployContract(auth, parsed, common.FromHex(bin), backend, params...)
}

func WaitMined(ctx context.Context, ec *ethclient.Client,
	tx *types.Transaction, blockDelay uint64) (*types.Receipt, error) {
	return WaitMinedWithTxHash(ctx, ec, tx.Hash().Hex(), blockDelay)
}

// WaitMined waits for tx to be mined on the blockchain
// It returns tx receipt when the tx has been mined and enough block confirmations have passed
func WaitMinedWithTxHash(ctx context.Context, ec *ethclient.Client,
	txHash string, blockDelay uint64) (*types.Receipt, error) {
	// an error possibly returned when a transaction is pending
	const missingFieldErr = "missing required field 'transactionHash' for Log"

	if ec == nil {
		return nil, errors.New("nil ethclient")
	}
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()
	// wait tx to be mined
	txHashBytes := common.HexToHash(txHash)
	for {
		receipt, rerr := ec.TransactionReceipt(ctx, txHashBytes)
		if rerr == nil {
			log.Infof("Transaction mined. Waiting for %d block confirmations", blockDelay)
			if blockDelay == 0 {
				return receipt, rerr
			}
			break
		} else if rerr == ethereum.NotFound || rerr.Error() == missingFieldErr {
			// Wait for the next round
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-queryTicker.C:
			}
		} else {
			return receipt, rerr
		}
	}
	// wait for enough block confirmations
	ddl := big.NewInt(0)
	latestBlockHeader, err := ec.HeaderByNumber(ctx, nil)
	if err == nil {
		ddl.Add(new(big.Int).SetUint64(blockDelay), latestBlockHeader.Number)
	}
	for {
		latestBlockHeader, err := ec.HeaderByNumber(ctx, nil)
		if err == nil && ddl.Cmp(latestBlockHeader.Number) < 0 {
			receipt, rerr := ec.TransactionReceipt(ctx, txHashBytes)
			if rerr == nil {
				log.Infoln("tx confirmed!")
				return receipt, rerr
			} else if rerr == ethereum.NotFound || rerr.Error() == missingFieldErr {
				return nil, errors.New("tx is dropped due to chain re-org")
			} else {
				return receipt, rerr
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}
