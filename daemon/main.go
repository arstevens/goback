package main

import (
  "github.com/arstevens/goback/daemon/reflector"
  "github.com/arstevens/goback/daemon/processor"
  "github.com/arstevens/goback/daemon/interactor"
  "path/filepath"
  "os/user"
  "log"
)

const (
  PlainReflectorCode processor.ReflectorCode = "pref"
)

var MetadataDBFile string = ".gobackdb"
var GobackPort int = 25000

func main() {
  refTypes := map[processor.ReflectorCode]interactor.ReflectorCreator{
    PlainReflectorCode: reflector.NewPlainReflector,
  }
  generator := interactor.NewReflectionGenerator(refTypes)

  curUser, err := user.Current()
  if err != nil {
    log.Fatalf("Failed to retrieve current user in main(): %v", err)
  }
  dbFile := filepath.Join(curUser.HomeDir, MetadataDBFile)
  mdb := NewFileMetadataDB(dbFile)

  uiChan := make(chan string)
  sysChan := make(chan string)
  go processor.CommandProcessor(generator, mdb, uiChan, sysChan)
  go processor.MonitorSystem(mdb, sysChan)
  go ListenAndRelay(GobackPort, uiChan)

  done := make(chan struct{})
  <-done
}
