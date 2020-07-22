package reflector

import (
  "io/ioutil"
)

type changeCode int

const (
  Delete changeCode = iota
  Create
  Update
)

type SHA1ChangeMap struct {
  root string
  cmFname string
  dirModel directoryTree
}

// Read data from cmFname into the dirModel
func (s *SHA1ChangeMap) Deserialize() error {
  raw, err := ioutil.ReadFile(s.cmFname)
  if err != nil {
    return err
  }

  s.dirModel.deserialize(raw)
  return nil
}

func (s SHA1ChangeMap) Serialize() error {
  raw := s.dirModel.serialize()
  return ioutil.WriteFile(s.cmFname, raw, 0644)
}

func (s *SHA1ChangeMap) Sync(cm SHA1ChangeMap) {
  s.dirModel = cm.dirModel.duplicate()
}

/*
func (s SHA1ChangeMap) ChangeLog(cm SHA1ChangeMap) [][]string {

}
*/
