package main

import (
  "github.com/arstevens/goback/daemon/processor"
  "strconv"
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
    rmCmd := processor.UnbackupCommand+":"+*originalDir+"\n"
    resp := executeCommand(rmCmd)
    fmt.Println(resp)
  } else {
    bkCmd := processor.NewBackupCommand+":"+*originalDir+","+*reflectDir+","+"pref"+"\n"
    resp := executeCommand(bkCmd)
    fmt.Println(resp)
  }

}

func executeCommand(cmd string) string {
  conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(GobackPort))
  if err != nil {
    log.Printf("Failed to connect to daemon on port %d", GobackPort)
    return ""
  }
  defer conn.Close()

  _, err = bufio.NewWriter(conn).WriteString(cmd)
  if err != nil {
    log.Printf("Failed to write to daemon on port %d", GobackPort)
    return ""
  }

  resp, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    log.Printf("Failed to read from daemon on port %d", GobackPort)
    return ""
  }
  return resp
}
