package processor

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
