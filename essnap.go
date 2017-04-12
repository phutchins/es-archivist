package main

import (
  "bytes"
  "net/http"
  "encoding/json"
  "es-archivist/config"
  //"io/ioutil"
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

type ESResponse struct {
  Error ESError `json:"error"`
  Type string `json:"type"`
  Reason string `json:"reason"`
  Status int `json:"status"`
}

type ESError struct {
  RootCause []RootCause `json:"root_cause"`
}

type RootCause struct {
  Type string `json:"type"`
  Reason string `json:"reason"`
}

// Get the response.body from the server and return that in some kind of generic struct

// Add a decode method on the struct that returns json given a new struct

func GetIndexList() []IndexItem {
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
    fmt.Println("Error getting node stats: ", err)
  } else {
    defer resp.Body.Close()

    //body, _ := ioutil.ReadAll(resp.Body)
    //out2, _ := json.Marshal(resp.Body)
    //fmt.Println("Response body is: ", string(body))

    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&nodeStats)

    if err != nil {
      fmt.Println("Error decoding JSON: ", err)
    }

    // Example for printing the JSON
    out, _ := json.Marshal(nodeStats)
    fmt.Println("Out", string(out))
  }

  return nodeStats
}

func TakeSnapshot(indexName string) string {
  myConfig := config.New("config.json")

  protocol := "http://"
  httpPath := "/_snapshot/" + myConfig.SnapshotRepositoryName + "/" + indexName
  jsonArgs := []byte(`{"indices": "` + indexName + `", "ignore_unavailable": true, "include_global_state": false}`)

  requestURI := protocol + myConfig.ESHost + httpPath

  req, err := http.NewRequest("PUT", requestURI, bytes.NewBuffer(jsonArgs))

  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)

  if err != nil {
    panic(err)
  } else {
    defer resp.Body.Close()

    fmt.Println("response status: ", resp.Status)
    fmt.Println("response headers: ", resp.Header)

    //body, _ := ioutil.ReadAll(resp.Body)
    //fmt.Println("response body: ", string(body))

    esResponse := ESResponse{}
    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&esResponse)

    if err != nil {
      fmt.Println("Error parsing JSON response body: ", err)
    } else {
      esResponseJson, _ := json.Marshal(esResponse)
      fmt.Println("response body: ", string(esResponseJson))

      //var rootCause RootCause
      //rootCause = esResponse.root_cause

      fmt.Println("Root Cause: ", esResponse.Error.RootCause[0].Type)
    }

    result := "unknown"
    // Get the initial response out of the returned body
    // tmp
    if resp.Status == "200" {
      result = "accepted"
    } else if resp.Status == "400" {
      result = "fail"
    } else {
      result = "unknown"
    }

    return result
  }
}

func GetSnapshotStatus(indexName string) string {
  return "ok"
}
