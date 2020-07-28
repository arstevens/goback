package interactor

type reflectorCreator func(processor.ChangeMap, processor.ChangeMap) (*processor.Reflector, error)
type changeMapLoader func(string) (*processor.ChangeMap, error)
type changeMapCreator func(string, string) (*processor.ChangeMap, error)

type Generator struct {
  reflectorTypes map[processor.ReflectorCode]reflectorCreator
  changeMapCreators map[processor.ChangeMapCode]changeMapCreator
  changeMapLoaders map[processor.ChangeMapCode]changeMapLoader
}

func (g Generator) Reflect(code processor.ReflectorCode, originalCM processor.ChangeMap, reflectingCM processor.ChangeMap) (*processor.Reflector, error) {
  reflect, ok := g.reflectorTypes[code]
  if !ok {
    return nil, fmt.Errorf("No reflector type with code %s", code)
  }

  reflector, err :=  reflect(originalCM, reflectingCM)
  if err != nil {
    return nil, fmt.Errorf("Failed to reflect using reflector of code %s: %v", code, err)
  }
  return reflector, nil
}

func (g Generator) OpenChangeMap(code processor.ChangeMapCode, cmFile string) (*processor.ChangeMap, error) {
  cmLoader, ok := g.changeMapLoaders[code]
  if !ok {
    return nil, fmt.Errorf("No change map with code %s", code)
  }

  changeMap, err := cmLoader(cmFile)
  if err != nil {
    return nil, fmt.Errorf("Failed to open change map with type %s stored at %s: %v", code, cmFile, err)
  }
  return changeMap, err
}

func (g Generator) NewChangeMap(code processor.ChangeMapCode, root string, serialPath string) (*processor.ChangeMap, error) {
  cmCreator, ok := g.changeMapCreators[code]
  if !ok {
    return nil, fmt.Errorf("No change map with code %s", code)
  }

  changeMap, err := cmCreator(root, serialPath)
  if err != nil {
    return nil, fmt.Errorf("Failed to create change map with type %s rooted at %s: %v", code, root, err)
  }
  return changeMap, nil
}
