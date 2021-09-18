package upnp

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"time"
)

const (
	maxRefreshRetries = 5
	configFilePath    = ".config.yml"
	mapTimeout        = 30 * time.Minute
)

var successMapPort = 0

func makeUPNPListener(intPort int, extPort int, nat NAT) error {
	desc := nat.(*upnpNAT).ourIP + "_TCP_" + strconv.Itoa(intPort)
	lifetime := mapTimeout.Seconds()
	port, err := nat.AddPortMapping("tcp", extPort, intPort, desc, int(lifetime))
	if err != nil {
		return errors.New(fmt.Sprintf("Port mapping error: %v", err))
	}

	successMapPort = port

	return nil
}

func RetryMapPort() {
	log.Info("Port Mapping for UPnP!")

	nat, discoverErr := Discover()
	if discoverErr != nil {
		log.Errorf(fmt.Sprintf("Discover NAT Gateway device failed: %v", discoverErr))
	}

	if nat == nil {
		log.Error("Nat is nil pointer")
		return
	}

	cachePort, cacheErr := getCachePort()
	if cacheErr != nil {
		log.Errorf("Create config file failed: %v", cacheErr)
		return
	}

	var randomNumber int
	if cachePort == 0 {
		rand.Seed(time.Now().UnixNano())
		randomNumber = rand.Intn(55001) + 10000
	} else {
		randomNumber = cachePort
		successMapPort = cachePort
	}

	intPort, extPort := 44158, randomNumber

	var err error
	for retryCnt := 0; retryCnt < maxRefreshRetries; retryCnt++ {
		err = makeUPNPListener(intPort, extPort, nat)
		if err == nil {
			writeCachePort(configFilePath, successMapPort)
			log.Infof("Retry=>NAT Traversal successful from external port %d to internal port %d", extPort, intPort)
			return
		}

		log.Errorf("Retry=>Renewing port mapping try #%d from external port %d to internal port %d failed with %s",
			retryCnt+1, extPort, intPort, err)
		time.Sleep(1 * time.Second)

	}

	if err != nil {
		log.Errorf("Port mapping all failed from external port %d to internal port %d with %s", extPort, intPort, err)

		rand.Seed(time.Now().UnixNano())
		intPort = rand.Intn(55001) + 10000
		writeCachePort(configFilePath, intPort)
	}
}

func getCachePort() (int, error) {
	path := configFilePath
	if !Exists(path) {
		err := createCacheFile(path)
		if err != nil {
			return 0, err
		}

		writeErr := writeCachePort(path, 0)
		if writeErr != nil {
			return 0, writeErr
		}
	}

	p, readErr := readCachePort(path)

	if readErr != nil {
		return 0, readErr
	}

	if p == 0 {
		return successMapPort, nil
	}

	return p, nil
}
