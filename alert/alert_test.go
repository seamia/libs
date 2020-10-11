package alert

import (
	"testing"
)

func TestAlpha(t *testing.T) {
	PostData("test", []byte("=)"))
}
