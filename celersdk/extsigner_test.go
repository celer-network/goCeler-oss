package celersdk

import (
	"bytes"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

var (
	signResult = []byte("sign result")
	toSign1    = []byte("to sign 1")
	toSign2    = []byte("to sign 2")
)

// impl ExternalSignerCallback
type extSigner struct {
	tosign map[int][]byte // reqid to tosign msg or rawtx
}

func (s *extSigner) OnSignMessage(reqid int, msg []byte) {
	s.tosign[reqid] = msg
}
func (s *extSigner) OnSignTransaction(reqid int, rawtx []byte) {
	s.tosign[reqid] = rawtx
}

func TestExtSigner(t *testing.T) {
	signer := &extSigner{
		tosign: make(map[int][]byte),
	}
	esm := newExtSignerMgr(signer)
	esm.seq = 0 // set to 0 so first reqid is 1
	go func() {
		time.Sleep(50 * time.Millisecond)
		esm.SendSignResult(1, signResult)
	}()
	result, _ := esm.SignEthMessage(toSign1)
	if !bytes.Equal(result, signResult) {
		t.Fatalf("sign result mismatch, exp: %s, got: %s", signResult, result)
	}
	// esm does hash for ext signer
	if !bytes.Equal(crypto.Keccak256(toSign1), signer.tosign[1]) {
		t.Fatalf("mismatch tosign msg, exp: %s, got: %s", toSign1, signer.tosign[1])
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		esm.SendSignResult(2, signResult)
	}()
	result, _ = esm.SignEthTransaction(toSign2)
	if !bytes.Equal(result, signResult) {
		t.Fatalf("sign result mismatch, exp: %s, got: %s", signResult, result)
	}
	if !bytes.Equal(toSign2, signer.tosign[2]) {
		t.Fatalf("mismatch tosign msg, exp: %s, got: %s", toSign2, signer.tosign[2])
	}
}
