package processor

import (
  "fmt"
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

// Returns label followed by part of path that is on the drive
func pathToLabel(path string) (string, string) {
  lsblk := exec.Command("lsblk", "-o", "mountpoint,label")
  grep := exec.Command("grep", "/")

  pipe, _ := lsblk.StdoutPipe()
  defer pipe.Close()

  grep.Stdin = pipe
  lsblk.Start()

  out, err := grep.Output()
  if err != nil {
    return "", ""
  }

  lines := strings.Split(string(out), "\n")
  for _, line := range lines {
    tokens := splitLsblkLine(line)
    if len(tokens) < 2 {
      continue
    }
    fmt.Println(tokens)
    mntPt := strings.TrimSpace(tokens[0])
    if strings.Contains(path, mntPt) && mntPt != "" {
      return strings.TrimSpace(tokens[1]), strings.Replace(path, mntPt, "", 1)
    }
  }
  return "", ""
}

func splitLsblkLine(line string) []string {
  tokens := strings.Split(line, " ")
  stripped := make([]string, 0)
  for _, token := range tokens {
    if token != "" && token != " " {
      stripped = append(stripped, strings.TrimSpace(token))
    }
  }
  return stripped
}
