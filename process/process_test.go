package process

import (
	"testing"
)

func TestSearchPath(t *testing.T) {
	actual := searchPath("ls")
	if actual != "/bin/ls" {
		t.Errorf("Expected searchPath('ls') to return /bin/ls, got %s", actual)
	}
}
