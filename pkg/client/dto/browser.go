package dto

import (
	"encoding/json"

	"github.com/simple-container-com/go-aws-lambda-sdk/pkg/service"
)

type Config struct {
	Browser        BrowserOpts `json:"browser" yaml:"browser"`
	SessionID      *string     `json:"sessionID" yaml:"sessionID"`                               // sessionID to use when running async requests (must be unique)
	MaxAttempts    *int        `json:"maxAttempts,omitempty" yaml:"maxAttempts,omitempty"`       // max amount of attempts to fetch/process (default: 3)
	UseRandomProxy *bool       `json:"useRandomProxy,omitempty" yaml:"useRandomProxy,omitempty"` // whether to use random proxy from the configured proxy pool (default: false)
}

type Result struct {
	Result    *BrowserResponse   `json:"result" yaml:"result"`                           // result returned by browser
	UsedProxy string             `json:"usedProxy,omitempty" yaml:"usedProxy,omitempty"` // which proxy server was used for fetching
	Meta      service.ResultMeta `json:"meta" yaml:"meta"`                               // metadata related to processing
}

type BrowserOpts struct {
	Headful             bool              `json:"headful" default:"false"`                              // run headful chrome (default: false)
	ReturnScreenshot    *bool             `json:"returnScreenshot" default:"false"`                     // whether to return screenshot after execution
	UserAgent           string            `json:"userAgent" default:""`                                 // use user-agent (default: undefined)
	UseProxy            *string           `json:"useProxy" default:""`                                  // use specific proxy server (default: undefined)
	Cookies             []BrowserCookie   `json:"cookies"`                                              // cookies to set before executing actions
	Program             string            `json:"program" required:"true"`                              // program to run (required)
	Secrets             map[string]string `json:"secrets" required:"false"`                             // program secrets to use (values can be obtained as getSecret('name'))
	Values              map[string]string `json:"values" required:"false"`                              // program values to use (values can be obtained as getValue('name'))
	WaitForFileDownload *bool             `json:"waitForFileDownload" required:"false" default:"false"` // whether to wait until file is downloaded
	Timeout             string            `json:"timeout" example:"60s" default:"60s"`                  // duration string in go duration format (e.g.: 10s)
	Width               *int              `json:"width" example:"1920" default:"1920"`                  // width of the browser window
	Height              *int              `json:"height" example:"1080" default:"1080"`                 // height of the browser window
}

type BrowserMessageIn struct {
	SessionID               string            `json:"sessionID" required:"true"`                               // sessionID to send event to
	RequestID               string            `json:"requestID"`                                               // ID of the current request (used for internal purposes)
	Program                 string            `json:"program" required:"true"`                                 // program to run (required)
	Secrets                 map[string]string `json:"secrets" required:"false"`                                // program secrets to use (values can be obtained as getSecret('name'))
	Values                  map[string]string `json:"values" required:"false"`                                 // program values to use (values can be obtained as getValue('name'))
	Timeout                 string            `json:"timeout" example:"10s" default:"10s"`                     // timeout to process message
	OperationTimeout        *string           `json:"operationTimeout" example:"20s" default:"20s"`            // timeout to execute a single command (default: 20s)
	StopSession             *bool             `json:"stopSession" example:"true" default:"false"`              // tells browser to stop session
	ErrorOnOperationTimeout *bool             `json:"errorOnOperationTimeout" required:"false" default:"true"` // whether to return error when a single operation times out (default: true)
}

func (i *BrowserMessageIn) Sanitized() any {
	bytesIn, _ := json.Marshal(i)
	inCopy := BrowserMessageIn{}
	_ = json.Unmarshal(bytesIn, &inCopy)
	inCopy.Secrets = nil
	return inCopy
}

type BrowserMessageOut struct {
	Timestamp          string             `json:"timestamp" required:"true"`    // timestamp of the response
	SessionID          string             `json:"sessionID" required:"true"`    // sessionID to send event to
	RequestID          string             `json:"requestID"`                    // ID of the current request (used for internal purposes)
	Meta               service.ResultMeta `json:"meta" yaml:"meta"`             // metadata related to processing
	Error              string             `json:"error,omitempty" yaml:"error"` // error happened when running program
	Value              any                `json:"value,omitempty" yaml:"value"` // return value
	Screenshots        map[string][]byte  `json:"screenshots,omitempty"`
	Log                []string           `json:"log,omitempty"`
	DownloadedFile     []byte             `json:"downloadedFile,omitempty"`
	DownloadedFileName string             `json:"downloadedFileName,omitempty"`
	OutHTML            string             `json:"outHtml"`
}

type BrowserCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type BrowserResponse struct {
	OutHTML        string            `json:"outHtml"`
	Screenshot     []byte            `json:"screenshot,omitempty"`
	Screenshots    map[string][]byte `json:"screenshots,omitempty"`
	DownloadedFile []byte            `json:"downloadedFile,omitempty"`
	Cookies        []BrowserCookie   `json:"cookies"`
	Log            []string          `json:"log,omitempty"`
	URL            string            `json:"url"`
	Error          *string           `json:"error"`
}
