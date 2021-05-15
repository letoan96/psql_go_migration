package adapter

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type DatabaseConfig map[string]*Adapter

func InitializeMultipleAdapter(path string, env string, databases []string) map[string]*Adapter {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Can't not read %v err #%v ", path, err)))
	}

	envConfig := make(map[string]DatabaseConfig)
	if err := yaml.Unmarshal(yamlFile, envConfig); err != nil {
		panic(errors.New(fmt.Sprintf("Unmarshal: %v", err)))
	}

	environment, found := envConfig[env]
	if !found {
		e := fmt.Sprintf(" ==== Can not read configurations for '%s' database ====", env)
		panic(errors.New(e))
	}

	adapters := make(map[string]*Adapter)
	if databases == nil || len(databases) == 0 {
		return environment
	}

	for _, dbName := range databases {
		if adapters[dbName], found = environment[dbName]; !found {
			e := fmt.Sprintf(" ==== Can not read configurations for database '%s' ====", dbName)
			panic(errors.New(e))
		}
	}

	return adapters
}
