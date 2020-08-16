package main

import (
  "github.com/arstevens/goback/daemon/processor"
  "strconv"
  "strings"
  "io/ioutil"
  "sync"
  "fmt"
  "log"
  "os"
)

const (
  dbSeparator string = ","
)

type FileMetadataDB struct {
  rowsByKey map[string]processor.MDBRow
  dbPath string
  mutex *sync.Mutex
}

func NewFileMetadataDB(dbLoc string) *FileMetadataDB {
  fdb :=  &FileMetadataDB{
    rowsByKey: make(map[string]processor.MDBRow),
    dbPath: dbLoc,
    mutex: &sync.Mutex{},
  }

  err := fdb.deserializeDB()
  if err != nil {
    log.Printf("Failed to deserialize file db in NewFileMetadataDB(): %v", err)
  }
  return fdb
}

func (f *FileMetadataDB) GetRow(key string) (processor.MDBRow, error) {
  f.mutex.Lock()
  defer f.mutex.Unlock()

  row, ok := f.rowsByKey[key]
  if !ok {
    return processor.MDBRow{}, fmt.Errorf("No row with key %s in FileMetadataDB.GetRow()", key)
  }
  return row, nil
}

func (f *FileMetadataDB) InsertRow(row processor.MDBRow) error {
  f.mutex.Lock()
  defer f.mutex.Unlock()

  if _, ok := f.rowsByKey[row.OriginalRoot]; ok {
    return fmt.Errorf("Row already exists with key %s", row.OriginalRoot)
  }
  f.rowsByKey[row.OriginalRoot] = row

  if err := f.writeToDisk(); err != nil {
    return fmt.Errorf("Failed to write to disk in FileMetadataDB.InsertRow(): %v", err)
  }
  return nil
}

func (f *FileMetadataDB) DeleteRow(key string) (processor.MDBRow, error) {
  f.mutex.Lock()
  defer f.mutex.Unlock()

  row, ok := f.rowsByKey[key]
  if !ok {
    return processor.MDBRow{}, fmt.Errorf("Key %s does not exist in FileMetadataDB.DeleteRow()", key)
  }

  delete(f.rowsByKey, key)
  if err := f.writeToDisk(); err != nil {
    return processor.MDBRow{}, fmt.Errorf("Failed to write to disk in FileMetadataDB.DeleteRow(): %v", err)
  }
  return row, nil
}

func (f *FileMetadataDB) UpdateRow(row processor.MDBRow) error {
  f.mutex.Lock()
  defer f.mutex.Unlock()

  _, ok := f.rowsByKey[row.OriginalRoot]
  if ok {
    return fmt.Errorf("Couldn't update row with key %s. Does not exist in FileMetadataDB.UpdateRow()", row.OriginalRoot)
  }

  f.rowsByKey[row.OriginalRoot] = row
  if err := f.writeToDisk(); err != nil {
    return fmt.Errorf("Failed to write to disk in FileMetadataDB.UpdateRow(): %v", err)
  }
  return nil
}

func (f *FileMetadataDB) Keys() []string {
  f.mutex.Lock()
  defer f.mutex.Unlock()

  keys := make([]string, 0, len(f.rowsByKey))
  for key, _ := range f.rowsByKey {
    keys = append(keys, key)
  }
  return keys
}

func (f *FileMetadataDB) writeToDisk() error {
  serial := f.serializeDB()
  err := ioutil.WriteFile(f.dbPath, serial, 0644)
  if err != nil {
    return fmt.Errorf("Failed to write file in FileMetadataDB.writeToDisk(): %v", err)
  }
  return nil
}

func (f *FileMetadataDB) serializeDB() []byte {
  serial := ""
  for _, row := range f.rowsByKey {
    serialRow := row.OriginalRoot+dbSeparator+row.ReflectionRoot+dbSeparator+
      row.ReflectionBase+dbSeparator+string(row.ReflectionCode)+dbSeparator+
      row.DriveLabel+dbSeparator+strconv.FormatBool(row.HasChanged)
    serial += serialRow+"\n"
  }
  return []byte(serial)
}

func (f *FileMetadataDB) deserializeDB() error {
  if _, err := os.Stat(f.dbPath); os.IsNotExist(err) {
    f, err := os.Create(f.dbPath)
    if err == nil {
      f.Close()
    }
    return nil
  }
  serial, err := ioutil.ReadFile(f.dbPath)
  if err != nil {
    return fmt.Errorf("Failed to read from %s in FileMetadataDB.deserializeDB(): %v", f.dbPath, err)
  }

  rawRows := strings.Split(string(serial), "\n")
  for _, rawRow := range rawRows {
    entries := strings.Split(rawRow, dbSeparator)
    if len(entries) < 5 {
      return fmt.Errorf("Not enough entries when reading row in deserializeDB()")
    }
    hasChanged, err := strconv.ParseBool(entries[5])
    if err != nil {
      return fmt.Errorf("Failed to parse bool field in deserializeDB(): %v", err)
    }

    f.rowsByKey[entries[0]] = processor.MDBRow{
      OriginalRoot: entries[0],
      ReflectionRoot: entries[1],
      ReflectionBase: entries[2],
      ReflectionCode: processor.ReflectorCode(entries[3]),
      DriveLabel: entries[4],
      HasChanged: hasChanged,
    }
  }
  return nil
}
