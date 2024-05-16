package maputils

import (
	"errors"
	"fmt"
	"maps"
)

func Merge(dst map[string]interface{}, source map[string]interface{}) map[string]interface{} {
	out := maps.Clone(dst)

	for k, v := range source {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = Merge(bv, v)
					continue
				}
			}
		}

		out[k] = v
	}

	return out
}

func Lookup(m map[string]interface{}, ks ...string) (interface{}, error) {
	if len(ks) == 0 { // degenerate input
		return nil, errors.New("lookup needs at least one key")
	}

	if rval, ok := m[ks[0]]; !ok {
		return nil, fmt.Errorf("key not found; remaining keys: %v", ks)
	} else if len(ks) == 1 { // we've reached the final key
		return rval, nil
	} else if m, ok := rval.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("malformed structure at %#v", rval)
	} else { // 1+ more keys
		return Lookup(m, ks[1:]...)
	}
}
