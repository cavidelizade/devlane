package handler

import "testing"

func TestIsServableObjectPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"uploads/2026/07/abc.png", true},
		{"attachments/11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222", true},
		{"", false},
		{"attachments/../uploads/secret", false},
		{"uploads/../../etc/passwd", false},
		{"secrets/key", false},
		{"attachmentsfoo/x", false},
		{"uploadsfoo/x", false},
	}
	for _, c := range cases {
		if got := isServableObjectPath(c.path); got != c.want {
			t.Errorf("isServableObjectPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
