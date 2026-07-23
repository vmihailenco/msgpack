package msgpack

import "testing"

func TestDecodeNilReader(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked: %v", r)
		}
	}()
	var v int
	err := NewDecoder(nil).Decode(&v)
	if err == nil {
		t.Fatal("want error")
	}
}
