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
  // Have watchStorageSpace return some sort of error or message?
  // loop here?
  for {
    err := watchStorageSpace(myConfig)
    fmt.Printf("Error while watching storage space: %v\n", err)
    fmt.Printf("Starting watch again in 5 seconds...\n")

    time.Sleep(5 * time.Second)
  }
}

func watchStorageSpace(myConf config.Config) string {
  var lowestNodeDiskPercent float64

  // Main loop to monitor disk usage
  for {
    var nodeFSDataArray []NodesFSData

    // Get current node stats
    nodeStats := GetNodeStats()
    //out, _ := json.Marshal(nodeStats)
    //fmt.Printf("nodeStats: %v", string(out))

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

      // Set lowestNodeDiskPercent with the percentage of the least disk space free
      if ( lowestNodeDiskPercent == 0 || percent < float64(lowestNodeDiskPercent) ) {
        lowestNodeDiskPercent = percent
      }

      // We will probably want to determine the percentage from storage space alotted to
      // es so that we dont' continue deleting indices due to disk usage from some other
      // process or application. We could set this in our config here.
      fmt.Printf("[%%%v] Node '%s' has %v free space left out of %v\n",
        ipct,
        node.Name,
        node.Totals.FreeInBytes,
        node.Totals.TotalInBytes,
      )
      //fmt.Printf("lowestNodeDiskPercent is [%%%v]\n", lowestNodeDiskPercent)
    }

    // If storage space drops below specified level, kick off a snapshot
    //   of the oldest index
    if ( lowestNodeDiskPercent < myConf.MinFreeSpacePercent) {
      //var initialResponse string
      var snapshotStatus string
      indexList := GetIndexList()
      indexArray := GetIndexArray(indexList)
      sortedIndexArray := SortIndexArray(indexArray)

      fmt.Printf("Free space of %v is less than the configured minimum of %v.\n",
        lowestNodeDiskPercent,
        myConf.MinFreeSpacePercent,
      )

      fmt.Printf("Starting oldest index archival\n")

      //fmt.Printf("indexList is: %v\n", indexList)
      //fmt.Printf("indexArray is: %v\n", indexArray)
      fmt.Printf("sortedIndexArray is: %v\n", sortedIndexArray)

      if len(sortedIndexArray) == 0 {
        fmt.Printf("Sorted index list is empty so unable to take snapshot\n")
        return "Sorted index array is empty :("
      }

      oldestIndexName := sortedIndexArray[0]

      // Take a snapshot one index at a time
      fmt.Printf("Oldest index is %v. Taking snapshot.\n", oldestIndexName)

      var moveAlong bool

      // Keep trying to take the snapshot until it is successful handling error cases
      // Possible results:
      // - accepted: continue
      // - fail_name_in_use_exception: check status of snapshot and delete if not complete
      //     if complete, continue on and delete the index
      moveAlong = false
      for moveAlong != true {
        initialResponse := TakeSnapshot(oldestIndexName)

        if initialResponse == "accepted" {
          moveAlong = true
          fmt.Printf("Snapshot request has been accepted\n")
        } else if initialResponse == "fail_name_in_use_exception" {
          // Check if the snapshot was successful
          fmt.Println("Failed to start snapshot: Snapshot name in use")
          moveAlong = true
        }

        if moveAlong != true {
          time.Sleep(5 * time.Second)
        }
      }

      // Wait for the snapshot(s) to complete successfully
      // Possible results:
      // - SUCCESS: Great, move along and delete the index that we archived
      // - IN_PROGRESS: Keep waiting until the status changes...
      // - ? FAIL ?
      moveAlong = false
      for moveAlong != true {
        snapshotStatus = GetSnapshotStatus(oldestIndexName)
        fmt.Println("Snapshot status is: " + snapshotStatus)

        if snapshotStatus == "SUCCESS" {
          moveAlong = true
        } else if snapshotStatus == "IN_PROGRESS" {
          fmt.Println("Snapshot is in progress, please wait...")
        } else if snapshotStatus == "PARTIAL" {
          fmt.Println("Snapshot finished with a PARTIAL result. Deleting and trying again")
          if myConf.SnapDryRun == true {
            fmt.Println("DRYRUN: Would have deleted snapshot " + oldestIndexName)
            return "snapdryrun"
          }

          deleteSnapResult, err := DeleteSnapshot(oldestIndexName)

          fmt.Println("Deleting snapshot resulted in response: " + deleteSnapResult)

          if err != nil {
            // failed to delete snapshot, freak out and run around
            panic(err)
          }

          if deleteSnapResult == "OK" {

          } else {
            // attempt delete again?
            return "Error deleting snapshot, not sure how to handle"
          }
        } else {
          fmt.Println("We have an unhandled snapshot status of " + snapshotStatus)
          // Delete the snapshot and try again here?
        }

        if moveAlong != true {
          time.Sleep(5 * time.Second)
        }
      }

      // When completed, delete the indices that were snapshotted
      if myConf.IndexDryRun == true {
        fmt.Println("DRYRUN: Would have deleted index " + oldestIndexName)
        return "indexdryrun"
      }

      deleteSuccess, err := DeleteIndex(oldestIndexName)

      if err != nil {
        panic(err)
      }

      if deleteSuccess {
        fmt.Println("Horray! The index was deleted successfully. We're done here.")
      }
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
    //fmt.Printf("Adding index %v to indexArray\n", element.Index)
  }
  return indexArray
}

func SortIndexArray(ia []string) []string {
  sort.Strings(ia)

  return ia
}
