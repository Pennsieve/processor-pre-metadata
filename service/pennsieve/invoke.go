package pennsieve

import (
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/service/logging"
	"io"
	"log/slog"
	"net/http"
)

var logger = logging.PackageLogger("pennsieve")

type Session struct {
	Token    string
	APIHost  string
	API2Host string
}

func NewSession(sessionToken, apiHost, api2Host string) *Session {
	return &Session{
		Token:    sessionToken,
		APIHost:  apiHost,
		API2Host: api2Host}
}

func (s *Session) newPennsieveRequest(method string, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating GET %s request: %w", url, err)
	}
	request.Header.Add("accept", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	return request, nil
}

func (s *Session) InvokePennsieve(method string, url string, body io.Reader) (*http.Response, error) {

	req, err := s.newPennsieveRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating %s %s request: %w", method, url, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error invoking %s %s: %w", method, url, err)
	}
	if err := checkHTTPStatus(res); err != nil {
		// if there was an error, checkHTTPStatus read the body
		if closeError := res.Body.Close(); closeError != nil {
			logger.Warn("error closing body from http status error",
				slog.String("method", method),
				slog.String("url", url),
				slog.Any("error", closeError))
		}
		return nil, err
	}
	return res, nil
}

// checkHTTPStatus returns an error if 400 <= response status code < 600. Otherwise, returns nil.
// If an error is being returned, this function will consume response.Body so it should be
// called before the caller has read the body.
func checkHTTPStatus(response *http.Response) error {
	readBody := func() []byte {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return []byte(fmt.Sprintf("<unable to read body: %s>", err.Error()))
		}
		return body
	}
	if http.StatusBadRequest <= response.StatusCode && response.StatusCode < 600 {
		responseBody := readBody()
		errorType := "client"
		if response.StatusCode >= http.StatusInternalServerError {
			errorType = "server"
		}
		return fmt.Errorf("%s error %s calling %s %s; response body: %s",
			errorType,
			response.Status,
			response.Request.Method,
			response.Request.URL,
			string(responseBody))
	}
	return nil
}
