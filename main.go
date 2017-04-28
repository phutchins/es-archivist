package main

import (
  "es-archivist/config"
  "log"
  "fmt"
  "io/ioutil"
  "os"
//  "encoding/json"
  "regexp"
  "sort"
  "strconv"
  "time"
)

type NodesFSData struct {
  Name string
  Totals FSTotals
}

func main() {
  myConfig := config.New("config.json")
  sleepAfterMainLoopSeconds := myConfig.SleepAfterMainLoopSeconds
  InitLogger(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

  // Watch the storage space left on each of the nodes
  // Have watchStorageSpace return some sort of error or message?
  // loop here?
  for {
    err := watchStorageSpace(myConfig)
    Info.Printf("%v\n", err)

    time.Sleep(time.Duration(sleepAfterMainLoopSeconds) * time.Second)
  }
}

func watchStorageSpace(myConf config.Config) string {
  // Main loop to monitor disk usage
  for {
    var nodeFSDataArray []NodesFSData
    var lowestNodeDiskPercent float64

    // Get current node stats
    nodeStats := GetNodeStats(myConf)
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
    if ( lowestNodeDiskPercent < myConf.MinFreeSpacePercent ) {
      var snapshotStatus string
      indexList := GetIndexList(myConf)
      indexArray := GetIndexArray(indexList)
      filteredIndexArray := Filter(indexArray, ContainsPrefixFilter, myConf.IndexIncludePrefix)
      sortedIndexArray := SortIndexArray(filteredIndexArray)
      indexCount := len(indexArray)

      //fmt.Printf("Total index count is %d\n", int(indexCount))

      if indexCount <= myConf.MinIndexCount {
        return fmt.Sprintf("Current index count of %d is less than MinIndexCount of %d. Not doing anything...", indexCount, myConf.MinIndexCount)
      }

      fmt.Printf("Free space of %v is less than the configured minimum of %v.\n",
        lowestNodeDiskPercent,
        myConf.MinFreeSpacePercent,
      )

      fmt.Printf("Starting oldest index archival\n")

      //fmt.Printf("indexList is: %v\n", indexList)
      //fmt.Printf("indexArray is: %v\n", indexArray)
      //fmt.Printf("sortedIndexArray is: %v\n", sortedIndexArray)

      if len(sortedIndexArray) == 0 {
        return "No indices available to archive."
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
        initialResponse := TakeSnapshot(myConf, oldestIndexName)

        if initialResponse == "accepted" {
          moveAlong = true
          fmt.Printf("Snapshot request has been accepted\n")
        } else if initialResponse == "fail_name_in_use_exception" {
          // Check if the snapshot was successful
          fmt.Println("Failed to start snapshot: Snapshot name in use")
          moveAlong = true
        } else if initialResponse == "concurrent_snapshot_execution_exception" {
          fmt.Println("Failed to start snapshot: Snapshot is already in progress")
          moveAlong = true
        } else {
          return fmt.Sprintf("Unhandled response from TakeSnapshot: %v\n", initialResponse)
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
        snapshotStatus = GetSnapshotStatus(myConf, oldestIndexName)
        fmt.Println("Snapshot status is: " + snapshotStatus)


        // Need to handle INTERNAL_SERVER_ERROR ?

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

          deleteSnapResult, err := DeleteSnapshot(myConf, oldestIndexName)

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

      deleteSuccess, err := DeleteIndex(myConf, oldestIndexName)

      if err != nil {
        panic(err)
      }

      if deleteSuccess {
        fmt.Println("Horray! The index was deleted successfully. We're done here.")
        fmt.Printf("Sleeping for %s seconds to wait for space to be freed up", myConf.SleepAfterDeleteIndexSeconds)

        // Possibly expunge deleted indices here
        // curl -XPOST 'localhost:9200/_optimize?only_expunge_deletes=true'

        time.Sleep(time.Duration(myConf.SleepAfterDeleteIndexSeconds) * time.Second)
      }
      // Wait some period of time for the disk usage to stabalize
      // Then continue to watch disk usage

      //fmt.Println("Sleeping...")
    } else {
      return fmt.Sprintf("Highest node disk usage of %%%v is not below the threshold of %%%v. Continuing to watch...", lowestNodeDiskPercent, myConf.MinFreeSpacePercent)
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

type ByLsTimeStamp []string

func (s ByLsTimeStamp) Len() int {
  return len(s)
}

func (s ByLsTimeStamp) Swap(i, j int) {
  s[i], s[j] = s[j], s[i]
}

func (s ByLsTimeStamp) Less(i, j int) bool {
  // Get the date substring
  rDate, err := regexp.Compile(".*-([0-9]{4}.[0-9]{2}.[0-9]{2})")

  if err != nil {
    log.Fatal(err)
  }

  rSep, err := regexp.Compile("[.]")

  if err != nil {
    log.Fatal(err)
  }

  //fmt.Printf("s[i]: %v s[j]: %v\n", s[i], s[j])
  //fmt.Println(rDate.FindStringSubmatch(s[i]))

  iDate := rDate.FindStringSubmatch(s[i])
  jDate := rDate.FindStringSubmatch(s[j])

  //fmt.Printf("iDate: %s jDate: %s\n", iDate[1], jDate[1])

  // Remove period separator from date so that we can convert to int
  iDateString := rSep.ReplaceAllString(iDate[1], "")
  jDateString := rSep.ReplaceAllString(jDate[1], "")

  iDateInt, _ := strconv.Atoi(iDateString)
  jDateInt, _ := strconv.Atoi(jDateString)

  //fmt.Printf("iDateInt: %d jDateInt: %d\n", iDateInt, jDateInt)

  return iDateInt < jDateInt
}

func Filter(s []string, fn func(string, []string) bool, pfx []string) []string {
  var p []string
  for _, v := range s {
    if fn(v, pfx) {
      p = append(p, v)
    }
  }
  return p
}

func ContainsPrefixFilter(s string, p []string) bool {
  var containsPrefix bool

  for _, prefix := range p {
    r, _ := regexp.Compile(".*" + prefix + ".*")
    match := r.MatchString(s)

    if match == true {
      containsPrefix = true
    }
  }

  return containsPrefix
}

func SortIndexArray(ia []string) []string {
  sort.Sort(ByLsTimeStamp(ia))

  return ia
}
