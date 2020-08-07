package reflector

import (
  "github.com/arstevens/goback/processor"
  "path/filepath"
  "crypto/sha1"
  "strings"
  "sort"
  "fmt"
  "os"
  "io"
)

type changeCode int

const (
  deleteCode changeCode = iota
  createCode
  updateCode
)

const (
  startOfText byte = 0x02
)

/* sha1FileHash() and sha1DirHash() implement fileHashFunction and
dirHashFunction defined along side directoryTree. These are used to
generate a new directoryTree from a filesystem path */
func sha1FileHash(fileName string) []byte {
  file, err := os.Open(fileName)
  if err != nil {
    panic(fmt.Errorf("Error opening file in sha1FileHash: %v", err))
  }
  defer file.Close()

  hash := sha1.New()
  if _, err = io.Copy(hash, file); err != nil {
    panic(fmt.Errorf("Error copying in sha1FileHash: %v", err))
  }
  baseName := []byte(filepath.Base(fileName))
  if _, err = hash.Write(baseName); err != nil {
    panic(fmt.Errorf("Error hashing in sha1FileHash: %v", err))
  }

  return hash.Sum(nil)
}

func sha1DirHash(fn fileNode) []byte {
  keys := make([]int, 0, len(fn.children))
  for _, child := range fn.children {
    keys = append(keys, child.id)
  }
  sort.Ints(keys)

  hash := sha1.New()
  for _, key := range keys {
    if _, err := hash.Write(fn.children[key].hash); err != nil {
      panic(fmt.Errorf("Error hashing in sha1DirHash: %v", err))
    }
  }

  if _, err := hash.Write([]byte(fn.name)); err != nil {
    panic(fmt.Errorf("Error name hashing in sha1DirHash: %v", err))
  }

  return hash.Sum(nil)
}

type SHA1ChangeMap struct {
  root string
  dirModel directoryTree
}

// Satisfies interactor.changeMapCreator
func NewSHA1ChangeMap(rootName string, serialPath string) (processor.ChangeMap, error) {
  dt, err := newDirectoryTree(rootName, sha1FileHash, sha1DirHash)
  if err != nil {
    return nil, fmt.Errorf("Failed constructing directory tree for %s in NewS1CM(): %v", rootName, err)
  }

  /* Need to remove actual directory being modeled
  from cmRoot because its name is already included in the
  directoryTree name field */
  rootNameSplit := strings.Split(rootName, "/")
  cmRoot := strings.Join(rootNameSplit[:len(rootNameSplit)-1], "/")
  return &SHA1ChangeMap {
    root: cmRoot,
    dirModel: *dt,
  }, nil
}

func LoadSHA1ChangeMap(serial string, root string) (processor.ChangeMap, error) {
  var cm SHA1ChangeMap
  err := cm.Deserialize(serial)
  if err != nil {
    return nil, fmt.Errorf("Couldn't load S1CM in LoadSHA1ChangeMap(): %v", err)
  }
  cm.root = root
  return &cm, err
}

func (s *SHA1ChangeMap) RootDir() string {
  return s.root
}

func (s *SHA1ChangeMap) RootName() string {
  return s.dirModel.root.name
}

func (s *SHA1ChangeMap) Deserialize(raw string) error {
  s.dirModel.deserialize([]byte(raw))
  s.dirModel.dirHash = sha1DirHash
  return nil
}

func (s SHA1ChangeMap) Serialize() string {
  return string(s.dirModel.serialize())
}

/* changes give to Update should be incremental. This means that
saying to create a directory simply creates it but puts nothing in
it. This is different from the format changelog returns where a create
command for a directory means all contents should be copied */
func (s *SHA1ChangeMap) Update(fileChanges [][]string, dirChanges [][]string) error {
  deletes := append(fileChanges[0], dirChanges[0]...)
  for _, del := range deletes {
    fmt.Println(del)
    s.dirModel.deleteChild(del)
  }

  fileCreates := fileChanges[1]
  for _, create := range fileCreates {
    fullPath := createFilesystemPath(s, create)
    hash := sha1FileHash(fullPath)
    err := s.dirModel.addChild(create, hash, false)
    if err != nil {
      return fmt.Errorf("Failed to add child file %s in S1CM.Update(): %v", create, err)
    }
  }

  dirCreates := dirChanges[1]
  for _, create := range dirCreates {
    err := s.dirModel.addChild(create, []byte{}, true)
    if err != nil {
      return fmt.Errorf("Failed to add child dir %s in S1CM.Update(): %v", create, err)
    }
  }

  fileUpdates := fileChanges[2]
  for _, update := range fileUpdates {
    parts := strings.Split(update, paramSep)
    if len(parts) < 2 {
      return fmt.Errorf("Invalid update format %s in SHA1ChangeMap.Update()", update)
    }
    path, newName := parts[0], parts[1]
    err := s.dirModel.renameChild(path, newName)
    if err != nil {
      return fmt.Errorf("Unable to rename child in S1CM.Update(): %v", err)
    }
    newPath := changePathBase(path, newName)
    fullPath := createFilesystemPath(s, newPath)
    hash := sha1FileHash(fullPath)
    err = s.dirModel.updateHash(newPath, hash)
    if err != nil {
      return fmt.Errorf("Unable to update hash in S1CM.Update(): %v", err)
    }
  }

  dirUpdates := dirChanges[2]
  for _, update := range dirUpdates {
    parts := strings.Split(update, paramSep)
    if len(parts) < 2 {
      return fmt.Errorf("Invalid update format %s in SHA1ChangeMap.Update()", update)
    }
    path, newName := parts[0], parts[1]
    err := s.dirModel.renameChild(path, newName)
    if err != nil {
      return fmt.Errorf("Unable to rename child in S1CM.Update(): %v", err)
    }
  }
  s.dirModel.hash()

  return nil
}

/* Changes the current SHA1ChangeMap into the foreign one */
func (s *SHA1ChangeMap) Sync(m processor.ChangeMap) error {
  cm, ok := m.(*SHA1ChangeMap)
  if !ok {
    return fmt.Errorf("Change maps not of the same type in S1CM.Sync()")
  }
  s.dirModel = cm.dirModel.duplicate()
  return nil
}

/* Creates a list of commands to turn cm s into cm */
func (s SHA1ChangeMap) ChangeLog(m processor.ChangeMap) ([][]string, error) {
  cm, ok := m.(*SHA1ChangeMap)
  if !ok {
    return nil, fmt.Errorf("Change maps not of the same type in S1CM.ChangeLog()")
  }
  return treeDifference(s.dirModel.root, cm.dirModel.root), nil
}
