package matechers

import (
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/onsi/gomega/format"
)

func ExtractJQ(expression string) func(in any) (any, error) {
	return func(in any) (any, error) {
		query, err := gojq.Parse(expression)
		if err != nil {
			return nil, err
		}

		actualString, ok := toString(in)
		if !ok {
			return false, fmt.Errorf("ExtractJQ requires a string, stringer, or []byte.  Got actual:\n%s", format.Object(in, 1))
		}

		if len(actualString) == 0 {
			return nil, nil
		}

		b := []byte(actualString)

		var it gojq.Iter

		// rough check for object vs array
		switch b[0] {
		case '{':
			data := make(map[string]any)
			if err := json.Unmarshal(b, &data); err != nil {
				return false, err
			}

			it = query.Run(data)
		case '[':
			var data []any
			if err := json.Unmarshal(b, &data); err != nil {
				return false, err
			}

			it = query.Run(data)
		default:
			return false, fmt.Errorf("ExtractJQ requires a Json Array or Object. Got actual:\n%s", format.Object(in, 1))
		}

		v, ok := it.Next()
		if !ok {
			return false, nil
		}

		if err, ok := v.(error); ok {
			return false, err
		}

		return v, nil
	}
}
