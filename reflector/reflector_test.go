package reflector
import (
  "testing"
)

func TestReflector(t *testing.T) {
  refRoot := "/run/media/aleksandr/AleksPersonal/testref"
  origRoot := "/home/aleksandr/Workspace/testzone"
  ref, err := NewPlainReflector(origRoot, refRoot)
  if err != nil {
    panic(err)
  }

  err = ref.Backup()
  if err != nil {
    panic(err)
  }
}
/*
func TestChangeMap(t *testing.T) {
  cm, err := NewSHA1ChangeMap("/home/aleksandr/Workspace/testzone")
  if err != nil {
    panic(err)
  }
  cm2, err := NewSHA1ChangeMap("/home/aleksandr/Workspace/newenv/testzone")
  if err != nil {
    panic(err)
  }

  clog, err := cm.ChangeLog(cm2)
  if err != nil {
    panic(err)
  }
  fmt.Println(clog)

  cm2.Deserialize(cm.Serialize())
  updates := [][]string{[]string{"Hive_Whitepaper_v3_Fluid.odt"},
  []string{"graphic_mockups/newfile"}, []string{"graphics/simple_exchange_v1.png,newname.png"}}
  err = cm2.Update(updates, make([][]string, 3))
  if err != nil {
    panic(err)
  }
  clog, err = cm.ChangeLog(cm2)
  if err != nil {
    panic(err)
  }
  fmt.Println(clog)

  ref, err := NewPlainReflector(cm2, cm)
  if err != nil {
    panic(err)
  }
  err = ref.Backup()
  if err != nil {
    panic(err)
  }
}
/*
func TestReflector(t *testing.T) {
  cm, err := NewSHA1ChangeMap("/home/aleksandr/Workspace/testzone")
  if err != nil {
    panic(err)
  }
  cm2, err := NewSHA1ChangeMap("/home/aleksandr/Workspace/newenv/testzone")
  if err != nil {
    panic(err)
  }

  rf, err := NewPlainReflector(cm, cm2)
  if err != nil {
    panic(err)
  }
  err = rf.Backup()
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
