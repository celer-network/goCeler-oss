package transactor

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrTxDropped = errors.New("onchain transaction dropped")
	ErrTxTimeout = errors.New("onchain transaction timeout")
	ErrTxReorg   = errors.New("onchain transaction reorg")
	// an error possibly returned when a transaction is pending
	ErrMissingField = errors.New("missing required field 'transactionHash' for Log")
)

func WaitMined(
	ctx context.Context,
	ec *ethclient.Client,
	tx *ethtypes.Transaction,
	blockDelay uint64,
	pollingIntervalSec uint64) (*ethtypes.Receipt, error) {
	return waitMined(ctx, ec, tx, tx.Hash(), blockDelay, pollingIntervalSec)
}

// WaitMinedWithTxHash only wait with given txhash, without other info such as nonce.
// Therefore, it cannot tell if a tx is dropped if not yet mined
func WaitMinedWithTxHash(
	ctx context.Context,
	ec *ethclient.Client,
	txHash string,
	blockDelay uint64,
	pollingIntervalSec uint64) (*ethtypes.Receipt, error) {
	return waitMined(ctx, ec, nil, ctype.Hex2Hash(txHash), blockDelay, pollingIntervalSec)
}

// waitMinedTx waits for tx to be mined on the blockchain
// It returns tx receipt when the tx has been mined and enough block confirmations have passed
func waitMined(
	ctx context.Context,
	ec *ethclient.Client,
	tx *ethtypes.Transaction,
	txHash ctype.Hash,
	blockDelay uint64,
	pollingIntervalSec uint64) (*ethtypes.Receipt, error) {
	if pollingIntervalSec == 0 {
		return nil, fmt.Errorf("invalid polling interval")
	}
	if ec == nil {
		return nil, fmt.Errorf("nil ethclient")
	}
	var txSender ctype.Addr
	if tx != nil {
		txHash = tx.Hash()
		msg, err := tx.AsMessage(ethtypes.NewEIP155Signer(tx.ChainId()))
		if err != nil {
			return nil, fmt.Errorf("AsMessage err: %w", err)
		}
		txSender = msg.From()
	}
	pollingInterval := time.Duration(pollingIntervalSec) * time.Second
	receipt, err := waitTxConfirmed(ctx, ec, tx, txSender, txHash, blockDelay, pollingInterval)
	for errors.Is(err, ErrTxReorg) { // retry if dropped due to chain reorg
		receipt, err = waitTxConfirmed(ctx, ec, tx, txSender, txHash, blockDelay, pollingInterval)
	}
	return receipt, err
}

func waitTxConfirmed(
	ctx context.Context,
	ec *ethclient.Client,
	tx *ethtypes.Transaction,
	txSender ctype.Addr,
	txHash ctype.Hash,
	blockDelay uint64,
	pollingInterval time.Duration) (*ethtypes.Receipt, error) {
	var receipt *ethtypes.Receipt
	var nonce uint64
	var err error
	txTimeout := time.Duration(rtconfig.GetWaitMinedTxTimeout()) * time.Second
	txQueryTimeout := time.Duration(rtconfig.GetWaitMinedTxQueryTimeout()) * time.Second
	txQueryRetryInterval := time.Duration(rtconfig.GetWaitMinedTxQueryRetryInterval()) * time.Second
	deadline := time.Now().Add(txTimeout)
	queryTicker := time.NewTicker(pollingInterval)
	defer queryTicker.Stop()
	var pending bool
	// wait tx to be mined
	for {
		if tx != nil {
			nonce, err = currentNonce(ctx, ec, txSender, txQueryTimeout, txQueryRetryInterval)
			if err != nil {
				return nil, fmt.Errorf("tx %x NonceAt err: %w", txHash, err)
			}
		}
		receipt, err = transactionReceipt(ctx, ec, txHash, txQueryTimeout, txQueryRetryInterval)
		if err == nil {
			log.Debugf("Transaction %x mined. Waiting for %d block confirmations", txHash, blockDelay)
			if blockDelay == 0 {
				return receipt, nil
			}
			break
		} else if err == ethereum.NotFound || err == ErrMissingField {
			if tx != nil {
				// tx is dropped if the account nonce is larger than the unmined tx nonce
				if tx.Nonce() < nonce {
					return nil, fmt.Errorf("tx %x err: %w", txHash, ErrTxDropped)
				}
			}
			if !pending && time.Now().After(deadline) {
				_, pending, err = transactionByHash(ctx, ec, txHash, txQueryTimeout, txQueryRetryInterval)
				if err != nil {
					return nil, fmt.Errorf("tx %x TransactionByHash err: %w", txHash, err)
				}
				if !pending {
					return nil, fmt.Errorf("tx %x err: %w", txHash, ErrTxTimeout)
				}
			}
			// Wait for the next round
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-queryTicker.C:
			}
		} else {
			return receipt, fmt.Errorf("tx %x get receipt err: %w", txHash, err)
		}
	}
	// wait for enough block confirmations
	confirmBlk := new(big.Int).Add(receipt.BlockNumber, new(big.Int).SetUint64(blockDelay))
	var header *ethtypes.Header
	for {
		header, err = blockHeader(ctx, ec, txQueryTimeout, txQueryRetryInterval)
		if err == nil && confirmBlk.Cmp(header.Number) < 0 {
			receipt, err = transactionReceipt(ctx, ec, txHash, txQueryTimeout, txQueryRetryInterval)
			if err == nil {
				log.Debugf("tx %x confirmed!", txHash)
				return receipt, nil
			} else if err == ethereum.NotFound || err == ErrMissingField {
				return nil, fmt.Errorf("tx %x err: %w", txHash, ErrTxReorg)
			} else {
				return receipt, fmt.Errorf("tx %x confirm receipt err: %w", txHash, err)
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}

func currentNonce(ctx context.Context, ec *ethclient.Client, account ctype.Addr,
	txQueryTimeout, txQueryRetryInterval time.Duration) (uint64, error) {
	nonce, err := ec.NonceAt(ctx, account, nil)
	if err != nil && retryOnErr(err) {
		deadline := time.Now().Add(txQueryTimeout)
		for time.Now().Before(deadline) {
			time.Sleep(txQueryRetryInterval)
			nonce, err = ec.NonceAt(ctx, account, nil)
			if err == nil {
				return nonce, nil
			}
			if retryOnErr(err) {
				log.Warnln("retry NonceAt err", err)
				continue
			} else {
				return nonce, err
			}
		}
	}
	return nonce, err
}

func transactionReceipt(ctx context.Context, ec *ethclient.Client, txHash ctype.Hash,
	txQueryTimeout, txQueryRetryInterval time.Duration) (*ethtypes.Receipt, error) {
	receipt, err := ec.TransactionReceipt(ctx, txHash)
	if err != nil && retryOnErr(err) {
		deadline := time.Now().Add(txQueryTimeout)
		for time.Now().Before(deadline) {
			time.Sleep(txQueryRetryInterval)
			receipt, err = ec.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			if retryOnErr(err) {
				log.Warnln("retry TransactionReceipt err", err)
				continue
			} else {
				return receipt, err
			}
		}
	}
	return receipt, err
}

func transactionByHash(ctx context.Context, ec *ethclient.Client, txHash ctype.Hash,
	txQueryTimeout, txQueryRetryInterval time.Duration) (*ethtypes.Transaction, bool, error) {
	tx, pending, err := ec.TransactionByHash(ctx, txHash)
	if err != nil && retryOnErr(err) {
		deadline := time.Now().Add(txQueryTimeout)
		for time.Now().Before(deadline) {
			time.Sleep(txQueryRetryInterval)
			tx, pending, err = ec.TransactionByHash(ctx, txHash)
			if err == nil {
				return tx, pending, nil
			}
			if retryOnErr(err) {
				log.Warnln("retry TransactionByHash err", err)
				continue
			} else {
				return tx, pending, err
			}
		}
	}
	return tx, pending, err
}

func blockHeader(ctx context.Context, ec *ethclient.Client,
	txQueryTimeout, txQueryRetryInterval time.Duration) (*ethtypes.Header, error) {
	header, err := ec.HeaderByNumber(ctx, nil)
	if err != nil && retryOnErr(err) {
		deadline := time.Now().Add(txQueryTimeout)
		for time.Now().Before(deadline) {
			time.Sleep(txQueryRetryInterval)
			header, err = ec.HeaderByNumber(ctx, nil)
			if err == nil {
				return header, nil
			}
			if retryOnErr(err) {
				log.Warnln("retry HeaderByNumber err", err)
				continue
			} else {
				return header, err
			}
		}
	}
	return header, err
}

func retryOnErr(err error) bool {
	retryErrPatterns := []string{"bad gateway", "write on closed buffer"}
	errMsg := strings.ToLower(err.Error())
	for _, pat := range retryErrPatterns {
		if strings.Contains(errMsg, pat) {
			return true
		}
	}
	return false
}
