package yamlpatch

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
func (o Operation) Value() *Node {
	return NewNode(o.RawValue)
}
