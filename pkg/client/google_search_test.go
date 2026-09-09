package client

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGoogleSearch(t *testing.T) {
	p, cancel := newLocalDebugProgram(t)
	defer cancel()

	s, err := p.NavigateStatus("https://google.com")
	Expect(err).To(BeNil())
	Expect(s).To(Equal(200))

	err = p.SaveScreenshot("google", "output/google.png")
	Expect(err).To(BeNil())

	err = p.LlmSetValue("Search textarea", "What is LLM?\\n")
	Expect(err).To(BeNil())

	err = p.SaveScreenshot("search", "output/search.png")
	Expect(err).To(BeNil())
}
