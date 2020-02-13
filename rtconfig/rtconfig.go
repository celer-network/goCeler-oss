// Copyright 2018-2019 Celer Network

// runtime config helpers, includes chan for os signal etc

package rtconfig

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var (
	rtc  = &RuntimeConfig{}        // pointer to actual runtime configs
	lock sync.RWMutex              // rw mutex to protect read and write rtc (easier than atomic.Value)
	c    = make(chan os.Signal, 1) // chan to receive os signal, sighup triggers reload
)

const (
	defaultStreamSendTimeoutS   = uint64(1)
	defaultOspDepositMultiplier = int64(10)
	defaultMaxDisputeTimeout    = uint64(20000)
	defaultMinDisputeTimeout    = uint64(8000)
	defaultColdBootstrapDeposit = uint64(1e18)
	defaultMaxPaymentTimeout    = uint64(10000)
	defaultMaxNumPendingPays    = uint64(200)
)

// Init parse the json config file at path and start a goroutine to reload upon syscall.SIGHUP
// errors are not critical because default values have no effect
func Init(path string) error {
	err := updateConfigFromFile(path)
	if err != nil {
		return err
	}
	signal.Notify(c, syscall.SIGHUP) // ask the os to notify us when sighup is received
	go func() {
		for {
			s := <-c // block reading from os.Signal chan
			switch s {
			// kill -SIGHUP pid or kill -s HUP pid or kill -1 pid
			case syscall.SIGHUP:
				log.Info("Receive SIGHUP signal")
				updateConfigFromFile(path)
			default:
				log.Warn("Unsupported OS signal. Do nothing")
			}
		}
	}()
	return nil
}

// updateConfigFromFile updates rtc to point to a new config
// on any err, no change to rtc
func updateConfigFromFile(path string) error {
	log.Info("Loading runtime config from ", path)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warnln("rtconfig: file read err", err)
		return err
	}
	var newCfg RuntimeConfig
	err = json.Unmarshal(raw, &newCfg)
	if err != nil {
		log.Warnln("rtconfig: json parse err", err)
		return err
	}
	lock.Lock()
	rtc = &newCfg
	lock.Unlock()
	setLogLevel(rtc.LogLevel)
	log.Info("New runtime config:", pb2json(&newCfg))
	return nil
}

func setLogLevel(level string) {
	switch level {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.Warn("invalid log level input, set log to InfoLevel by default")
		log.SetLevel(log.InfoLevel)
	}
}

// GetOpenChanWaitSecond returns open_chan_wait_s
// int64 for easy use with time.XX funcs
func GetOpenChanWaitSecond() int64 {
	lock.RLock()
	defer lock.RUnlock()
	return rtc.OpenChanWaitS
}

// GetMinGasGwei returns min_gas_gwei
func GetMinGasGwei() uint64 {
	lock.RLock()
	defer lock.RUnlock()
	return rtc.MinGasGwei
}

// GetMaxGasGwei returns max_gas_gwei
func GetMaxGasGwei() uint64 {
	lock.RLock()
	defer lock.RUnlock()
	return rtc.MaxGasGwei
}

// GetOspDepositMultiplier returns osp_deposit_multiplier
// If not set in rtconfig, returns 10.
func GetOspDepositMultiplier() int64 {
	lock.RLock()
	defer lock.RUnlock()
	if rtc.OspDepositMultiplier == 0 {
		return defaultOspDepositMultiplier
	}
	return rtc.OspDepositMultiplier
}

// GetStreamSendTimeoutSecond returns stream_send_timeout_s
func GetStreamSendTimeoutSecond() uint64 {
	lock.RLock()
	defer lock.RUnlock()
	if rtc.StreamSendTimeoutS == 0 {
		return defaultStreamSendTimeoutS
	}
	return rtc.StreamSendTimeoutS
}

// GetMaxDisputeTimeout returns max_dispute_timeout
func GetMaxDisputeTimeout() uint64 {
	lock.RLock()
	defer lock.RUnlock()
	if rtc.MaxDisputeTimeout == 0 {
		return defaultMaxDisputeTimeout
	}
	return rtc.MaxDisputeTimeout
}

// GetMinDisputeTimeout returns max_dispute_timeout
func GetMinDisputeTimeout() uint64 {
	lock.RLock()
	defer lock.RUnlock()
	if rtc.MinDisputeTimeout == 0 {
		return defaultMinDisputeTimeout
	}
	return rtc.MinDisputeTimeout
}

// GetEthColdBootstrapDeposit returns osp eth deposit cap for cold bootstrap
// Return a hard-coded default if it's not set in rtconfig.
func GetEthColdBootstrapDeposit() *big.Int {
	lock.RLock()
	defer lock.RUnlock()
	ret := big.NewInt(0)
	if rtc.EthColdBootstrapDeposit != "" {
		ret.SetString(rtc.EthColdBootstrapDeposit, 10)
	} else {
		ret.SetUint64(defaultColdBootstrapDeposit)
	}
	return ret
}

// GetErc20ColdBootstrapDeposit returns osp erc20 deposit cap for cold bootstrap
// If it's configured in erc20_cold_bootstrap_deposit_map, return the value corresponding
// to the token addr. Otherwise, returns the default for rtconfig. If rtconfig is empty,
// return a hard-coded default
func GetErc20ColdBootstrapDeposit(addr []byte) *big.Int {
	lock.RLock()
	defer lock.RUnlock()
	ret := big.NewInt(0)
	if deposit, ok := rtc.Erc20ColdBootstrapDepositMap[ctype.Bytes2Hex(addr)]; ok {
		ret.SetString(deposit, 10)
	} else if rtc.Erc20ColdBootstrapDepositDefault != "" {
		ret.SetString(rtc.Erc20ColdBootstrapDepositDefault, 10)
	} else {
		ret.SetUint64(defaultColdBootstrapDeposit)
	}
	return ret
}
func GetStandardConfigs() *StandardConfigs {
	return rtc.StandardConfigs
}
func GetOspToOspOpenConfigs() *OspToOspOpenConfigs {
	return rtc.OspToOspOpenConfigs
}

func GetMaxPaymentTimeout() uint64 {
	lock.RLock()
	defer lock.RUnlock()
	if rtc.MaxPaymentTimeout == 0 {
		return defaultMaxPaymentTimeout
	}
	return rtc.MaxPaymentTimeout
}

func GetMaxNumPendingPays() uint64 {
	lock.RLock()
	defer lock.RUnlock()
	if rtc.MaxNumPendingPays == 0 {
		return defaultMaxNumPendingPays
	}
	return rtc.MaxNumPendingPays
}

func pb2json(pb proto.Message) string {
	m := jsonpb.Marshaler{}
	ret, err := m.MarshalToString(pb)
	if err != nil {
		log.Error("pb2json err: ", err)
	}
	return ret
}
