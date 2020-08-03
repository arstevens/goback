package processor

import (
  "strconv"
  "bufio"
  "log"
  "net"
  "fmt"
  "io"
)

var (
  FailCode string = "fail"
  SuccessCode = "success"
)

type ReflectorCode string
type ChangeMapCode string

type Reflector interface {
  Backup() error
  Recover() error
}

type ChangeMap interface {
  Serialize() error
  Deserialize(fname string) error
  Update([][]string, [][]string) error
  Sync(ChangeMap) error
  ChangeLog(ChangeMap) ([][]string, error)
  RootDir() string
  RootName() string
}

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

func executeCommand(cmd string) error {

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
