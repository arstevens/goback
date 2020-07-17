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
