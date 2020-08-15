package reflector
import (
  "testing"
)

func TestReflector(t *testing.T) {
  refRoot := "/run/media/aleksandr/AleksPersonal/testref"
  origRoot := "/home/aleksandr/Workspace/testzone"
  ref, err := NewPlainReflector(origRoot, refRoot)
  if err != nil {
    panic(err)
  }

  err = ref.Backup()
  if err != nil {
    panic(err)
  }
}
