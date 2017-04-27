package yamlpatch

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Patch is an ordered collection of operations.
type Patch []Operation

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
	var c Container = &NodeMap{}

	err := yaml.Unmarshal(doc, c)
	if err != nil {
		c = &NodeSlice{}
		err = yaml.Unmarshal(doc, c)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshaling doc: %s\n\n%s", string(doc), err)
		}
	}

	for _, op := range p {
		switch op.Op {
		case opAdd:
			err = tryAdd(c, op)
		case opRemove:
			err = tryRemove(c, op)
		case opReplace:
			err = tryReplace(c, op)
		case opMove:
			err = tryMove(c, op)
		case opCopy:
			err = tryCopy(c, op)
		default:
			err = fmt.Errorf("Unexpected op: %s", op.Op)
		}

		if err != nil {
			return nil, err
		}
	}

	return yaml.Marshal(c)
}

func tryAdd(doc Container, op Operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch add operation does not apply: doc is missing path: %s", op.Path)
	}

	return con.Add(key, op.Value())
}

func tryRemove(doc Container, op Operation) error {
	con, key := findContainer(doc, op.Path)

	if con == nil {
		return fmt.Errorf("yamlpatch remove operation does not apply: doc is missing path: %s", op.Path)
	}

	return con.Remove(key)
}

func tryReplace(doc Container, op Operation) error {
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

func tryMove(doc Container, op Operation) error {
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

func tryCopy(doc Container, op Operation) error {
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
