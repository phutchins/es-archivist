package config

// Automatically archive the oldest indices in elasticsearch

import (
  "encoding/json"
  "os"
  "fmt"
)

// Elasticsearch Configuration
type Config struct {
  ESHost string
  MinStorageBytes int
}

func New(c string) Config {
  configFileName := c
  configFile, _ := os.Open(configFileName)
  decoder := json.NewDecoder(configFile)
  config := Config{}
  err := decoder.Decode(&config)

  if err != nil {
    fmt.Println("error decoding config file: ", err)
  }

  fmt.Println(config.ESHost)

  return config
}
