package common

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type ShutdownHook func()

var hooks []ShutdownHook

func init() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		for _, hook := range hooks {
			log.Infof("Executing shutdown hook %v", hook)
			hook()
		}
		os.Exit(0)
	}()
}

func RegisterShutdownHook(hook ShutdownHook) {
	hooks = append(hooks, hook)
}
