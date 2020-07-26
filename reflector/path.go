package reflector

import (
  "os"
  "strings"
  "path/filepath"
)

func cleanPath(path string) string {
  return strings.Trim(path, string(os.PathSeparator))
}

func createFilesystemPath(s *SHA1ChangeMap, relPath string) string {
  return filepath.Join(s.root, s.dirModel.root.name, relPath)
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

func swapRootDir(path string, s *SHA1ChangeMap) string {
  parts := strings.Split(path, string(os.PathSeparator))
  base := filepath.Join(parts[1:]...)
  newPath := filepath.Join(s.dirModel.root.name, base)
  return newPath
}
