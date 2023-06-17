package main

import (
	"bytes"
	"testing"

	"github.com/naivary/objst/random"
)

func TestWrite(t *testing.T) {
	o := tEnv.emptyObj()
	pl := random.String(10)
	if _, err := o.Write([]byte(pl)); err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal([]byte(pl), o.pl.Bytes()) {
		t.Fatalf("bytes are not equal. Got: %s. Expected: %s", o.pl.String(), pl)
	}
}

func TestRead(t *testing.T) {
	o := tEnv.emptyObj()
	pl := random.String(10)
	if _, err := o.Write([]byte(pl)); err != nil {
		t.Error(err)
		return
	}
	d := make([]byte, len(pl))
	if _, err := o.Read(d); err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal([]byte(pl), d) {
		t.Fatalf("payload isn't the same. Got: %s. Expected: %s", string(d), pl)
	}
}
