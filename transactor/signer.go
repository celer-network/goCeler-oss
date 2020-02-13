// Copyright 2016 The go-ethereum Authors
// Copyright 2019 Celer Network
//
// This is a fork of go-ethereum/accounts/bind/abi/auth.go to enforce EIP-155 signers to protect
// against replay attacks across chain IDs. This should be deprecated and removed once
// https://github.com/ethereum/go-ethereum/issues/16484 is resolved.

package transactor

import (
	"crypto/ecdsa"
	"errors"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func newTransactOpts(
	keyin io.Reader, passphrase string, chainId *big.Int) (*bind.TransactOpts, error) {
	json, err := ioutil.ReadAll(keyin)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(json, passphrase)
	if err != nil {
		return nil, err
	}
	return newKeyedTransactOpts(key.PrivateKey, chainId), nil
}

func newKeyedTransactOpts(key *ecdsa.PrivateKey, chainId *big.Int) *bind.TransactOpts {
	keyAddr := crypto.PubkeyToAddress(key.PublicKey)
	eip155Signer := types.NewEIP155Signer(chainId)
	return &bind.TransactOpts{
		From: keyAddr,
		// Ignore the passed in Signer to enforce EIP-155
		Signer: func(
			signer types.Signer,
			address common.Address,
			tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, errors.New("not authorized to sign this account")
			}
			signature, err := crypto.Sign(eip155Signer.Hash(tx).Bytes(), key)
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(eip155Signer, signature)
		},
	}
}
