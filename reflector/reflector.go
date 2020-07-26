package reflector

import (
  "strings"
  "fmt"
  "os"
)

const bufferSize = 4096

type PlainReflector struct {
  directoryMap SHA1ChangeMap
  reflectingMap SHA1ChangeMap
}

func NewPlainReflector(original SHA1ChangeMap, reflecting SHA1ChangeMap) *PlainReflector {
  pr := PlainReflector{
    directoryMap: original,
    reflectingMap: reflecting,
  }
  return &pr
}

/* PlainReflector.Backup() finds the differences between the
reflecting map and the original map and performs the necessary
operations to turn the reflecting directory into the original directory */
func (p PlainReflector) Backup() error {
  differences := p.reflectingMap.ChangeLog(p.directoryMap)

  // Handle deletions
  deletes := differences[deleteCode]
  err := handleDeletions(deletes, p.reflectingMap.root)
  if err != nil {
    return err
  }
  // Handle Creations
  creates := differences[createCode]
  err = handleCreations(creates, &p.reflectingMap, &p.directoryMap)
  if err != nil {
    return err
  }
  // Handle Updates
  updates := differences[updateCode]
  err = handleUpdates(updates, p.reflectingMap.root)
  if err != nil {
    return err
  }

  // Sync change maps
  p.reflectingMap.Sync(p.directoryMap)
  err = p.reflectingMap.Serialize()
  if err != nil {
    return fmt.Errorf("Failed to serialize in PlainReflector.Recover(): %v", err)
  }
  return nil
}

func handleDeletions(deletes []string, root string) error {
  for _, deletion := range deletes {
    removalPath := extendPath(root, deletion)
    err := os.RemoveAll(removalPath)
    if err != nil {
      return fmt.Errorf("Issue removing in handleDeletions(): %v", err)
    }
  }
  return nil
}

func handleCreations(creates []string, reflecting *SHA1ChangeMap, original *SHA1ChangeMap) error {
  fmt.Println(creates)
  for _, creation := range creates {
    relative := swapRootDir(creation, original)
    creationPath := extendPath(reflecting.root, creation)
    copyPath := extendPath(original.root, relative)

    stat, err := os.Lstat(copyPath)
    if err != nil {
      return fmt.Errorf("Issue receiving stat in handleCreations(): %v", err)
    }

    if stat.IsDir() {
      err = copyDir(copyPath, creationPath)
      if err != nil {
        return fmt.Errorf("Issue copying directory in handleCreations(): %v", err)
      }
    } else {
      fmt.Println(copyPath, creationPath)
      err = copyFile(copyPath, creationPath)
      if err != nil {
        fmt.Errorf("Issue copying file in handleCreations(): %v", err)
      }
    }
  }

  return nil
}

func handleUpdates(updates []string, root string) error {
  for _, update := range updates {
    updateParts := strings.Split(update, paramSep)
    if len(updateParts) < 2 {
      return fmt.Errorf("Update(%s) too small in handleUpdates(): ", update)
    }

    oldPath := extendPath(root, updateParts[0])
    newPath := changePathBase(oldPath, updateParts[1])
    err := os.Rename(oldPath, newPath)
    if err != nil {
      return fmt.Errorf("Failed to rename file in handleUpdates(): %v", err)
    }
  }

  return nil
}

/* Recover() traverses the reflecting directory and copies all of the files
over to the original directory */
func (p PlainReflector) Recover() error {
  reflectingDir := createFilesystemPath(&p.reflectingMap, "")
  originalDir := createFilesystemPath(&p.directoryMap, "")
  err := os.RemoveAll(originalDir)
  if err != nil {
    return fmt.Errorf("Couldn't remove original directory in PR.Recover(): %v", err)
  }

  err = copyDir(reflectingDir, originalDir)
  if err != nil {
    return fmt.Errorf("Couldn't copy directory contents in PR.Recover(): %v", err)
  }
  p.directoryMap.Sync(p.reflectingMap)

  err = p.directoryMap.Serialize()
  if err != nil {
    return fmt.Errorf("Issue serializing in PR.Recover(): %v", err)
  }
  return nil
}
/*
func copyFile(src, dst string) error {
  buf := make([]byte, bufferSize)

  dstFile, err  := os.Create(dst)
  if err != nil {
    return err
  }
  defer dstFile.Close()

  srcFile, err := os.Open(src)
  if err != nil {
    return err
  }
  defer srcFile.Close()

  for {
    n, err := srcFile.Read(buf)
    if err != nil && err != io.EOF {
      return err
    }
    if n == 0 {
      return nil
    }
    if _, err = dstFile.Write(buf[:n]); err != nil {
      return err
    }
  }
  return nil
}
*/
