package interactor

import (
  "github.com/arstevens/goback/processor"
  "github.com/arstevens/goback/reflector"
  "testing"
)

func TestInteractor(t *testing.T) {
  refTypes := map[processor.ReflectorCode]reflectorCreator{
    "1":reflector.NewPlainReflector,
  }

}
