// Copyright 2018-2020 Celer Network

package rtconfig

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestLoadFromFile(t *testing.T) {
	err := updateConfigFromFile("test_cfg.json")
	if err != nil {
		t.Fatal(err)
	}
	ocw := GetOpenChanWaitSecond()
	if ocw != 10 {
		t.Error("mismatch open_chan_wait_s: ", ocw, " expect: ", 10)
	}
}

func TestInitAndSignal(t *testing.T) {
	err := Init("test_cfg.json")
	if err != nil {
		t.Fatal(err)
	}
	ocw := GetOpenChanWaitSecond()
	if ocw != 10 {
		t.Error("mismatch open_chan_wait_s: ", ocw, " expect: ", 10)
	}
	tcbConfig := GetTcbConfigs()
	if tcbConfig == nil {
		t.Error("tcbConfig is null")
	}
	standardConfig := GetStandardConfigs()
	if standardConfig == nil {
		t.Error("standardConfig is null")
	}

	swap2Files("test_cfg.json", "test_cfg2.json")       // cfg2 values will be loaded in next sighup
	defer swap2Files("test_cfg.json", "test_cfg2.json") // swap back after test
	c <- syscall.SIGHUP
	time.Sleep(100 * time.Millisecond) // sleep 0.1s to give updateConfigFromFile enough time to run
	ocw = GetOpenChanWaitSecond()
	if ocw != 20 { // cfg2 value
		t.Error("mismatch open_chan_wait_s: ", ocw, " expect: ", 20)
	}
	tcbConfig = GetTcbConfigs()
	if tcbConfig != nil {
		t.Error("tcbConfig is not null")
	}
	standardConfig = GetStandardConfigs()
	if standardConfig != nil {
		t.Error("standardConfig is not null")
	}
}

// helper util to swap 2 files by renaming
func swap2Files(f1, f2 string) {
	os.Rename(f1, f1+"_tmp")
	os.Rename(f2, f1)
	os.Rename(f1+"_tmp", f2)
}

func chkEq(v, exp string, t *testing.T) {
	if v != exp {
		t.Errorf("mismatch string exp: %s, got %s", exp, v)
	}
}
