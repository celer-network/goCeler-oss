// Copyright 2018-2020 Celer Network

package celersdk

import (
	"hash/fnv"
	"math/rand"
	"sync"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/celersdkintf"
	"github.com/celer-network/goCeler/client"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

// newClient creates a new Celer Client based on provided config
// @param keyStore the keystore json file content
// @param pass	passphrase for keystore
// @param profile common.CProfile
// @return a Celer client or null if error
// @throws Exception
func newClient(
	keyStore string,
	pass string,
	profile *common.CProfile,
	clientCallback ClientCallback) (*Client, error) {
	log.Infof("%+v", profile)
	cclient, err := client.NewCelerClient(keyStore, pass, *profile, clientCallback)
	if cclient == nil {
		return nil, err
	}
	err = cclient.RegisterStream()
	if err != nil {
		cclient.Close()
		return nil, err
	}
	return &Client{
		c:       cclient,
		datadir: profile.StoreDir,
	}, nil
}

func createXfer(tk *Token, receiver, amtWei string) *entity.TokenTransfer {
	xfer := &entity.TokenTransfer{
		Token: sdkToken2entityToken(tk),
		Receiver: &entity.AccountAmtPair{
			Account: ctype.Hex2Bytes(receiver),
			Amt:     utils.Wei2BigInt(amtWei).Bytes(),
		},
	}
	return xfer
}

func bc2c(bc *BooleanCondition) (*entity.Condition, error) {
	var cond *entity.Condition
	if bc.OnChainDeployed {
		sessionQuery := &app.SessionQuery{
			Session: ctype.Hex2Bytes(bc.SessionID),
			Query:   bc.ArgsForQueryOutcome,
		}
		seralizedSessionQuery, err := proto.Marshal(sessionQuery)
		if err != nil {
			return nil, err
		}
		cond = &entity.Condition{
			ConditionType:           entity.ConditionType_DEPLOYED_CONTRACT,
			DeployedContractAddress: ctype.Hex2Addr(bc.OnChainAddress).Bytes(),
			ArgsQueryFinalization:   ctype.Hex2Bytes(bc.SessionID),
			ArgsQueryOutcome:        seralizedSessionQuery,
		}
	} else {
		cond = &entity.Condition{
			ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
			VirtualContractAddress: ctype.Hex2Bytes(bc.SessionID),
			ArgsQueryOutcome:       bc.ArgsForQueryOutcome,
		}
	}
	return cond, nil
}

func sdkToken2entityToken(tk *Token) *entity.TokenInfo {
	var token *entity.TokenInfo
	if tk == nil { // ETH case
		token = &entity.TokenInfo{
			TokenType: entity.TokenType_ETH,
		}
	} else {
		token = &entity.TokenInfo{
			TokenType:    entity.TokenType_ERC20,
			TokenAddress: ctype.Hex2Bytes(tk.Addr),
		}
	}
	return token
}

func conditionToEntityCondition(condition *Condition) *entity.Condition {
	var conditionType entity.ConditionType
	var deployedContractAddress []byte
	var virtualContractAddress []byte
	if condition.OnChainDeployed {
		conditionType = entity.ConditionType_DEPLOYED_CONTRACT
		deployedContractAddress = condition.ContractAddress
	} else {
		conditionType = entity.ConditionType_VIRTUAL_CONTRACT
		virtualContractAddress = condition.ContractAddress
	}
	return &entity.Condition{
		ConditionType:           conditionType,
		DeployedContractAddress: deployedContractAddress,
		VirtualContractAddress:  virtualContractAddress,
		ArgsQueryFinalization:   condition.IsFinalizedArgs,
		ArgsQueryOutcome:        condition.GetOutcomeArgs,
	}
}

type CSharedRandom struct {
	matchID string
	rng     *rand.Rand
	mu      sync.Mutex
}

func NewSharedRandom(matchID string) *CSharedRandom {
	h := fnv.New64a()
	h.Write([]byte(matchID))
	seed := h.Sum64()

	sr := &CSharedRandom{
		matchID: matchID,
		rng:     rand.New(rand.NewSource(int64(seed))),
	}
	return sr
}

func (sr *CSharedRandom) GetSharedRandom() float64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.rng.Float64()
}

// version is set by ldflags during sdk build
var version string

// Version returns CelerSDK version
func Version() string {
	return version
}

// GetCelerErrCode is the helper util to return errcode if e is a celersdkintf.E
// other wise returns -1. meaning -1 is reserved and shouldn't be used by other systems
func GetCelerErrCode(e error) int {
	if ce, ok := e.(*celersdkintf.E); ok {
		return ce.Code
	}
	return -1
}

type LogCallback interface {
	// msg is the log output
	OnLog(msg string)
}

// SetLogCallback set the self-defined writer in the log module.
// Once set, logs will be written to cb.Onlog() instead of os.Stderr
func SetLogCallback(cb LogCallback) {
	writer := &logCallbackWriter{cb: cb}
	log.SetOutput(writer)
}

type logCallbackWriter struct {
	cb LogCallback
}

func (w *logCallbackWriter) Write(output []byte) (n int, err error) {
	w.cb.OnLog(string(output))
	return len(output), nil
}
