package yamlpatch

import (
	"errors"
	"fmt"
	"strconv"
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
	var iface interface{}
	err := yaml.Unmarshal(doc, &iface)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling doc: %s\n\n%s", string(doc), err)
	}

	var c Container
	node := NewNode(&iface)
	if node.IsNodeSlice() {
		c, err = node.NodeSlice()
	} else {
		c, err = node.NodeMap()
	}
	if err != nil {
		return nil, err
	}

	for _, op := range p {
		if op.Path.ContainsExtendedSyntax() {
			var paths []string
			paths, err = canonicalPaths(op.Path, iface)
			if err != nil {
				return nil, err
			}

			for _, path := range paths {
				newOp := op
				newOp.Path = OpPath(path)
				err = newOp.Perform(c)
				if err != nil {
					return nil, err
				}
			}
		} else {
			err = op.Perform(c)
			if err != nil {
				return nil, err
			}
		}
	}

	return yaml.Marshal(c)
}

func canonicalPaths(opPath OpPath, obj interface{}) ([]string, error) {
	var prefix string
	var paths []string

OUTER_LOOP:
	for _, part := range strings.Split(string(opPath), "/") {
		if part == "" {
			continue
		}

		if kv := strings.Split(part, "="); len(kv) == 2 {
			_, newPaths := findAllPaths(kv[0], kv[1], obj)
			for _, path := range newPaths {
				paths = append(paths, fmt.Sprintf("%s/%s", prefix, path))
			}
			continue
		}

		// this is an optimization to reduce recursive calls to findAllPaths
		switch ot := obj.(type) {
		case []interface{}:
			if idx, err := strconv.Atoi(part); err == nil && idx >= 0 && idx <= len(ot)-1 {
				prefix = fmt.Sprintf("%s/%d", prefix, idx)
				obj = ot[idx]
				continue
			}
			return nil, fmt.Errorf("invalid index given for array at %s", prefix)
		case map[interface{}]interface{}:
			for k, v := range ot {
				if ks, ok := k.(string); ok && ks == part {
					prefix = fmt.Sprintf("%s/%s", prefix, k)
					obj = v
					continue OUTER_LOOP
				}
			}
			return nil, errors.New("path does not match structure")
		}
	}

	return paths, nil
}

func findAllPaths(findKey, findValue string, data interface{}) (interface{}, []string) {
	var paths []string

	switch dt := data.(type) {
	case map[interface{}]interface{}:
		for k, v := range dt {
			switch vs := v.(type) {
			case string:
				if ks, ok := k.(string); ok && ks == findKey && vs == findValue {
					return dt, nil
				}
			default:
				_, subPaths := findAllPaths(findKey, findValue, v)
				for i := range subPaths {
					paths = append(paths, fmt.Sprintf("%s/%s", k, subPaths[i]))
				}
			}
		}
	case []interface{}:
		for i, v := range dt {
			if f, subPaths := findAllPaths(findKey, findValue, v); f != nil {
				if len(subPaths) > 0 {
					for _, subPath := range subPaths {
						paths = append(paths, fmt.Sprintf("%d/%s", i, subPath))
					}
					continue
				}
				paths = append(paths, fmt.Sprintf("%d", i))
			}
		}
	default:
		panic(fmt.Sprintf("don't know how to handle %T", data))
	}

	if paths != nil {
		return data, paths
	}

	return nil, nil
}
