// Copyright 2018-2020 Celer Network

package cobj

import (
	"crypto/ecdsa"

	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

type CelerSigner struct {
	key *ecdsa.PrivateKey
}

func NewCelerSigner(privateKey string) (*CelerSigner, error) {
	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	c := &CelerSigner{
		key: key,
	}
	return c, nil
}

// input data: a byte array of raw message to be signed
// return a byte array signature in the R,S,V format
func (s *CelerSigner) SignEthMessage(data []byte) ([]byte, error) {
	sig, err := crypto.Sign(utils.GeneratePrefixedHash(data), s.key)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// input rawTx: a byte array of a RLP-encoded unsigned Ethereum raw transaction
// return a byte array signed raw tx in RLP-encoded format
func (s *CelerSigner) SignEthTransaction(rawTx []byte) ([]byte, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(rawTx, tx); err != nil {
		return nil, err
	}
	eip155Signer := types.NewEIP155Signer(config.ChainID)
	signature, err := crypto.Sign(eip155Signer.Hash(tx).Bytes(), s.key)
	if err != nil {
		return nil, err
	}
	tx, err = tx.WithSignature(eip155Signer, signature)
	if err != nil {
		return nil, err
	}
	return rlp.EncodeToBytes(tx)
}
