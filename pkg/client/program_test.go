package client

import "testing"

func TestFunctionCallHelpers(t *testing.T) {
	p := &program{}

	testCases := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "functionCall0 no options",
			got:  p.functionCall0("foo"),
			want: p.functionCallN("foo"),
		},
		{
			name: "functionCall0 with options",
			got:  p.functionCall0("foo", WithTimeout("2s")),
			want: p.functionCallN("foo", WithTimeout("2s")),
		},
		{
			name: "functionCall1 no options",
			got:  p.functionCall1("click", ".button"),
			want: p.functionCallN("click", ".button"),
		},
		{
			name: "functionCall1 with options",
			got:  p.functionCall1("click", ".button", WithoutTimeout(), WithIncludeInvisible()),
			want: p.functionCallN("click", ".button", WithoutTimeout(), WithIncludeInvisible()),
		},
		{
			name: "functionCall2 no options",
			got:  p.functionCall2("setValue", "#input", "value"),
			want: p.functionCallN("setValue", "#input", "value"),
		},
		{
			name: "functionCall2 with options",
			got:  p.functionCall2("setValue", "#input", "value", WithSelector(".form")),
			want: p.functionCallN("setValue", "#input", "value", WithSelector(".form")),
		},
	}

	for _, tc := range testCases {
		if tc.got != tc.want {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.want, tc.got)
		}
	}
}
