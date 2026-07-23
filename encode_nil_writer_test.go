package msgpack

import "testing"

func TestEncodeNilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked: %v", r)
		}
	}()
	err := NewEncoder(nil).Encode(1)
	if err == nil {
		t.Fatal("want error")
	}
}
