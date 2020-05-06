// Copyright 2018-2020 Celer Network

package testing

import (
	"os"
	"time"
)

type ServerController struct {
	process *os.Process
}

// note address isn't used
func StartServerController(path string, args ...string) *ServerController {
	args = append(args,
		"-routerbcastinterval", "5",
		"-routerbuildinterval", "10",
		"-routeralivetimeout", "12",
		"-ospclearpayinterval", "10")
	process := StartProcess(path, args...)
	time.Sleep(2 * time.Second)
	return &ServerController{process}
}

func (sc *ServerController) Kill() error {
	KillProcess(sc.process)
	return nil
}
