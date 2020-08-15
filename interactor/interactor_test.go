package interactor

import (
  "github.com/arstevens/goback/processor"
  "github.com/arstevens/goback/reflector"
  "testing"
)

func TestInteractor(t *testing.T) {
  refTypes := map[processor.ReflectorCode]ReflectorCreator{
    "rf1":reflector.NewPlainReflector,
  }

  g := ReflectionGenerator{
    reflectorTypes:refTypes,
  }

  ref1, err := g.Reflect("rf1", "/home/aleksandr/Workspace/testzone", "/home/aleksandr/Workspace/testzone2")
  if err != nil {
    panic(err)
  }

  err = ref1.Backup()
  if err != nil {
    panic(err)
  }
}
