package admiralproto

import (
	"bytes"
	"io"
	"testing"
)

func TestBasicReadWriteNetwork(t *testing.T) {
	ep := NewExampleProtocol("tcp", "www.google.com:80")
	io.WriteString(ep, "GET / \n")

	bb := &bytes.Buffer{}

	if _, err := io.CopyN(bb, ep, int64(100)); err != nil && err != io.EOF {
		t.Error("Unable to copy buffer", err)
	}
	ep.Close()

}

func TestBadNetworkReadWriteNetwork(t *testing.T) {
	if foo := NewExampleProtocol("tcp", "fdsjflksdjflks:90"); foo != nil {
		t.Error("Foo is not equal to nil")
	}
}
