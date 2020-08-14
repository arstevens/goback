package processor

import (
  "strings"
  "strconv"
  "bufio"
  "log"
  "net"
  "fmt"
  "io"
)

type CommandCode string

const (
  BackupCommand CommandCode = "bak"
  NewBackupCommand = "n_bak"
  UpdateCommand = "upd"
)

const (
  RenameCommand = "ren"
  WriteCommand = "wrt"
  DeleteCommand = "del"
  CreateCommand = "cre"
)

const (
  dirTrue = "true"
)

func CommandProcessor(gen Generator, mdb MetadataDB, comChan chan string, updateChan <-chan string) {
  for {
    select {
      case cmd, ok := <-updateChan:
        if !ok {
          return
        }
        if err := executeCommand(cmd, gen, mdb); err != nil {
          log.Printf("Failed to execute command in CommandProcessor: %v\n", err)
        }
      case cmd, ok := <-comChan:
        if !ok {
          return
        }
        if err := executeCommand(cmd, gen, mdb); err != nil {
          log.Printf("Failed to execute command(%s) in CommandProcessor: %v\n", cmd, err)
          comChan<-FailCode
        } else {
          comChan<-SuccessCode
        }
    }
  }
}

/* Command format: command_code:param1,param2,... */
func executeCommand(cmd string, gen Generator, mdb MetadataDB) error {
  cmdComponents := strings.Split(cmd, ":")
  if len(cmdComponents) < 2 {
    return fmt.Errorf("Invalid command input(%s) in executeCommand()", cmd)
  }
  cmdType := CommandCode(cmdComponents[0])
  params := strings.Split(cmdComponents[1], ",")
  var err error

  switch cmdType {
    case BackupCommand:
      err = backupCommand(params, gen, mdb)
    case NewBackupCommand:
      err = newBackupCommand(params, gen, mdb)
    case UpdateCommand:
      err = updateCommand(params, gen, mdb)
    default:
      return fmt.Errorf("Unknown command(%s) in executeCommand()", cmd)
  }

  if err != nil {
    return fmt.Errorf("Couldn't process command in executeCommand(): %v", err)
  }
  return nil
}

func backupCommand(params []string, gen Generator, mdb MetadataDB) error {
  if len(params) < 1 {
    return fmt.Errorf("Not enough params in backupCommand()")
  }
  backupRoot := params[0]
  mdbRow, err := mdb.GetRow(backupRoot)
  if err != nil {
    return fmt.Errorf("Couldn't retrieve row in backupCommand(): %v", err)
  }

  origCm, err := gen.OpenChangeMap(mdbRow.CMCode, mdbRow.OriginalRoot, mdbRow.OriginalCM)
  if err != nil {
    return fmt.Errorf("Couldn't open change map at %s with code %s in backupCommand(): %v", backupRoot, mdbRow.CMCode, err)
  }
  refCm, err := gen.OpenChangeMap(mdbRow.CMCode, mdbRow.ReflectionRoot, mdbRow.ReflectionCM)
  if err != nil {
    return fmt.Errorf("Couldn't open change map at %s with code %s in backupCommand(): %v", backupRoot, mdbRow.CMCode, err)
  }

  reflector, err := gen.Reflect(mdbRow.ReflectionCode, origCm, refCm)
  if err != nil {
    return fmt.Errorf("Couldn't reflect %s in backupCommand(): %v", backupRoot, err)
  }
  err = reflector.Backup()
  if err != nil {
    return fmt.Errorf("Failed to backup %s in backupCommand(): %v", backupRoot, err)
  }

  mdbRow.ReflectionCM = mdbRow.OriginalCM
  _, err = mdb.DeleteRow(mdbRow.OriginalCM)
  if err != nil {
    return fmt.Errorf("Failed to delete stale row in backupCommand(): %v", err)
  }

  err = mdb.InsertRow(mdbRow.OriginalRoot, mdbRow)
  if err != nil {
    return fmt.Errorf("Failed to insert updated row in backupCommand(): %v", err)
  }
  return nil
}

func newBackupCommand(params []string, gen Generator, mdb MetadataDB) error {
  if len(params) < 4 {
    return fmt.Errorf("Not enough paramaters in newBackupCommand()")
  }
  origRoot, refRoot := params[0], params[1]
  refCode := ReflectorCode(params[2])
  cmCode := ChangeMapCode(params[3])

  origCm, err := gen.NewChangeMap(cmCode, origRoot)
  if err != nil {
    return fmt.Errorf("Couldn't create new change map in newBackupCommand(): %v", err)
  }
  refCm, err := gen.NewChangeMap(cmCode, refRoot)
  if err != nil {
    return fmt.Errorf("Couldn't create new change map in newBackupCommand(): %v", err)
  }

  reflector, err := gen.Reflect(refCode, origCm, refCm)
  if err != nil {
    return fmt.Errorf("Couldn't reflect in newBackupCommand(): %v", err)
  }
  err = reflector.Backup()
  if err != nil {
    return fmt.Errorf("Couldn't backup in newBackupCommand(): %v", err)
  }

  origSerial := origCm.Serialize()
  refSerial := refCm.Serialize()

  mdbRow := MDBRow{
    OriginalRoot: origRoot,
    ReflectionRoot: refRoot,
    OriginalCM: origSerial,
    ReflectionCM: refSerial,
    ReflectionCode: refCode,
    CMCode: cmCode,
  }
  err = mdb.InsertRow(origRoot, mdbRow)
  if err != nil {
    fmt.Errorf("Couldnt insert row in newBackupCommand(): %v", err)
  }

  return nil
}

func updateCommand(params []string, gen Generator, mdb MetadataDB) error {
  if len(params) < 3 {
    return fmt.Errorf("Not enough params in updateCommand()")
  }

  origRoot := params[1]
  mdbRow, err := mdb.GetRow(origRoot)
  if err != nil {
    return fmt.Errorf("Could not get row with key %s in updateCommand(): %v", origRoot, err)
  }

  origCm, err := gen.OpenChangeMap(mdbRow.CMCode, origRoot, mdbRow.OriginalCM)
  if err != nil {
    return fmt.Errorf("Failed to open change map in updateCommand(): %v", err)
  }

  switch (params[0]) {
  case RenameCommand:
    err = renameUpdateCommand(params[2], params[3], origCm)
    if err != nil {
      return fmt.Errorf("Failed to complete rename in updateCommand(): %v", err)
    }
  case WriteCommand:
    err = writeUpdateCommand(params[2], origCm)
    if err != nil {
      return fmt.Errorf("Failed to complete write in updateCommand(): %v", err)
    }
  case DeleteCommand:
    dir := dirTrueToBool(params[3])
    err = deleteUpdateCommand(params[2], dir, origCm)
    if err != nil {
      return fmt.Errorf("Failed to complete delete in updateCommand(): %v", err)
    }
  case CreateCommand:
    dir := dirTrueToBool(params[3])
    err = createUpdateCommand(params[2], dir, origCm)
    if err != nil {
      return fmt.Errorf("Failed to complete create in updateCommand(): %v", err)
    }
  }

  mdbRow.OriginalCM = origCm.Serialize()
  _, err = mdb.DeleteRow(origRoot)
  if err != nil {
    return fmt.Errorf("Failed to delete row in updateCommand(): %v", err)
  }

  err = mdb.InsertRow(origRoot, mdbRow)
  if err != nil {
    return fmt.Errorf("Failed to insert row in updateCommand(): %v", err)
  }
  return nil
}

func dirTrueToBool(s string) bool {
  if s == dirTrue {
    return true
  }
  return false
}

func createUpdateCommand(path string, dir bool, cm ChangeMap) error {
  fileUpdates := make([][]string, 3)
  dirUpdates := make([][]string, 3)
  if !dir {
    fileUpdates = [][]string{CreateCode: []string{path}, RenameCode:[]string{}}
  } else {
    dirUpdates = [][]string{CreateCode: []string{path}, RenameCode:[]string{}}
  }

  err := cm.Update(fileUpdates, dirUpdates)
  if err != nil {
    return fmt.Errorf("Couldn't delete in deleteUpdateCommand(): %v", err)
  }
  return nil
}

func deleteUpdateCommand(path string, dir bool, cm ChangeMap) error {
  fileUpdates := make([][]string, 3)
  dirUpdates := make([][]string, 3)
  if !dir {
    fileUpdates = [][]string{DeleteCode: []string{path}, RenameCode:[]string{}}
  } else {
    dirUpdates = [][]string{DeleteCode: []string{path}, RenameCode:[]string{}}
  }

  err := cm.Update(fileUpdates, dirUpdates)
  if err != nil {
    return fmt.Errorf("Couldn't delete in deleteUpdateCommand(): %v", err)
  }
  return nil
}

func renameUpdateCommand(path string, newName string, cm ChangeMap) error {
  fileUpdates := [][]string{RenameCode: []string{path+","+newName}}
  dirUpdates := make([][]string, 3)

  err := cm.Update(fileUpdates, dirUpdates)
  if err != nil {
    return fmt.Errorf("Couldn't rename in renameUpdateCommand(): %v", err)
  }
  return nil
}

func writeUpdateCommand(writePath string, cm ChangeMap) error {
  fileUpdates := [][]string{DeleteCode: []string{writePath}, CreateCode: []string{writePath}, RenameCode:[]string{}}
  dirUpdates := make([][]string, 3)

  err := cm.Update(fileUpdates, dirUpdates)
  if err != nil {
    return fmt.Errorf("Failed to update in writeUpdateCommand(): %v", err)
  }
  return nil
}

/* listenAndRelay() connects and communicates with anyone
on the local port. It will receive all strings and send them
accross the channel. A response must then come accross the
channel to be written to the client */
func listenAndRelay(port int, ch chan string) {
  defer close(ch)

  addr := "localhost:"+strconv.Itoa(port)
  ln, err := net.Listen("tcp", addr)
  if err != nil {
    log.Printf("Failed to create listener on port %d in listenAndRelay(): %v\n", port, err)
    return
  }
  defer ln.Close()

  for {
    conn, err := ln.Accept()
    if err != nil {
      log.Printf("Failed to accept connection in listenAndRelay(): %v\n", err)
      continue
    }

    err = relayMsgAndResponse(conn, ch)
    if err != nil {
      log.Printf("Failed to relay in listenAndRelay(): %v\n", err)
    }
  }
}

func relayMsgAndResponse(conn net.Conn, ch chan string) error {
    defer conn.Close()
    rConn := bufio.NewReader(conn)
    msg, err := rConn.ReadString(byte('\n'))
    if err != nil && err != io.EOF {
      return fmt.Errorf("Failed to read msg from client in relayMsgAndResponse(): %v\n", err)
    }

    // Process message and send response back
    ch<-msg
    resp := <-ch
    wConn := bufio.NewWriter(conn)
    _, err = wConn.WriteString(resp)
    if err != nil {
      return fmt.Errorf("Failed to write msg to client in relayMsgAndResponse(): %v\n", err)
    }
    return nil
}
