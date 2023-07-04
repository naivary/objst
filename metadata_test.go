package objst

import (
	"fmt"
	"testing"
)

func TestMetadataMarshal(t *testing.T) {
	m := NewMetadata()
	data, err := m.Marshal()
	fmt.Println(data, err)
}
