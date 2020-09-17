package adapter

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type DatabaseConfig map[string]*Adapter

func InitializeMutipleAdapter(path string, env string, databases []string) []*Adapter {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Can't not read %v  err   #%v ", path, err)
	}

	envConfig := make(map[string]DatabaseConfig)
	err = yaml.Unmarshal(yamlFile, envConfig)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	enviroment, found := envConfig[env]
	if !found {
		panic(errors.New(fmt.Sprintf(" ========== Can not read configurations of '%s' ᕙ(⇀‸↼‶)ᕗ =========", env)))
	}

	adapters := []*Adapter{}
	for _, dbName := range databases {
		adapter, found := enviroment[dbName]
		if !found {
			panic(errors.New(fmt.Sprintf(" ========== Can not read configurations for database '%s' =========", dbName)))
		}

		adapters = append(adapters, adapter)
	}

	return adapters
}

func InitializeAdapter(path string, env string, dbName string) *Adapter {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Can't not read %v  err   #%v ", path, err)
	}

	envConfig := make(map[string]DatabaseConfig)
	err = yaml.Unmarshal(yamlFile, envConfig)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	enviroment, found := envConfig[env]
	if !found {
		panic(errors.New(fmt.Sprintf(" ========== Can not read configurations of '%s' ᕙ(⇀‸↼‶)ᕗ =========", env)))
	}

	adapter, found := enviroment[dbName]
	if !found {
		panic(errors.New(fmt.Sprintf(" ========== Can not read configurations for database '%s' =========", dbName)))
	}

	return adapter
}
