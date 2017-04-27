package yamlpatch

import "strings"

// Container is the interface for performing operations on Nodes
type Container interface {
	Get(key string) (*Node, error)
	Set(key string, val *Node) error
	Add(key string, val *Node) error
	Remove(key string) error
}

func findContainer(c Container, path string) (Container, string) {
	foundContainer := c

	split := strings.Split(path, "/")

	if len(split) < 2 {
		return nil, ""
	}

	parts := split[1 : len(split)-1]

	key := split[len(split)-1]

	for _, part := range parts {
		node, err := foundContainer.Get(decodePatchKey(part))
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
