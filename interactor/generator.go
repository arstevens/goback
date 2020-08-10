package interactor

import (
  "github.com/arstevens/goback/processor"
  "fmt"
)

type ReflectorCreator func(processor.ChangeMap, processor.ChangeMap) (processor.Reflector, error)
type ChangeMapLoader func(string, string) (processor.ChangeMap, error)
type ChangeMapCreator func(string) (processor.ChangeMap, error)

type ReflectionGenerator struct {
  reflectorTypes map[processor.ReflectorCode]ReflectorCreator
  changeMapCreators map[processor.ChangeMapCode]ChangeMapCreator
  changeMapLoaders map[processor.ChangeMapCode]ChangeMapLoader
}

func NewReflectionGenerator(refTypes map[processor.ReflectorCode]ReflectorCreator,
  cmCreators map[processor.ChangeMapCode]ChangeMapCreator, cmLoaders map[processor.ChangeMapCode]ChangeMapLoader) ReflectionGenerator {
  return ReflectionGenerator{
      reflectorTypes: refTypes,
      changeMapCreators: cmCreators,
      changeMapLoaders: cmLoaders,
  }
}

func (g ReflectionGenerator) Reflect(code processor.ReflectorCode, originalCM processor.ChangeMap, reflectingCM processor.ChangeMap) (processor.Reflector, error) {
  reflect, ok := g.reflectorTypes[code]
  if !ok {
    return nil, fmt.Errorf("No reflector type with code %s", code)
  }

  reflector, err := reflect(originalCM, reflectingCM)
  if err != nil {
    return nil, fmt.Errorf("Failed to reflect using reflector of code %s: %v", code, err)
  }
  return reflector, nil
}

/* Root is the path to the directory the change map reflects without excluding
the root of the tree directory tree */
func (g ReflectionGenerator) OpenChangeMap(code processor.ChangeMapCode, root string, serial string) (processor.ChangeMap, error) {
  cmLoader, ok := g.changeMapLoaders[code]
  if !ok {
    return nil, fmt.Errorf("No change map with code %s", code)
  }

  changeMap, err := cmLoader(serial, root)
  if err != nil {
    return nil, fmt.Errorf("Failed to open change map with type %s: %v", code, err)
  }
  return changeMap, err
}

/* Root is the path to the directory the change map will reflect including the root
of the directory tree */
func (g ReflectionGenerator) NewChangeMap(code processor.ChangeMapCode, root string) (processor.ChangeMap, error) {
  cmCreator, ok := g.changeMapCreators[code]
  if !ok {
    return nil, fmt.Errorf("No change map with code %s", code)
  }

  changeMap, err := cmCreator(root)
  if err != nil {
    return nil, fmt.Errorf("Failed to create change map with type %s rooted at %s: %v", code, root, err)
  }
  return changeMap, nil
}
