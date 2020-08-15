package processor

import (
  "time"
  "path/filepath"
)

var PollSpeed time.Duration = time.Second

func MonitorSystem(mdb MetadataDB, c chan<- string) {
  NextChangeTimeout = time.Second
  watching := make(map[string]bool)
  mounted := make(map[string]bool)
  detector := newFsDetector()
  pollForNewBackups(mdb, watching, detector)

  for {
    // Check for any changes to backup points
    changeRoot, err := detector.NextChange()
    _, isTimeout := err.(*TimeoutErr)
    if !isTimeout && err != nil {
      panic(err)
    }
    if !isTimeout {
      row, err := mdb.GetRow(changeRoot)
      if err != nil {
        panic(err)
      }
      if !row.HasChanged {
        row.HasChanged = true
        err = mdb.UpdateRow(row)
        if err != nil {
          panic(err)
        }
        backupCmd := string(BackupCommand)+":"+changeRoot
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
  keys := mdb.Keys()
  for _, key := range keys {
    if watching[key] == false {
      err := detector.Watch(key)
      if err != nil {
        panic(err)
      }
      watching[key] = true
    }
  }

  for key, watched := range watching {
    if watched && !contains(keys, key) {
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
      panic(err)
    }

    mountPoint := labelToMountPoint(row.DriveLabel)
    _, isMounted := mounted[key]
    if !isMounted && mountPoint != "" {
      mounted[key] = true
      refRoot := filepath.Join(mountPoint, row.ReflectionBase)
      row.ReflectionRoot = refRoot
      err = mdb.UpdateRow(row)
      if err != nil {
        panic(err)
      }
      newMounts = append(newMounts, key)
    } else if isMounted && mountPoint == "" {
      row.ReflectionRoot = ""
      err = mdb.UpdateRow(row)
      if err != nil {
        panic(err)
      }
      delete(mounted, key)
    }
  }

  return newMounts
}
