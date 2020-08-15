package processor

var (
  FailCode string = "fail"
  SuccessCode = "success"
)

type ReflectorCode string
type ChangeMapCode string

type Reflector interface {
  Backup() error
}

type Generator interface {
  Reflect(ReflectorCode, string, string) (Reflector, error)
}

type MDBRow struct {
  OriginalRoot string
  ReflectionRoot string
  ReflectionBase string
  HasChanged bool
  ReflectionCode ReflectorCode
  DriveLabel string
}

type MetadataDB interface {
  Keys() []string
  GetRow(string) (MDBRow, error)
  DeleteRow(string) (MDBRow, error)
  InsertRow(string, MDBRow) error
}
