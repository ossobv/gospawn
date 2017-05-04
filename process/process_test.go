package process

import (
	"testing"
)

func TestSearchPathEnv(t *testing.T) {
	actual := searchPathEnv("ls")
	if actual != "/bin/ls" {
		t.Errorf("Expected searchPath('ls') to return /bin/ls, got %s", actual)
	}
}
