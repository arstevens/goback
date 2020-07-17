package reflector

import (
  "path/filepath"
  "strings"
  "fmt"
  "os"
)

const (
  jumpBack = '|'
  nodeSep = ")"
)

type directoryTree struct {
  root fileNode
}

func (d *directoryTree) addChild(idPath []string) error {
  parent := d.root
  for i := 0; i < len(idPath) - 1; i++ {
    var ok bool
    parent, ok = parent.children[idPath[i]]
    if !ok {
      return fmt.Errorf("No child with id %s", idPath[i])
    }
  }

  childId := idPath[len(idPath) - 1]
  parent.children[childId] = *newFileNode(childId)
  return nil
}

type fileNode struct {
  id string
  children map[string]fileNode
}

func newDirectoryTree(rootPath string) *directoryTree {
  dt := directoryTree{
    root: *newFileNode(filepath.Base(rootPath)),
  }

  err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
    idPath := strings.Split(path, "/")
    return dt.addChild(idPath)
  })
  if err != nil {
    panic(err)
  }
  return &dt
}

func newFileNode(id string) *fileNode {
  return &fileNode{
    id: id,
    children: make(map[string]fileNode),
  }
}

func (f fileNode) serialize() string {
  return f.id
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
  serial += string(jumpBack)
  return serial
}

func (d directoryTree) deserialize(data []byte) {
  tokens := tokenizeSerial(data)
  stack := make([]fileNode, 0)
  d.root = *newFileNode(tokens[0])
  parent := &d.root

  for i := 1; i < len(tokens); i++ {
    tk := tokens[i]
    if tk[0] == byte(jumpBack) {
      for jumpLen := len(tk); jumpLen > 0 && len(stack) > 0; jumpLen-- {
        parent = &stack[len(stack)]
        stack = stack[:len(stack)-1]
      }
    } else {
      child := newFileNode(tk)
      (*parent).children[tk] = *child
      stack = append(stack, *parent)
      parent = child
    }
  }
}

func tokenizeSerial(data []byte) []string {
  strData := string(data)
  tokens := strings.Split(strData, nodeSep)
  return tokens
}
