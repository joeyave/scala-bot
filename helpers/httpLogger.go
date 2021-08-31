package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type transportWithLogger struct {
	Transport http.RoundTripper
}

func NewTransportWithLogger(transport http.RoundTripper) *transportWithLogger {
	return &transportWithLogger{Transport: transport}
}

func (t *transportWithLogger) RoundTrip(req *http.Request) (*http.Response, error) {

	ctx := context.WithValue(req.Context(), "reqStart", time.Now())
	req = req.WithContext(ctx)

	var reqBodyBytes []byte
	if req.Body != nil {
		reqBodyBytes, _ = ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes))
	}

	event := log.Info().
		Str("method", req.Method).
		Str("url", req.URL.String())

	if len(reqBodyBytes) > 0 {
		if json.Valid(reqBodyBytes) {
			event = event.RawJSON("body", reqBodyBytes)
		} else {
			event = event.Bytes("body", reqBodyBytes)
		}
	}

	if !strings.Contains(req.URL.Path, "/getUpdates") {
		event.Msg("API request:")
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	var respBodyBytes []byte
	if resp.Body != nil {
		respBodyBytes, _ = ioutil.ReadAll(resp.Body)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBodyBytes))
	}

	switch {
	case resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError:
		event = log.Warn()
	case resp.StatusCode >= http.StatusInternalServerError:
		event = log.Error()
	default:
		event = log.Info()
	}

	event = event.Str("method", resp.Request.Method).
		Str("url", resp.Request.URL.String()).
		Int("status", resp.StatusCode).
		Dur("latency", time.Now().Sub(ctx.Value("reqStart").(time.Time)))

	if len(respBodyBytes) > 0 {
		if json.Valid(respBodyBytes) {
			event = event.RawJSON("body", respBodyBytes)
		} else {
			event = event.Bytes("body", respBodyBytes)
		}
	}

	if !strings.Contains(req.URL.Path, "/getUpdates") {
		event.Msg("API response:")
	}

	return resp, err
}
