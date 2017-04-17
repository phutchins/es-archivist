package config

// Automatically archive the oldest indices in elasticsearch

import (
  "encoding/json"
  "os"
  "fmt"
  "log"
)

// Elasticsearch Configuration
type Config struct {
  ESHost string
  MinStorageBytes int
  SleepSeconds int
  SleepAfterMainLoopSeconds int
  SleepAfterDeleteIndexSeconds int
  MinFreeSpacePercent float64
  SnapshotRepositoryName string
  MinIndexCount int
  SnapDryRun bool
  IndexDryRun bool
  IndexIncludePrefix []string
  Logger log.Logger
}

func New(c string) Config {
  configFileName := c
  configFileDefaultDir := "/etc/esa/"
  config := Config{}
  var configFileFound bool

  // Try current directory for config file first, then default dir
  if _, err := os.Stat(configFileName); err == nil {
    configFileFound = true
    _, config = LoadConfigFile(configFileName)
  } else if _, err := os.Stat(configFileDefaultDir + configFileName); err == nil {
    configFileFound = true
    _, config = LoadConfigFile(configFileDefaultDir + configFileName)
  }

  if configFileFound == false {
    fmt.Printf("No config file found at %s. Using defaults\n", configFileName)
  }

  if config.ESHost == "" {
    config.ESHost = "localhost:9200"
  }

  if config.MinStorageBytes == 0 {
    config.MinStorageBytes = 109951162777
  }

  if config.SleepSeconds == 0 {
    config.SleepSeconds = 5
  }

  if config.SleepAfterDeleteIndexSeconds == 0 {
    config.SleepAfterDeleteIndexSeconds = 60
  }

  if config.SleepAfterMainLoopSeconds == 0 {
    config.SleepAfterMainLoopSeconds = 60
  }

  // Set defaults for MinIndexCount and DryRun here

  //fmt.Println(config.ESHost)

  return config
}

func LoadConfigFile(f string) (string, Config) {
  config := Config{}

  fmt.Printf("Loading config from file %v\n", f)

  configFile, _ := os.Open(f)
  decoder := json.NewDecoder(configFile)
  err := decoder.Decode(&config)

  if err != nil {
    fmt.Println("error decoding config file: ", err)
    return fmt.Sprintf("Error decoding config file: ", err), config
  }

  fmt.Printf("Successfully loaded config file %v\n", f)

  return "", config
}
