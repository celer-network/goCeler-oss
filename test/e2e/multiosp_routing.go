// Copyright 2020 Celer Network

package e2e

import (
	"testing"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/storage"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/testing/testapp"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

func multiOspRouting(args ...*tf.ServerController) func(*testing.T) {
	return func(t *testing.T) {
		log.Info("============== start test multiOspRouting ==============")
		defer log.Info("============== end test multiOspRouting ==============")
		/* OSPs connect with each other with topology:
		  o4---o5
		 /  \    \
		o3---o1---o2
		*/
		// Let osp2 initiate openning channel with osp1.
		err := requestOpenChannel(o2AdminWeb, osp1EthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
		if err != nil {
			log.Warn(err)
		}
		// Let osp3 initiate openning channel with osp1.
		err = requestOpenChannel(o3AdminWeb, osp1EthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		// Let osp4 initiate openning channel with osp1.
		err = requestOpenChannel(o4AdminWeb, osp1EthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		// Let osp4 initiate openning channel with osp3.
		err = requestOpenChannel(o4AdminWeb, osp3EthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		// Let osp4 initiate openning channel with osp5.
		err = requestOpenChannel(o4AdminWeb, osp5EthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		// Let osp5 initiate openning channel with osp2.
		err = requestOpenChannel(o5AdminWeb, osp2EthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}

		ks, addrs, err := tf.CreateAccountsWithBalance(5, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
		c3KeyStore := ks[2]
		c3EthAddr := addrs[2]
		c5KeyStore := ks[4]
		c5EthAddr := addrs[4]
		c3, err := startClientForOsp3(c3KeyStore)
		if err != nil {
			t.Error(err)
			return
		}
		defer c3.Kill()
		c5, err := startClientForOsp5(c5KeyStore)
		if err != nil {
			t.Error(err)
			return
		}
		defer c5.Kill()
		res, err := c3.OpenChannel(c3EthAddr, entity.TokenType_ETH, tokenAddrEth, initialBalance, initialBalance)
		if err != nil {
			t.Error(err)
			return
		}
		c3cid := ctype.Hex2Cid(res.GetChannelId())
		log.Infoln("channel id for c3:", ctype.Cid2Hex(c3cid))
		res, err = c5.OpenChannel(c5EthAddr, entity.TokenType_ETH, tokenAddrEth, initialBalance, initialBalance)
		if err != nil {
			t.Error(err)
			return
		}
		c5cid := ctype.Hex2Cid(res.GetChannelId())
		log.Infoln("channel id for c5:", ctype.Cid2Hex(c5cid))

		dal1, dal2, dal3, dal4, dal5 := getAllOspDALs()
		token := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tokenAddrEth))
		cid12, found, err := dal1.GetCidByPeerToken(ctype.Hex2Addr(osp2EthAddr), token)
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("channel cid12 not found")
			return
		}
		log.Infoln("channel id for o1 o2:", ctype.Cid2Hex(cid12))
		cid13, found, err := dal1.GetCidByPeerToken(ctype.Hex2Addr(osp3EthAddr), token)
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("channel cid13 not found")
			return
		}
		log.Infoln("channel id for o1 o3:", ctype.Cid2Hex(cid13))
		cid14, found, err := dal1.GetCidByPeerToken(ctype.Hex2Addr(osp4EthAddr), token)
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("channel cid14 not found")
			return
		}
		log.Infoln("channel id for o1 o4:", ctype.Cid2Hex(cid14))
		cid25, found, err := dal2.GetCidByPeerToken(ctype.Hex2Addr(osp5EthAddr), token)
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("channel cid25 not found")
			return
		}
		log.Infoln("channel id for o2 o5:", ctype.Cid2Hex(cid25))
		cid34, found, err := dal3.GetCidByPeerToken(ctype.Hex2Addr(osp4EthAddr), token)
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("channel cid34 not found")
			return
		}
		log.Infoln("channel id for o3 o4:", ctype.Cid2Hex(cid34))
		cid45, found, err := dal4.GetCidByPeerToken(ctype.Hex2Addr(osp5EthAddr), token)
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("channel cid45 not found")
			return
		}
		log.Infoln("channel id for o4 o5:", ctype.Cid2Hex(cid45))

		sleep(10)

		log.Infoln("p1: c3 pay c5, should go through c3->o3->o4->o5->c5")
		p1, err := c3.SendPayment(c5EthAddr, "1", entity.TokenType_ETH, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		err = waitForPaymentCompletion(p1, c3, c5)
		if err != nil {
			t.Error(err)
			return
		}
		err = c3.AssertBalance(
			tokenAddrEth,
			tf.AddAmtStr(initialBalance, "-1"),
			"0",
			tf.AddAmtStr(initialBalance, "1"))
		if err != nil {
			t.Error(err)
			return
		}
		err = c5.AssertBalance(
			tokenAddrEth,
			tf.AddAmtStr(initialBalance, "1"),
			"0",
			tf.AddAmtStr(initialBalance, "-1"))
		if err != nil {
			t.Error(err)
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err := dal3.GetPaymentInfo(ctype.Hex2PayID(p1))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p1 not found in o3")
			return
		}
		expInCid, expInState, expOutCid, expOutState := c3cid, structs.PayState_COSIGNED_PAID, cid34, structs.PayState_COSIGNED_PAID
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal4.GetPaymentInfo(ctype.Hex2PayID(p1))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p1 not found in o4")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid34, structs.PayState_COSIGNED_PAID, cid45, structs.PayState_COSIGNED_PAID
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal5.GetPaymentInfo(ctype.Hex2PayID(p1))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p1 not found in o5")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid45, structs.PayState_COSIGNED_PAID, c5cid, structs.PayState_COSIGNED_PAID
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		log.Info("------------------ kill o4, route should change ------------------")
		o4 := args[3]
		o4.Kill()
		sleep(20)

		log.Infoln("p2: o3 pay o5, should go through o3->o1->o2->o5")
		p2, err := requestSendToken(o3AdminWeb, osp5EthAddr, "1", tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		sleep(5)

		_, _, inCid, inState, outCid, outState, _, found, err = dal3.GetPaymentInfo(ctype.Hex2PayID(p2))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p2 not found in o3")
			return
		}
		expInCid, expInState, expOutCid, expOutState = ctype.ZeroCid, structs.PayState_NULL, cid13, structs.PayState_COSIGNED_PAID
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal1.GetPaymentInfo(ctype.Hex2PayID(p2))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p2 not found in o1")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid13, structs.PayState_COSIGNED_PAID, cid12, structs.PayState_COSIGNED_PAID
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal2.GetPaymentInfo(ctype.Hex2PayID(p2))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p2 not found in o2")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid12, structs.PayState_COSIGNED_PAID, cid25, structs.PayState_COSIGNED_PAID
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal5.GetPaymentInfo(ctype.Hex2PayID(p2))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p2 not found in o5")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid25, structs.PayState_COSIGNED_PAID, ctype.ZeroCid, structs.PayState_NULL
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		log.Infoln("p3: c5 pay c3, should go through c5->o5->o2->o1->o3->c3")
		p3, err := c5.SendPayment(c3EthAddr, "1", entity.TokenType_ETH, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		err = waitForPaymentCompletion(p3, c5, c3)
		if err != nil {
			t.Error(err)
			return
		}
		err = c3.AssertBalance(tokenAddrEth, initialBalance, "0", initialBalance)
		if err != nil {
			t.Error(err)
			return
		}
		err = c5.AssertBalance(tokenAddrEth, initialBalance, "0", initialBalance)
		if err != nil {
			t.Error(err)
			return
		}

		log.Info("------------------ restart o4 ------------------")
		o4 = tf.StartServerController(outRootDir+toBuild["server"],
			"-profile", noProxyProfile,
			"-port", o4Port,
			"-storedir", sStoreDir,
			"-adminrpc", o4AdminRPC,
			"-adminweb", o4AdminWeb,
			"-ks", osp4Keystore,
			"-nopassword",
			"-rtc", rtConfigMultiOSP,
			"-defaultroute", osp1EthAddr,
			"-svrname", "o4",
			"-logcolor",
			"-logprefix", "o4_"+osp4EthAddr[:4])
		defer o4.Kill()
		utils.RequestRegisterStream(o4AdminWeb, ctype.Hex2Addr(osp1EthAddr), localhost+o1Port)
		utils.RequestRegisterStream(o4AdminWeb, ctype.Hex2Addr(osp3EthAddr), localhost+o3Port)
		utils.RequestRegisterStream(o4AdminWeb, ctype.Hex2Addr(osp5EthAddr), localhost+o5Port)

		log.Info("p4: c5 pay random addr, expect loop route c5->o5->o4->o1->o3->o4")
		// Use the default route settings to create a pay loop.
		randAddr := "7a6d2a97da1c453a4e099e8054865a0a59728863"
		p4, err := c5.SendPayment(randAddr, "1", entity.TokenType_ETH, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		err = waitForPaymentCompletion(p4, c5, nil)
		if err != nil {
			t.Error(err)
			return
		}

		// pay should be rolled back, remaining balance shouldn't change.
		err = c5.AssertBalance(tokenAddrEth, initialBalance, "0", initialBalance)
		if err != nil {
			t.Error(err)
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal5.GetPaymentInfo(ctype.Hex2PayID(p4))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p4 not found in o5")
			return
		}
		expInCid, expInState, expOutCid, expOutState = c5cid, structs.PayState_COSIGNED_CANCELED, cid45, structs.PayState_COSIGNED_CANCELED
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal4.GetPaymentInfo(ctype.Hex2PayID(p4))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p4 not found in o4")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid45, structs.PayState_COSIGNED_CANCELED, cid14, structs.PayState_COSIGNED_CANCELED
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal1.GetPaymentInfo(ctype.Hex2PayID(p4))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p4 not found in o1")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid14, structs.PayState_COSIGNED_CANCELED, cid13, structs.PayState_COSIGNED_CANCELED
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal3.GetPaymentInfo(ctype.Hex2PayID(p4))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p4 not found in o3")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid13, structs.PayState_COSIGNED_CANCELED, cid34, structs.PayState_COSIGNED_CANCELED
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		sleep(15)
		log.Infoln("p5: c5 pay c3, should go through c5->o5->o4->o3->c3")
		p5, err := c5.SendPayment(c3EthAddr, "1", entity.TokenType_ETH, tokenAddrEth)
		if err != nil {
			t.Error(err)
			return
		}
		err = waitForPaymentCompletion(p5, c5, c3)
		if err != nil {
			t.Error(err)
			return
		}
		err = c3.AssertBalance(
			tokenAddrEth,
			tf.AddAmtStr(initialBalance, "1"),
			"0",
			tf.AddAmtStr(initialBalance, "-1"))
		if err != nil {
			t.Error(err)
			return
		}
		err = c5.AssertBalance(
			tokenAddrEth,
			tf.AddAmtStr(initialBalance, "-1"),
			"0",
			tf.AddAmtStr(initialBalance, "1"))
		if err != nil {
			t.Error(err)
			return
		}
		_, _, inCid, inState, outCid, outState, _, found, err = dal4.GetPaymentInfo(ctype.Hex2PayID(p5))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p5 not found in o4")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid45, structs.PayState_COSIGNED_PAID, cid34, structs.PayState_COSIGNED_PAID
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		log.Info("------------------ test auto clear pays ------------------")
		constructor := testapp.GetSingleSessionConstructor(
			[]ctype.Addr{
				ctype.Hex2Addr(c3EthAddr),
				ctype.Hex2Addr(c5EthAddr),
			})
		appChanID, err := c3.NewAppChannelOnVirtualContract(
			testapp.AppCode,
			constructor,
			testapp.Nonce.Uint64(),
			testapp.Timeout.Uint64())
		if err != nil {
			t.Error(err)
			return
		}
		c3Cond1 := &entity.Condition{
			ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
			VirtualContractAddress: ctype.Hex2Bytes(appChanID),
			ArgsQueryFinalization:  []byte{},
			ArgsQueryOutcome:       []byte{2},
		}
		timeout := uint64(3)
		p6, err := c3.SendPaymentWithBooleanConditions(
			c5EthAddr, sendAmt, entity.TokenType_ETH, tokenAddrEth, []*entity.Condition{c3Cond1}, timeout)
		if err != nil {
			t.Error(err)
			return
		}
		payTime, err := c3.GetCurrentBlockNumber()
		if err != nil {
			t.Error(err)
			return
		}
		err = waitForPaymentPending(p6, c3, c5)
		if err != nil {
			t.Error(err)
			return
		}
		err = c3.AssertBalance(
			tokenAddrEth,
			tf.AddAmtStr(initialBalance, "0"),
			"1",
			tf.AddAmtStr(initialBalance, "-1"))
		if err != nil {
			t.Error(err)
			return
		}
		err = c5.AssertBalance(
			tokenAddrEth,
			tf.AddAmtStr(initialBalance, "-1"),
			"0",
			tf.AddAmtStr(initialBalance, "0"))
		if err != nil {
			t.Error(err)
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal3.GetPaymentInfo(ctype.Hex2PayID(p6))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p6 not found in o3")
			return
		}
		expInCid, expInState, expOutCid, expOutState = c3cid, structs.PayState_COSIGNED_PENDING, cid34, structs.PayState_COSIGNED_PENDING
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal4.GetPaymentInfo(ctype.Hex2PayID(p6))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p6 not found in o4")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid34, structs.PayState_COSIGNED_PENDING, cid45, structs.PayState_COSIGNED_PENDING
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal5.GetPaymentInfo(ctype.Hex2PayID(p6))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p6 not found in o5")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid45, structs.PayState_COSIGNED_PENDING, c5cid, structs.PayState_COSIGNED_PENDING
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		log.Info("wait till pay expired")
		err = c3.WaitUntilDeadline(payTime + timeout + 10)
		if err != nil {
			t.Error(err)
			return
		}
		sleep(5)

		_, _, inCid, inState, outCid, outState, _, found, err = dal3.GetPaymentInfo(ctype.Hex2PayID(p6))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p6 not found in o3")
			return
		}
		expInCid, expInState, expOutCid, expOutState = c3cid, structs.PayState_COSIGNED_PENDING, cid34, structs.PayState_COSIGNED_CANCELED
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal4.GetPaymentInfo(ctype.Hex2PayID(p6))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p6 not found in o4")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid34, structs.PayState_COSIGNED_CANCELED, cid45, structs.PayState_COSIGNED_CANCELED
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}

		_, _, inCid, inState, outCid, outState, _, found, err = dal5.GetPaymentInfo(ctype.Hex2PayID(p6))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Error("p6 not found in o5")
			return
		}
		expInCid, expInState, expOutCid, expOutState = cid45, structs.PayState_COSIGNED_CANCELED, c5cid, structs.PayState_COSIGNED_PENDING
		if inCid != expInCid || inState != expInState || outCid != expOutCid || outState != expOutState {
			t.Errorf("pay states error get in: %x %s out: %x %s, exp int: %x %s out: %x %s",
				inCid, fsm.PayStateName(inState), outCid, fsm.PayStateName(outState),
				expInCid, fsm.PayStateName(expInState), expOutCid, fsm.PayStateName(expOutState))
			return
		}
	}
}

func getAllOspDALs() (dal1, dal2, dal3, dal4, dal5 *storage.DAL) {
	profile1 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp1EthAddr,
	}
	dal1 = toolsetup.NewDAL(profile1)

	profile2 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp2EthAddr,
	}
	dal2 = toolsetup.NewDAL(profile2)

	profile3 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp3EthAddr,
	}
	dal3 = toolsetup.NewDAL(profile3)

	profile4 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp4EthAddr,
	}
	dal4 = toolsetup.NewDAL(profile4)

	profile5 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp5EthAddr,
	}
	dal5 = toolsetup.NewDAL(profile5)
	return dal1, dal2, dal3, dal4, dal5
}

func startClientForOsp2(keystorePath string) (*tf.ClientController, error) {
	return tf.StartClientController(
		tf.GetNextClientPort(),
		keystorePath,
		outRootDir+"c2o2.json",
		outRootDir+"c2Store",
		"c2")
}
func startClientForOsp3(keystorePath string) (*tf.ClientController, error) {
	return tf.StartClientController(
		tf.GetNextClientPort(),
		keystorePath,
		outRootDir+"c3o3.json",
		outRootDir+"c3Store",
		"c3")
}
func startClientForOsp4(keystorePath string) (*tf.ClientController, error) {
	return tf.StartClientController(
		tf.GetNextClientPort(),
		keystorePath,
		outRootDir+"c4o4.json",
		outRootDir+"c4Store",
		"c4")
}
func startClientForOsp5(keystorePath string) (*tf.ClientController, error) {
	return tf.StartClientController(
		tf.GetNextClientPort(),
		keystorePath,
		outRootDir+"c5o5.json",
		outRootDir+"c5Store",
		"c5")
}
