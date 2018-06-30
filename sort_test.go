package msgpack

import (
	"bytes"
	"strconv"
	"testing"
)

func TestSort(t *testing.T) {
	tests := []struct {
		in  [][]byte
		out [][]byte
	}{
		{
			in:  nil,
			out: nil,
		},
		{
			in: [][]byte{
				[]byte{},
			},
			out: [][]byte{
				[]byte{},
			},
		},
		{
			in: [][]byte{
				[]byte{},
				[]byte{},
			},
			out: [][]byte{
				[]byte{},
				[]byte{},
			},
		},
		{
			in: [][]byte{
				[]byte{1},
				[]byte{1},
			},
			out: [][]byte{
				[]byte{1},
				[]byte{1},
			},
		},
		{
			in: [][]byte{
				[]byte{1, 1},
				[]byte{1},
			},
			out: [][]byte{
				[]byte{1},
				[]byte{1, 1},
			},
		},
		{
			in: [][]byte{
				[]byte{1},
				[]byte{1, 1},
			},
			out: [][]byte{
				[]byte{1},
				[]byte{1, 1},
			},
		},
		{
			in: [][]byte{
				[]byte{1, 2},
				[]byte{1, 1},
			},
			out: [][]byte{
				[]byte{1, 1},
				[]byte{1, 2},
			},
		},
		{
			in: [][]byte{
				[]byte{1, 1},
				[]byte{1, 2},
			},
			out: [][]byte{
				[]byte{1, 1},
				[]byte{1, 2},
			},
		},
		{
			in: [][]byte{
				[]byte{1, 1, 1},
				[]byte{1, 2, 3},
				[]byte{1, 1, 3},
			},
			out: [][]byte{
				[]byte{1, 1, 1},
				[]byte{1, 1, 3},
				[]byte{1, 2, 3},
			},
		},
		{
			in: [][]byte{
				[]byte{1, 1, 1},
				[]byte{1, 2, 3},
				[]byte{1, 1, 3},
				[]byte{3, 1, 3},
				[]byte{4, 1, 3},
				[]byte{9, 1, 3},
				[]byte{6, 1, 3},
				[]byte{2, 1, 3},
				[]byte{7, 1, 3},
				[]byte{1, 4, 3},
				[]byte{1, 78, 3},
				[]byte{1, 32, 3},
				[]byte{1, 14, 3},
			},
			out: [][]byte{
				[]byte{1, 1, 1},
				[]byte{1, 1, 3},
				[]byte{1, 2, 3},
				[]byte{1, 4, 3},
				[]byte{1, 14, 3},
				[]byte{1, 32, 3},
				[]byte{1, 78, 3},
				[]byte{2, 1, 3},
				[]byte{3, 1, 3},
				[]byte{4, 1, 3},
				[]byte{6, 1, 3},
				[]byte{7, 1, 3},
				[]byte{9, 1, 3},
			},
		},
	}

	for i, tcase := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			sortBytesArray(tcase.in)
			if len(tcase.out) != len(tcase.in) {
				t.Fatalf("Error sorting. Invalid length (%d vs %d", len(tcase.out), len(tcase.in))
			}
			for i := range tcase.in {
				if bytes.Compare(tcase.in[i], tcase.out[i]) != 0 {
					t.Fatalf("Error sorting. Different data at pos %d", i)
				}
			}
		})
	}
}

var unsorted = [][]byte{
	[]byte{24, 3, 4, 2, 3},
	[]byte{1, 2, 3, 4, 5},
	[]byte{2, 5, 6, 2, 3},
	[]byte{8, 4, 5, 7, 8},
	[]byte{8, 4, 5, 7, 7},
	[]byte{8, 4, 5, 7, 9},
	[]byte{7, 4, 6, 7, 8, 9, 24, 35},
	[]byte{7, 4, 6, 7, 8, 9, 24, 36},
	[]byte{7, 4, 6, 7, 8, 9, 24, 37},
	[]byte{7, 4, 6, 7, 8, 9, 24, 39},
	[]byte{1, 2, 3},
	[]byte{1, 2, 3, 89},
}

func BenchmarkMarshalSort(b *testing.B) {
	data := make([][]byte, len(unsorted))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		copy(data, unsorted)
		sortBytesArray(data)
	}
}
