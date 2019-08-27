// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

import (
	"bytes"
	"testing"
	"time"
)

func TestAddAndWrite(t *testing.T) {
	builder, err := NewBuilder(Header{
		Author:      "devblok",
		DateCreated: time.Now().Unix(),
		Version:     1,
	})
	if err != nil {
		t.Error(err)
	}
	builder.Add("test", bytes.NewReader([]byte("idunvovkjnreovmegihjbrqlkmfrjnb")))
	builder.Add("test2", bytes.NewReader([]byte("idunvovkjnreovmsdvwrvnervnreegihjbrqlkmfrjnb")))

	if len(builder.files) != 2 {
		t.Error("incorrect number of files present")
	}

	var data []byte
	buf := bytes.NewBuffer(data)
	if written, err := builder.WriteTo(buf); err != nil {
		t.Error(err)
	} else {
		t.Logf("written %d", written)
	}
}
