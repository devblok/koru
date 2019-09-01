package kar_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/devblok/koru/utility/kar"
	"golang.org/x/exp/mmap"
)

func readFileAndCompare(f *kar.Reader, expected string, t *testing.T) error {
	result := make([]byte, len(expected))
	n, err := f.Read(result)
	if err != nil {
		t.Error(err)
	}
	if n < len(expected) {
		return errors.New("incorrect number of bytes read")
	}

	if strings.Compare(string(result), expected) != 0 {
		return errors.New("test string does not match up")
	}

	return nil
}

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

func TestOpenmmap(t *testing.T) {
	r, err := mmap.Open("testdata/opentest.kar")
	if err != nil {
		t.Error(err)
	}

	ar, err := kar.Open(r)
	if err != nil {
		t.Error(err)
	}

	t.Log(ar)
}

func TestOpenAndRead(t *testing.T) {
	r, err := os.Open("testdata/opentest.kar")
	if err != nil {
		t.Error(err)
	}

	ar, err := kar.Open(r)
	if err != nil {
		t.Error(err)
	}

	if f, err := ar.Open("test/test1.txt"); err != nil {
		t.Error(err)
	} else if err := readFileAndCompare(f, "this is a test", t); err != nil {
		t.Error(err)
	}

	if f, err := ar.Open("test/test2.txt"); err != nil {
		t.Error(err)
	} else if err := readFileAndCompare(f, "this is another test", t); err != nil {
		t.Error(err)
	}
}

func TestOpenAndReadAll(t *testing.T) {
	r, err := os.Open("testdata/opentest.kar")
	if err != nil {
		t.Error(err)
	}

	ar, err := kar.Open(r)
	if err != nil {
		t.Error(err)
	}

	if f, err := ar.ReadAll("test/test1.txt"); err != nil {
		t.Error(err)
	} else if strings.Compare("this is a test", string(f)) != 0 {
		t.Error(errors.New("result is not expected value"))
	}

	if f, err := ar.ReadAll("test/test2.txt"); err != nil {
		t.Error(err)
	} else if strings.Compare("this is another test", string(f)) != 0 {
		t.Error(errors.New("result is not expected value"))
	}
}
