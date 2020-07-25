package reflector

import (
  "testing"
  "fmt"
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
  cm2.dirModel.root.children = make(map[int]fileNode)
  cm2.dirModel.idCount = 1
  cm2.dirModel.idMap = make(map[string]int)

  changeLog := cm2.ChangeLog(*cm)
  fmt.Println(changeLog)

  dirChanges := [][]string{1:[]string{"newdir"}, 2:[]string{}}
  fileChanges := [][]string{1:[]string{"newfile"}, 2:[]string{}}

  err = cm2.Update(fileChanges, dirChanges)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(cm2.dirModel.serialize()))
  fmt.Printf("\n%t and %t", true, false)


}
/*
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
func TestDifference(t *testing.T) {
  var dt directoryTree
  dt.deserialize([]byte(`0,Hive_Whitepaper,Hive_Whitepaper_dhash)2,Hive_Whitepaper_v1.odt,Hive_Whitepaper_v1.odt_fhash)|)3,graphic_mockups,)4,honeycoin_load_change.drawio,honeycoin_load_change.drawio_fhash)|)5,honeycoin_mining_v1.drawio,honeycoin_mining_v1.drawio_fhash)|)6,load_distribution_v1.drawio,load_distribution_v1.drawio_fhash)|)7,load_pickup_v1.drawio,load_pickup_v1.drawio_fhash)|)8,topology_graph_v1.drawio,topology_graph_v1.drawio_fhash)|)|)9,graphics,)13,load_pickup_v1.png,load_pickup_v1.png_fhash)|)14,topology_graph_v1.png,topology_graph_v1.png_fhash)|)10,honeycoin_load_change.png,honeycoin_load_change.png_fhash)|)11,honeycoin_mining_v1.png,honeycoin_mining_v1.png_fhash)|)12,load_distribution_v1.png,load_distribution_v1.png_fhash)|)|)|)`))
  dt2 := dt.duplicate()

  dt2.deleteChild([]int{0, 9, 13})
  updatedNode := duplicate(dt2.root.children[9])
  updatedNode.hash = "deleted_hash"
  dt2.root.children[9] = updatedNode
  dt2.addChild([]int{0, 9}, 19, "newfile", "newfile_hash")

  newNode := duplicate(dt2.root.children[2])
  newNode.name = "newName"
  newNode.hash = "newName_fhash"
  dt2.root.children[2] = newNode

  fmt.Println(string(dt.serialize()))
  fmt.Println(string(dt2.serialize()))

  diffs := treeDifference(dt.root, dt2.root)
  for code, diffSlice := range diffs {
    fmt.Println(code)
    fmt.Println(diffSlice)
  }
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
*/
