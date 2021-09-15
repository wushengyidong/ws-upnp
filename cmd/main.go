package main

import (
	"github.com/sirupsen/logrus"
	"time"
	"ws-upnp/internal/app/upnp"
)

const (
	mapTimeout = 10 * time.Second
)

func main() {
	startUPNPMapping()

	select {}
}

func startUPNPMapping() {
	upnp.RetryMapPort()
	go keepPortMapping()
}

func keepPortMapping() {
	refresh := time.NewTimer(mapTimeout)
	defer func() {
		refresh.Stop()
	}()

	for {
		select {
		case <-refresh.C:
			logrus.Info("Refreshing port mapping")
			upnp.RetryMapPort()
			refresh.Reset(mapTimeout)
		}
	}
}
