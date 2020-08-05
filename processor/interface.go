package processor

var (
  FailCode string = "fail"
  SuccessCode = "success"
)

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

type Generator interface {
  Reflect(ReflectorCode, ChangeMap, ChangeMap) (Reflector, error)
  OpenChangeMap(ChangeMapCode, string) (ChangeMap, error)
  NewChangeMap(ChangeMapCode, string, string) (ChangeMap, error)
}

type MetadataDB interface {
  ChangeMapFile(string) string
  RefChangeMapFile(string) string
  ReflectorType(string) ReflectorCode
  ChangeMapType(string) ChangeMapCode
}
