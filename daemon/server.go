package main

import (
  "strconv"
  "strings"
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
    fmt.Println("new connection")
    err = relayMsgAndResponse(conn, ch)
    if err != nil {
      log.Printf("Failed to relay in listenAndRelay(): %v\n", err)
    }
  }
}

func relayMsgAndResponse(conn net.Conn, ch chan string) error {
    msg, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil && err != io.EOF {
      return fmt.Errorf("Failed to read msg from client in relayMsgAndResponse(): %v\n", err)
    }
    fmt.Printf("Message received: %s\n", msg)
    msg = strings.Trim(msg, "\n")

    // Process message and send response back
    ch<-msg
    resp := <-ch
    fmt.Printf("Message response: %s\n", resp)
    resp += "\n"
    if _, err = conn.Write([]byte(resp)); err != nil {
      return fmt.Errorf("Failed to write msg to client in relayMsgAndResponse(): %v\n", err)
    }
    return nil
}
