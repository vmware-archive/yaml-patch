package yamlpatch

import (
	"fmt"

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
		err = op.Perform(c)
		if err != nil {
			return nil, err
		}
	}

	return yaml.Marshal(c)
}
