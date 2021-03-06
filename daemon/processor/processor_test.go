package processor
import (
  "fmt"
  "time"
  "testing"
  "github.com/arstevens/goback/daemon/interactor"
  "github.com/arstevens/goback/daemon/reflector"
  "github.com/arstevens/goback/daemon/processor"
)

type TestMDB struct {
  db map[string]processor.MDBRow
}

func (mdb *TestMDB) GetRow(key string) (processor.MDBRow, error) {
  row, ok := mdb.db[key]
  if !ok {
    return processor.MDBRow{}, fmt.Errorf("Unknown key %s in TestDB.GetRow()", key)
  }
  return row, nil
}

func (mdb *TestMDB) DeleteRow(key string) (processor.MDBRow, error) {
  row := mdb.db[key]
  delete(mdb.db, key)
  return row, nil
}

func (mdb *TestMDB) InsertRow(row processor.MDBRow) error {
  mdb.db[row.OriginalRoot] = row
  return nil
}

func (mdb *TestMDB) UpdateRow(row processor.MDBRow) error {
  mdb.db[row.OriginalRoot] = row
  return nil
}

func (mdb *TestMDB) Keys() []string {
  keys := make([]string, 0, len(mdb.db))
  for key, _ := range mdb.db {
    keys = append(keys, key)
  }
  return keys
}

func TestMonitorSystem(t *testing.T) {

}
/*
func TestProcessor(t *testing.T) {
  refTypes := map[processor.ReflectorCode]interactor.ReflectorCreator{
    "sh1ref":reflector.NewPlainReflector,
  }

  generator := interactor.NewReflectionGenerator(refTypes)
  mdb := &TestMDB{db:make(map[string]processor.MDBRow)}

  comChan := make(chan string)
  updateChan := make(chan string)
  go processor.MonitorSystem(mdb, updateChan)
  go processor.CommandProcessor(generator, mdb, comChan, updateChan)

  origRoot := "/home/aleksandr/Workspace/testzone"
  refRoot := "/run/media/aleksandr/AleksPersonal/testref"
  bakCmd := NewBackupCommand+":"+origRoot+","+refRoot+","+"sh1ref"+","+"cm1"
  comChan<-bakCmd
  resp := <-comChan
  fmt.Println(resp)
  row, _ := mdb.GetRow(origRoot)
  fmt.Println(row)
  <-time.After(time.Second*60)
  row, _ = mdb.GetRow(origRoot)
  fmt.Println(row)

}
/*
func TestLabel(t *testing.T) {
  resp := labelToMountPoint("AleksPersonal")
  fmt.Println("Label to Mountpoint: "+resp)
  resp, extra := pathToLabel("/run/media/aleksandr/AleksPersonal/newdir")
  fmt.Println("Path to label: "+resp+","+extra)

}
/*

func TestBackup(t *testing.T) {
  // New Backup
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
  mdb := TestMDB{db:make(map[string]processor.MDBRow)}

  origRoot := "/home/aleksandr/Workspace/testzone"
  refRoot := "/home/aleksandr/Workspace/newenv/testzone"
  refCode := "sh1ref"
  cmCode := "cm1"
  nback := NewBackupCommand+":"+origRoot+","+refRoot+","+refCode+","+cmCode

  comChan := make(chan string)
  updateChan := make(chan processor.UpdatePackage)
  defer close(comChan)
  defer close(updateChan)

  go processor.CommandProcessor(generator, &mdb, comChan, updateChan)
  comChan<-nback
  resp := <-comChan
  fmt.Println(resp)

  // Update
  fileUpdates := [][]string{[]string{"Hive_Whitepaper_v3_Fluid.odt"},2:[]string{}}
  update := processor.UpdatePackage{
    Backup: true,
    OriginalRoot: origRoot,
    FileUpdates: fileUpdates,
    DirUpdates: make([][]string, 3),
  }

  updateChan<-update
  time.Sleep(time.Second*5)

}
/*
func TestBackup(t *testing.T) {
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

  origRoot := "/home/aleksandr/Workspace/testzone"
  refRoot := "/home/aleksandr/Workspace/testzone2"
  refCode := "sh1ref"
  cmCode := "cm1"

  mdb := TestMDB{db:make(map[string]processor.MDBRow)}
  row := processor.MDBRow{
    OriginalRoot: origRoot,
    ReflectionRoot: refRoot,
    OriginalCM: `0,testzone,CbOue9P6wkR2SyeceD28O0ippzs=,true)3,graphic_mockups,UjYLX8f9kyEAnjfsX4SG3n+Ef4A=,true)4,honeycoin_load_change.drawio,z3s12tTqZEgwoz1jNia3B4oksYM=,false)|)5,honeycoin_mining_v1.drawio,awVf4lhtZtpqyr790J+9trtwwu4=,false)|)6,load_distribution_v1.drawio,HHkH1VueYC7aYUDkTjMlTTulLbc=,false)|)7,load_pickup_v1.drawio,M031xhXfr9laZ75H6ZRwWq6bAKU=,false)|)8,simple_exchange_v1.drawio,z8ld9Lfx3MbdKa34RelzzMLP/0k=,false)|)9,topology_graph_v1.drawio,kO9QBq3i9QDm2wQb9ENkr+NkLSU=,false)|)|)10,graphics,zuiDznL+4eDuCYzwSFmhkWQ2xtg=,true)11,honeycoin_load_change.png,/ow/GtiyJo3oh1r0wpcCkPWaF54=,false)|)12,honeycoin_mining_v1.png,7PABOxbIMhYTHg9RNEEBcugazW4=,false)|)13,load_distribution_v1.png,JgzHHoLznr1pW4huCV71J7hoOdY=,false)|)14,load_pickup_v1.png,UfljFkCsLmCMLOXQG0kUkw6AWkc=,false)|)15,simple_exchange_v1.png,eSj6Zmp8nBkml15LzTAuEBpggmU=,false)|)16,topology_graph_v1.png,eLdHhek5e7wj0pC7/1QwyG36rR4=,false)|)|)2,Hive_Whitepaper_v3_Fluid.odt,q23GRVlHATRR7QPhEZPYoVkLwqQ=,false)|)|)`,
    ReflectionCM: `0,testzone2,random,true)`,
    ReflectionCode: processor.ReflectorCode(refCode),
    CMCode: processor.ChangeMapCode(cmCode),
  }
  err := mdb.InsertRow(origRoot, row)
  if err != nil {
    panic(err)
  }
  back := string(BackupCommand)+":"+origRoot

  comChan := make(chan string)
  updateChan := make(chan processor.UpdatePackage)
  defer close(comChan)
  defer close(updateChan)

  go processor.CommandProcessor(generator, &mdb, comChan, updateChan)
  comChan<-back
  resp := <-comChan
  fmt.Println(resp)
}
*/
/*
func TestDetector(t *testing.T) {
  d := newFsDetector()
  err := d.Watch("/home/aleksandr/Workspace/testzone")
  if err != nil {
    panic(err)
  }

  resp, err := d.NextChange()
  if err != nil {
    panic(err)
  }

  fmt.Println(resp)
}
*/
