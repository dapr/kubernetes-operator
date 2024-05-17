package customizers

import (
	"fmt"

	"github.com/itchyny/gojq"
)

func JQ(expression string) func(map[string]any) (map[string]any, error) {
	return func(in map[string]any) (map[string]any, error) {
		query, err := gojq.Parse(expression)
		if err != nil {
			return nil, fmt.Errorf("unable to parse expression %s: %w", expression, err)
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
			return r, nil
		}

		return in, nil
	}
}
