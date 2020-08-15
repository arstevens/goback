package reflector

import (
  "github.com/arstevens/goback/processor"
  "fmt"
  "os"
)

type PlainReflector struct {
  originalDirectory string
  reflectingDirectory string
}

// Satisfies interactor.reflectorCreator
func NewPlainReflector(original, reflecting string) (processor.Reflector, error) {
  pr := PlainReflector{
    originalDirectory: original,
    reflectingDirectory: reflecting,
  }
  return &pr, nil
}

/* PlainReflector.Backup() finds the differences between the
reflecting map and the original map and performs the necessary
operations to turn the reflecting directory into the original directory */
func (p PlainReflector) Backup() error {
  err := os.RemoveAll(p.reflectingDirectory)
  if err != nil {
    return fmt.Errorf("Couldn't delete old contents of directory in Backup(): %v", err)
  }

  err = copyDir(p.originalDirectory, p.reflectingDirectory)
  if err != nil {
    return fmt.Errorf("Couldn't copy directory over in Backup(): %v", err)
  }
  return nil
}
