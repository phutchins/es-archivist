package main

import (
  "fmt"
  "es-archivist/config"
)

func main() {
  myConfig := config.New("config.json")

  fmt.Println("Got config, ES Host is: " + myConfig.ESHost)

  // Get the list of all indices
  indexList := GetIndicesList()

  fmt.Println("First element: " + indexList[0].Health)

  // Create a date ordered array of indices
  indexArray := getIndexArray(indexList)

  fmt.Println("Index array: ", indexArray)
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
