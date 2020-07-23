package reflector

import (
  "path/filepath"
  "strings"
  "strconv"
  "fmt"
  "os"
)

const (
  jumpBack = "|"
  nodeSep = ")"
  paramSep = ","
)

type fileHashFunction func(*os.File) string
type dirHashFunction func(fileNode) string

type directoryTree struct {
  root fileNode
}

/* Entries for updates and creations are in the form
id_0/id_1/.../id_n,name where ids are the path down the tree
and the name is the name of the new file or updated file name
*/
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
    } else if child.hash != foreignChild.hash {
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

/* Add and Delete child take the path down the tree
to the node to be inserted/deleted without the root node
included in the path */
func (d *directoryTree) addChild(idPath []int, id int, name string, hash string) error {
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

  parent.children[id] = *newFileNode(id, name, hash)
  return nil
}

func (d *directoryTree) deleteChild(idPath []int) error {
  if len(idPath) > 0 {
    if idPath[0] != d.root.id {
      return fmt.Errorf("Invalid root id %d", idPath[0])
    } else if len(idPath) == 1 {
      d.root = *newFileNode(-1, "", "")
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

type fileNode struct {
  id int
  name string
  hash string
  children map[int]fileNode
}

func newDirectoryTree(rootPath string, fHash fileHashFunction, dHash dirHashFunction) *directoryTree {
  var idCount int
  dt := directoryTree{
    root: *newFileNode(idCount, filepath.Base(rootPath), ""),
  }
  idCount++

  idMap := make(map[string]int)
  idMap[filepath.Base(rootPath)] = 0
  err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
    basePath := path
    path = strings.Replace(path, rootPath, "", 1)
    if path == "" {
      return nil
    }

    splitPath := strings.Split(path, "/")
    idPath, err := pathToIdPath(splitPath, idMap)
    if err != nil {
      panic(err)
    }
    idCount++
    idMap[splitPath[len(splitPath) - 1]] = idCount

    hash := ""
    if !info.IsDir() {
      file, err := os.Open(basePath)
      if err != nil {
        panic(err)
      }
      defer file.Close()
      hash = fHash(file)
    }

    return dt.addChild(idPath, idCount, splitPath[len(splitPath) - 1], hash)
  })
  if err != nil {
    panic(err)
  }
  dt.root.applyHash(dHash)

  return &dt
}

func pathToIdPath(splitPath []string, idMap map[string]int) ([]int, error) {
  fmt.Println(splitPath)
  idPath := make([]int, 0, len(splitPath))

  for i := 0; i < len(splitPath); i++ {
    loc := splitPath[i]
    id, ok := idMap[loc]
    if ok {
      idPath = append(idPath, id)
    }
    /*
    if !ok {
      return nil, fmt.Errorf("Reference to non-existent id")
    }
    idPath = append(idPath, id)
    */
  }
  return idPath, nil
}

func newFileNode(id int, name string, hash string) *fileNode {
  return &fileNode{
    id: id,
    name: name,
    hash: hash,
    children: make(map[int]fileNode),
  }
}

func (f *fileNode) applyHash(dHash dirHashFunction) {
  for _, node := range f.children {
    if node.hash == "" {
      node.applyHash(dHash)
    }
  }
  f.hash = dHash(*f)
}

func (f fileNode) serialize() string {
  return strconv.Itoa(f.id) + paramSep + f.name + paramSep + f.hash
}

func (f *fileNode) deserialize(data []byte) error {
  f.children = make(map[int]fileNode)
  tokens := strings.Split(string(data), paramSep)
  if len(tokens) < 3 {
    return fmt.Errorf("Invalid file node input. unable to deserialize")
  }

  id, err := strconv.Atoi(tokens[0])
  if err != nil {
    return err
  }
  f.id, f.name, f.hash = id, tokens[1], tokens[2]
  return nil
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

func (d directoryTree) duplicate() directoryTree {
  var dtree directoryTree
  dtree.root = duplicate(d.root)
  return dtree
}

func duplicate(root fileNode) fileNode {
  newRoot := *newFileNode(root.id, root.name, root.hash)
  for id, node := range root.children {
    newRoot.children[id] = duplicate(node)
  }
  return newRoot
}

func tokenizeSerial(data []byte) []string {
  strData := string(data)
  tokens := strings.Split(strData, nodeSep)
  return tokens
}
