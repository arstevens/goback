package reflector

import (
  "encoding/base64"
  "strconv"
  "strings"
)

const (
  jumpBack = '|'
  nodeSep = ")"
  paramSep = ","
)

type directoryTree struct {
  root fileNode
}

type fileNode struct {
  hash []byte
  children map[int]fileNode
}

func newFileNode(hash []byte) *fileNode {
  return &fileNode{
    hash: hash,
    children: make(map[int]fileNode),
  }
}

func (f fileNode) serialize(id int) string {
  encodedHash := base64.StdEncoding.EncodeToString(f.hash)
  encodedId := strconv.Itoa(id)
  return encodedHash + paramSep + encodedId
}

func (d directoryTree) serialize() []byte {
  treeSerial := serializeTree(d.root, 0)
  return []byte(treeSerial)
}

func serializeTree(root fileNode, id int) string {
  serial := root.serialize(id) + ")"
  for childId, node := range root.children {
    serial += serializeTree(node, childId)
  }
  serial += string(jumpBack)
  return serial
}

func (d directoryTree) deserialize(data []byte) {
  tokens := tokenizeSerial(data)
  stack := make([]fileNode, 0)
  d.root = *newFileNode([]byte{})
  parent := &d.root

  for _, tk := range tokens {
    if tk[0] == byte(jumpBack) {
      for jumpLen := len(tk); jumpLen > 0 && len(stack) > 0; jumpLen-- {
        parent = &stack[len(stack)]
        stack = stack[:len(stack)-1]
      }
    } else {
      nodeParams := strings.Split(tk, paramSep)
      hash, _ := base64.StdEncoding.DecodeString(nodeParams[0])
      id, _ := strconv.Atoi(nodeParams[1])

      child := newFileNode(hash)
      (*parent).children[id] = *child
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
