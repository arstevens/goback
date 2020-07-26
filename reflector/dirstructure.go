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
  idCount int
  idMap map[string]int
  dirHash dirHashFunction
}

/* Generates all steps required to turn root into foreignRoot */
func treeDifference(root fileNode, foreignRoot fileNode) [][]string {
  differences := make([][]string, 3)

  // Handle case of new file creation
  for fId, fChild := range foreignRoot.children {
    if _, ok := root.children[fId]; !ok {
      differences[createCode] = append(differences[createCode], extendPath(root.name, fChild.name))
    }
  }

  // Handle cases of file deletion or name change
  for id, child := range root.children {
    foreignChild, ok := foreignRoot.children[id]
    if !ok {
      differences[deleteCode] = append(differences[deleteCode], extendPath(root.name, child.name))
    } else if !bytes.Equal(child.hash, foreignChild.hash) {
      if child.name != foreignChild.name {
        differences[updateCode] = append(differences[updateCode], extendPath(root.name, child.name + paramSep + foreignChild.name))
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
      newPath := extendPath(prefix, path)
      highLevel[code] = append(highLevel[code], newPath)
    }
  }
  return highLevel
}

/* addChild() and deleteChild() accept an idPath for the designated node
including the root id(0) even though there can only be a single root */
func (d *directoryTree) addChild(path string, hash []byte, isDir bool) error {
  cleanPath := splitPath(path)
  cleanPath = cleanPath[:len(cleanPath) - 1]
  idPath, err := pathToIdPath(cleanPath, d.idMap)
  if err != nil {
    return fmt.Errorf("Failed to create idPath in directoryTree.addChild(): %v", err)
  }

  parent := d.root
  for i := 0; i < len(idPath); i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %d", idPath[i])
    }
  }

  name := filepath.Base(path)

  d.idCount++
  d.idMap[path] = d.idCount
  parent.children[d.idCount] = *newFileNode(d.idCount, name, hash, isDir)
  return nil
}

func (d *directoryTree) deleteChild(path string) error {
  idPath, err := pathToIdPath(splitPath(path), d.idMap)
  if err != nil {
    return fmt.Errorf("Failed to create idPath in directoryTree.deleteChild(): %v", err)
  }
  if len(idPath) == 0 {
    d.root = *newFileNode(-1, "", []byte{}, false)
  }

  parent := d.root
  for i := 0; i < len(idPath) - 1; i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %d", idPath[i])
    }
  }

  delete(parent.children, idPath[len(idPath) - 1])
  delete(d.idMap, path)
  return nil
}

func (d *directoryTree) renameChild(path string, name string) error {
  idPath, err := pathToIdPath(splitPath(path), d.idMap)
  if err != nil {
    return fmt.Errorf("Failed to create idPath in directoryTree.renameChild(): %v", err)
  }

  parent := d.root
  for i := 0; i < len(idPath) - 1; i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %d in directoryTree.renameChild(): %v", idPath[i], err)
    }
  }

  child := parent.children[idPath[len(idPath) - 1]]
  id := d.idMap[path]
  delete(d.idMap, child.name)
  d.idMap[changePathBase(path, name)] = id

  child.name = name
  parent.children[idPath[len(idPath) - 1]] = child
  return nil
}

func (d *directoryTree) updateHash(path string, hash []byte) error {
  idPath, err := pathToIdPath(splitPath(path), d.idMap)
  if err != nil {
    return fmt.Errorf("Failed to create idPath in directoryTree.renameChild(): %v", err)
  }

  parent := d.root
  for i := 0; i < len(idPath) - 1; i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %d in directoryTree.renameChild(): %v", idPath[i], err)
    }
  }

  child := parent.children[idPath[len(idPath) - 1]]
  child.hash = hash
  parent.children[idPath[len(idPath) - 1]] = child
  fmt.Println(child.id, base64.StdEncoding.EncodeToString(child.hash))
  return nil
}

func (d *directoryTree) hash() {
  d.root.applyDirHash(d.dirHash)
}

func newDirectoryTree(rootPath string, fHash fileHashFunction, dHash dirHashFunction) (*directoryTree, error) {
  dt := directoryTree{
    root: *newFileNode(0, filepath.Base(rootPath), []byte{}, true),
    dirHash: dHash,
    idMap: make(map[string]int),
    idCount: 1,
  }

  dt.idMap[filepath.Base(rootPath)] = 0
  err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
    // Remove irrelavent filesystem path information
    relativePath := createRelativePath(path, rootPath)
    if relativePath == "" {
      return nil
    }
    relativePath = cleanPath(relativePath)

    hash := []byte{}
    if !info.IsDir() {
      hash = fHash(path)
    }

    return dt.addChild(relativePath, hash, info.IsDir())
  })
  if err != nil {
    return nil, fmt.Errorf("Walking path failed: %v", err)
  }
  dt.hash()

  return &dt, nil
}

/* Convertes a filesystem-type path into a slice of ids that can be
used for traversing a directory tree */
func pathToIdPath(splitPath []string, idMap map[string]int) ([]int, error) {
  idPath := make([]int, 0, len(splitPath))

  path := ""
  if len(splitPath) > 0 {
    path = splitPath[0]
    id, ok := idMap[path]
    if ok {
      idPath = append(idPath, id)
    } else {
      return nil, fmt.Errorf("Reference to non-existent node (%s) in pathToIdPath()", path)
    }
  }

  for i := 1; i < len(splitPath); i++ {
    path = extendPath(path, splitPath[i])
    id, ok := idMap[path]
    if ok {
      idPath = append(idPath, id)
    } else {
      return nil, fmt.Errorf("Reference to non-existent node (%s) in pathToIdPath()", path)
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
  d.idMap = make(map[string]int)

  var root fileNode
  root.deserialize([]byte(tokens[0]))
  d.idMap[root.name] = root.id
  d.idCount = root.id
  var prefix string

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
      prefix = filepath.Dir(prefix)
    } else {
      var child fileNode
      child.deserialize([]byte(tk))
      (*parent).children[child.id] = child
      stack = append(stack, *parent)
      parent = &child
      prefix = extendPath(prefix, child.name)

      d.idMap[prefix] = child.id
      if child.id > d.idCount {
        d.idCount = child.id
      }
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

  for key, _ := range f.children {
    if f.children[key].isDir {
      child := f.children[key]
      child.applyDirHash(dHash)
      f.children[key] = child
    }
  }
  f.hash = dHash(*f)
}

func (f fileNode) serialize() string {
  hashSerial := base64.StdEncoding.EncodeToString(f.hash)
  dirMark := fmt.Sprintf("%t", f.isDir)
  return strconv.Itoa(f.id) + paramSep + f.name + paramSep + hashSerial + paramSep + dirMark
}

func (f *fileNode) deserialize(data []byte) error {
  f.children = make(map[int]fileNode)
  tokens := strings.Split(string(data), paramSep)
  if len(tokens) < 4 {
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
  f.isDir = false
  if tokens[3] == "true" {
    f.isDir = true
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
