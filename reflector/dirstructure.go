package reflector

import (
  "encoding/base64"
  "path/filepath"
  "strings"
  "strconv"
  "bytes"
  "fmt"
  "os"
)

const (
  jumpBack = "|"
  nodeSep = ")"
  paramSep = ","
)

type fileHashFunction func(string) []byte
type dirHashFunction func(fileNode) []byte

type directoryTree struct {
  root fileNode
}

/* Generates all steps required to turn root into foreignRoot */
func treeDifference(root fileNode, foreignRoot fileNode) [][]string {
  differences := make([][]string, 3)

  // Handle case of new file creation
  for fId, fChild := range foreignRoot.children {
    if _, ok := root.children[fId]; !ok {
      differences[createCode] = append(differences[createCode], root.name + "/" + fChild.name)
    }
  }

  // Handle cases of file deletion or name change
  for id, child := range root.children {
    foreignChild, ok := foreignRoot.children[id]
    if !ok {
      differences[deleteCode] = append(differences[deleteCode], root.name+"/"+child.name)
    } else if !bytes.Equal(child.hash, foreignChild.hash) {
      if child.name != foreignChild.name {
        differences[updateCode] = append(differences[updateCode], root.name+"/"+child.name + paramSep + foreignChild.name)
      }
      subDifferences := treeDifference(child, foreignChild)
      differences = mergeDifferenceMaps(differences, subDifferences, root.name)
    }
  }
  return differences
}

func mergeDifferenceMaps(highLevel [][]string, lowLevel [][]string, prefix string) [][]string {
  for code, diffs := range lowLevel {
    for _, path := range diffs {
      newPath := prefix + "/" + path
      highLevel[code] = append(highLevel[code], newPath)
    }
  }
  return highLevel
}

/* addChild() and deleteChild() accept an idPath for the designated node
including the root id(0) even though there can only be a single root */
func (d *directoryTree) addChild(idPath []int, id int, name string, hash []byte, isDir bool) error {
  if len(idPath) > 0 && idPath[0] != d.root.id {
    return fmt.Errorf("Invalid root id %d", idPath[0])
  }

  parent := d.root
  for i := 1; i < len(idPath); i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %d", idPath[i])
    }
  }

  parent.children[id] = *newFileNode(id, name, hash, isDir)
  return nil
}

func (d *directoryTree) deleteChild(idPath []int) error {
  if len(idPath) > 0 {
    if idPath[0] != d.root.id {
      return fmt.Errorf("Invalid root id %d", idPath[0])
    } else if len(idPath) == 1 {
      d.root = *newFileNode(-1, "", []byte{}, false)
    }
  }

  parent := d.root
  for i := 1; i < len(idPath) - 1; i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %d", idPath[i])
    }
  }

  delete(parent.children, idPath[len(idPath) - 1])
  return nil
}

func newDirectoryTree(rootPath string, fHash fileHashFunction, dHash dirHashFunction) (*directoryTree, error) {
  var idCount int
  dt := directoryTree{
    root: *newFileNode(idCount, filepath.Base(rootPath), []byte{}, true),
  }
  idCount++

  idMap := make(map[string]int)
  idMap[filepath.Base(rootPath)] = 0
  err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
    // Remove irrelavent filesystem path information
    basePath := path
    path = strings.Replace(path, rootPath, "", 1)

    splitPath := strings.Split(path, "/")
    idPath, err := pathToIdPath(splitPath, idMap)
    if err != nil {
      return fmt.Errorf("Couldn't construct id for %s path: %v", path, err)
    }
    idCount++
    idMap[splitPath[len(splitPath) - 1]] = idCount

    hash := []byte{}
    if !info.IsDir() {
      hash = fHash(basePath)
    }

    return dt.addChild(idPath, idCount, splitPath[len(splitPath) - 1], hash, info.IsDir())
  })
  if err != nil {
    return nil, fmt.Errorf("Walking path failed: %v", err)
  }
  dt.root.applyDirHash(dHash)

  return &dt, nil
}

/* Convertes a filesystem-type path into a slice of ids that can be
used for traversing a directory tree */
func pathToIdPath(splitPath []string, idMap map[string]int) ([]int, error) {
  fmt.Println(splitPath)
  idPath := make([]int, 0, len(splitPath))

  for i := 0; i < len(splitPath); i++ {
    loc := splitPath[i]
    id, ok := idMap[loc]
    if ok {
      idPath = append(idPath, id)
    } else {
      return nil, fmt.Errorf("Reference to non-existent node (%d) in pathToIdPath()", id)
    }
  }
  return idPath, nil
}

func (d directoryTree) serialize() []byte {
  treeSerial := serializeTree(d.root)
  return []byte(treeSerial)
}

func serializeTree(root fileNode) string {
  serial := root.serialize() + nodeSep
  for _, node := range root.children {
    serial += serializeTree(node)
  }
  serial += string(jumpBack) + nodeSep
  return serial
}

func (d *directoryTree) deserialize(data []byte) {
  tokens := tokenizeSerial(data)
  stack := make([]fileNode, 0)

  var root fileNode
  root.deserialize([]byte(tokens[0]))
  d.root = root
  parent := &d.root

  for i := 1; i < len(tokens); i++ {
    tk := tokens[i]
    if tk == jumpBack {
      /* If we've escaped out of the root directory
      then we've already processed the last node
      in the tree structure */
      if len(stack) < 1 {
        return
      }

      parent = &stack[len(stack) - 1]
      stack = stack[:len(stack) - 1]
    } else {
      var child fileNode
      child.deserialize([]byte(tk))
      (*parent).children[child.id] = child
      stack = append(stack, *parent)
      parent = &child
    }
  }
}

func tokenizeSerial(data []byte) []string {
  strData := string(data)
  tokens := strings.Split(strData, nodeSep)
  return tokens
}

/* Creates a deep copy of a directoryTree */
func (d directoryTree) duplicate() directoryTree {
  var dtree directoryTree
  dtree.root = duplicate(d.root)
  return dtree
}

/* Represents a single file in the modeled directory structure
holding references to all child files */
type fileNode struct {
  id int
  name string
  hash []byte
  isDir bool
  children map[int]fileNode
}

func newFileNode(id int, name string, hash []byte, isDir bool) *fileNode {
  return &fileNode{
    id: id,
    name: name,
    hash: hash,
    isDir: isDir,
    children: make(map[int]fileNode),
  }
}

/* Applies directory hash function to all children who
do not already have hashes and then applies hash to self. Only
nodes that should */
func (f *fileNode) applyDirHash(dHash dirHashFunction) {
  if !f.isDir {
    return
  }

  for _, node := range f.children {
    if node.isDir {
      node.applyDirHash(dHash)
    }
  }
  f.hash = dHash(*f)
}

func (f fileNode) serialize() string {
  hashSerial := base64.StdEncoding.EncodeToString(f.hash)
  return strconv.Itoa(f.id) + paramSep + f.name + paramSep + hashSerial
}

func (f *fileNode) deserialize(data []byte) error {
  f.children = make(map[int]fileNode)
  tokens := strings.Split(string(data), paramSep)
  if len(tokens) < 3 {
    return fmt.Errorf("Invalid filenode input in fileNode.deserialize(). unable to deserialize")
  }

  id, err := strconv.Atoi(tokens[0])
  if err != nil {
    return fmt.Errorf("Unable to convert %s to int in fileNode.deserialize(): %v", tokens[0], err)
  }
  f.id, f.name = id, tokens[1]
  f.hash, err = base64.StdEncoding.DecodeString(tokens[2])
  if err != nil {
    return fmt.Errorf("Unable to decode hash in fileNode.deserialize(): %v", err)
  }
  return nil
}

/* Creates a deep copy of a fileNode */
func duplicate(root fileNode) fileNode {
  newRoot := *newFileNode(root.id, root.name, root.hash, root.isDir)
  for id, node := range root.children {
    newRoot.children[id] = duplicate(node)
  }
  return newRoot
}
