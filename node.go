package yamlpatch

import (
	"fmt"
	"strconv"
)

type NodeType int

const (
	NodeTypeRaw NodeType = iota
	NodeTypeMap
	NodeTypeSlice
)

type Node struct {
	raw       *interface{}
	nodeMap   NodeMap
	nodeSlice NodeSlice
	nodeType  NodeType
}

func NewNode(raw *interface{}) *Node {
	return &Node{raw: raw, nodeType: NodeTypeRaw}
}

func (n *Node) MarshalYAML() (interface{}, error) {
	switch n.nodeType {
	case NodeTypeRaw:
		return *n.raw, nil
	case NodeTypeMap:
		return n.nodeMap, nil
	case NodeTypeSlice:
		return n.nodeSlice, nil
	default:
		return nil, fmt.Errorf("Unknown type")
	}
}

func (n *Node) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data interface{}

	err := unmarshal(&data)
	if err != nil {
		return err
	}

	n.raw = &data
	n.nodeType = NodeTypeRaw
	return nil
}

func (n *Node) IsNodeSlice() bool {
	if n.nodeType == NodeTypeRaw {
		switch (*n.raw).(type) {
		case []interface{}:
			return true
		default:
			return false
		}
	}

	return n.nodeType == NodeTypeSlice
}

func (n *Node) NodeMap() (*NodeMap, error) {
	if n.nodeMap != nil {
		return &n.nodeMap, nil
	}

	raw := *n.raw

	switch rt := raw.(type) {
	case map[interface{}]interface{}:
		doc := map[interface{}]*Node{}

		for k := range rt {
			v := rt[k]
			doc[k] = NewNode(&v)
		}

		n.nodeMap = doc
		n.nodeType = NodeTypeMap
		return &n.nodeMap, nil
	default:
		return nil, fmt.Errorf("don't know how to convert %T into doc", raw)
	}
}

func (n *Node) NodeSlice() (*NodeSlice, error) {
	if n.nodeSlice != nil {
		return &n.nodeSlice, nil
	}

	raw := *n.raw

	switch rt := raw.(type) {
	case []interface{}:
		array := make(NodeSlice, len(rt))

		for i := range rt {
			array[i] = NewNode(&rt[i])
		}

		n.nodeSlice = array
		n.nodeType = NodeTypeSlice
		return &n.nodeSlice, nil
	default:
		return nil, fmt.Errorf("don't know how to convert %T into ary", raw)
	}
}

type NodeMap map[interface{}]*Node

func (n *NodeMap) set(key string, val *Node) error {
	(*n)[key] = val
	return nil
}

func (n *NodeMap) add(key string, val *Node) error {
	(*n)[key] = val
	return nil
}

func (n *NodeMap) get(key string) (*Node, error) {
	return (*n)[key], nil
}

func (n *NodeMap) remove(key string) error {
	_, ok := (*n)[key]
	if !ok {
		return fmt.Errorf("Unable to remove nonexistent key: %s", key)
	}

	delete(*n, key)
	return nil
}

type NodeSlice []*Node

func (n *NodeSlice) set(key string, val *Node) error {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	sz := len(*n)
	if idx+1 > sz {
		sz = idx + 1
	}

	ary := make([]*Node, sz)

	cur := *n

	copy(ary, cur)

	if idx >= len(ary) {
		return fmt.Errorf("Unable to access invalid index: %d", idx)
	}

	ary[idx] = val

	*n = ary
	return nil
}

func (n *NodeSlice) add(key string, val *Node) error {
	if key == "-" {
		*n = append(*n, val)
		return nil
	}

	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	ary := make([]*Node, len(*n)+1)

	cur := *n

	copy(ary[0:idx], cur[0:idx])
	ary[idx] = val
	copy(ary[idx+1:], cur[idx:])

	*n = ary
	return nil
}

func (n *NodeSlice) get(key string) (*Node, error) {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return nil, err
	}

	if idx >= len(*n) {
		return nil, fmt.Errorf("Unable to access invalid index: %d", idx)
	}

	return (*n)[idx], nil
}

func (n *NodeSlice) remove(key string) error {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	cur := *n

	if idx >= len(cur) {
		return fmt.Errorf("Unable to remove invalid index: %d", idx)
	}

	ary := make([]*Node, len(cur)-1)

	copy(ary[0:idx], cur[0:idx])
	copy(ary[idx:], cur[idx+1:])

	*n = ary
	return nil

}
