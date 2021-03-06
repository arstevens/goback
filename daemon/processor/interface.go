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
  ReflectionCode ReflectorCode
  DriveLabel string
  HasChanged bool
}

type MetadataDB interface {
  Keys() []string
  GetRow(string) (MDBRow, error)
  DeleteRow(string) (MDBRow, error)
  InsertRow(MDBRow) error
  UpdateRow(MDBRow) error
}
