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
  Serialize() string
  Deserialize(fname string) error
  Update([][]string, [][]string) error
  Sync(ChangeMap) error
  ChangeLog(ChangeMap) ([][]string, error)
  RootDir() string
  RootName() string
}

type Generator interface {
  Reflect(ReflectorCode, ChangeMap, ChangeMap) (Reflector, error)
  OpenChangeMap(ChangeMapCode, string, string) (ChangeMap, error)
  NewChangeMap(ChangeMapCode, string) (ChangeMap, error)
}

type MDBRow struct {
  OriginalRoot string
  ReflectionRoot string
  OriginalCM string
  ReflectionCM string
  ReflectionCode ReflectorCode
  CMCode ChangeMapCode
  DriveLabel string
}

type MetadataDB interface {
  Keys() []string
  GetLabel(string) string
  GetRow(string) (MDBRow, error)
  DeleteRow(string) (MDBRow, error)
  InsertRow(string, MDBRow) error
}
