package processor

import (
  "log"
  "time"
  "path/filepath"
)

var PollSpeed time.Duration = time.Second

func MonitorSystem(mdb MetadataDB, c chan<- string) {
  defer close(c)

  NextChangeTimeout = time.Second
  watching := make(map[string]bool)
  mounted := make(map[string]bool)
  detector := newFsDetector()
  pollForNewBackups(mdb, watching, detector)

  for {
    // Check for any changes to backup points
    changeRoot, err := detector.NextChange()
    _, isTimeout := err.(*TimeoutErr)
    if !isTimeout {
      if err != nil {
        log.Printf("Failed to receive next change in MonitorSystem(): %v", err)
      } else {
        row, err := mdb.GetRow(changeRoot)
        if err != nil {
          log.Printf("Failed to retrieve row in MonitorSystem(): %v", err)
        } else if !row.HasChanged {
          row.HasChanged = true
          err = mdb.UpdateRow(row)
          if err != nil {
            log.Printf("Failed to update row in MonitorSystem(): %v", err)
          } else {
            backupCmd := string(BackupCommand)+":"+changeRoot
            c<-backupCmd
          }
        }
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
  keys := mdb.Keys()
  for _, key := range keys {
    if watching[key] == false {
      err := detector.Watch(key)
      if err != nil {
        log.Printf("Failed to set watch on %s in pollForNewBackups(): %v", key, err)
        continue
      }
      watching[key] = true
    }
  }

  for key, watched := range watching {
    if watched && !contains(keys, key) {
      err := detector.Unwatch(key)
      if err != nil {
        log.Printf("Failed to unwatch %s in pollForNewBackups(): %v", key, err)
        continue
      }
      watching[key] = false
    }
  }
}

func contains(slice []string, key string) bool {
  for _, val := range slice {
    if val == key {
      return true
    }
  }
  return false
}

func pollForNewDrives(mdb MetadataDB, mounted map[string]bool) []string {
  newMounts := make([]string, 0)
  for _, key := range mdb.Keys() {
    row, err  := mdb.GetRow(key)
    if err != nil {
      log.Printf("Failed to get row in pollForNewDrives(): %v", err)
      return []string{}
    }

    mountPoint := labelToMountPoint(row.DriveLabel)
    _, isMounted := mounted[key]
    if !isMounted && mountPoint != "" {
      mounted[key] = true
      refRoot := filepath.Join(mountPoint, row.ReflectionBase)
      row.ReflectionRoot = refRoot
      err = mdb.UpdateRow(row)
      if err != nil {
        log.Printf("Failed to update row for %s in pollForNewDrives(): %v", key, err)
        continue
      }
      newMounts = append(newMounts, key)
    } else if isMounted && mountPoint == "" {
      row.ReflectionRoot = ""
      err = mdb.UpdateRow(row)
      if err != nil {
        log.Printf("Failed to update row for %s in pollForNewDrives(): %v", key, err)
        continue
      }
      delete(mounted, key)
    }
  }

  return newMounts
}
