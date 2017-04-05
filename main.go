package main

import (
  "es-archivist/config"
  "fmt"
//  "encoding/json"
  "sort"
  "time"
)

type NodesFSData struct {
  Name string
  Totals FSTotals
}

func main() {
  myConfig := config.New("config.json")

  // Watch the storage space left on each of the nodes
  watchStorageSpace(myConfig)
}

func watchStorageSpace(myConf config.Config) {
  var lowestNodeDiskPercent float64

  for {
    // Get current node stats
    nodeStats := GetNodeStats()

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


    // Log the current storage space used and free for each node in the cluster
    for _, node := range nodeFSDataArray {
      // Calculate the percentage of storage space used
      percent := float64(node.Totals.FreeInBytes) / float64(node.Totals.TotalInBytes) * float64(100)
      ipct := int(percent / float64(1))

      if ( lowestNodeDiskPercent == 0 || percent < float64(lowestNodeDiskPercent) ) {
        lowestNodeDiskPercent = percent
      }

      fmt.Printf("[%%%v] Node '%s' has %v free space left out of %v\n", ipct, node.Name, node.Totals.FreeInBytes, node.Totals.TotalInBytes)
    }

    // If storage space drops below specified level, kick off a snapshot
    //   of the oldest index
    if ( lowestNodeDiskPercent < myConf.MinFreeSpacePercent) {
      indexList := GetIndexList()
      indexArray := GetIndexArray(indexList)
      sortedIndexArray := SortIndexArray(indexArray)
      oldestIndexName := sortedIndexArray[0]

      // Take a snapshot one (or more) indices at a time
      initialResponse := TakeSnapshot(oldestIndexName)

      // If initialResponse is not OK, do something, report the failure?
      if initialResponse != "accepted" {
      }

      // Wait for the snapshot(s) to complete
      snapshotStatus := GetSnapshotStatus(oldestIndexName)

      for snapshotStatus != "IN_PROGRESS" {
        fmt.Println("Snapshot status is: " + snapshotStatus)

        time.Sleep(5 * time.Second)
      }


      // Ensure that the snapshot was successful
      // If it was not, delete the failed or partial attempt and retry

      // When completed, delete the indices that were snapshotted

      // Wait some period of time for the disk usage to stabalize
      // Then continue to watch disk usage

      //fmt.Println("Sleeping...")
    }
    time.Sleep(time.Duration(myConf.SleepSeconds) * time.Second)
  }
}

func GetIndexArray(il []IndexItem) []string {
  var indexArray []string

  for _,element := range il {
    indexArray = append(indexArray,element.Index)
  }
  return indexArray
}

func SortIndexArray(ia []string) []string {
  sort.Strings(ia)

  return ia
}
