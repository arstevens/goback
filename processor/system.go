package processor

import (
  "fmt"
  "time"
  "path/filepath"
)

var PollSpeed time.Duration = time.Second
const (
  DrivePlaceholder = "[DRIVE]"
)

func MonitorSystem(mdb MetadataDB, c chan<- string) {
  NextChangeTimeout = time.Second
  watching := make(map[string]bool)
  mounted := make(map[string]bool)
  detector := newFsDetector()
  pollForNewBackups(mdb, watching, detector)

  for {
    // Check for any changes to backup points
    change, err := detector.NextChange()
    _, isTimeout := err.(*TimeoutErr)
    if !isTimeout && err != nil {
      panic(err)
    }
    if !isTimeout {
      row := mdb.GetRow(change.Root)
      if !mdb.HasChanged {
        row.HasChanged = true
        err = mdb.DeleteRow(change.Root)
        if err != nil {
          panic(err)
        }
        err = mdb.InsertRow(change.Root, row)
        if err != nil {
          panic(err)
        }
        backupCmd := string(BackupCommand)+":"+change.Root
        c<-backupCmd
      }
    }

    // Check for any new backups created
    pollForNewBackups(mdb, watching, detector)

    // Check if backup reflections are mounted
    newlyMounted := pollForNewDrives(mdb, mounted)
    for _, origRoot := range newlyMounted {
      backupCmd := string(BackupCommand)+":"+origRoot
      c<-backupCmd
    }

    // Wait to check again
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

func pollForNewDrives(mdb MetadataDB, mounted map[string]bool) []string {
  newMounts := make([]string, 0)
  for _, key := range mdb.Keys() {
    row, err  := mdb.GetRow(key)
    if err != nil {
      panic(err)
    }

    mountPoint := labelToMountPoint(row.DriveLabel)
    _, exists := mounted[key]
    if !exists && mountPoint != "" {
      mounted[key] = true
      refRoot := filepath.Join(mountPoint, row.ReflectionBase)
      row.ReflectionRoot = refRoot
      mdb.DeleteRow(key)
      mdb.InsertRow(key, row)
      newMounts = append(newMounts, key)
    } else if exists && mountPoint == "" {
      row.ReflectionRoot = ""
      mdb.DeleteRow(key)
      mdb.InsertRow(key, row)
      delete(mounted, key)
    }
  }

  return newMounts
}
