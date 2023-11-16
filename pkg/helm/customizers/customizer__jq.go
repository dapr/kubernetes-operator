package customizers

import (
	"github.com/itchyny/gojq"
)

func JQ(expression string) func(map[string]any) (map[string]any, error) {
	return func(in map[string]any) (map[string]any, error) {
		query, err := gojq.Parse(expression)
		if err != nil {
			return nil, err
		}

		it := query.Run(in)

		v, ok := it.Next()
		if !ok {
			return in, nil
		}

		if err, ok := v.(error); ok {
			return in, err
		}

		if r, ok := v.(map[string]any); ok {
			return r, err
		}

		return in, nil
	}
}
