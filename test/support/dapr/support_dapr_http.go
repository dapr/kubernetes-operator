package dapr

import (
	"bytes"
	"net/http"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	"github.com/onsi/gomega"
)

func GET(t support.Test, url string) func(g gomega.Gomega) (*http.Response, error) {
	return func(g gomega.Gomega) (*http.Response, error) {
		req, err := http.NewRequestWithContext(t.Ctx(), http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, err
		}

		return t.HTTPClient().Do(req)
	}
}

func POST(t support.Test, url string, contentType string, content []byte) func(g gomega.Gomega) (*http.Response, error) {
	return func(g gomega.Gomega) (*http.Response, error) {
		data := content
		if data == nil {
			data = []byte{}
		}

		req, err := http.NewRequestWithContext(t.Ctx(), http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		if contentType != "" {
			req.Header.Add("Content-Type", contentType)
		}

		return t.HTTPClient().Do(req)
	}
}
