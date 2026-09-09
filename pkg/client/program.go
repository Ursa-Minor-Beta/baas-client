package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/savioxavier/termlink"

	"github.com/Ursa-Minor-Beta/baas-client/pkg/client/dto"
)

// ActionOption augments a browser action invocation by appending option flags
// that the remote browser service parses in service/browser.go.
type ActionOption func(args []string) []string

// WithoutTimeout disables the remote action timeout by adding the
// "withoutTimeout" option expected by the browser service.
func WithoutTimeout() ActionOption {
	return func(args []string) []string {
		return append(args, "withoutTimeout")
	}
}

// WithAllowTags limits sanitization during HTML replacement by forwarding the
// allowed tags list to the remote browser action.
func WithAllowTags(tags ...string) ActionOption {
	return func(args []string) []string {
		return append(args, fmt.Sprintf("allowTags:%s", strings.Join(tags, ",")))
	}
}

// WithAllowAttrs limits sanitization during HTML replacement by forwarding the
// allowed attribute list to the remote browser action.
func WithAllowAttrs(attrs ...string) ActionOption {
	return func(args []string) []string {
		return append(args, fmt.Sprintf("allowAttributes:%s", strings.Join(attrs, ",")))
	}
}

// WithTimeout overrides the remote action timeout by passing the timeout value
// understood by the browser service.
func WithTimeout(timeout string) ActionOption {
	return func(args []string) []string {
		return append(args, fmt.Sprintf("timeout:%s", timeout))
	}
}

// WithSelector scopes LLM-assisted actions to a specific selector understood by
// the browser service.
func WithSelector(selector string) ActionOption {
	return func(args []string) []string {
		return append(args, fmt.Sprintf("selector:%s", selector))
	}
}

// WithSecretArgs instructs the browser service to redact arguments in logs for
// sensitive operations.
func WithSecretArgs() ActionOption {
	return func(args []string) []string {
		return append(args, "secretArgs")
	}
}

// WithIncludeInvisible allows actions that iterate elements to include hidden
// nodes, matching the service action options.
func WithIncludeInvisible() ActionOption {
	return func(args []string) []string {
		return append(args, "includeInvisible")
	}
}

// WithIframe executes the action within the iframe identified by selector as
// supported by the browser service.
func WithIframe(selector string) ActionOption {
	return func(args []string) []string {
		return append(args, fmt.Sprintf("iframe:%s", selector))
	}
}

// Program describes the client façade over browser actions that the server
// implements in service/browser.go.
type Program interface {
	Error() error
	NavigateStatus(url string, opts ...ActionOption) (int, error)
	TakeScreenshot(name string, opts ...ActionOption) ([]byte, error)
	LlmSetValue(desc, value string, opts ...ActionOption) error
	LlmSetValueSkipVerify(desc, value string, opts ...ActionOption) error
	LlmLogin(username, password string, opts ...ActionOption) error
	GetURL(opts ...ActionOption) (string, error)
	Click(selector string, opts ...ActionOption) error
	ClickN(selector string, index int, opts ...ActionOption) error
	GetSecret(name string, opts ...ActionOption) (string, error)
	GetValue(name string, opts ...ActionOption) (string, error)
	OuterHtml(selector string, opts ...ActionOption) (string, error)
	InnerHtml(selector string, opts ...ActionOption) (string, error)
	IsElementPresent(selector string, opts ...ActionOption) (bool, error)
	CountElements(selector string, opts ...ActionOption) (int, error)
	LlmClick(description string, opts ...ActionOption) error
	LlmClickElement(elems []string, description string, opts ...ActionOption) error
	LlmSendKeys(description, value string, opts ...ActionOption) error
	LlmText(description string, opts ...ActionOption) (string, error)
	Log(message string, opts ...ActionOption) error
	LogURL(opts ...ActionOption) error
	Navigate(url string, opts ...ActionOption) error
	Reload(opts ...ActionOption) error
	ScrollToBottom(opts ...ActionOption) error
	EvaluateJS(script string, opts ...ActionOption) (any, error)
	ReplaceInnerHtml(selector, html string, opts ...ActionOption) error
	GetElementValueN(selector string, index int, opts ...ActionOption) (string, error)
	SetValueN(selector string, index int, value string, opts ...ActionOption) error
	GetInnerText(selector string, opts ...ActionOption) (string, error)
	SendKeysToElement(selector string, keys string, opts ...ActionOption) error
	SendKeys(text string, opts ...ActionOption) error
	Sleep(duration string, opts ...ActionOption) error
	Submit(selector string, opts ...ActionOption) error
	Text(selector string, opts ...ActionOption) (string, error)
	WaitFileDownload(duration string, opts ...ActionOption) (bool, error)
	ExecuteAndDownloadFile(program string, fileName string, waitStarted, waitDownloaded string, opts ...ActionOption) ([]byte, error)
	DownloadFile(fileName string, waitStarted, waitDownloaded string, opts ...ActionOption) ([]byte, error)
	UploadFileFromURL(fileURL, selector string, opts ...ActionOption) error
	WaitReady(selector string, opts ...ActionOption) error
	WaitVisible(selector string, opts ...ActionOption) error
	SaveScreenshot(name string, fileName string, opts ...ActionOption) error
	FindVisibleElements(elements []string, attributeName string, opts ...ActionOption) (string, error)
	Execute(program string, opts ...ActionOption) (any, error)
	DragAndDropBySelectors(from, to string, opts ...ActionOption) error
	ScrollIntoView(selector string, opts ...ActionOption) error
	ScrollIntoViewN(selector string, index int, opts ...ActionOption) error
	WaitForHtml(textOrHtml string, opts ...ActionOption) error
	WaitForText(text string, opts ...ActionOption) error
	GetElementInnerTextN(selector string, index int, opts ...ActionOption) (string, error)
}

type Reporter interface {
	Report(msg string)
}

type Config struct {
	UseProxy       bool                `json:"useProxy" yaml:"useProxy"`
	LocalDebug     bool                `json:"localDebug" yaml:"localDebug"`
	Url            string              `json:"url" yaml:"url"`
	ApiKey         string              `json:"apiKey" yaml:"apiKey"`
	Timeout        string              `json:"timeout" yaml:"timeout"`
	MessageTimeout string              `json:"messageTimeout" yaml:"messageTimeout"`
	Secrets        []string            `json:"secrets" yaml:"secrets"`
	Values         []string            `json:"values" yaml:"values"`
	Cookies        []dto.BrowserCookie `json:"cookies" yaml:"cookies"`
}

// Option customizes the client program before the backing browser session is
// created.
type Option func(p *program)

// WithSecrets injects static secret values that the remote automation layer
// reads through supportedActions["getSecret"] in service/browser.go.
func WithSecrets(secrets map[string]string) Option {
	return func(p *program) {
		p.secrets = secrets
	}
}

// WithValues injects static values that the remote automation layer reads via
// supportedActions["getValue"] in service/browser.go.
func WithValues(values map[string]string) Option {
	return func(p *program) {
		p.values = values
	}
}

// NewProgram establishes a remote browser session through Client.RunAsync and
// keeps the session identifier for subsequent calls into service/browser.go.
func NewProgram(ctx context.Context, cfg Config, reporter Reporter, opts ...Option) (Program, error) {
	client := NewClient(cfg.Url, cfg.ApiKey, time.Second*30)
	ctx, cancel := context.WithCancel(ctx)

	p := &program{
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
		reporter: reporter,
		cfg:      cfg,
	}

	for _, opt := range opts {
		opt(p)
	}

	go func() {
		defer cancel()
		res, wait, err := client.RunAsync(ctx, dto.Config{
			Browser: dto.BrowserOpts{
				Headful:          cfg.LocalDebug,
				ReturnScreenshot: lo.ToPtr(true),
				Timeout:          cfg.Timeout,
				Cookies:          cfg.Cookies,
			},
			UseRandomProxy: lo.ToPtr(cfg.UseProxy),
		})
		if err != nil {
			p.exitWithError(err)
			return
		}
		if res.Error != "" {
			p.exitWithError(errors.Errorf("%s", res.Error))
			return
		}
		p.sessionID = res.SessionID
		wait()
	}()

	// wait until ready
	for p.sessionID == "" {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			reporter.Report("Waiting for sessionID...")
			time.Sleep(200 * time.Millisecond)
		}
	}
	reporter.Report("Got sessionID: " + p.sessionID)

	return p, nil
}

type program struct {
	client    Client
	err       error
	ctx       context.Context
	cancel    func()
	sessionID string
	reporter  Reporter
	secrets   map[string]string
	values    map[string]string
	cfg       Config
}

func (p *program) Error() error {
	return p.err
}

func (p *program) exitWithError(err error) {
	p.err = err
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *program) runProgram(prog string) (*dto.BrowserMessageOut, error) {
	p.reporter.Report(fmt.Sprintf("Executing %q...", prog))
	res, err := p.client.Message(p.ctx, dto.BrowserMessageIn{
		SessionID: p.sessionID,
		Program:   prog,
		Secrets:   p.secrets,
		Values:    p.values,
		Timeout:   p.cfg.MessageTimeout,
	})
	p.reporter.Report(fmt.Sprintf("Got result: %v (%s), %v", lo.FromPtr(res).Value, lo.FromPtr(res).Error, err))
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, errors.Errorf("%s", res.Error)
	}
	return res, nil
}

// SetValueN executes supportedActions["setValueN"] to set the value of the nth
// element matching selector in service/browser.go.
func (p *program) SetValueN(selector string, index int, value string, opts ...ActionOption) error {
	_, err := p.runProgram(fmt.Sprintf("setValueN('%s', %d, '%s'%s)", selector, index, value, p.addArgs(opts)))
	if err != nil {
		return err
	}
	return nil
}

// GetElementValueN proxies supportedActions["getElementValueN"] to read the
// value of the nth element matching selector from service/browser.go.
func (p *program) GetElementValueN(selector string, index int, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(fmt.Sprintf("getElementValueN('%s', %d%s)", selector, index, p.addArgs(opts)))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// ClickN triggers supportedActions["clickN"] to click the nth element matching
// selector in service/browser.go.
func (p *program) ClickN(selector string, index int, opts ...ActionOption) error {
	_, err := p.runProgram(fmt.Sprintf("clickN('%s', %d%s)", selector, index, p.addArgs(opts)))
	if err != nil {
		return err
	}
	return nil
}

// Click calls supportedActions["click"] to click the first element matching the
// selector in service/browser.go.
func (p *program) Click(selector string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("click", selector, opts...))
	if err != nil {
		return err
	}
	return nil
}

// GetElementInnerTextN awaits supportedActions["getElementInnerTextN"] to read
// the innerText from the nth matching element in service/browser.go.
func (p *program) GetElementInnerTextN(selector string, index int, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(fmt.Sprintf("getElementInnerTextN('%s', %d%s)", selector, index, p.addArgs(opts)))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// GetInnerText maps to supportedActions["getInnerText"] to retrieve text content
// via service/browser.go.
func (p *program) GetInnerText(selector string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall1("getInnerText", selector, opts...))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// ScrollIntoViewN wraps supportedActions["scrollIntoViewN"] to scroll to the nth
// element in service/browser.go.
func (p *program) ScrollIntoViewN(selector string, index int, opts ...ActionOption) error {
	_, err := p.runProgram(fmt.Sprintf("scrollIntoViewN('%s', %d%s)", selector, index, p.addArgs(opts)))
	return err
}

// ScrollIntoView uses supportedActions["scrollIntoView"] to bring the element
// into view through service/browser.go.
func (p *program) ScrollIntoView(selector string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("scrollIntoView", selector, opts...))
	return err
}

// WaitForHtml calls supportedActions["waitForHtml"] to wait for HTML/text
// presence logic implemented in service/browser.go.
func (p *program) WaitForHtml(textOrHtml string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("waitForHtml", textOrHtml, opts...))
	return err
}

// WaitForText relies on supportedActions["waitForText"] to wait for text to
// appear using service/browser.go.
func (p *program) WaitForText(text string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("waitForText", text, opts...))
	return err
}

// GetSecret resolves supportedActions["getSecret"] to fetch a stored secret in
// service/browser.go.
func (p *program) GetSecret(name string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall1("getSecret", name, opts...))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// GetValue resolves supportedActions["getValue"] to fetch a stored value in
// service/browser.go.
func (p *program) GetValue(name string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall1("getValue", name, opts...))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// SendKeysToElement reaches supportedActions["sendKeysToElement"] to type into a
// specific element using service/browser.go.
func (p *program) SendKeysToElement(selector string, keys string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("sendKeysToElement", selector, keys, opts...))
	return err
}

// CountElements forwards to supportedActions["countElements"] to count matching
// nodes in service/browser.go.
func (p *program) CountElements(selector string, opts ...ActionOption) (int, error) {
	res, err := p.runProgram(p.functionCall1("countElements", selector, opts...))
	if err != nil {
		return 0, err
	}
	return int(res.Value.(float64)), nil
}

// IsElementPresent calls supportedActions["isElementPresent"] to check for a
// selector match via service/browser.go.
func (p *program) IsElementPresent(selector string, opts ...ActionOption) (bool, error) {
	res, err := p.runProgram(p.functionCall1("isElementPresent", selector, opts...))
	if err != nil {
		return false, err
	}
	return res.Value.(bool), nil
}

// LlmClick invokes supportedActions["llmClick"] for LLM-driven element
// selection handled in service/browser.go.
func (p *program) LlmClick(description string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("llmClick", description, opts...))
	return err
}

// LlmSendKeys delegates to supportedActions["llmSendKeys"] to let the backend
// locate inputs through LLM heuristics.
func (p *program) LlmSendKeys(description, value string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("llmSendKeys", description, value, opts...))
	return err
}

// LlmClickElement proxies supportedActions["llmClickElement"] to attempt LLM
// guided clicks among candidate selectors.
func (p *program) LlmClickElement(elements []string, description string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("llmClickElement", strings.Join(elements, ","), description, opts...))
	if err != nil {
		return err
	}
	return nil
}

// FindVisibleElements mirrors supportedActions["findVisibleElements"] to collect
// element markup from service/browser.go.
func (p *program) FindVisibleElements(elements []string, addAttributeName string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall2("findVisibleElements", strings.Join(elements, ","), addAttributeName, opts...))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// LlmText ties into supportedActions["llmText"] to extract text via LLM
// guidance in service/browser.go.
func (p *program) LlmText(description string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall1("llmText", description, opts...))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// Log uses supportedActions["log"] to append a message to the server logs.
func (p *program) Log(message string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("log", message, opts...))
	return err
}

// LogURL wraps supportedActions["logURL"] to log the current page URL via the
// browser service helper.
func (p *program) LogURL(opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall0("logURL", opts...))
	return err
}

// ScrollToBottom binds to supportedActions["scrollToBottom"] to move to the end
// of the document.
func (p *program) ScrollToBottom(opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall0("scrollToBottom", opts...))
	return err
}

// EvaluateJS forwards to supportedActions["evaluateJS"] to execute arbitrary
// JavaScript in the remote browser.
func (p *program) EvaluateJS(script string, opts ...ActionOption) (any, error) {
	return p.runProgram(p.functionCall1("evaluateJS", script, opts...))
}

// Reload proxies supportedActions["reload"] to refresh the current page in the
// remote browser.
func (p *program) Reload(opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall0("reload", opts...))
	return err
}

// Navigate uses supportedActions["navigate"] to open the provided URL via
// service/browser.go.
func (p *program) Navigate(url string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("navigate", url, opts...))
	return err
}

// OuterHtml captures the recorded outer HTML from supportedActions["outerHtml"].
func (p *program) OuterHtml(selector string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall1("outerHtml", selector, opts...))
	if err != nil {
		return "", err
	}
	return res.OutHTML, nil
}

// InnerHtml captures the recorded inner HTML from supportedActions["innerHtml"].
func (p *program) InnerHtml(selector string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall1("innerHtml", selector, opts...))
	if err != nil {
		return "", err
	}
	return res.OutHTML, nil
}

// ReplaceInnerHtml mirrors supportedActions["replaceInnerHtml"] to patch page
// content on the remote browser.
func (p *program) ReplaceInnerHtml(selector, html string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("replaceInnerHtml", selector, html, opts...))
	return err
}

// SendKeys executes supportedActions["sendKeys"] to dispatch keystrokes to the
// page context in service/browser.go.
func (p *program) SendKeys(text string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("sendKeys", text, opts...))
	return err
}

// Sleep uses supportedActions["sleep"] to block for a duration remotely.
func (p *program) Sleep(duration string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("sleep", duration, opts...))
	return err
}

// Submit leverages supportedActions["submit"] to submit a form via the backend.
func (p *program) Submit(selector string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("submit", selector, opts...))
	return err
}

// Text corresponds to supportedActions["text"] to capture element text values.
func (p *program) Text(selector string, opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall1("text", selector, opts...))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// WaitFileDownloadStarted checks supportedActions["waitFileDownloadStarted"] for
// download initiation handled in service/browser.go.
func (p *program) WaitFileDownloadStarted(duration string, opts ...ActionOption) (bool, error) {
	res, err := p.runProgram(p.functionCall1("waitFileDownloadStarted", duration, opts...))
	if err != nil {
		return false, err
	}
	return res.Value.(bool), nil
}

// WaitFileDownload awaits supportedActions["waitFileDownload"] for completion
// state managed in service/browser.go.
func (p *program) WaitFileDownload(duration string, opts ...ActionOption) (bool, error) {
	res, err := p.runProgram(p.functionCall1("waitFileDownload", duration, opts...))
	if err != nil {
		return false, err
	}
	return res.Value.(bool), nil
}

// ExecuteAndDownloadFile orchestrates client-side execution and download helpers
// by composing supportedActions["waitFileDownloadStarted"] and
// supportedActions["waitFileDownload"] in service/browser.go.
func (p *program) ExecuteAndDownloadFile(program string, fileName string, waitStarted, waitDownloaded string, opts ...ActionOption) ([]byte, error) {
	res, err := p.runProgram(fmt.Sprintf(`
			%s
			if (!%s) {
				throw 'File download did not start within %s';
			}
			%s`,
		program,
		p.functionCall1("waitFileDownloadStarted", waitStarted, opts...),
		waitStarted,
		p.functionCall1("waitFileDownload", waitDownloaded, opts...)))
	if err != nil {
		return nil, err
	}
	if len(res.DownloadedFile) == 0 {
		return nil, errors.Errorf("downloaded file size is zero")
	}
	var message string
	if err := os.WriteFile(fileName, res.DownloadedFile, 0o644); err != nil {
		message = fmt.Sprintf("failed to save file %s: %q", fileName, err.Error())
		p.reporter.Report(message)
		return nil, err
	} else {
		message = fmt.Sprintf("%q saved to ", fileName) +
			termlink.ColorLink(fileName, fmt.Sprintf("file://%s", fileName), "italic green")
		p.reporter.Report(message)
	}
	return res.DownloadedFile, nil
}

// DownloadFile is a helper over ExecuteAndDownloadFile without preliminary
// script execution.
func (p *program) DownloadFile(fileName string, waitStarted, waitDownloaded string, opts ...ActionOption) ([]byte, error) {
	return p.ExecuteAndDownloadFile("", fileName, waitStarted, waitDownloaded, opts...)
}

// UploadFileFromURL proxies supportedActions["uploadFileFromUrl"] to download a
// remote file and attach it to an input element.
func (p *program) UploadFileFromURL(fileURL, selector string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("uploadFileFromUrl", fileURL, selector, opts...))
	return err
}

// DragAndDropBySelectors bridges to supportedActions["dragAndDropBySelectors"]
// for drag-and-drop automation written in service/browser.go.
func (p *program) DragAndDropBySelectors(from, to string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("dragAndDropBySelectors", from, to, opts...))
	return err
}

// WaitReady delegates to supportedActions["waitReady"] to block until the
// element is ready as implemented in service/browser.go.
func (p *program) WaitReady(selector string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("waitReady", selector, opts...))
	return err
}

// WaitVisible binds to supportedActions["waitVisible"] to wait for visibility
// conditions in service/browser.go.
func (p *program) WaitVisible(selector string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall1("waitVisible", selector, opts...))
	return err
}

// NavigateStatus returns the HTTP status reported by supportedActions["navigateStatus"].
func (p *program) NavigateStatus(url string, opts ...ActionOption) (int, error) {
	res, err := p.runProgram(p.functionCall1("navigateStatus", url, opts...))
	if err != nil {
		return 0, err
	}
	status, ok := res.Value.(float64)
	if !ok {
		return 0, errors.Errorf("Failed to convert status code to int %v", res.Value)
	}
	return int(status), nil
}

// TakeScreenshot calls supportedActions["takeScreenshot"] and returns the
// stored bytes from service/browser.go.
func (p *program) TakeScreenshot(name string, opts ...ActionOption) ([]byte, error) {
	res, err := p.runProgram(p.functionCall1("takeScreenshot", name, opts...))
	if err != nil {
		return nil, err
	}
	if len(res.Screenshots[name]) == 0 {
		return nil, errors.Errorf("screenshot with name %s wasn't returned", name)
	}
	return res.Screenshots[name], nil
}

// SaveScreenshot wraps TakeScreenshot and persists the captured bytes locally.
func (p *program) SaveScreenshot(name string, fileName string, opts ...ActionOption) error {
	screenshot, err := p.TakeScreenshot(name, opts...)
	if err != nil {
		return err
	}
	var message string
	if err := os.WriteFile(fileName, screenshot, 0o644); err != nil {
		message = fmt.Sprintf("failed to save %q to %s: %q", name, fileName, err.Error())
		p.reporter.Report(message)
		return err
	} else {
		message = fmt.Sprintf("%q saved to ", name) +
			termlink.ColorLink(name, fmt.Sprintf("file://%s", fileName), "italic green")
		p.reporter.Report(message)
	}
	return nil
}

// LlmSetValue proxies supportedActions["llmSetValue"] for described input
// fields.
func (p *program) LlmSetValue(desc, value string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("llmSetValue", desc, value, opts...))
	if err != nil {
		return err
	}
	return nil
}

// LlmSetValueSkipVerify keeps compatibility with deprecated
// supportedActions["llmSetValueSkipVerify"].
func (p *program) LlmSetValueSkipVerify(desc, value string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("llmSetValueSkipVerify", desc, value, opts...))
	if err != nil {
		return err
	}
	return nil
}

// LlmLogin reuses supportedActions["llmLogin"] for credential entry flows.
func (p *program) LlmLogin(username, password string, opts ...ActionOption) error {
	_, err := p.runProgram(p.functionCall2("llmLogin", username, password, opts...))
	if err != nil {
		return err
	}
	return nil
}

// Execute sends an arbitrary program string to runProgram, which ultimately
// executes through service/browser.go.
func (p *program) Execute(program string, opts ...ActionOption) (any, error) {
	res, err := p.runProgram(program)
	if err != nil {
		return "", err
	}
	return res.Value, nil
}

// GetURL resolves supportedActions["getURL"] to report the current page URL.
func (p *program) GetURL(opts ...ActionOption) (string, error) {
	res, err := p.runProgram(p.functionCall0("getURL", opts...))
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

// functionCall0 delegates to functionCallN, which is the only variant that
// drops the separator addArgs prepends. Formatting the options directly here
// emits "name(, 'timeout:2s')" — invalid JavaScript — as soon as the caller
// passes an option.
func (p *program) functionCall0(name string, opts ...ActionOption) string {
	return p.functionCallN(name, opts)
}

func (p *program) functionCall1(name, arg1 string, opts ...ActionOption) string {
	return fmt.Sprintf("%s('%s'%s)", name, arg1, p.addArgs(opts))
}

func (p *program) functionCall2(name, arg1, arg2 string, opts ...ActionOption) string {
	return fmt.Sprintf("%s('%s', '%s'%s)", name, arg1, arg2, p.addArgs(opts))
}

func (p *program) functionCallN(name string, args ...any) string {
	var (
		argStrings []string
		argOpts    []ActionOption
	)
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			argStrings = append(argStrings, v)
		case ActionOption:
			argOpts = append(argOpts, v)
		case []ActionOption:
			argOpts = append(argOpts, v...)
		default:
			argStrings = append(argStrings, fmt.Sprint(v))
		}
	}

	var b strings.Builder
	b.Grow(len(name) + len(argStrings)*8)
	b.WriteString(name)
	b.WriteByte('(')
	for i, arg := range argStrings {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('\'')
		b.WriteString(arg)
		b.WriteByte('\'')
	}
	extra := p.addArgs(argOpts)
	if extra != "" {
		if len(argStrings) == 0 && len(extra) > 2 {
			b.WriteString(extra[2:])
		} else {
			b.WriteString(extra)
		}
	}
	b.WriteByte(')')
	return b.String()
}

func (p *program) addArgs(opts []ActionOption) string {
	var addArgs []string
	for _, opt := range opts {
		addArgs = opt(addArgs)
	}
	addArgsString := ""
	if len(addArgs) > 0 {
		addArgsString = ", " + fmt.Sprintf("'%s'", strings.Join(addArgs, "','"))
	}
	return addArgsString
}
