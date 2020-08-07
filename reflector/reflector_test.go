package reflector
/*
import (
  "testing"
)
func TestReflector(t *testing.T) {
  cm, err := NewSHA1ChangeMap("/home/aleksandr/Workspace/Hive_Whitepaper", "/home/aleksandr/Workspace/cmfile")
  if err != nil {
    panic(err)
  }
  cm2, err := NewSHA1ChangeMap("/home/aleksandr/Workspace/testzone", "/home/aleksandr/Workspace/cm2file")
  if err != nil {
    panic(err)
  }

  rf := NewPlainReflector(*cm2, *cm)
  err = rf.Recover()
  if err != nil {
    panic(err)
  }
}
/*
func TestChangeMap(t *testing.T) {
  /*
  cm, err := NewSHA1ChangeMap("/home/aleksandr/Workspace/Hive_Whitepaper", "/home/aleksandr/Workspace/cmtestserial")
  if err != nil {
    panic(err)
  }

  err = cm.Serialize()
  if err != nil {
    panic(err)
  }
  fmt.Println(string(cm.dirModel.serialize()))
  var cm SHA1ChangeMap
  err := cm.Deserialize("/home/aleksandr/Workspace/cmtestserial")
  if err != nil {
    panic(err)
  }

  var cm2 SHA1ChangeMap
  err = cm2.Deserialize("/home/aleksandr/Workspace/cmtestserial")
  if err != nil {
    panic(err)
  }

  dirChanges := [][]string{1:[]string{"newdir"}, 2:[]string{}}
  fileChanges := [][]string{0: []string{"Hive_Whitepaper_v1.odt"}, 1:[]string{"newfile"}, 2:[]string{"graphics/load_pickup_v1.png,load_pickup.png"}}

  err = cm2.Update(fileChanges, dirChanges)
  if err != nil {
    panic(err)
  }
  cm2.dirModel.hash()
  fmt.Println(string(cm.dirModel.serialize()))
  fmt.Println(string(cm2.dirModel.serialize()))
  cm.Sync(cm2)

  fmt.Println(cm.ChangeLog(cm2))
}
*/
