package prom

import (
	"io"
	"net/http"
	"net/url"

	"github.com/mmihic/golib/src/pkg/httpclient"
)

// WithHeader returns an httpclient.CallOption that sets an outbound
// header to the given value.
func WithHeader(k, v string) httpclient.CallOption {
	return func(r *http.Request) error {
		r.Header.Add(k, v)
		return nil
	}
}

// FormURLEncoded returns an httpclient.Marshaller that encodes a set
// of key, vlaue pairs using the form-urlencoded scheme.
func FormURLEncoded(vals url.Values) httpclient.Marshaller {
	return formEncoded{
		vals: vals,
	}
}

type formEncoded struct {
	vals url.Values
}

func (m formEncoded) ContentType() string {
	return "application/x-www-form-urlencoded"
}

func (m formEncoded) Marshal(w io.Writer) error {
	_, err := w.Write([]byte(m.vals.Encode()))
	return err
}
