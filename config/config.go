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

  /* print config for debug
  configJson, err:= json.Marshal(config)
  if err != nil {
    panic(err)
  }

  fmt.Printf("Config: \n", string(configJson))
  */

  return config
}

func LoadConfigFile(f string) (string, Config) {
  config := Config{}

  // Set default values
  config.ESHost = "localhost:9200"
  config.MinStorageBytes = 109951162777
  config.MinFreeSpacePercent = 22
  config.SleepAfterDeleteIndexSeconds = 60
  config.SleepSeconds = 5
  config.SleepAfterMainLoopSeconds = 60
  config.MinIndexCount = 30
  config.IndexDryRun = true
  config.SnapDryRun = true

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
