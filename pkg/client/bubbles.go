package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/savioxavier/termlink"

	"github.com/simple-container-com/go-aws-lambda-sdk/pkg/service"

	"github.com/Ursa-Minor-Beta/baas-client/pkg/client/dto"
	"github.com/Ursa-Minor-Beta/baas-client/pkg/util"
)

type (
	errMsg error
)

var headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF88")).Background(lipgloss.Color("#444444"))

type CliClient struct {
	viewport              viewport.Model
	messages              []string
	textarea              textarea.Model
	senderStyle           lipgloss.Style
	responseStyle         lipgloss.Style
	errorStyle            lipgloss.Style
	err                   error
	baas                  Client
	ctx                   context.Context
	sessionID             string
	sessionMeta           *service.ResultMeta
	outDir                string
	loader                spinner.Model
	inProgress            atomic.Bool
	programHistory        []string
	programHistoryPointer int
	cfg                   Config
}

func BubbleClient(ctx context.Context, cfg Config) (tea.Model, error) {
	ta := textarea.New()
	ta.Placeholder = "Start typing program... (or press Ctrl^C to exit, use Up and Down to navigate)"
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 1024

	ta.SetWidth(128)
	ta.SetHeight(6)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(160, 30)
	vp.SetContent(`Welcome to the BaaS client! Type a program and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	fmt.Printf("Connecting to %s...\n", cfg.Url)
	baas := NewClient(cfg.Url, cfg.ApiKey, time.Second*30)
	loader := spinner.New(
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
		spinner.WithSpinner(spinner.Dot),
	)
	ctx, cancel := context.WithCancel(ctx)
	c := &CliClient{
		ctx:           ctx,
		baas:          baas,
		textarea:      ta,
		messages:      []string{},
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		responseStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		errorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("330000")).Foreground(lipgloss.Color("#FF3333")),
		loader:        loader,
		err:           nil,
		cfg:           cfg,
	}

	if outDir, err := os.MkdirTemp(os.TempDir(), "baas-response"); err == nil {
		c.outDir = outDir
	} else {
		cancel()
		return nil, errors.Wrapf(err, "failed to init temp dir")
	}

	c.inProgress.Store(true)
	go func() {
		defer cancel()
		defer c.updateMessages()
		res, wait, err := baas.RunAsync(ctx, dto.Config{
			Browser: dto.BrowserOpts{
				Headful:          cfg.LocalDebug,
				ReturnScreenshot: lo.ToPtr(true),
				Timeout:          cfg.Timeout,
			},
			UseRandomProxy: lo.ToPtr(cfg.UseProxy),
		})
		if err != nil {
			c.messages = append(c.messages, c.errorStyle.Render("Browser: ")+"Failed to start session: "+err.Error())
			c.err = errors.Wrapf(err, "failed to start session")
			return
		}
		if res.Error != "" {
			c.err = errors.Errorf("%s", res.Error)
			c.messages = append(c.messages, c.errorStyle.Render("Browser: ")+"Failed to start session: "+res.Error)
			return
		}
		c.sessionID = res.SessionID
		c.messages = append(c.messages, c.responseStyle.Render("Browser: ")+fmt.Sprintf("Started session %s at %s", res.SessionID, cfg.Url))
		c.updateMessages()
		c.inProgress.Store(false)
		wait()
		c.messages = append(c.messages, c.errorStyle.Render("Browser: ")+fmt.Sprintf("Session %s has been terminated", res.SessionID))
		c.updateMessages()
	}()
	c.displaySpinner()

	return c, nil
}

func (m *CliClient) Init() tea.Cmd {
	return textarea.Blink
}

func (m *CliClient) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	if m.ctx.Err() != nil {
		return m, tea.Quit
	}
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyUp:
			if m.programHistoryPointer < len(m.programHistory) {
				m.programHistoryPointer++
				m.textarea.SetValue(m.programHistory[len(m.programHistory)-m.programHistoryPointer])
			}
		case tea.KeyDown:
			if m.programHistoryPointer > 0 {
				m.programHistoryPointer--
				m.textarea.SetValue(m.programHistory[len(m.programHistory)-m.programHistoryPointer-1])
			} else {
				m.textarea.SetValue("")
			}
		case tea.KeyEnter:
			m.inProgress.Store(true)
			currentValue := m.textarea.Value()
			m.loader.Tick()
			go func() {
				defer m.inProgress.Store(false)
				res, err := m.baas.Message(m.ctx, dto.BrowserMessageIn{
					SessionID: m.sessionID,
					Program:   currentValue,
					Timeout:   m.cfg.MessageTimeout,
					Values:    util.SliceToMap(m.cfg.Values),
					Secrets:   util.SliceToMap(m.cfg.Secrets),
				})
				m.processResponse(res, err)
			}()
			m.displaySpinner()
			m.programHistory = append(m.programHistory, currentValue)
			m.programHistoryPointer = 0
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+currentValue)
			m.updateMessages()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *CliClient) displaySpinner() {
	go func() {
		for m.inProgress.Load() {
			time.Sleep(50 * time.Millisecond)
			m.Update(m.loader.Tick())
		}
	}()
}

func (m *CliClient) updateMessages() {
	if len(m.messages) > 10 {
		m.messages = m.messages[1:]
	}
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
	m.textarea.Reset()
	m.viewport.GotoBottom()
}

func (m *CliClient) processResponse(res *dto.BrowserMessageOut, err error) {
	defer m.updateMessages()
	if err != nil {
		m.err = err
		m.messages = append(m.messages, m.errorStyle.Render("ERROR: "+err.Error()))
		return
	}
	if res.Error != "" {
		m.err = errors.Errorf("%s", res.Error)
		m.messages = append(m.messages, m.errorStyle.Render("ERROR: "+res.Error))
		return
	}
	m.sessionMeta = lo.ToPtr(res.Meta)
	m.messages = append(m.messages, m.responseStyle.Render("Browser: ")+fmt.Sprintf("%v", res.Value))
	if len(res.Screenshots) > 0 {
		for name, screenshot := range res.Screenshots {
			m.saveFile(m.outDir, "screenshot", name, screenshot)
		}
	}
	if len(res.DownloadedFile) > 0 {
		m.saveFile(m.outDir, "file", res.DownloadedFileName, res.DownloadedFile)
	}
}

func (m *CliClient) saveFile(outDir string, fileType, name string, screenshot []byte) {
	fileName := filepath.Join(outDir, fmt.Sprintf("%s.png", name))
	if fileType != "screenshot" {
		fileName = filepath.Join(outDir, name)
	}
	var message string
	if err := os.WriteFile(fileName, screenshot, 0o644); err != nil {
		message = fmt.Sprintf("failed to save %s %q to %s: %q", fileType, name, fileName, err.Error())
	} else {
		message = fmt.Sprintf("%s %q saved to ", fileType, name) +
			termlink.ColorLink(name, fmt.Sprintf("file://%s", fileName), "italic green")
	}
	m.messages = append(m.messages, m.responseStyle.Render("Browser: ")+message)
}

func (m *CliClient) View() string {
	dialogView := m.textarea.View()
	if m.inProgress.Load() {
		dialogView = m.loader.View()
	}
	header := headerStyle.Render("SessionID: " + m.sessionID)
	if m.sessionMeta != nil {
		header += headerStyle.Render(fmt.Sprintf("; duration: %fs, cost: %f", m.sessionMeta.RequestTime.Seconds(), m.sessionMeta.Cost))
	}
	return header + fmt.Sprintf(
		"\n\n%s\n\n%s",
		m.viewport.View(),
		dialogView,
	) + "\n\n"
}
