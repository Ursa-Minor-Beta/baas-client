package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/Ursa-Minor-Beta/baas-client/pkg/client/dto"
)

type baasClient struct {
	baasURL    string
	baasApiKey string
	timeout    time.Duration
}

type Meta struct {
	Error     *string `json:"error,omitempty" yaml:"error,omitempty"`         // whether error happened whilst processing
	UsedProxy string  `json:"usedProxy,omitempty" yaml:"usedProxy,omitempty"` // which proxy server was used for fetching
	JobUID    string  `json:"jobUID" yaml:"jobUID"`                           // unique identifier of job for debugging purposes
}

type Client interface {
	RunAsync(ctx context.Context, baasRequest dto.Config) (*dto.BrowserMessageOut, func(), error)
	Message(ctx context.Context, message dto.BrowserMessageIn) (*dto.BrowserMessageOut, error)
}

func NewClient(baasURL, baasKey string, timeout time.Duration) Client {
	return &baasClient{
		baasURL:    baasURL,
		baasApiKey: baasKey,
		timeout:    timeout,
	}
}

func (o *baasClient) runClient(ctx context.Context, headers map[string]string, endpoint string, timeout string, body any) (*http.Response, error) {
	timeoutDuration := o.timeout
	if dur, err := time.ParseDuration(timeout); err != nil {
		// nothing to do
	} else {
		timeoutDuration = dur
	}

	// Fetch page from URL
	client := &http.Client{Timeout: timeoutDuration}

	baasURL := fmt.Sprintf("%s%s", o.baasURL, endpoint)

	reqBodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal baas request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baasURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to init request for page: %v", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", o.baasApiKey))
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the page: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch the page: status code %d: %s", resp.StatusCode, string(readBytes(resp.Body)))
	}
	return resp, nil
}

func (o *baasClient) Message(ctx context.Context, msg dto.BrowserMessageIn) (*dto.BrowserMessageOut, error) {
	// generate random request ID
	msg.RequestID = lo.RandomString(20, lo.LettersCharset)
	resp, err := o.runClient(ctx, map[string]string{}, "/api/async/message", msg.Timeout, msg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make baas request")
	}
	var baasResponseObjects []dto.BrowserMessageOut
	respBytes := readBytes(resp.Body)
	// hack to prevent multiple messages to be unmarshalled (TODO: figure out when this may happen)
	if strings.Contains(string(respBytes), "}\n{") {
		respBytes = []byte(strings.ReplaceAll(strings.ReplaceAll(string(respBytes), "}{", "},{"), "}\n{", "},\n{"))
	}
	respBytes = []byte("[" + string(respBytes) + "]")

	err = json.Unmarshal(respBytes, &baasResponseObjects)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal baas response: %s", string(respBytes))
	}

	baasResponse, found := lo.Find(baasResponseObjects, func(msgOut dto.BrowserMessageOut) bool {
		return msg.RequestID == msgOut.RequestID
	})
	if !found {
		return nil, errors.Errorf("failed to find message with the same RequestID: %q", msg.RequestID)
	}

	if lo.FromPtr(baasResponse.Meta.Error) != "" {
		return nil, errors.Errorf("baas returned error: %s, baas RequestUID: %q", lo.FromPtr(baasResponse.Meta.Error), baasResponse.Meta.RequestUID)
	}
	return &baasResponse, nil
}

func (o *baasClient) RunAsync(ctx context.Context, baasRequest dto.Config) (*dto.BrowserMessageOut, func(), error) {
	resp, err := o.runClient(ctx, map[string]string{
		"Accept": "text/event-stream",
	}, "/api/async/start", baasRequest.Browser.Timeout, baasRequest)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to make baas request")
	}
	var baasResponse dto.BrowserMessageOut

	// Use a buffered reader to read the response line by line
	reader := bufio.NewReader(resp.Body)

	// Read a line from the response
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error reading response")
	}

	// Trim whitespace from the line
	line = strings.TrimSpace(line)
	err = json.Unmarshal([]byte(line), &baasResponse)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to unmarshal baas response: %s", line)
	}
	if lo.FromPtr(baasResponse.Meta.Error) != "" {
		return nil, nil, errors.Errorf("baas returned error: %s, baas RequestUID: %q", lo.FromPtr(baasResponse.Meta.Error), baasResponse.Meta.RequestUID)
	}
	return &baasResponse, func() {
		for {
			_, err := reader.ReadString('\n')
			if err != nil {
				break
			}
		}
		resp.Body.Close()
	}, nil
}

// nolint: unused
func readBytes(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(stream)
	return buf.Bytes()
}
