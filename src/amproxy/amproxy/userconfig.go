package main

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
)

type config struct {
  Apikeys map[string]Creds `yaml:"apikeys"`
}

type Creds struct {
    SecretKey string `yaml:"secret_key"`
    Metrics map[string]uint8 `yaml:"metrics"`
}

func loadUserConfigFile(filename string) map[string]Creds {
    var f config

    yamlFile, err := ioutil.ReadFile(filename)
    err = yaml.Unmarshal(yamlFile, &f)
    if err != nil {
        panic(err)
    }

    return f.Apikeys
}
