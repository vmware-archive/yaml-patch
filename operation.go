package yamlpatch

import "fmt"

// Op is a type alias
type Op string

// Ops
const (
	opAdd     Op = "add"
	opRemove  Op = "remove"
	opReplace Op = "replace"
	opMove    Op = "move"
	opCopy    Op = "copy"
)

// Operation is an RFC6902 'Operation'
// https://tools.ietf.org/html/rfc6902#section-4
type Operation struct {
	Op       Op           `yaml:"op,omitempty"`
	Path     string       `yaml:"path,omitempty"`
	From     string       `yaml:"from,omitempty"`
	RawValue *interface{} `yaml:"value,omitempty"`
}

// Value returns the operation's value as a *Node
func (o *Operation) Value() *Node {
	return NewNode(o.RawValue)
}

// Perform executes the operation on the given container
func (o *Operation) Perform(c Container) error {
	var err error

	switch o.Op {
	case opAdd:
		err = tryAdd(c, o)
	case opRemove:
		err = tryRemove(c, o)
	case opReplace:
		err = tryReplace(c, o)
	case opMove:
		err = tryMove(c, o)
	case opCopy:
		err = tryCopy(c, o)
	default:
		err = fmt.Errorf("Unexpected op: %s", o.Op)
	}

	return err
}

func tryAdd(doc Container, op *Operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch add operation does not apply: doc is missing path: %s", op.Path)
	}

	return con.Add(key, op.Value())
}

func tryRemove(doc Container, op *Operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch remove operation does not apply: doc is missing path: %s", op.Path)
	}

	return con.Remove(key)
}

func tryReplace(doc Container, op *Operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch replace operation does not apply: doc is missing path: %s", op.Path)
	}

	val, err := con.Get(key)
	if val == nil || err != nil {
		return fmt.Errorf("yamlpatch replace operation does not apply: doc is missing key: %s", op.Path)
	}

	return con.Set(key, op.Value())
}

func tryMove(doc Container, op *Operation) error {
	con, key := findContainer(doc, op.From)
	if con == nil {
		return fmt.Errorf("yamlpatch move operation does not apply: doc is missing from path: %s", op.From)
	}

	val, err := con.Get(key)
	if err != nil {
		return err
	}

	err = con.Remove(key)
	if err != nil {
		return err
	}

	con, key = findContainer(doc, op.Path)
	if con == nil {
		return fmt.Errorf("yamlpatch move operation does not apply: doc is missing destination path: %s", op.Path)
	}

	return con.Set(key, val)
}

func tryCopy(doc Container, op *Operation) error {
	con, key := findContainer(doc, op.From)
	if con == nil {
		return fmt.Errorf("copy operation does not apply: doc is missing from path: %s", op.From)
	}

	val, err := con.Get(key)
	if err != nil {
		return err
	}

	con, key = findContainer(doc, op.Path)
	if con == nil {
		return fmt.Errorf("copy operation does not apply: doc is missing destination path: %s", op.Path)
	}

	return con.Set(key, val)
}
