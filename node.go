package yamlpatch

import (
	"fmt"
	"strconv"
)

// Node holds a YAML document that has not yet been processed into a NodeMap or
// NodeSlice
type Node struct {
	raw       *interface{}
	container Container
}

// NewNode returns a new Node. It expects a pointer to an interface{}
func NewNode(raw *interface{}) *Node {
	return &Node{
		raw: raw,
	}
}

// MarshalYAML implements yaml.Marshaler, and returns the correct interface{}
// to be marshaled
func (n *Node) MarshalYAML() (interface{}, error) {
	if n.container != nil {
		return n.container, nil
	}

	return *n.raw, nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (n *Node) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data interface{}

	err := unmarshal(&data)
	if err != nil {
		return err
	}

	n.raw = &data
	return nil
}

// RawValue returns the interface that the YAML was unmarshaled into
func (n *Node) RawValue() *interface{} {
	return n.raw
}

// Container returns the node as a Container
func (n *Node) Container() (Container, error) {
	if n.container != nil {
		return n.container, nil
	}

	switch rt := (*n.raw).(type) {
	case []interface{}:
		c := make(nodeSlice, len(rt))
		n.container = &c

		for i := range rt {
			c[i] = NewNode(&rt[i])
		}
	case map[interface{}]interface{}:
		c := make(nodeMap, len(rt))
		n.container = &c

		for k := range rt {
			v := rt[k]
			c[k] = NewNode(&v)
		}
	}

	return n.container, nil
}

// Equal compares the values of the raw interfaces that the YAML was
// unmarshaled into
func (n *Node) Equal(other *Node) bool {
	return *n.raw == *other.raw
}

type nodeMap map[interface{}]*Node

// Set or replace the Node at key with the provided Node
func (n *nodeMap) Set(key string, val *Node) error {
	(*n)[key] = val
	return nil
}

// Add the provided Node at the given key
func (n *nodeMap) Add(key string, val *Node) error {
	(*n)[key] = val
	return nil
}

// Get the node at the given key
func (n *nodeMap) Get(key string) (*Node, error) {
	return (*n)[key], nil
}

// Remove the node at the given key
func (n *nodeMap) Remove(key string) error {
	_, ok := (*n)[key]
	if !ok {
		return fmt.Errorf("Unable to remove nonexistent key: %s", key)
	}

	delete(*n, key)
	return nil
}

type nodeSlice []*Node

// Set the Node at the given index with the provided Node
func (n *nodeSlice) Set(index string, val *Node) error {
	i, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	sz := len(*n)
	if i+1 > sz {
		sz = i + 1
	}

	ary := make([]*Node, sz)

	cur := *n

	copy(ary, cur)

	if i >= len(ary) {
		return fmt.Errorf("Unable to access invalid index: %d", i)
	}

	ary[i] = val

	*n = ary
	return nil
}

// Add the provided Node at the given index
func (n *nodeSlice) Add(index string, val *Node) error {
	if index == "-" {
		*n = append(*n, val)
		return nil
	}

	i, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	ary := make([]*Node, len(*n)+1)

	cur := *n

	copy(ary[0:i], cur[0:i])
	ary[i] = val
	copy(ary[i+1:], cur[i:])

	*n = ary
	return nil
}

// Get the node at the given index
func (n *nodeSlice) Get(index string) (*Node, error) {
	i, err := strconv.Atoi(index)
	if err != nil {
		return nil, err
	}

	if i >= len(*n) {
		return nil, fmt.Errorf("Unable to access invalid index: %d", i)
	}

	return (*n)[i], nil
}

// Remove the node at the given index
func (n *nodeSlice) Remove(index string) error {
	i, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	cur := *n

	if i >= len(cur) {
		return fmt.Errorf("Unable to remove invalid index: %d", i)
	}

	ary := make([]*Node, len(cur)-1)

	copy(ary[0:i], cur[0:i])
	copy(ary[i:], cur[i+1:])

	*n = ary
	return nil

}
