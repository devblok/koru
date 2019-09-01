package kar_test

import (
	"errors"
	"io"
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

func BenchmarkReadFromMemoryMapped(b *testing.B) {
	r, err := mmap.Open("testdata/assets.kar")
	if err != nil {
		b.Error(err)
	}

	ar, err := kar.Open(r)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		info, err := ar.GetFileInfo("assets/Bricks_COLOR.png")
		if err != nil {
			b.Error(err)
		}

		f, err := ar.Open("assets/Bricks_COLOR.png")
		if err != nil {
			b.Error(err)
		}

		fileContents := make([]byte, info.Size)
		if _, err := f.Read(fileContents); err != nil && err != io.EOF {
			b.Error(err)
		}
	}
}

func BenchmarkReadAllFromMemoryMapped(b *testing.B) {
	r, err := mmap.Open("testdata/assets.kar")
	if err != nil {
		b.Error(err)
	}

	ar, err := kar.Open(r)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, err := ar.ReadAll("assets/Bricks_COLOR.png")
		if err != nil {
			b.Error(err)
		}
	}
}
