package processor

import (
  "strings"
  "os/exec"
)


func labelToMountPoint(label string) string {
  lsblk := exec.Command("lsblk", "-o", "label,mountpoint")
  grep := exec.Command("grep", label)

  pipe, _ := lsblk.StdoutPipe()
  defer pipe.Close()

  grep.Stdin = pipe
  lsblk.Start()

  out, err := grep.Output()
  if err != nil {
    return ""
  }

  tokens := strings.Split(string(out), " ")
  mount := strings.TrimSpace(tokens[len(tokens) - 1])
  return mount
}
