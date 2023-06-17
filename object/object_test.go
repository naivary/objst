package object

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/naivary/objst/random"
	"golang.org/x/exp/slog"
)

var (
	tOwner string
	tName  string
)

func setup() error {
	tOwner = uuid.NewString()
	tName = fmt.Sprintf("obj_name_%s", tOwner)
	return nil
}

func destroy() error {
	return nil
}

func TestMain(t *testing.M) {
	if err := setup(); err != nil {
		slog.Error("something went wrong while setting up the test", slog.String("msg", err.Error()))
		return
	}
	code := t.Run()
	if err := destroy(); err != nil {
		slog.Error("something went wrong while setting up the test", slog.String("msg", err.Error()))
		return
	}
	os.Exit(code)
}

func TestWrite(t *testing.T) {
	o := New(tName, tOwner)
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
	o := New(tName, tOwner)
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


