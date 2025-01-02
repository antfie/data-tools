package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
)

type yamlConfig struct {
	IsDebug                     bool     `yaml:"debug"`
	DBPath                      string   `yaml:"db_path"`
	LogFilePath                 string   `yaml:"log_file_path"`
	BatchSize                   int64    `yaml:"batch_size"`
	MaxConcurrentFileOperations int64    `yaml:"max_concurrent_file_operations"`
	FileNamesToIgnore           []string `yaml:"file_names_to_ignore"`
	FolderNamesToIgnore         []string `yaml:"folder_names_to_ignore"`
}
type Config struct {
	IsDebug                     bool
	DBPath                      string
	LogFilePath                 string
	BatchSize                   int64
	MaxConcurrentFileOperations int64
	FileNamesToIgnore           []string
	FolderNamesToIgnore         []string
}

func Load(defaultConfigData []byte) (*Config, error) {
	configFile := "config.yaml"
	_, err := os.Stat(configFile)

	if err != nil {
		log.Print("No config file found. Creating a new config file...")
		err := os.WriteFile(configFile, defaultConfigData, 0600)

		if err != nil {
			return nil, err
		}
	}

	return parseConfigFile(configFile)
}

func parseConfigFile(configFilePath string) (*Config, error) {
	yamlFile, err := os.ReadFile(path.Clean(configFilePath))

	if err != nil {
		return nil, err
	}

	config := &yamlConfig{}

	err = yaml.Unmarshal(yamlFile, config)

	if err != nil {
		return nil, err
	}

	return &Config{
		IsDebug:                     config.IsDebug,
		DBPath:                      config.DBPath,
		LogFilePath:                 config.LogFilePath,
		BatchSize:                   config.BatchSize,
		MaxConcurrentFileOperations: config.MaxConcurrentFileOperations,
		FileNamesToIgnore:           config.FileNamesToIgnore,
		FolderNamesToIgnore:         config.FolderNamesToIgnore,
	}, nil
}
