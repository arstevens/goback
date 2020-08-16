package interactor

import (
  "github.com/arstevens/goback/daemon/processor"
  "fmt"
)

type ReflectorCreator func(string, string) (processor.Reflector, error)

type ReflectionGenerator struct {
  reflectorTypes map[processor.ReflectorCode]ReflectorCreator
}

func NewReflectionGenerator(refTypes map[processor.ReflectorCode]ReflectorCreator) ReflectionGenerator {
  return ReflectionGenerator{
      reflectorTypes: refTypes,
  }
}

func (g ReflectionGenerator) Reflect(code processor.ReflectorCode, originalRoot string, reflectingRoot string) (processor.Reflector, error) {
  reflect, ok := g.reflectorTypes[code]
  if !ok {
    return nil, fmt.Errorf("No reflector type with code %s", code)
  }

  reflector, err := reflect(originalRoot, reflectingRoot)
  if err != nil {
    return nil, fmt.Errorf("Failed to reflect using reflector of code %s: %v", code, err)
  }
  return reflector, nil
}
