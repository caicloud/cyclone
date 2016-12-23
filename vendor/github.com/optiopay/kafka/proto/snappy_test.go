package proto

import (
	"bytes"
	"testing"
)

var snappyChunk = []byte("\x03\x08foo") // snappy encoding of "foo"

func TestSnappyDecodeNormal(t *testing.T) {
	got, err := snappyDecode(snappyChunk)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte("foo"); !bytes.Equal(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}

func TestSnappyDecodeJava(t *testing.T) {
	javafied := []byte{
		0x82, 'S', 'N', 'A', 'P', 'P', 'Y', 0x0, // magic
		0, 0, 0, 1, // version
		0, 0, 0, 1, // compatible version
		0, 0, 0, 5, // chunk size
		0x3, 0x8, 'f', 'o', 'o', // chunk data
		0, 0, 0, 5, // chunk size
		0x3, 0x8, 'f', 'o', 'o', // chunk data
	}
	got, err := snappyDecode(javafied)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte("foofoo"); !bytes.Equal(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}
