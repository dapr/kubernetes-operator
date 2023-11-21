package matchers

import (
	"encoding/json"
)

func AsJSON() func(in any) (any, error) {
	return func(in any) (any, error) {
		data, err := json.Marshal(in)
		if err != nil {
			return nil, err
		}

		return data, nil
	}
}
