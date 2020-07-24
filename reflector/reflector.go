package reflector

import (
  "path/filepath"
  "strings"
  "fmt"
  "io"
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
  err = handleCreations(creates, p.reflectingMap.root, p.directoryMap.root)
  if err != nil {
    return err
  }
  // Handle Updates
  updates := differences[updateCode]
  err = handleUpdates(updates)
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
    removalPath := root + "/" + deletion
    err := os.RemoveAll(removalPath)
    if err != nil {
      return fmt.Errorf("Issue removing in handleDeletions(): %v", err)
    }
  }
  return nil
}

func handleCreations(creates []string, reflectingRoot string, originalRoot string) error {
  for _, creation := range creates {
    creationPath := reflectingRoot + "/" + creation
    copyPath := originalRoot + "/" + creation

    stat, err := os.Lstat(copyPath)
    if err != nil {
      return fmt.Errorf("Issue receiving stat in handleCreations(): %v", err)
    }

    if stat.IsDir() {
      err = os.Mkdir(creationPath, stat.Mode())
      if err != nil {
        return fmt.Errorf("Issue creating directory in handleCreations(): %v", err)
      }
    } else {
      err = copyFile(creationPath, copyPath)
      if err != nil {
        fmt.Errorf("Issue copying file in handleCreations(): %v", err)
      }
    }
  }

  return nil
}

func handleUpdates(updates []string) error {
  for _, update := range updates {
    updateParts := strings.Split(update, paramSep)
    if len(updateParts) < 2 {
      return fmt.Errorf("Update too small in handleUpdates(): ", updateParts)
    }

    oldPath := updateParts[0]
    newPathBase := strings.Split(oldPath, "/")
    newPath := strings.Join(newPathBase[:len(newPathBase)-1], "/") + "/" + updateParts[1]
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
  pathToWalk := p.reflectingMap.root + "/" + p.reflectingMap.dirModel.root.name
  err := filepath.Walk(pathToWalk, func(path string, info os.FileInfo, err error) error {
    fmt.Println(path)
    basePath := strings.Replace(path, p.reflectingMap.root, "", 1)
    newFilePath := p.directoryMap.root + "/" + basePath
    if info.IsDir() {
      return os.Mkdir(newFilePath, info.Mode())
    }

    return copyFile(newFilePath, path)
  })
  if err != nil {
    return fmt.Errorf("Error walking in Recover(): %v", err)
  }

  err = p.directoryMap.Serialize()
  if err != nil {
    err = fmt.Errorf("Error serializing in Recover(): %v", err)
  }
  return err
}

func copyFile(dst string, src string) error {
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
