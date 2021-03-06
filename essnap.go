package main

import (
  "bytes"
  "encoding/json"
  "errors"
  "es-archivist/config"
  "fmt"
  //"io/ioutil"
  "log"
  "net/http"
  "regexp"
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
  Snapshots []ESSnapshot `json:"snapshots"`
}

type ESSnapshot struct {
  Snapshot string `json:"snapshot"`
  Uuid string `json:"uuid"`
  VersionId int `json:"version_id"`
  Version string `json:"version"`
  Indices []string `json:"indices"`
  State string `json:"state"`
  StartTime string `json:"start_time"`
  StartTimeInMillis int `json:"start_time_in_millis"`
  EndTime string `json:"end_time"`
  EndTimeInMillis int `json:"end_time_in_millis"`
  DurationInMillis int `json:"duration_in_millis"`
  Failures []ESFailure `json:"failures'`
  Shards ESSnapShards `json:"shards"`
}

type ESFailure struct {
  Index string `json:"index"`
  IndexUUID string `json:"index_uuid"`
  ShardID int `json:"shard_id"`
  Reason string `json:"reason"`
  NodeID string `json:"node_id"`
  Status string `json:"status"`
}

type ESSnapShards struct {
  Total int `json:"total"`
  Failed int `json:"failed"`
  Successful int `json:"successful"`
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

func GetIndexList(myConfig config.Config) []IndexItem {
  indexList := []IndexItem{}

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

func GetNodeStats(myConfig config.Config) NodeStats {
  nodeStats := NodeStats{}

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
    //out, _ := json.Marshal(nodeStats)
    //fmt.Println("Out", string(out))
  }

  return nodeStats
}

func TakeSnapshot(myConfig config.Config, indexName string) string {

  protocol := "http://"
  httpPath := "/_snapshot/" + myConfig.SnapshotRepositoryName + "/" + indexName
  jsonArgs := []byte(`{"indices": "` + indexName + `", "ignore_unavailable": true, "include_global_state": false}`)
  requestURI := protocol + myConfig.ESHost + httpPath

  req, err := http.NewRequest("PUT", requestURI, bytes.NewBuffer(jsonArgs))

  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)

  if err != nil {
    // Return error here instead and retry
    panic(err)
  } else {
    defer resp.Body.Close()

    //fmt.Println("response status: ", resp.Status)
    //fmt.Println("response headers: ", resp.Header)

    //body, _ := ioutil.ReadAll(resp.Body)
    //fmt.Println("response body: ", string(body))

    esResponse := ESResponse{}
    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&esResponse)

    var errorString string
    var errorReason string
    var statusCode int

    if err != nil {
      fmt.Println("Error parsing JSON response body: ", err)
    } else {
      //esResponseJson, _ := json.Marshal(esResponse)

      if esResponse.Error.RootCause != nil {
        //fmt.Println("response body: ", string(esResponseJson))

        errorString = esResponse.Error.RootCause[0].Type
        errorReason = esResponse.Error.RootCause[0].Reason
        statusCode = resp.StatusCode

        //fmt.Println("Root Cause: ", errorString)
      }
    }

    result := "unknown"
    // Get the initial response out of the returned body

    rNameAlreadyExists, err := regexp.Compile(".*name already exists.*")

    if err != nil {
      fmt.Printf("Failed to create regex: %v", err)
      result = "fail"
    }

    if resp.StatusCode == 200 {
      result = "accepted"
    } else if errorString == "invalid_snapshot_name_exception" {
      result = "fail_name_in_use_exception"
      errorMessage := "Snapshot name is already in use"
      fmt.Println("Fail: ", errorMessage)

    } else if rNameAlreadyExists.MatchString(errorReason) {
      result = "fail_name_in_use_exception"
      //errorMessage := "Snapshot name is already in use"
      //fmt.Println("Fail: ", errorMessage)
    } else if errorString == "concurrent_snapshot_execution_exception" {
      result = "concurrent_snapshot_execution_exception"
    } else {
      fmt.Printf("Status code [%v] - Unknown response: %v\n", statusCode, errorString)
      result = "fail"
    }

    return result
  }
}

func GetSnapshotStatus(myConfig config.Config, indexName string) string {
  protocol := "http://"
  httpPath := "/_snapshot/" + myConfig.SnapshotRepositoryName + "/" + indexName
  requestURI := protocol + myConfig.ESHost + httpPath

  resp, err := http.Get(requestURI)

  if err != nil {
    fmt.Println("Caught error: ", err)
    return "Error while executing http request"
  } else {
    defer resp.Body.Close()

    // When uncommented, this will break the JSON decoder due to resp.Body being at EOF
    //body, _ := ioutil.ReadAll(resp.Body)
    //fmt.Printf("Snapshot response body: %v\n", string(body))

    esResponse := ESResponse{}
    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&esResponse)

    //esResponseJson, _ := json.Marshal(esResponse)
    //fmt.Println("Snapshot Response json is: ", string(esResponseJson))

    var snapState string

    if err != nil {
      fmt.Println("Error parsing JSON response body for SnapshotStatus: ", err)
      return "Error while parsing JSON"
    } else {
      snapState = esResponse.Snapshots[0].State
      return snapState
    }
  }
}

func DeleteSnapshot(myConfig config.Config, snapshotName string) (string, error) {
  var deleteSnapResult string
  client := &http.Client{}

  protocol := "http://"
  httpPath := "/_snapshot/" + myConfig.SnapshotRepositoryName + "/" + snapshotName
  requestURI := protocol + myConfig.ESHost + httpPath

  req, err := http.NewRequest("DELETE", requestURI, nil)
  resp, err := client.Do(req)

  if err != nil {
    fmt.Println("Caught error: ", err)
    return "fail", errors.New("Error while executing http request")
  } else {
    defer resp.Body.Close()

    esResponse := ESResponse{}
    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&esResponse)


    if err != nil {
      fmt.Println("Error parsing JSON response body for deleteStatus: ", err)
      return "fail", errors.New("Error while parsing JSON")
    }

    //esResponseJson, _ := json.Marshal(esResponse)
    //deleteSnapResult = esResponse.Snapshots[0].State
  }

  return deleteSnapResult, nil
}

func DeleteIndex(myConfig config.Config, indexName string) (bool, error) {
  deleteSuccess := false

  client := &http.Client{}

  protocol := "http://"
  httpPath := "/" + indexName
  requestURI := protocol + myConfig.ESHost + httpPath

  req, err := http.NewRequest("DELETE", requestURI, nil)
  resp, err := client.Do(req)

  if err != nil {
    fmt.Println("Caught error: ", err)
    return deleteSuccess, errors.New("Error while executing http request")
  } else {
    defer resp.Body.Close()

    esResponse := ESResponse{}
    decoder := json.NewDecoder(resp.Body)
    err := decoder.Decode(&esResponse)

    if err != nil {
      fmt.Println("Error parsing JSON response body for deleteIndexResult: ", err)
      return deleteSuccess, errors.New("Error while parsing JSON")
    }

    fmt.Println("Delete index response status is " + resp.Status)
    if resp.Status == "200" {
      deleteSuccess = true
    }
  }

  return deleteSuccess, nil
}
