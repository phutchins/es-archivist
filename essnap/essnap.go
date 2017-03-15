package essnap

import (
  "net/http"
  "es-archivist/config"
  "encoding/json"
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

type IndexList struct {
  Collection []IndexItem
}

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
