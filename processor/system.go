package processor

import (
  "time"
  "strings"
)

var PollSpeed time.Duration = time.Second
const (
  DrivePlaceholder = "[DRIVE]"
)

func MonitorSystem(mdb MetadataDB, c chan<- string) {
  watching := make(map[string]bool)
  detector := newFsDetector()
  pollForNewBackups(mdb, watching, detector)

  for {
    change, err := detector.NextChange()
    if err != nil {
      panic(err)
    }
    cmd := fsChangeToCommand(change)
    c<-cmd


    <-time.After(PollSpeed)
  }


}

func pollForNewBackups(mdb MetadataDB, watching map[string]bool, detector *fsDetector) {
  for _, key := range mdb.Keys() {
    if watching[key] == false {
      err := detector.Watch(key)
      if err != nil {
        panic(err)
      }
      watching[key] = true
    }
  }
}

func pollForNewDrives(mdb MetadataDB, mounted map[string]string) []string {
  newMounts := make([]string, 0)
  for _, key := range mdb.Keys() {
    row, err  := mdb.GetRow(key)
    if err != nil {
      panic(err)
    }

    mountPoint := labelToMountPoint(row.DriveLabel)
    _, exists := mounted[key]
    if !exists && mountPoint != "" {
      mounted[key] = row.ReflectionRoot
      row.ReflectionRoot = strings.Replace(row.ReflectionRoot, DrivePlaceholder, mountPoint, 1)
      mdb.DeleteRow(key)
      mdb.InsertRow(key, row)

      newMounts = append(newMounts, key)
    } else if exists && mountPoint == "" {
      row.ReflectionRoot = mounted[key]
      mdb.DeleteRow(key)
      mdb.InsertRow(key, row)
      delete(mounted, key)
    }
  }

  return newMounts
}

func fsChangeToCommand(change fsChange) string {
  command := ""
  paramSep := ","
  switch (change.Operation) {
    case DeleteCode:
      dir := "false"
      if change.Dir {
        dir = dirTrue
      }
      command = UpdateCommand+":"+DeleteCommand+paramSep+change.Root+paramSep+
                change.Filepath+paramSep+dir
    case CreateCode:
      dir := "false"
      if change.Dir {
        dir = dirTrue
      }

      command = UpdateCommand+":"+CreateCommand+paramSep+change.Root+paramSep+
                change.Filepath+paramSep+dir
    case WriteCode:
      command = UpdateCommand+":"+WriteCommand+paramSep+change.Root+paramSep+
                change.Filepath
    case RenameCode:
      command = UpdateCommand+":"+RenameCommand+paramSep+change.Root+paramSep+
                change.Filepath
  }
  return command
}
