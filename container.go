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
