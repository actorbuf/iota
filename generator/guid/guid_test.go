package guid

import (
	"fmt"
	"testing"
)

func TestGuid(t *testing.T) {
	factory := NewFactory()
	id, err := factory.Get()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(id)
	id, err = factory.Get()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(id)
}
