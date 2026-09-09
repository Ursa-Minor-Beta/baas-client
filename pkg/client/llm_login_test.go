package client

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

// TestLlmLogin drives the LLM-assisted login flow against a real site. Point
// LOGIN_TEST_URL at a page with a username/password form and supply the
// credentials; the test is skipped when they are absent.
func TestLlmLogin(t *testing.T) {
	loginURL := os.Getenv("LOGIN_TEST_URL")
	username := os.Getenv("LOGIN_TEST_USERNAME")
	password := os.Getenv("LOGIN_TEST_PASSWORD")
	if loginURL == "" || username == "" || password == "" {
		t.Skip("set LOGIN_TEST_URL, LOGIN_TEST_USERNAME and LOGIN_TEST_PASSWORD to run this test")
	}

	secrets := map[string]string{
		"username": username,
		"password": password,
	}
	p, cancel := newLocalDebugProgram(t, WithSecrets(secrets))
	defer cancel()

	s, err := p.NavigateStatus(loginURL)
	Expect(err).To(BeNil())
	Expect(s).To(Equal(200))

	err = p.LlmLogin(secrets["username"], secrets["password"])
	Expect(err).To(BeNil())

	err = p.WaitReady("body")
	Expect(err).To(BeNil())

	err = p.Sleep("5s")
	Expect(err).To(BeNil())

	html, err := p.FindVisibleElements([]string{"p", "div", "span", "input"}, "data-llm-id")
	Expect(err).To(BeNil())
	Expect(html).NotTo(BeEmpty())
	fmt.Println(html)

	err = p.SaveScreenshot("login", "output/login.png")
	Expect(err).To(BeNil())
}
