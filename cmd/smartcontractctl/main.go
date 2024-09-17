package main

import (
	"fmt"
	"sync/atomic"

	"github.com/bugnanetwork/bugnad/infrastructure/os/signal"
	"github.com/bugnanetwork/bugnad/util/panics"
)

var shutdown int32 = 0

func main() {
	interrupt := signal.InterruptListener()
	defer panics.HandlePanic(log, "main", nil)

	cmd, cfg := parseCommandLine()
	fmt.Println(cmd)

	switch cmd {
	case deploySubCmd:
		deployCfg := cfg.(*deployConfig)
		err := deploySmartContract(deployCfg)
		if err != nil {
			printErrorAndExit(err.Error())
		}
	}

	<-interrupt
	atomic.AddInt32(&shutdown, 1)
}
