package processor

import (
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
  RecoverCommand = "rec"
  UpdateCommand = "upd"
)

func CommandProcessor(port int, sysChan <-chan string) {
  netChan := make(chan string)
  go listenAndRelay(port, netChan)

  for {
    select {
      case cmd, ok := <-sysChan:
        if !ok {
          return
        }
        if err := executeCommand(cmd); err != nil {
          log.Printf("Failed to execute command(%s) in CommandProcessor: %v\n", cmd, err)
        }
      case cmd, ok := <-netChan:
        if !ok {
          return
        }
        if err := executeCommand(cmd); err != nil {
          log.Printf("Failed to execute command(%s) in CommandProcessor: %v\n", cmd, err)
          netChan<-FailCode
        } else {
          netChan<-SuccessCode
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

  switch cmdType {
    case BackupCommand:
      err := backupCommand(params, gen, mdb)
    case NewBackupCommand:
    case RecoverCommand:
    case UpdateCommand:
    default:
  }


}

func backupCommand(params []string, gen Generator, mdb MetadataDB) error {
  if len(params) < 1 {
    return fmt.Errorf("Not enough params in backupCommand()")
  }
  backupRoot := params[0]
  cmCode := mdb.ChangeMapType(backupRoot)
  refCode := mdb.ReflectorType(backupRoot)

  cmFile := mdb.ChangeMapFile(backupRoot)
  fCmFile := mdb.RefChangeMapFile(backupRoot)

  origCm, err := gen.OpenChangeMap(cmCode, cmFile)
  if err != nil {
    return fmt.Errorf("Couldn't open change map at %s with code %s in backupCommand(): %v", backupRoot, cmCode, err)
  }
  refCm, err := gen.OpenChangeMap(cmCode, fCmFile)
  if err != nil {
    return fmt.Errorf("Couldn't open change map at %s with code %s in backupCommand(): %v", backupRoot, cmCode, err)
  }

  reflector, err := gen.Reflect(refCode, orignCm, refCm)
  if err != nil {
    return fmt.Errorf("Couldn't reflect %s in backupCommand(): %v", backupRoot, err)
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
