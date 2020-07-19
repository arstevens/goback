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

type directoryTree struct {
  root fileNode
}

/* Add and Delete child take the path down the tree
to the node to be inserted/deleted without the root node
included in the path */
func (d *directoryTree) addChild(idPath []int, id int, name string, hash string) error {
  parent := d.root
  for i := 0; i < len(idPath); i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %s", idPath[i])
    }
  }

  parent.children[id] = *newFileNode(id, name, hash)
  return nil
}

func (d *directoryTree) deleteChild(idPath []int) error {
  parent := d.root
  for i := 0; i < len(idPath) - 1; i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %s", idPath[i])
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

func newDirectoryTree(rootPath string) *directoryTree {
  dt := directoryTree{
    root: *newFileNode(filepath.Base(rootPath)),
  }

  err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
    path = strings.Replace(path, rootPath, "", 1)
    if path == "" {
      return nil
    }
    idPath := strings.Split(path, "/")
    fmt.Println(idPath)
    return dt.addChild(idPath)
  })
  if err != nil {
    panic(err)
  }
  return &dt
}

func newFileNode(id int, name string, hash string) *fileNode {
  return &fileNode{
    id: id,
    name: name,
    hash: hash,
    children: make(map[int]fileNode),
  }
}

func (f fileNode) serialize() string {
  return strconv.Itoa(f.id) + paramSep + f.name + paramSep + f.hash
}

func (f *fileNode) deserialize(data []byte) error {
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
