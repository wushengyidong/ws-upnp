package upnp

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type CacheMapPort struct {
	Port int `yaml:"Port"`
}

func writeCachePort(src string, port int) error {
	c := &CacheMapPort{
		Port: port,
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(src, data, 0777)
	if err != nil {
		return err
	}

	return nil
}

func readCachePort(src string) (int, error) {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return 0, err
	}
	c := &CacheMapPort{}
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		return 0, err
	}
	return c.Port, nil
}

func createCacheFile(src string) error {
	_, err := os.Create(src)
	if err != nil {
		return errors.New(fmt.Sprintf("Create file error: %v", err))
	}
	return nil
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
