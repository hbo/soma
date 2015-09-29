package main

import (
  "bytes"
  "encoding/json"
  "github.com/nahanni/go-ucl"
  "io/ioutil"
  "log"
)

type Config struct {
  Environment string `json:"environment"`
  Timeout     string `json:"timeout"`
  TlsMode     string `json:"tlsmode"`
  Database    DbConfig `json:"database"`
}

type DbConfig struct {
  Host string `json:"host"`
  User string `json:"user"`
  Name string `json:"name"`
  Port string `json:"port"`
  Pass string `json:"password"`
}

func (c *Config) populateFromFile(fname string) error {
  file, err := ioutil.ReadFile(fname)
  if err != nil {
    return err
  }

  log.Printf( "Loading configuration from %s", fname)

  // UCL parses into map[string]interface{}
  fileBytes := bytes.NewBuffer([]byte(file))
  parser := ucl.NewParser(fileBytes)
  uclData, err := parser.Ucl()
  if err != nil {
    log.Fatal("UCL error: ", err)
  }

  // take detour via JSON to load UCL into struct
  uclJson, err := json.Marshal(uclData)
  if err != nil {
    log.Fatal(err)
  }
  json.Unmarshal([]byte(uclJson), &c)

  return nil
}
