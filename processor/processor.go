package processor

import (
  "strings"
  "log"
  "fmt"
)

type CommandCode string

const (
  BackupCommand CommandCode = "bak"
  NewBackupCommand = "n_bak"
  UnbackupCommand = "u_bak"
)

func CommandProcessor(gen Generator, mdb MetadataDB, comChan chan string, updateChan <-chan string) {
  for {
    select {
      case cmd, ok := <-updateChan:
        if !ok {
          return
        }
        if err := executeCommand(cmd, gen, mdb); err != nil {
          log.Printf("Failed to execute command in CommandProcessor: %v\n", err)
        }
      case cmd, ok := <-comChan:
        if !ok {
          return
        }
        if err := executeCommand(cmd, gen, mdb); err != nil {
          log.Printf("Failed to execute command(%s) in CommandProcessor: %v\n", cmd, err)
          comChan<-FailCode
        } else {
          comChan<-SuccessCode
        }
    }
  }
}

/* Command format: command_code:param1,param2,... */
func executeCommand(cmd string, gen Generator, mdb MetadataDB) error {
  cmdComponents := strings.Split(cmd, ":")
  if len(cmdComponents) < 2 {
    return fmt.Errorf("Invalid command input(%s) in executeCommand()", cmd)
  }
  cmdType := CommandCode(cmdComponents[0])
  params := strings.Split(cmdComponents[1], ",")
  var err error

  switch cmdType {
    case BackupCommand:
      err = backupCommand(params, gen, mdb)
    case NewBackupCommand:
      err = newBackupCommand(params, gen, mdb)
    case UnbackupCommand:
      err = unbackupCommand(params, gen, mdb)
    default:
      return fmt.Errorf("Unknown command(%s) in executeCommand()", cmd)
  }

  if err != nil {
    return fmt.Errorf("Couldn't process command in executeCommand(): %v", err)
  }
  return nil
}

func backupCommand(params []string, gen Generator, mdb MetadataDB) error {
  if len(params) < 1 {
    return fmt.Errorf("Not enough params in backupCommand()")
  }
  backupRoot := params[0]
  mdbRow, err := mdb.GetRow(backupRoot)
  if err != nil {
    return fmt.Errorf("Couldn't retrieve row in backupCommand(): %v", err)
  }

  if !mdbRow.HasChanged {
    log.Printf("No need to backup. Directory unchanged")
    return nil
  }
  if mdbRow.ReflectionRoot == "" {
    log.Printf("No need to backup. Device not mounted")
    return nil
  }

  reflector, err := gen.Reflect(mdbRow.ReflectionCode, mdbRow.OriginalRoot, mdbRow.ReflectionRoot)
  if err != nil {
    return fmt.Errorf("Failed to create reflector in backupCommand(): %v", err)
  }
  err = reflector.Backup()
  if err != nil {
    return fmt.Errorf("Failed to reflect in backupCommand(): %v", err)
  }

  mdbRow.HasChanged = false
  err = mdb.UpdateRow(mdbRow)
  if err != nil {
    return fmt.Errorf("Failed to update row in backupCommand(): %v", err)
  }
  return nil
}

func newBackupCommand(params []string, gen Generator, mdb MetadataDB) error {
  if len(params) < 3 {
    return fmt.Errorf("Not enough paramaters in newBackupCommand()")
  }
  origRoot, refRoot := params[0], params[1]
  refCode := ReflectorCode(params[2])

  reflector, err := gen.Reflect(refCode, origRoot, refRoot)
  if err != nil {
    return fmt.Errorf("Couldn't reflect in newBackupCommand(): %v", err)
  }
  err = reflector.Backup()
  if err != nil {
    return fmt.Errorf("Couldn't backup in newBackupCommand(): %v", err)
  }

  driveLabel, refBase := pathToLabel(refRoot)

  mdbRow := MDBRow{
    OriginalRoot: origRoot,
    ReflectionRoot: refRoot,
    ReflectionCode: refCode,
    ReflectionBase: refBase,
    DriveLabel: driveLabel,
    HasChanged: false,
  }
  err = mdb.InsertRow(mdbRow)
  if err != nil {
    fmt.Errorf("Couldnt insert row in newBackupCommand(): %v", err)
  }

  return nil
}

func unbackupCommand(params []string, gen Generator, mdb MetadataDB) error {
  if len(params) < 1 {
    return fmt.Errorf("Not enough parameters in unbackupCommand()")
  }

  origRoot := params[0]
  if _, err := mdb.DeleteRow(origRoot); err != nil {
    return fmt.Errorf("Failed to remove %s for database in unbackupCommand(): %v", origRoot, err)
  }
  return nil
}
