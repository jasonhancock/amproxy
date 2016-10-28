package amproxy

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type config struct {
	Apikeys map[string]Creds `yaml:"apikeys"`
}

type Creds struct {
	SecretKey string           `yaml:"secret_key"`
	Metrics   map[string]uint8 `yaml:"metrics"`
}

func LoadUserConfigFile(filename string) (map[string]Creds, time.Time, error) {
	var f config

	info, err := os.Stat(filename)
	if err != nil {
		return nil, time.Unix(0, 0), err
	}

	bytes, err := ioutil.ReadFile(filename)
	err = yaml.Unmarshal(bytes, &f)
	if err != nil {
		return nil, time.Unix(0, 0), err
	}

	return f.Apikeys, info.ModTime(), err
}
