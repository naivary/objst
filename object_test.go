package objst

import (
	"bytes"
	"errors"
	"os"
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

func TestWriteTo(t *testing.T) {
	o1 := tEnv.obj()
	o2 := tEnv.emptyObj()

	if _, err := o1.WriteTo(o2); err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(o1.Payload(), o2.Payload()) {
		t.Fatalf("payload should be equal. Got: %s. Expected: %s", o2.Payload(), o1.Payload())
	}
}

func TestNamePattern(t *testing.T) {
	o1 := tEnv.obj()
	o1.meta.set(MetaKeyName, "invalid#name")
	if err := o1.isValid(); !errors.Is(err, ErrInvalidNamePattern) {
		t.Fatalf("the name '%s' should not be valid.", o1.Name())
	}
	o1.meta.set(MetaKeyName, "valid/name/musti.jpg")
	if err := o1.isValid(); err != nil {
		t.Fatalf("the name %s should be valid.", o1.Name())
	}
}

func TestWriteLargeFile(t *testing.T) {
	o1 := tEnv.emptyObj()
	image, err := os.ReadFile("./testdata/images/large.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := o1.Write(image); err != nil {
		t.Error(err)
		return
	}
}

func BenchmarkWriteLargeFile(b *testing.B) {
	image, err := os.ReadFile("./testdata/images/large.jpg")
	if err != nil {
		b.Error(err)
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o := tEnv.emptyObj()
		if _, err := o.Write(image); err != nil {
			b.Error(err)
			return
		}
	}
	b.ReportAllocs()
}
