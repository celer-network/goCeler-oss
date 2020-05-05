// Copyright 2018-2020 Celer Network

package e2e

import (
	"os"
	"testing"

	tf "github.com/celer-network/goCeler/testing"
)

// Killable is object that has Kill() func
type Killable interface {
	Kill() error
}

func setUp() []Killable {
	os.RemoveAll(sStoreDir)
	// use default 11000 and 8090 for adminrpc/web port
	s1 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
		"-port", sPort,
		"-selfrpc", sSelfRPC,
		"-storedir", sStoreDir,
		"-ks", ospKeystore,
		"-depositks", depositKeystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "s1",
		"-logprefix", "s1_"+ospEthAddr[:4],
		"-logcolor")

	return []Killable{s1}
}

func tearDown(tokill []Killable) {
	for _, p := range tokill {
		p.Kill()
	}
}
func tearDownMultiSvr(tokill []Killable) {
	tearDown(tokill)
	os.RemoveAll(sStoreDir)
}
func TestE2E(t *testing.T) {
	toKill := setUp()
	defer tearDownMultiSvr(toKill)

	t.Run("e2e-grp1", func(t *testing.T) {
		t.Run("adminSendToken", adminSendToken)
		t.Run("clientDepositEth", clientDepositEth)
		t.Run("clientDepositErc20WithRestart", clientDepositErc20WithRestart)
		t.Run("clientRecovery", clientRecovery)
		t.Run("concurrentOpenChannel", concurrentOpenChannel)
		t.Run("coldBootstrap", coldBootstrap)
		t.Run("cooperativeWithdrawErc20", cooperativeWithdrawErc20)
		t.Run("cooperativeWithdrawEthWithRestart", cooperativeWithdrawEthWithRestart)
		t.Run("cooperativeWithdrawAfterSendPay", cooperativeWithdrawAfterSendPay)
		t.Run("cooperativeWithdrawAndSendInvalidPay", cooperativeWithdrawAndSendInvalidPay)
		t.Run("cooperativeWithdrawInsufficient", cooperativeWithdrawInsufficient)
		t.Run("clientIntendWithdrawErc20", clientIntendWithdrawErc20)
		t.Run("ospIntendWithdrawErc20", ospIntendWithdrawErc20)
	})

	t.Run("e2e-grp2", func(t *testing.T) {
		t.Run("sendCondPayWithErc20", sendCondPayWithErc20)
		t.Run("sendCondPayWithEth", sendCondPayWithEth)
		t.Run("sendCondPayWithEthDstOffline", sendCondPayWithEthDstOffline)
		t.Run("sendEthOnVirtualContractCondition", sendEthOnVirtualContractCondition)
		t.Run("sendCondPayNoEnoughErc20AtSrc", sendCondPayNoEnoughErc20AtSrc)
		t.Run("sendCondPayNoEnoughErc20AtOsp", sendCondPayNoEnoughErc20AtOsp)
		t.Run("delegateSendEth", delegateSendEth)
		t.Run("delegateSendErc20", delegateSendErc20)
		t.Run("tcbOpenChannel", tcbOpenChannel)
		t.Run("sendEthPayTimeout", sendEthPayTimeout)
		t.Run("sendPaySettleWithEthDstReconnect", sendPaySettleWithEthDstReconnect)
		t.Run("sendCondPayWithEthToOSP", sendCondPayWithEthToOSP)
		t.Run("slidingWindowEth", slidingWindowEth)
		t.Run("authSync", authSync)
	})

	t.Run("e2e-grp3", func(t *testing.T) {
		t.Run("disputeEthPayWithVirtualContract", disputeEthPayWithVirtualContract)
		t.Run("disputeEthPayWithDeployedContract", disputeEthPayWithDeployedContract)
		t.Run("disputeEthPaySrcOffline", disputeEthPaySrcOffline)
		t.Run("disputeEthPayWithDeployedGomoku", disputeEthPayWithDeployedGomoku)
		t.Run("disputePayBySigTimeoutWithDeployedContract", disputeEthPayBySigTimeoutWithDeployedContract)
		t.Run("settleErc20ChannelEmpty", settleErc20ChannelEmpty)
		t.Run("settleErc20ChannelOneSimplex", settleErc20ChannelOneSimplex)
		t.Run("settleErc20ChannelFullDuplex", settleErc20ChannelFullDuplex)
		t.Run("settleErc20ChannelWithReopen", settleErc20ChannelWithReopen)
		t.Run("settleErc20ChannelWithDispute", settleErc20ChannelWithDispute)
		t.Run("ospIntendSettleErc20Channel", ospIntendSettleErc20Channel)
		t.Run("ospDepositAndRefill", ospDepositAndRefill)
	})

	/*// following tests for tools do not need to be run with CI
	t.Run("e2e-tools", func(t *testing.T) {
		t.Run("ethChannelView", ethChannelView)
		t.Run("erc20ChannelView", erc20ChannelView)
		t.Run("ospAdminTest", ospAdminTest)
	})*/
}
