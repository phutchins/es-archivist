package main

import (
  "fmt"
//  "encoding/json"
  "time"
)

type NodesFSData struct {
  Name string
  Totals FSTotals
}

func main() {
  //myConfig := config.New("config.json")

  //fmt.Println("Got config, ES Host is: " + myConfig.ESHost)

  // Get the list of all indices
  //indexList := GetIndicesList()
  //fmt.Println("First element: " + indexList[0].Health)

  // Create a date ordered array of indices
  //indexArray := getIndexArray(indexList)
  //fmt.Println("Index array: ", indexArray)


  // Watch the storage space left on each of the nodes
  watchStorageSpace()
}

func watchStorageSpace() {
  for {
    nodeStats := GetNodeStats()
    // Get current node stats

    //out, _ := json.Marshal(nodeStats)
    //fmt.Printf("nodeStats: %v", string(out))

    var nodeFSDataArray []NodesFSData

    for key, element := range nodeStats.Nodes {
      var fsData NodesFSData

      fsData.Name = key
      fsData.Totals = element.FS.Total

      //data, _ := json.Marshal(element.FS.Total)

      //fmt.Printf("Adding %+v to data\n", string(data))
      nodeFSDataArray = append(nodeFSDataArray, fsData)
    }

    for _, node := range nodeFSDataArray {
      fmt.Printf("Node '%s' has %v free space left out of %v\n", node.Name, node.Totals.FreeInBytes, node.Totals.TotalInBytes)
    }

    // If storage space drops below specified level, kick off a snapshot
    //   of the oldest index

    // Take a snapshot one (or more) indices at a time
    initialResponse := snapshotOldestIndex()

    // If initialResponse is not OK, do something, report the failure?
    if initialResponse != "accepted" {
    }

    // Wait for the snapshot(s) to complete

    // Ensure that the snapshot was successful
    // If it was not, delete the failed or partial attempt and retry

    // When completed, delete the indices that were snapshotted

    // Wait some period of time for the disk usage to stabalize
    // Then continue to watch disk usage

    fmt.Println("Sleeping...")
    time.Sleep(5000 * time.Millisecond)
  }
}

func snapshotOldestIndex() string {

  return "stub no error"
}

func getIndexArray(il []IndexItem) []string {
  var indexArray []string

  for _,element := range il {
    indexArray = append(indexArray,element.Index)
  }
  return indexArray
}

func sortIndexArray(ia []string) []string {
  var sortedArray []string
  return sortedArray
}
