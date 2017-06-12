package yamlpatch

import (
	"fmt"
	"strings"
)

// Container is the interface for performing operations on Nodes
type Container interface {
	Get(key string) (*Node, error)
	Set(key string, val *Node) error
	Add(key string, val *Node) error
	Remove(key string) error
}

func findContainer(c Container, path *OpPath) (Container, string, error) {
	parts, key, err := path.Decompose()
	if err != nil {
		return nil, "", err
	}

	foundContainer := c

	for _, part := range parts {
		node, err := foundContainer.Get(decodePatchKey(part))
		if err != nil {
			return nil, "", err
		}

		if node == nil {
			return nil, "", fmt.Errorf("key does not exist: %s", key)
		}

		if node.IsNodeSlice() {
			foundContainer, err = node.NodeSlice()
			if err != nil {
				return nil, "", err
			}
		} else {
			foundContainer, err = node.NodeMap()
			if err != nil {
				return nil, "", err
			}
		}
	}

	return foundContainer, decodePatchKey(key), nil
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
