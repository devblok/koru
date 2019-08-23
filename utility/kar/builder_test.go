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

	builder.Add("test", []byte("idunvovkjnreovmegihjbrqlkmfrjnb"))
	builder.Add("test2", []byte("idunvovkjnreovmsdvwrvnervnreegihjbrqlkmfrjnb"))

	if len(builder.files) != 2 {
		t.Error("incorrect number of files present")
	}

	data := make([]byte, 5*1024)
	buf := bytes.NewBuffer(data)
	num, err := builder.WriteTo(buf)
	if err != nil {
		t.Error(err)
	}
	t.Logf("written %d \n", num)
}
