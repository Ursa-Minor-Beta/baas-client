package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type testReporter struct{}

func (r *testReporter) Report(msg string) {
	fmt.Println(msg)
}

func newLocalDebugProgram(t *testing.T, opts ...Option) (Program, context.CancelFunc) {
	RegisterTestingT(t)
	baasURL := os.Getenv("BAAS_URL")
	if baasURL == "" {
		t.Skip("BAAS_URL is not set; start a BaaS instance and point BAAS_URL at it to run this test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	p, err := NewProgram(ctx, Config{
		UseProxy:       true,
		LocalDebug:     strings.HasPrefix(baasURL, "http://localhost"),
		Url:            baasURL,
		ApiKey:         os.Getenv("BAAS_API_KEY"),
		Timeout:        "600s",
		MessageTimeout: "120s",
	}, &testReporter{}, opts...)
	Expect(err).To(BeNil())

	return p, cancel
}
