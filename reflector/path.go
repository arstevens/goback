package reflector

import (
  "os"
  "strings"
  "path/filepath"
  "github.com/arstevens/goback/processor"
)

func cleanPath(path string) string {
  return strings.Trim(path, string(os.PathSeparator))
}

func createFilesystemPath(s processor.ChangeMap, relPath string) string {
  return filepath.Join(s.RootDir(), s.RootName(), relPath)
}

func createRelativePath(path string, root string) string {
  return strings.Replace(path, root, "", 1)
}

func changePathBase(path string, newName string) string {
  return filepath.Join(filepath.Dir(path), newName)
}

func extendPath(path string, ext string) string {
  return filepath.Join(path, ext)
}

func splitPath(path string) []string {
  path = strings.Trim(path, string(os.PathSeparator))
  return strings.Split(path, string(os.PathSeparator))
}

func swapRootDir(path string, newRoot string) string {
  parts := strings.Split(path, string(os.PathSeparator))
  base := filepath.Join(parts[1:]...)
  newPath := filepath.Join(newRoot, base)
  return newPath
}
