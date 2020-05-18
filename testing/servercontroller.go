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
		"-routerbcastinterval", "4",
		"-routerbuildinterval", "5",
		"-routeralivetimeout", "8",
		"-ospclearpayinterval", "8")
	process := StartProcess(path, args...)
	time.Sleep(2 * time.Second)
	return &ServerController{process}
}

func (sc *ServerController) Kill() error {
	KillProcess(sc.process)
	return nil
}
