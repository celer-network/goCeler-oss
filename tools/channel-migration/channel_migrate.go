// Copyright 2019-2020 Celer Network

/**
 * This tool is used for submitting onchain transactions for channel migration
 * with low gas cost. It submits limited onchain transactions in a way to avoid
 * bringing congestion to the ethereum. The whole process could be abstracted
 * into following steps:
 * Loop 1. Fetch a bunch of onchain migration information from database.
 * Loop 2. For each channel migration transaction, query ethereum client for suggested
 *    gas price util finding a gas price within limitation.
 * 3. Submit each channel migration transaction and wait to be mined, then update
 *    channel migration state in database.
 *
 * SUGGESTION: Before running this tool, user should check the current ethereum clog situation
 * online. If the ethereum network is good, then run this tool. If the ethereum is clogging or
 * the suggested gas price is high(maybe around 7 gwei), then find another time to run this tool.
 */

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/migrate"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/proto"
)

var (
	profile   = flag.String("profile", "", "Path to profile json file")
	ks        = flag.String("ks", "", "Path to keystore json file")
	password  = flag.String("pw", "", "keystore file's password")
	storesql  = flag.String("storesql", "", "sql database URL")
	storedir  = flag.String("storedir", "", "local database directory")
	chanLimit = flag.Int("limit", 50, "limits of channel number per migration")
	maxGas    = flag.Int("maxgas", 4, "maximum gas price allowed in gwei")
	blkdelay  = flag.Int("blkdelay", 0, "block delay for wait mined")
)

const (
	gasPriceQueryInterval = 30 * time.Second // estimated 2 blocks production time

	waitMinedTimeout    = 200 * time.Second
	findGasPriceTimeout = 180 * time.Second

	gasPriceTolerance uint64 = 2 // 2 Gwei gas price tolerance

	txSucceeded int = 0
	txFailed    int = 1
	txExpired   int = 2
	txCancelled int = 3
)

func main() {
	flag.Parse()
	if *profile == "" {
		log.Fatalln("Please specify the profile file path")
	}
	if *ks == "" {
		log.Fatalln("Please specify the keysotre file path")
	}
	if *storesql == "" && *storedir == "" {
		log.Fatalln("Please specify the database path")
	}
	if *maxGas <= 1 {
		log.Fatalln("Gas price limit is too low:", *maxGas)
	}
	if *chanLimit > 100 {
		log.Fatalln("Channel number are too big, please set limit below 100")
	}

	cp := common.ParseProfile(*profile)
	overrideProfile(cp)
	config.SetGlobalConfigFromProfile(cp)

	dal := toolsetup.NewDAL(cp)
	client, err := ethclient.Dial(cp.ETHInstance)
	if err != nil {
		log.Fatalln(err)
	}

	ksFile, err := os.Open(*ks)
	if err != nil {
		log.Fatalln(err)
	}
	auth, err := bind.NewTransactor(ksFile, *password)
	if err != nil {
		log.Fatalln(err)
	}
	ksFile.Close()

	latestLedger := ctype.Hex2Addr(cp.LedgerAddr)
	ledgerContract, err := ledger.NewCelerLedger(latestLedger, client)
	if err != nil {
		log.Fatalln(err)
	}

	reqs, deadlines, err := dal.GetChanMigrationReqByLedgerAndStateWithLimit(latestLedger, migrate.MigrationStateInitialized, *chanLimit)
	if err != nil {
		log.Fatalln(err)
	}

	succeed := 0 // count to record successful mined transactions
	failed := 0  // count to record failed migration request
	expired := 0 // count to record expired migration request
	total := len(reqs)
	if total == 0 {
		log.Infoln("no channel needs to be upgraded")
		return
	}
	defer func(succeed, expired, failed, total *int) {
		log.Infof("Migration done! %d succeed, %d expired, %d failed, %d undone out of total %d channels",
			*succeed, *expired, *failed, (*total - *succeed - *expired - *failed), *total)
	}(&succeed, &expired, &failed, &total)

	for cid, req := range reqs {
		res := handleSingleOnchainTx(cid, req, deadlines, auth, client, ledgerContract, cp.BlockDelayNum, dal, latestLedger)
		switch res {
		case txSucceeded:
			succeed++
		case txFailed:
			failed++
		case txExpired:
			expired++
		case txCancelled:
			log.Infof("Migration has been cancelled due to high gas price")
			return
		}
	}
}

func handleSingleOnchainTx(
	cid ctype.CidType,
	onchainReq []byte,
	deadlines map[ctype.CidType]uint64,
	auth *bind.TransactOpts,
	client *ethclient.Client,
	contract *ledger.CelerLedger,
	blockDelay uint64,
	dal *storage.DAL,
	latestLedger ctype.Addr,
) int {
	log.Infof("Start to migrate for channel %x", cid)
	ctx := context.Background()
	// find a relatively lower gas price with limited time
	gasPrice, err := findGasPriceWithLimit(ctx, uint64(*maxGas), client)
	if err != nil {
		log.Error(err)
		return txCancelled // cancel this channel migration due to unable to find low gas price
	}

	auth.GasPrice = gasPrice
	auth.GasLimit = 0

	log.Infof("Found low gas price for channel %x migration: %d", cid, gasPrice.Uint64())
	// check if current migration request is expired
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Errorln(err)
		return txFailed
	}
	currentBlockNum := header.Number.Uint64()
	if deadlines[cid] <= currentBlockNum {
		log.Infof("migration request has been expired for channel %x", cid)
		return txExpired
	}

	log.Infof("Start to submit onchain transaction for channel %x", cid)
	tx, err := migrateChannelFrom(auth, onchainReq, contract)
	if err != nil {
		log.Errorln(err)
		return txFailed
	}
	log.Infof("migration request has been submitted for channel %x", cid)
	ctx2, cancel := context.WithTimeout(ctx, waitMinedTimeout)
	defer cancel()

	receipt, err := eth.WaitMined(ctx2, client, tx, blockDelay, config.BlockIntervalSec)
	if err != nil {
		log.Errorf("channel migration tx failed for channel(%x): %w", cid, err)
		return txFailed
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("Channel migration tx 0x%x succeeded for channel %x", receipt.TxHash, cid)
		if err = dal.Transactional(processChannelMigrationTx, cid, latestLedger, migrate.MigrationStateSubmitted); err != nil {
			log.Errorf("!!!Fail to update state for channel %x: %w", cid, err)
			return txFailed
		}
		return txSucceeded
	}

	log.Errorf("Channel migration tx 0x%x failed for channel(%x)", receipt.TxHash, cid)
	return txFailed
}

func processChannelMigrationTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	latestLedger := args[1].(ctype.Addr)
	state := args[2].(int)

	_, state, _, found, err := tx.GetChanMigration(cid, latestLedger)
	if err != nil {
		return fmt.Errorf("Fail to find migration info for channel %x: %w", cid, err)
	}
	if !found {
		return nil // might already be deleted by cNode triggered by monitor
	}
	if state == migrate.MigrationStateSubmitted {
		return nil
	}

	if err = tx.UpdateChanMigrationState(cid, latestLedger, state); err != nil {
		return fmt.Errorf("Fail to store migration info for channel %x: %w", cid, err)
	}

	return nil
}

func migrateChannelFrom(auth *bind.TransactOpts, onchainReq []byte, ledgerContract *ledger.CelerLedger) (*types.Transaction, error) {
	fromLedger, err := getFromLedger(onchainReq)
	if err != nil {
		return nil, err
	}

	return ledgerContract.MigrateChannelFrom(auth, fromLedger, onchainReq)
}

func findGasPriceWithLimit(ctx context.Context, maxGas uint64, client *ethclient.Client) (*big.Int, error) {
	timer := time.NewTimer(findGasPriceTimeout)
	defer timer.Stop()
	maxGas += gasPriceTolerance // we allow
	capPrice := new(big.Int).SetUint64(maxGas * 1e9)

	for {
		select {
		case <-timer.C:
			return nil, errors.New("cannot find low gas price, run this tool in other time")
		default:
			price, err := client.SuggestGasPrice(ctx)
			if err != nil {
				log.Fatalln(err)
			}
			if capPrice.Cmp(price) < 0 {
				time.Sleep(gasPriceQueryInterval)
			} else {
				// found a proper gas price, send transaction
				return price, nil
			}
		}
	}
}

func getFromLedger(onchainReq []byte) (ctype.Addr, error) {
	var req chain.ChannelMigrationRequest
	err := proto.Unmarshal(onchainReq, &req)
	if err != nil {
		return ctype.ZeroAddr, err
	}

	var info entity.ChannelMigrationInfo
	err = proto.Unmarshal(req.GetChannelMigrationInfo(), &info)
	if err != nil {
		return ctype.ZeroAddr, err
	}

	return ctype.Bytes2Addr(info.GetFromLedgerAddress()), nil
}

func overrideProfile(profile *common.CProfile) {
	profile.BlockDelayNum = uint64(*blkdelay)
	if *storesql != "" {
		profile.StoreSql = *storesql
	} else if *storedir != "" {
		profile.StoreDir = *storedir
	}
}
