package interactor

import (
  "github.com/arstevens/goback/processor"
  "github.com/arstevens/goback/reflector"
  "testing"
)

func TestInteractor(t *testing.T) {
  refTypes := map[processor.ReflectorCode]reflectorCreator{
    "rf1":reflector.NewPlainReflector,
  }
  cmLoaders := map[processor.ChangeMapCode]changeMapLoader{
    "cm1":reflector.LoadSHA1ChangeMap,
  }
  cmCreator := map[processor.ChangeMapCode]changeMapCreator{
    "cm1":reflector.NewSHA1ChangeMap,
  }

  g := Generator{
    reflectorTypes:refTypes,
    changeMapCreators:cmCreator,
    changeMapLoaders:cmLoaders,
  }

  cm1, err := g.OpenChangeMap("cm1", "/home/aleksandr/Workspace/cmfile")
  if err != nil {
    panic(err)
  }

  cm2, err := g.NewChangeMap("cm1", "/home/aleksandr/Workspace/testzone", "/home/aleksandr/Workspace/cmfile2")
  if err != nil {
    panic(err)
  }

  ref1, err := g.Reflect("rf1", cm1, cm2)
  if err != nil {
    panic(err)
  }

  err = ref1.Backup()
  if err != nil {
    panic(err)
  }


}
