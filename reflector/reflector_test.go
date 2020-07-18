package reflector

import (
  "testing"
  "fmt"
)

func TestDirstruct(t *testing.T) {
  dt := newDirectoryTree("/home/aleksandr/Workspace/Hive_Whitepaper/")
  i := 0
  for id, _ := range dt.root.children {
    fmt.Printf("%s : %d\n", id, i)
    i++
  }

  serial := dt.serialize()
  fmt.Println(string(serial))

  var dt2 directoryTree
  dt2.deserialize(serial)

  i = 0
  for id, _ := range dt2.root.children {
    fmt.Printf("%s : %d\n", id, i)
    i++
  }
  err := dt2.deleteChild([]string{"Hive_Whitepaper_v1.odt"})
  if err != nil {
    fmt.Println(err)
  }
  fmt.Println(string(dt2.serialize()))
}

func TestChangeMap(t *testing.T) {
  root := "/home/aleksandr/Workspace/Hive_Whitepaper/"
  dt := newDirectoryTree(root)
  cm := SHA1ChangeMap{
    root: root,
    cmFname: "/home/aleksandr/Workspace/goback/reflector/cmfile",
    dirModel: *dt,
  }

  var cm2 SHA1ChangeMap
  cm2.Sync(cm)
  fmt.Println(string(cm2.dirModel.serialize()))
}
