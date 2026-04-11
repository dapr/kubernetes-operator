package maputils

import (
	"maps"
)

func Merge(dst map[string]any, source map[string]any) map[string]any {
	out := maps.Clone(dst)

	for k, v := range source {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = Merge(bv, v)
					continue
				}
			}
		}

		out[k] = v
	}

	return out
}
