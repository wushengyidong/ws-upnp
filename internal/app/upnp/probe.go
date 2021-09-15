package upnp

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"
)

const (
	maxRefreshRetries = 5
	configFilePath    = "./conf/config.yml"
)

var successMapPort = 0

func makeUPNPListener(intPort int, extPort int, nat NAT) error {

	desc := nat.(*upnpNAT).ourIP + "_TCP_" + strconv.Itoa(intPort)
	port, err := nat.AddPortMapping("tcp", extPort, intPort, desc, 0)
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

	cachePort := getCachePort()

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
		err := makeUPNPListener(intPort, extPort, nat)
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
		log.Errorf("NAT Traversal failed from external port %d to internal port %d with %s", extPort, intPort, err)
		if successMapPort != 0 && nat != nil {
			err := nat.DeletePortMapping("TCP", extPort, successMapPort)
			if err != nil {
				log.WithField("error", err).Error("Port mapping delete error")
				return
			}
		}

		rand.Seed(time.Now().UnixNano())
		intPort = rand.Intn(55001) + 10000
		writeCachePort(configFilePath, intPort)
	} else {
		log.Infof("NAT Traversal successful from external port %d to internal port %d", extPort, intPort)
	}
}

func getCachePort() int {
	p := readCachePort(configFilePath)

	if p == 0 {
		return successMapPort
	}

	return p
}

type CacheMapPort struct {
	Port int `yaml:"Port"`
}

func writeCachePort(src string, port int) {
	c := &CacheMapPort{
		Port: port,
	}
	data, err := yaml.Marshal(c)
	checkError(err)
	err = ioutil.WriteFile(src, data, 0777)
	checkError(err)
}

func readCachePort(src string) int {
	content, err := ioutil.ReadFile(src)
	checkError(err)
	c := &CacheMapPort{}
	err = yaml.Unmarshal(content, &c)
	checkError(err)
	return c.Port
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
