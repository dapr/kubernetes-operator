package matchers

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func MatchJQ(expression string) types.GomegaMatcher {
	return &jqMatcher{
		Expression: expression,
	}
}

func MatchJQf(format string, args ...any) types.GomegaMatcher {
	return &jqMatcher{
		Expression: fmt.Sprintf(format, args...),
	}
}

var _ types.GomegaMatcher = &jqMatcher{}

type jqMatcher struct {
	Expression       string
	firstFailurePath []interface{}
}

func (matcher *jqMatcher) Match(actual interface{}) (bool, error) {
	query, err := gojq.Parse(matcher.Expression)
	if err != nil {
		return false, err
	}

	actualString, ok := toString(actual)
	if !ok {
		return false, fmt.Errorf("MatchJQMatcher matcher requires a string, stringer, or []byte.  Got actual:\n%s", format.Object(actual, 1))
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(actualString), &data); err != nil {
		return false, err
	}

	it := query.Run(data)

	v, ok := it.Next()
	if !ok {
		return false, nil
	}

	if err, ok := v.(error); ok {
		return false, err
	}

	if match, ok := v.(bool); ok {
		return match, nil
	}

	return false, nil
}

func (matcher *jqMatcher) FailureMessage(actual interface{}) string {
	actualString, expectedString, _ := matcher.prettyPrint(actual)
	return formattedMessage(format.Message(actualString, "to match JQ Expression", expectedString), matcher.firstFailurePath)
}

func (matcher *jqMatcher) NegatedFailureMessage(actual interface{}) string {
	actualString, expectedString, _ := matcher.prettyPrint(actual)
	return formattedMessage(format.Message(actualString, "not to match JQ expression", expectedString), matcher.firstFailurePath)
}

func (matcher *jqMatcher) prettyPrint(actual interface{}) (string, string, error) {
	actualString, ok := toString(actual)
	if !ok {
		return "", "", fmt.Errorf("the MatchJQMatcher matcher requires a string, stringer, or []byte.  Got actual:\n%s", format.Object(actual, 1))
	}

	abuf := new(bytes.Buffer)

	if err := json.Indent(abuf, []byte(actualString), "", "  "); err != nil {
		return "", "", fmt.Errorf("actual '%s' should match '%s', but it does not.\nUnderlying error: %w", actualString, matcher.Expression, err)
	}

	return abuf.String(), matcher.Expression, nil
}
