package kar_test

import (
	"os"
	"testing"

	"github.com/devblok/koru/utility/kar"
)

func TestOpen(t *testing.T) {
	r, err := os.Open("testdata/opentest.kar")
	if err != nil {
		t.Error(err)
	}

	ar, err := kar.Open(r)
	if err != nil {
		t.Error(err)
	}

	t.Log(ar)
}
