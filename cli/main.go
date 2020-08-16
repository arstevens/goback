package main

import (
  "github.com/arstevens/goback/daemon/processor"
  "strconv"
  "strings"
  "bufio"
  "flag"
  "fmt"
  "net"
  "log"
)

var GobackPort int = 25000

func main() {
  originalDir := flag.String("o", "", "Directory to backup")
  reflectDir := flag.String("c", "", "Location to backup to")
  remove := flag.Bool("r", false, "Stop backing up provided directory")

  flag.Parse()
  if *remove {
    rmCmd := processor.UnbackupCommand+":"+*originalDir
    fmt.Printf("Sending message %s\n", rmCmd)
    resp := executeCommand(rmCmd)
    fmt.Println(resp)
  } else {
    bkCmd := processor.NewBackupCommand+":"+*originalDir+","+*reflectDir+","+"pref"
    fmt.Printf("Sending message %s\n", bkCmd)
    resp := executeCommand(bkCmd)
    fmt.Println(resp)
  }

}

func executeCommand(cmd string) string {
  cmd += "\n"
  conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(GobackPort))
  if err != nil {
    log.Printf("Failed to connect to daemon on port %d", GobackPort)
    return ""
  }
  defer conn.Close()

  if _, err = conn.Write([]byte(cmd)); err != nil {
    log.Printf("Failed to write to daemon on port %d", GobackPort)
    return ""
  }

  resp, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    log.Printf("Failed to read from daemon on port %d", GobackPort)
    return ""
  }
  resp = strings.Trim(resp, "\n")
  return resp
}
