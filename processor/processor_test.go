package processor
import (
  "testing"
  "github.com/arstevens/goback/interactor"
  "github.com/arstevens/goback/reflector"
  "github.com/arstevens/goback/processor"
)

type TestMDB struct {
  db map[string]MDBRow
}

func (mdb *TestMDB) GetRow(key string) (MDBRow, error) {
  return mdb.db[key], nil
}

func (mdb *TestMDB) DeleteRow(key string) (MDBRow, error) {
  row := mdb.db[key]
  delete(mdb.db, key)
  return row, nil
}

func (mdb *TestMDB) InsertRow(key string, row MDBRow) error {
  mdb.db[key] = row
  return nil
}

func TestProcessor(t *testing.T) {
  refTypes := map[processor.ReflectorCode]interactor.ReflectorCreator{
    "sh1ref":reflector.NewPlainReflector,
  }
  cmLoaders := map[processor.ChangeMapCode]interactor.ChangeMapLoader{
    "cm1":reflector.LoadSHA1ChangeMap,
  }
  cmCreators := map[processor.ChangeMapCode]interactor.ChangeMapCreator{
    "cm1":reflector.NewSHA1ChangeMap,
  }

  generator := interactor.NewReflectionGenerator(refTypes, cmCreators, cmLoaders)
  mdb := TestMDB{db:make(map[string]MDBRow)}

  origRoot := "/home/aleksandr/Workspace/testzone"
  refRoot := "/home/aleksandr/Workspace/testzone2"
  refCode := "sh1ref"
  cmCode := "cm1"
  nback := NewBackupCommand+":"+origRoot+","+refRoot+","+refCode+","+cmCode

  comChan := make(chan string)
  updateChan := make(chan UpdatePackage)
  defer close(comChan)
  defer close(updateChan)

  go CommandProcessor(generator, &mdb, comChan, updateChan)
  comChan<-nback
  resp := <-comChan
  fmt.Println(resp)
}
