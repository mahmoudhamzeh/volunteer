package lock

import (
	"errors"
	"testing"
)

func TestIsReadOnly(t *testing.T) {
	if !isReadOnly(errors.New("READONLY You can't write against a read only replica.")) {
		t.Fatal("want readonly")
	}
	if isReadOnly(errors.New("connection refused")) {
		t.Fatal("connection error is not readonly")
	}
}
