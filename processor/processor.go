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

type UpdatePackage struct {
  Backup bool
  OriginalRoot string
  FileUpdates [][]string
  DirUpdates [][]string
}

func CommandProcessor(gen Generator, mdb MetadataDB, comChan chan string, updateChan <-chan UpdatePackage) {
  for {
    select {
      case cmd, ok := <-updateChan:
        if !ok {
          return
        }
        if err := updateCommand(cmd, gen, mdb); err != nil {
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

func updateCommand(pack UpdatePackage, gen Generator, mdb MetadataDB) error {
  mdbRow, err := mdb.GetRow(pack.OriginalRoot)
  if err != nil {
    return fmt.Errorf("Could not get row with key %s in updateCommand(): %v", mdbRow.OriginalRoot, err)
  }
  cm, err := gen.OpenChangeMap(mdbRow.CMCode, mdbRow.OriginalRoot, mdbRow.OriginalCM)
  if err != nil {
    return fmt.Errorf("Couldn't open change map in updateCommand(): %v", err)
  }
  fmt.Println(mdbRow)

  err = cm.Update(pack.FileUpdates, pack.DirUpdates)
  if err != nil {
    return fmt.Errorf("Couldn't update change map in updateCommand(): %v", err)
  }
  fmt.Println(cm.Serialize())

  if pack.Backup {
    refcm, err := gen.OpenChangeMap(mdbRow.CMCode, mdbRow.ReflectionRoot, mdbRow.ReflectionCM)
    if err != nil {
      return fmt.Errorf("Couldn't open reflecting CM in updateCommand(): %v", err)
    }
    ref, err := gen.Reflect(mdbRow.ReflectionCode, cm, refcm)
    if err != nil {
      return fmt.Errorf("Couldn't reflect in updateCommand(): %v", err)
    }

    err = ref.Backup()
    if err != nil {
      return fmt.Errorf("Couldn't backup in updateCommand(): %v", err)
    }
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
