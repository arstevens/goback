package processor

type ReflectorCode string
type ChangeMapCode string

type Reflector interface {
  Backup() error
  Recover() error
}

type ChangeMap interface {
  Serialize() error
  Deserialize(fname string) error
  Update([][]string, [][]string) error
  Sync(ChangeMap) error
  ChangeLog(ChangeMap) ([][]string, error)
  RootDir() string
  RootName() string
}
