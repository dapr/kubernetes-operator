package matchers

import (
	"encoding/json"
	"fmt"
	"strings"
)

func formattedMessage(comparisonMessage string, failurePath []interface{}) string {
	var diffMessage string
	if len(failurePath) == 0 {
		diffMessage = ""
	} else {
		diffMessage = fmt.Sprintf("\n\nfirst mismatched key: %s", formattedFailurePath(failurePath))
	}
	return fmt.Sprintf("%s%s", comparisonMessage, diffMessage)
}

func formattedFailurePath(failurePath []interface{}) string {
	formattedPaths := []string{}
	for i := len(failurePath) - 1; i >= 0; i-- {
		switch p := failurePath[i].(type) {
		case int:
			formattedPaths = append(formattedPaths, fmt.Sprintf(`[%d]`, p))
		default:
			if i != len(failurePath)-1 {
				formattedPaths = append(formattedPaths, ".")
			}
			formattedPaths = append(formattedPaths, fmt.Sprintf(`"%s"`, p))
		}
	}
	return strings.Join(formattedPaths, "")
}

func toString(a interface{}) (string, bool) {
	aString, isString := a.(string)
	if isString {
		return aString, true
	}

	aBytes, isBytes := a.([]byte)
	if isBytes {
		return string(aBytes), true
	}

	aStringer, isStringer := a.(fmt.Stringer)
	if isStringer {
		return aStringer.String(), true
	}

	aJSONRawMessage, isJSONRawMessage := a.(json.RawMessage)
	if isJSONRawMessage {
		return string(aJSONRawMessage), true
	}

	return "", false
}
