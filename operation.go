package yamlpatch

type Op string

const (
	opAdd     Op = "add"
	opRemove  Op = "remove"
	opReplace Op = "replace"
	opMove    Op = "move"
	opCopy    Op = "copy"
)

type Operation struct {
	Op       Op           `yaml:"op,omitempty"`
	Path     string       `yaml:"path,omitempty"`
	From     string       `yaml:"from,omitempty"`
	RawValue *interface{} `yaml:"value,omitempty"`
}

func (o Operation) Value() *Node {
	return NewNode(o.RawValue)
}
