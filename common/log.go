package common

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func init() {
	file, err := os.OpenFile("./log.go.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Unable to open/create log file './log.go.txt': %s", err.Error())
	}
	RegisterShutdownHook(func() {
		log.Infof("Closing log file")
		_ = file.Close()
	})
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetLevel(log.DebugLevel)
}
