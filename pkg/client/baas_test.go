package client

import (
	"encoding/json"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

type testObj struct {
	ID string `json:"ID"`
}

func TestUnmarshalMulti(t *testing.T) {
	RegisterTestingT(t)

	respBytes := []byte(`{"ID":"1"}` + "\n" + `{"ID":"2"}`)

	// hack to prevent multiple messages to be unmarshalled (only keep the last one)
	if strings.Contains(string(respBytes), "}\n{") {
		respBytes = []byte(strings.Replace(string(respBytes), "}\n{", "},\n{", 1))
	}
	respBytes = []byte("[" + string(respBytes) + "]")

	var objects []testObj
	err := json.Unmarshal(respBytes, &objects)
	Expect(err).To(BeNil())
	Expect(objects[len(objects)-1].ID).To(Equal("2"))
}
