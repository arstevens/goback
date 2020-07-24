package reflector

import (
  "path/filepath"
  "crypto/sha1"
  "io/ioutil"
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

/* sha1FileHash() and sha1DirHash() implement fileHashFunction and
dirHashFunction defined along side directoryTree. These are used to
generate a new directoryTree from a filesystem path */
func sha1FileHash(fileName string) []byte {
  file, err := os.Open(fileName)
  if err != nil {
    panic(fmt.Errorf("Error opening file in fhash: %v", err))
  }
  defer file.Close()

  hash := sha1.New()
  if _, err := io.Copy(hash, file); err != nil {
    panic(fmt.Errorf("Error copying in fhash: %v", err))
  }
  baseName := []byte(filepath.Base(fileName))
  if _, err = hash.Write(baseName); err != nil {
    panic(fmt.Errorf("Error hashing in fhash: %v", err))
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
      panic(fmt.Errorf("Error hashing in dirhash: %v", err))
    }
  }

  if _, err := hash.Write([]byte(fn.name)); err != nil {
    panic(fmt.Errorf("Error name hashing in dirhash: %v", err))
  }

  return hash.Sum(nil)
}

type SHA1ChangeMap struct {
  root string
  cmFname string
  dirModel directoryTree
}

func NewSHA1ChangeMap(rootName string, serialPath string) (*SHA1ChangeMap, error) {
  dt, err := newDirectoryTree(rootName, sha1FileHash, sha1DirHash)
  if err != nil {
    return nil, fmt.Errorf("Failed constructing sha1changemap: %v", err)
  }

  /* Need to remove actual directory being modeled
  from cmRoot because its name is already included in the
  directoryTree name field */
  rootNameSplit := strings.Split(rootName, "/")
  cmRoot := strings.Join(rootNameSplit[:len(rootNameSplit)-1], "/")
  return &SHA1ChangeMap {
    root: cmRoot,
    cmFname: serialPath,
    dirModel: *dt,
  }, nil
}

func (s *SHA1ChangeMap) Deserialize(fname string) error {
  raw, err := ioutil.ReadFile(s.cmFname)
  if err != nil {
    return fmt.Errorf("Issue reading SHA1ChangeMap serial file %s: %v", s.cmFname, err)
  }

  s.dirModel.deserialize(raw)
  return nil
}

func (s SHA1ChangeMap) Serialize() error {
  raw := s.dirModel.serialize()
  err := ioutil.WriteFile(s.cmFname, raw, 0644)
  if err != nil {
    return fmt.Errorf("Failed to serialize in SHA1ChangeMap.Serialize(): %v", err)
  }
  return nil
}

/* Changes the current SHA1ChangeMap into the foreign one */
func (s *SHA1ChangeMap) Sync(cm SHA1ChangeMap) {
  s.dirModel = cm.dirModel.duplicate()
}

func (s SHA1ChangeMap) ChangeLog(cm SHA1ChangeMap) [][]string {
  return treeDifference(s.dirModel.root, cm.dirModel.root)
}
