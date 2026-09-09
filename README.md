# BaaS Client

A Go client library for Browser as a Service (BaaS) - a remote browser automation service that provides programmatic control over web browsers through a simple API.

## Overview

The BaaS Client library allows you to control remote browser instances for web automation tasks such as:
- Web scraping and data extraction
- Automated testing of web applications
- Form filling and submission
- Screenshot capture
- File downloads and uploads
- LLM-assisted web interactions

## Requirements

This is a client. It talks to a running BaaS server, which is a separate
project: https://github.com/Ursa-Minor-Beta/baas. You need its URL and API key
before anything here does something useful.

## Installation

```bash
go get github.com/Ursa-Minor-Beta/baas-client
```

## Command-line client

The repository also ships `baas-cli`, an interactive terminal client for
driving a session by hand — useful for working out a program before putting it
in code.

```bash
go build -o baas-cli ./cmd/baas-cli
BAAS_URL=http://localhost:8090 BAAS_API_KEY=your-key ./baas-cli
```

| Flag | Short | Description |
|---|---|---|
| `--url` | `-u` | BaaS backend URL (defaults to `BAAS_URL`, else `http://localhost:8090`) |
| `--key` | `-k` | API key (defaults to `BAAS_API_KEY`) |
| `--debug` | `-d` | Local debug mode |
| `--proxy` | `-p` | Route browser traffic through the server's proxy |
| `--timeout` | `-t` | Max session length, e.g. `10m` |
| `--message-timeout` | `-M` | Max wait per message, e.g. `30s` |
| `--secret` | `-S` | Secret to send with each request, repeatable |
| `--value` | `-V` | Value to send with each request, repeatable |
| `--cookie` | `-C` | Cookie to send with each request, repeatable |
| `--cookie-domain` | `-D` | Domain to set the cookies on |

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/Ursa-Minor-Beta/baas-client/pkg/client"
    "github.com/Ursa-Minor-Beta/baas-client/pkg/client/dto"
)

type SimpleReporter struct{}

func (r *SimpleReporter) Report(msg string) {
    fmt.Println(msg)
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
    defer cancel()

    config := client.Config{
        Url:            "https://your-baas-service.com",
        ApiKey:         "your-api-key",
        LocalDebug:     false,
        Timeout:        "300s",
        MessageTimeout: "75s",
        UseProxy:       false,
    }

    reporter := &SimpleReporter{}
    program, err := client.NewProgram(ctx, config, reporter)
    if err != nil {
        panic(err)
    }

    // Navigate to a website
    err = program.Navigate("https://example.com")
    if err != nil {
        panic(err)
    }

    // Take a screenshot
    screenshot, err := program.TakeScreenshot("example-page")
    if err != nil {
        panic(err)
    }

    // Save screenshot to file
    err = program.SaveScreenshot("example-page", "example.png")
    if err != nil {
        panic(err)
    }

    fmt.Println("Screenshot saved successfully!")
}
```

## Configuration

### Config Structure

```go
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
```

### Configuration Options

- **Url**: The BaaS service endpoint URL
- **ApiKey**: Authentication key for the BaaS service
- **LocalDebug**: Enable headful browser mode for debugging (shows browser window)
- **Timeout**: Overall browser session timeout
- **MessageTimeout**: Timeout for individual browser actions
- **UseProxy**: Enable proxy usage for requests
- **Secrets**: Static secret values accessible via `GetSecret()`
- **Values**: Static values accessible via `GetValue()`
- **Cookies**: Pre-set cookies for the browser session

## Core Features

### Navigation

```go
// Navigate to a URL
err := program.Navigate("https://example.com")

// Navigate and get HTTP status
status, err := program.NavigateStatus("https://example.com")

// Get current URL
url, err := program.GetURL()

// Reload the page
err := program.Reload()
```

### Element Interaction

```go
// Click an element
err := program.Click("button#submit")

// Click the nth element matching selector
err := program.ClickN("button.item", 2)

// Send keys to an element
err := program.SendKeysToElement("input#username", "john@example.com")

// Send keys to the page (global)
err := program.SendKeys("Hello World")

// Submit a form
err := program.Submit("form#login")
```

### Element Information

```go
// Check if element exists
exists, err := program.IsElementPresent("div.content")

// Count matching elements
count, err := program.CountElements("li.item")

// Get element text content
text, err := program.Text("h1.title")

// Get inner text
innerText, err := program.GetInnerText("div.content")

// Get inner HTML
html, err := program.InnerHtml("div.content")

// Get outer HTML
outerHtml, err := program.OuterHtml("div.content")

// Get element value (for inputs)
value, err := program.GetElementValueN("input.field", 0)
```

### LLM-Assisted Actions

The library supports AI-powered element detection and interaction:

```go
// LLM-guided clicking based on description
err := program.LlmClick("Sign in button")

// LLM-guided text input
err := program.LlmSendKeys("email field", "user@example.com")

// LLM-guided login
err := program.LlmLogin("username", "password")

// LLM-guided value setting
err := program.LlmSetValue("search box", "golang tutorial")

// LLM-guided text extraction
text, err := program.LlmText("page title")
```

### Screenshots and Files

```go
// Take screenshot and return bytes
screenshot, err := program.TakeScreenshot("page-name")

// Save screenshot to file
err := program.SaveScreenshot("page-name", "screenshot.png")

// Download a file
fileData, err := program.DownloadFile("document.pdf", "5s", "30s")

// Upload file from URL
err := program.UploadFileFromURL("https://example.com/file.pdf", "input[type=file]")
```

### Waiting and Timing

```go
// Wait for element to be ready
err := program.WaitReady("div.content")

// Wait for element to be visible
err := program.WaitVisible("button.submit")

// Wait for specific text to appear
err := program.WaitForText("Success!")

// Wait for HTML content
err := program.WaitForHtml("<div>Complete</div>")

// Sleep for duration
err := program.Sleep("2s")

// Wait for file download
downloaded, err := program.WaitFileDownload("30s")
```

### Advanced Actions

```go
// Execute JavaScript
result, err := program.EvaluateJS("return document.title;")

// Drag and drop
err := program.DragAndDropBySelectors("#source", "#target")

// Scroll to element
err := program.ScrollIntoView("footer")

// Scroll to bottom
err := program.ScrollToBottom()

// Replace element HTML
err := program.ReplaceInnerHtml("div.content", "<p>New content</p>")
```

## Action Options

Many methods accept optional `ActionOption` parameters to customize behavior:

```go
// Disable timeout for long-running actions
err := program.Navigate("https://slow-site.com", client.WithoutTimeout())

// Set custom timeout
err := program.Click("button", client.WithTimeout("10s"))

// Include invisible elements in searches
count, err := program.CountElements("div", client.WithIncludeInvisible())

// Scope action to specific selector
err := program.LlmClick("submit", client.WithSelector("form.login"))

// Execute within iframe
err := program.Click("button", client.WithIframe("iframe#content"))

// Mark arguments as secret (for logging)
err := program.LlmLogin("user", "pass", client.WithSecretArgs())

// Allow specific HTML tags during sanitization
err := program.ReplaceInnerHtml("div", "<b>Bold</b>", client.WithAllowTags("b", "i"))

// Allow specific attributes during sanitization
err := program.ReplaceInnerHtml("div", `<a href="/">Link</a>`, client.WithAllowAttrs("href"))
```

## Error Handling

```go
program, err := client.NewProgram(ctx, config, reporter)
if err != nil {
    // Handle program creation error
}

// Check for program errors
if err := program.Error(); err != nil {
    // Handle runtime error
}
```

## Secrets and Values

You can inject static secrets and values that are accessible during automation:

```go
secrets := map[string]string{
    "username": "john@example.com",
    "password": "secret123",
}

values := map[string]string{
    "base_url": "https://staging.example.com",
    "timeout": "30s",
}

program, err := client.NewProgram(ctx, config, reporter,
    client.WithSecrets(secrets),
    client.WithValues(values),
)

// Access in automation
username, err := program.GetSecret("username")
baseUrl, err := program.GetValue("base_url")
```

## Testing Example

Here's a complete example of using the library for automated testing:

```go
package main

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/Ursa-Minor-Beta/baas-client/pkg/client"
    "github.com/Ursa-Minor-Beta/baas-client/pkg/client/dto"
)

func TestWebsiteLogin(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
    defer cancel()

    config := client.Config{
        Url:            os.Getenv("BAAS_URL"),
        ApiKey:         os.Getenv("BAAS_API_KEY"),
        LocalDebug:     false,
        Timeout:        "300s",
        MessageTimeout: "75s",
        UseProxy:       false,
        Cookies: []dto.BrowserCookie{
            {
                Name:     "test-session",
                Value:    "test-value",
                Domain:   "example.com",
                Path:     "/",
                HTTPOnly: false,
                Secure:   true,
            },
        },
    }

    secrets := map[string]string{
        "username": os.Getenv("TEST_USERNAME"),
        "password": os.Getenv("TEST_PASSWORD"),
    }

    reporter := &SimpleReporter{}
    program, err := client.NewProgram(ctx, config, reporter,
        client.WithSecrets(secrets),
    )
    if err != nil {
        t.Fatalf("Failed to create program: %v", err)
    }

    // Navigate to login page
    status, err := program.NavigateStatus("https://example.com/login")
    if err != nil {
        t.Fatalf("Navigation failed: %v", err)
    }
    if status != 200 {
        t.Fatalf("Expected status 200, got %d", status)
    }

    // Wait for login form
    err = program.WaitReady("form#login")
    if err != nil {
        t.Fatalf("Login form not found: %v", err)
    }

    // Perform login using LLM assistance
    username, _ := program.GetSecret("username")
    password, _ := program.GetSecret("password")
    err = program.LlmLogin(username, password)
    if err != nil {
        t.Fatalf("Login failed: %v", err)
    }

    // Verify successful login
    err = program.WaitForText("Welcome")
    if err != nil {
        t.Fatalf("Login verification failed: %v", err)
    }

    // Take screenshot of success page
    err = program.SaveScreenshot("login-success", "test-results/login-success.png")
    if err != nil {
        t.Fatalf("Screenshot failed: %v", err)
    }

    t.Log("Login test completed successfully")
}
```

## Environment Variables

Common environment variables used with the library:

```bash
# BaaS service the client and CLI talk to
BAAS_URL=https://your-baas-service.com
BAAS_API_KEY=your-api-key

# Only read by TestLlmLogin; without all three it skips
LOGIN_TEST_URL=https://example.com/login
LOGIN_TEST_USERNAME=
LOGIN_TEST_PASSWORD=
```

## Browser Capabilities

The remote browser service supports:
- Modern web standards (HTML5, CSS3, ES6+)
- JavaScript execution
- File downloads and uploads
- Screenshots and screen recording
- Mobile device emulation
- Proxy support
- Cookie management
- Local storage access

## Best Practices

1. **Use appropriate timeouts**: Set reasonable timeouts for different types of operations
2. **Handle errors gracefully**: Always check for errors and implement proper error handling
3. **Take screenshots for debugging**: Capture screenshots at key points for troubleshooting
4. **Use LLM features wisely**: LLM-assisted actions are powerful but may be slower than direct selectors
5. **Clean up resources**: Ensure proper context cancellation to clean up browser sessions
6. **Use secrets for sensitive data**: Never hardcode credentials; use the secrets mechanism
7. **Wait for elements**: Always wait for elements to be ready before interacting with them

## Development

```bash
go build ./...
go test ./...
```

Tests that need a live BaaS instance skip themselves unless `BAAS_URL` is set.
`TestLlmLogin` additionally needs `LOGIN_TEST_URL`, `LOGIN_TEST_USERNAME` and
`LOGIN_TEST_PASSWORD`.

Formatting is `gofumpt`; `.golangci.yml` holds the lint configuration.

## Licence

Apache 2.0 — see [LICENSE](LICENSE).
