package main

import (
  "net/http"
  "encoding/json"
  "es-archivist/config"
  "fmt"
  "log"
)

type IndexItem struct {
  Health string
  Status string
  Index string
  Pri string
  Rep string
  DocsCount string `json:"docs.count"`
  DocsDeleted string `json:"docs.deleted"`
  StoreSize string `json:"store.size"`
  PriStoreSize string `json:"pri.store"`
}

type NodeStats struct {
  NodesInfo NodeInfo
  ClusterName string `json:"cluster_name"`
  Nodes map[string]Node
}

type Node struct {
  Timestamp int
  Name string
  FS FS
}

type FS struct {
  Timestamp int
  Total FSTotals
}

type FSTotals struct {
  TotalInBytes int `json:"total_in_bytes"`
  FreeInBytes int `json:"free_in_bytes"`
  AvailableInBytes int `json:"available_in_bytes"`
}

type NodeInfo struct {
  Total int
  Successful int
  Failed int
}

type Uri struct {
  Proto string
  Path string
  Settings string
}

type ApiResponse struct {
  Body string
}

// Get the response.body from the server and return that in some kind of generic struct

// Add a decode method on the struct that returns json given a new struct

func GetIndicesList() []IndexItem {
  indexList := []IndexItem{}

  myConfig := config.New("config.json")

  protocol := "http://"
  httpPath := "/_cat/indices/*"
  settings := "?format=json&pretty"

  requestURI := protocol + myConfig.ESHost + httpPath + settings

  resp, err := http.Get(requestURI)

  if err != nil {
    fmt.Println("Error getting index list: ", err)
  } else {
    defer resp.Body.Close()

    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&indexList)

    if err != nil {
      fmt.Println("Error decoding JSON: ", err)
    }

    if err != nil {
      log.Fatal(err)
    }
  }

  return indexList
}

func GetNodeStats() NodeStats {
  nodeStats := NodeStats{}

  myConfig := config.New("config.json")

  protocol := "http://"
  httpPath := "/_nodes/stats/fs"
  settings := "?format=json&pretty"

  requestURI := protocol + myConfig.ESHost + httpPath + settings

  resp, err := http.Get(requestURI)

  if err != nil {
    fmt.Println("Error getting index list: ", err)
  } else {
    defer resp.Body.Close()

    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&nodeStats)

    if err != nil {
      fmt.Println("Error decoding JSON: ", err)
    }

    // Example for printing the JSON
    //out, _ := json.Marshal(nodeStats)
    //fmt.Println("Out", string(out))
  }

  return nodeStats
}
