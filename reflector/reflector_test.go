package reflector

import (
  "testing"
  "fmt"
  "os"
)

func fHash(file *os.File) string {
  stat, err := file.Stat()
  if err != nil {
    panic(err)
  }
  return stat.Name() + "_fhash"
}

func dHash(f fileNode) string {
  return f.name + "_dhash"
}

func TestDirstruct(t *testing.T) {
  dt := newDirectoryTree("/home/aleksandr/Workspace/Hive_Whitepaper/", fHash, dHash)
  i := 0
  for id, _ := range dt.root.children {
    fmt.Printf("%d : %d\n", id, i)
    i++
  }

  serial := dt.serialize()
  fmt.Println(string(serial))

  var dt2 directoryTree
  dt2.deserialize(serial)
  err := dt2.deleteChild([]int{0,5})
  if err != nil {
    fmt.Println(err)
  }
  fmt.Println(string(dt2.serialize()))
}
/*
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
*/
