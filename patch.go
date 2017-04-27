package yamlpatch

import (
	"fmt"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const (
	eRaw = iota
	eDoc
	eAry
)

type lazyNode struct {
	raw   *interface{}
	doc   partialDoc
	ary   partialArray
	which int
}

func newLazyNode(raw *interface{}) *lazyNode {
	return &lazyNode{raw: raw, which: eRaw}
}

func (n *lazyNode) MarshalYAML() (interface{}, error) {
	switch n.which {
	case eRaw:
		if n.raw == nil {
			panic(fmt.Sprintf("type is raw but raw is nil: %p", n))
		}
		return *n.raw, nil
	case eDoc:
		return n.doc, nil
	case eAry:
		return n.ary, nil
	default:
		return nil, fmt.Errorf("Unknown type")
	}
}

func (n *lazyNode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data interface{}

	err := unmarshal(&data)
	if err != nil {
		return err
	}

	n.raw = &data
	n.which = eRaw
	return nil
}

func (n *lazyNode) intoDoc() (*partialDoc, error) {
	if n.which == eDoc {
		return &n.doc, nil
	}

	raw := *n.raw

	switch rt := raw.(type) {
	case map[interface{}]interface{}:
		doc := map[interface{}]*lazyNode{}

		for k := range rt {
			v := rt[k]
			doc[k] = newLazyNode(&v)
		}

		n.doc = doc
		n.which = eDoc
		return &n.doc, nil
	default:
		return nil, fmt.Errorf("don't know how to convert %T into doc", raw)
	}
}

func (n *lazyNode) intoAry() (*partialArray, error) {
	if n.which == eAry {
		return &n.ary, nil
	}

	raw := *n.raw

	switch rt := raw.(type) {
	case []interface{}:
		array := make(partialArray, len(rt))

		for i := range rt {
			array[i] = newLazyNode(&rt[i])
		}

		n.ary = array
		n.which = eAry
		return &n.ary, nil
	default:
		return nil, fmt.Errorf("don't know how to convert %T into ary", raw)
	}
}

type operation struct {
	Op       string       `yaml:"op,omitempty"`
	Path     string       `yaml:"path,omitempty"`
	From     string       `yaml:"from,omitempty"`
	RawValue *interface{} `yaml:"value,omitempty"`
}

func (o operation) value() *lazyNode {
	if o.RawValue == nil {
		panic(fmt.Sprintf("value is nil: %#v", o))
	}
	return newLazyNode(o.RawValue)
}

// Patch is an ordered collection of operations.
type Patch []operation

type container interface {
	get(key string) (*lazyNode, error)
	set(key string, val *lazyNode) error
	add(key string, val *lazyNode) error
	remove(key string) error
}

func isArray(iface interface{}) bool {
	_, ok := iface.([]interface{})
	return ok
}

func findContainer(c container, path string) (container, string) {
	foundContainer := c

	split := strings.Split(path, "/")

	if len(split) < 2 {
		return nil, ""
	}

	parts := split[1 : len(split)-1]

	key := split[len(split)-1]

	for _, part := range parts {
		node, err := foundContainer.get(decodePatchKey(part))
		if node == nil || err != nil {
			return nil, ""
		}

		if isArray(*node.raw) {
			foundContainer, err = node.intoAry()
			if err != nil {
				return nil, ""
			}
		} else {
			foundContainer, err = node.intoDoc()
			if err != nil {
				return nil, ""
			}
		}
	}

	return foundContainer, decodePatchKey(key)
}

type partialDoc map[interface{}]*lazyNode

func (d *partialDoc) set(key string, val *lazyNode) error {
	(*d)[key] = val
	return nil
}

func (d *partialDoc) add(key string, val *lazyNode) error {
	(*d)[key] = val
	return nil
}

func (d *partialDoc) get(key string) (*lazyNode, error) {
	return (*d)[key], nil
}

func (d *partialDoc) remove(key string) error {
	_, ok := (*d)[key]
	if !ok {
		return fmt.Errorf("Unable to remove nonexistent key: %s", key)
	}

	delete(*d, key)
	return nil
}

type partialArray []*lazyNode

func (d *partialArray) set(key string, val *lazyNode) error {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	sz := len(*d)
	if idx+1 > sz {
		sz = idx + 1
	}

	ary := make([]*lazyNode, sz)

	cur := *d

	copy(ary, cur)

	if idx >= len(ary) {
		return fmt.Errorf("Unable to access invalid index: %d", idx)
	}

	ary[idx] = val

	*d = ary
	return nil
}

func (d *partialArray) add(key string, val *lazyNode) error {
	if key == "-" {
		*d = append(*d, val)
		return nil
	}

	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	ary := make([]*lazyNode, len(*d)+1)

	cur := *d

	copy(ary[0:idx], cur[0:idx])
	ary[idx] = val
	copy(ary[idx+1:], cur[idx:])

	*d = ary
	return nil
}

func (d *partialArray) get(key string) (*lazyNode, error) {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return nil, err
	}

	if idx >= len(*d) {
		return nil, fmt.Errorf("Unable to access invalid index: %d", idx)
	}

	return (*d)[idx], nil
}

func (d *partialArray) remove(key string) error {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	cur := *d

	if idx >= len(cur) {
		return fmt.Errorf("Unable to remove invalid index: %d", idx)
	}

	ary := make([]*lazyNode, len(cur)-1)

	copy(ary[0:idx], cur[0:idx])
	copy(ary[idx:], cur[idx+1:])

	*d = ary
	return nil

}

func add(doc container, op operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch add operation does not apply: doc is missing path: %s", op.Path)
	}

	return con.add(key, op.value())
}

func remove(doc container, op operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch remove operation does not apply: doc is missing path: %s", op.Path)
	}

	return con.remove(key)
}

func replace(doc container, op operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch replace operation does not apply: doc is missing path: %s", op.Path)
	}

	val, err := con.get(key)
	if val == nil || err != nil {
		return fmt.Errorf("yamlpatch replace operation does not apply: doc is missing key: %s", op.Path)
	}

	return con.set(key, op.value())
}

func move(doc container, op operation) error {
	con, key := findContainer(doc, op.From)
	if con == nil {
		return fmt.Errorf("yamlpatch move operation does not apply: doc is missing from path: %s", op.From)
	}

	val, err := con.get(key)
	if err != nil {
		return err
	}

	err = con.remove(key)
	if err != nil {
		return err
	}

	con, key = findContainer(doc, op.Path)
	if con == nil {
		return fmt.Errorf("yamlpatch move operation does not apply: doc is missing destination path: %s", op.Path)
	}

	return con.set(key, val)
}

func copyOp(doc container, op operation) error {
	con, key := findContainer(doc, op.From)
	if con == nil {
		return fmt.Errorf("copy operation does not apply: doc is missing from path: %s", op.From)
	}

	val, err := con.get(key)
	if err != nil {
		return err
	}

	con, key = findContainer(doc, op.Path)
	if con == nil {
		return fmt.Errorf("copy operation does not apply: doc is missing destination path: %s", op.Path)
	}

	return con.set(key, val)
}

// DecodePatch decodes the passed YAML document as if it were an RFC 6902 patch
func DecodePatch(bs []byte) (Patch, error) {
	var p Patch

	err := yaml.Unmarshal(bs, &p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Apply returns a YAML document that has been mutated per the patch
func (p Patch) Apply(doc []byte) ([]byte, error) {
	var c container = &partialDoc{}

	err := yaml.Unmarshal(doc, c)
	if err != nil {
		c = &partialArray{}
		err = yaml.Unmarshal(doc, c)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshaling doc: %s\n\n%s", string(doc), err)
		}
	}

	for _, op := range p {
		switch op.Op {
		case "add":
			err = add(c, op)
		case "remove":
			err = remove(c, op)
		case "replace":
			err = replace(c, op)
		case "move":
			err = move(c, op)
		case "copy":
			err = copyOp(c, op)
		default:
			err = fmt.Errorf("Unexpected op: %s", op.Op)
		}

		if err != nil {
			return nil, err
		}
	}

	return yaml.Marshal(c)
}

// From http://tools.ietf.org/html/rfc6901#section-4 :
//
// Evaluation of each reference token begins by decoding any escaped
// character sequence.  This is performed by first transforming any
// occurrence of the sequence '~1' to '/', and then transforming any
// occurrence of the sequence '~0' to '~'.

var (
	rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")
)

func decodePatchKey(k string) string {
	return rfc6901Decoder.Replace(k)
}
