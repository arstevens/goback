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
  rootDirectory string
  reflectingDirectory string
}

func (p PlainReflector) Backup() {
  differences := p.reflectingMap.ChangeLog(p.directoryMap)

  // Handle deletions
  deletes := differences[deleteCode]
  for _, deletion := range deletes {
    removalPath := p.rootDirectory + "/" + deletion
    err := os.RemoveAll(removalPath)
    if err != nil {
      panic(err)
    }
  }

  // Handle Creations
  creates := differences[createCode]
  for _, creation := range creates {
    creationPath := p.rootDirectory + "/" + creation
    copyPath := p.reflectingDirectory + "/" + creation

    stat, err := os.Lstat(copyPath)
    if err != nil {
      panic(err)
    }

    if stat.IsDir() {
      err = os.Mkdir(creationPath, stat.Mode())
      if err != nil {
        panic(err)
      }
    } else {
      err = copyFile(creationPath, copyPath)
      if err != nil {
        panic(err)
      }
    }
  }

  // Handle Updates
  updates := differences[updateCode]
  for _, update := range updates {
    updateParts := strings.Split(update, paramSep)
    if len(updateParts) < 2 {
      panic(fmt.Errorf("Update too small"))
    }

    oldPath := updateParts[0]
    newPathBase := strings.Split(oldPath, "/")
    newPath := strings.Join(newPathBase[:len(newPathBase)-1], "/") + "/" + updateParts[1]
    err := os.Rename(oldPath, newPath)
    if err != nil {
      panic(err)
    }
  }

  // Sync change maps
  p.reflectingMap.Sync(p.directoryMap)
  err := p.reflectingMap.Serialize()
  if err != nil {
    panic(err)
  }
}

func (p PlainReflector) Recover() error {
  err := filepath.Walk(p.reflectingDirectory, func(path string, info os.FileInfo, err error) error {
    basePath := strings.Replace(path, p.reflectingDirectory, "", 1)
    newFilePath := p.rootDirectory + "/" + basePath
    if info.IsDir() {
      return os.Mkdir(newFilePath, info.Mode())
    }

    return copyFile(newFilePath, path)
  })
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
