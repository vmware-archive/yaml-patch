package yamlpatch

import (
	"fmt"
	"strconv"
)

// NodeType is a type alias
type NodeType int

// NodeTypes
const (
	NodeTypeRaw NodeType = iota
	NodeTypeMap
	NodeTypeSlice
)

// Node holds a YAML document that has not yet been processed into a NodeMap or
// NodeSlice
type Node struct {
	raw       *interface{}
	nodeMap   NodeMap
	nodeSlice NodeSlice
	nodeType  NodeType
}

// NewNode returns a new Node. It expects a pointer to an interface{}
func NewNode(raw *interface{}) *Node {
	return &Node{raw: raw, nodeType: NodeTypeRaw}
}

// MarshalYAML implements yaml.Marshaler, and returns the correct interface{}
// to be marshaled
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

// UnmarshalYAML implements yaml.Unmarshaler
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

// IsNodeSlice returns whether the contents it holds is a slice or not
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

// NodeMap returns the node as a NodeMap, if the raw interface{} it holds is
// indeed a map[interface{}]interface{}
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

// NodeSlice returns the node as a NodeSlice, if the raw interface{} it holds
// is indeed a []interface{}
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

// NodeMap represents a YAML object
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

// NodeSlice represents a YAML array
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
