// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/devblok/koru/utility/kar"
)

var (
	testString1 = "idunvovkjnreovmegihjbrqlkmfrjnb"
	testString2 = "idunvovkjnreovmsdvwrvnervnreegihjbrqlkmfrjnb"
)

func TestCreateAndRead(t *testing.T) {
	builder, err := kar.NewBuilder(kar.Header{
		Author:      "devblok",
		DateCreated: time.Now().Unix(),
		Version:     1,
	})
	if err != nil {
		t.Error(err)
	}
	builder.Add("test", bytes.NewReader([]byte(testString1)))
	builder.Add("test2", bytes.NewReader([]byte(testString2)))

	buf := bytes.NewBuffer([]byte{})
	if written, err := builder.WriteTo(buf); err != nil {
		t.Error(err)
	} else {
		t.Logf("written %d", written)
	}

	ar, err := kar.Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
	}

	f, err := ar.Open("test")
	if err != nil {
		t.Error(err)
	}

	result := make([]byte, len(testString1))
	n, err := f.Read(result)
	if err != nil {
		t.Error(err)
	}
	t.Log(n)

	if strings.Compare(string(result), testString1) != 0 {
		t.Error("test string does not match up")
	}
}

func TestCreateAndReadAll(t *testing.T) {
	builder, err := kar.NewBuilder(kar.Header{
		Author:      "devblok",
		DateCreated: time.Now().Unix(),
		Version:     1,
	})
	if err != nil {
		t.Error(err)
	}
	builder.Add("test", bytes.NewReader([]byte(testString1)))
	builder.Add("test2", bytes.NewReader([]byte(testString2)))

	buf := bytes.NewBuffer([]byte{})
	if written, err := builder.WriteTo(buf); err != nil {
		t.Error(err)
	} else {
		t.Logf("written %d", written)
	}

	ar, err := kar.Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
	}

	f, err := ar.ReadAll("test")
	if err != nil {
		t.Error(err)
	}

	if strings.Compare(string(f), testString1) != 0 {
		t.Error("test string does not match up")
	}
}
