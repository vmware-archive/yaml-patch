package yamlpatch

import "strings"

type container interface {
	get(key string) (*Node, error)
	set(key string, val *Node) error
	add(key string, val *Node) error
	remove(key string) error
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

		if node.IsNodeSlice() {
			foundContainer, err = node.NodeSlice()
			if err != nil {
				return nil, ""
			}
		} else {
			foundContainer, err = node.NodeMap()
			if err != nil {
				return nil, ""
			}
		}
	}

	return foundContainer, decodePatchKey(key)
}
