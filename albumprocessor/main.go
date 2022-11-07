package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Run AlbumProcessor service")

	go DoKafkaProcessing()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logrus.Info("AlbumProcessor service exist")
}
