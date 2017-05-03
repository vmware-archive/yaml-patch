package yamlpatch

import (
	"fmt"
	"strconv"
	"strings"
)

// PathFinder can be used to find RFC6902-standard paths given non-standard
// (key=value) pointer syntax
type PathFinder struct {
	root interface{}
}

// NewPathFinder takes an interface that represents a YAML document and returns
// a new PathFinder
func NewPathFinder(iface interface{}) *PathFinder {
	return &PathFinder{
		root: iface,
	}
}

// Find expands the given path into all matching paths, returning the canonical
// versions of those matching paths
func (p *PathFinder) Find(path string) []string {
	parts := strings.Split(path, "/")

	if parts[1] == "" {
		return []string{"/"}
	}

	routes := map[string]interface{}{
		"": p.root,
	}

	for _, part := range parts[1:] {
		routes = find(part, routes)
	}

	var paths []string
	for k := range routes {
		paths = append(paths, k)
	}

	return paths
}

func find(part string, routes map[string]interface{}) map[string]interface{} {
	matches := map[string]interface{}{}

	for prefix, iface := range routes {
		if strings.Contains(part, "=") {
			kv := strings.Split(part, "=")
			if newMatches := findAll(prefix, kv[0], kv[1], iface); len(newMatches) > 0 {
				matches = newMatches
			}
			continue
		}

		switch it := iface.(type) {
		case map[interface{}]interface{}:
			for k, v := range it {
				if ks, ok := k.(string); ok && ks == part {
					path := fmt.Sprintf("%s/%s", prefix, ks)
					matches[path] = v
				}
			}
		case []interface{}:
			if idx, err := strconv.Atoi(part); err == nil && idx >= 0 && idx <= len(it)-1 {
				path := fmt.Sprintf("%s/%d", prefix, idx)
				matches[path] = it[idx]
			}
		default:
			panic(fmt.Sprintf("don't know how to handle %T: %s", iface, iface))
		}
	}

	return matches
}

func findAll(prefix, findKey, findValue string, iface interface{}) map[string]interface{} {
	matches := map[string]interface{}{}

	switch it := iface.(type) {
	case map[interface{}]interface{}:
		for k, v := range it {
			if ks, ok := k.(string); ok {
				switch vs := v.(type) {
				case string:
					if ks == findKey && vs == findValue {
						return map[string]interface{}{
							prefix: it,
						}
					}
				default:
					for route, match := range findAll(fmt.Sprintf("%s/%s", prefix, ks), findKey, findValue, v) {
						matches[route] = match
					}
				}
			}
		}
	case []interface{}:
		for i, v := range it {
			for route, match := range findAll(fmt.Sprintf("%s/%d", prefix, i), findKey, findValue, v) {
				matches[route] = match
			}
		}
	default:
		panic(fmt.Sprintf("don't know how to handle %T: %s", iface, iface))
	}

	return matches
}
