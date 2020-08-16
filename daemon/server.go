package main

import (
  "strconv"
  "bufio"
  "fmt"
  "log"
  "net"
  "io"
)

/* listenAndRelay() connects and communicates with anyone
on the local port. It will receive all strings and send them
accross the channel. A response must then come accross the
channel to be written to the client */
func ListenAndRelay(port int, ch chan string) {
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
