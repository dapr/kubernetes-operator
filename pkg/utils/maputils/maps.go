package maputils

import (
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
