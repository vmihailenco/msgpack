package msgpack

// Security regression tests for the unbounded-allocation DoS (CWE-789/CWE-770).
//
// A MessagePack array/map header declares an element count in as few as 5 bytes
// (e.g. array32 = 0xdd + 4-byte big-endian count, up to ~4.29 billion). The
// decoder must NOT size an allocation directly from that untrusted count.
//
// Before the fix, decodeSliceValue used `d.flags&disableAllocLimitFlag != 1`.
// disableAllocLimitFlag == 8, so `flags & 8` is only ever 0 or 8 and never 1,
// making the expression always true: the sliceAllocLimit cap was dead and a
// 5-byte header pre-allocated the full declared slice. The interface{} path
// (decodeSlice) and DecodeMap/DecodeUntypedMap had no cap at all.
//
// Run:  go test -run 'TestSecAlloc' -v ./
// To watch these fail on the unpatched code, revert the edits in
// decode_slice.go / decode_map.go and re-run.

import (
	"runtime"
	"testing"
)

// secArray32Header builds a 5-byte array32 header declaring n elements, with NO
// element bytes following it.
func secArray32Header(n uint32) []byte {
	return []byte{0xdd, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}

// secMap32Header builds a 5-byte map32 header declaring n entries, with no body.
func secMap32Header(n uint32) []byte {
	return []byte{0xdf, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}

func secMeasureAlloc(fn func()) uint64 {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	fn()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// A correctly-bounded decoder allocates around the library's own caps
// (sliceAllocLimit / maxMapSize = 1e6 elements), far below the declared 10M.
const secAllocCeiling = 32 << 20 // 32 MiB

const secDeclared = 10_000_000 // declared by a 5-byte header

// Typed-slice path (decodeSliceValue). Vulnerable: ~80 MiB for []int64.
func TestSecAllocBounded_TypedSlice(t *testing.T) {
	data := secArray32Header(secDeclared)
	var dst []int64

	got := secMeasureAlloc(func() { _ = Unmarshal(data, &dst) })

	t.Logf("declared=%d elems, allocated=%.1f MiB (cap=%d elems)",
		secDeclared, float64(got)/(1<<20), int(sliceAllocLimit))
	if got > secAllocCeiling {
		t.Fatalf("typed slice allocation not bounded: a %d-byte header allocated %d bytes (> %d). "+
			"sliceAllocLimit bypassed (CWE-789).", len(data), got, secAllocCeiling)
	}
}

// Interface{} path (decodeSlice). Vulnerable: ~160 MiB for []interface{}.
func TestSecAllocBounded_InterfaceSlice(t *testing.T) {
	data := secArray32Header(secDeclared)
	var dst interface{}

	got := secMeasureAlloc(func() { _ = Unmarshal(data, &dst) })

	t.Logf("declared=%d elems, allocated=%.1f MiB", secDeclared, float64(got)/(1<<20))
	if got > secAllocCeiling {
		t.Fatalf("interface{} allocation not bounded: allocated %d bytes (> %d)", got, secAllocCeiling)
	}
}

// Map path (DecodeMap). Bounded by maxMapSize after the fix. The ceiling here is
// higher because Go pre-sizes map buckets (~70 MiB at the 1e6 cap), but that is
// still an order of magnitude below the unbounded ~700 MiB for 10M entries.
func TestSecAllocBounded_Map(t *testing.T) {
	data := secMap32Header(secDeclared)
	var dst map[string]interface{}

	got := secMeasureAlloc(func() { _ = Unmarshal(data, &dst) })

	const mapCeiling = 150 << 20 // 150 MiB
	t.Logf("declared=%d entries, allocated=%.1f MiB (cap=%d)",
		secDeclared, float64(got)/(1<<20), int(maxMapSize))
	if got > mapCeiling {
		t.Fatalf("map allocation not bounded: allocated %d bytes (> %d)", got, mapCeiling)
	}
}

// Regression: an HONEST slice larger than the cap must still round-trip fully.
// The cap limits only the initial allocation; growth as real elements arrive
// must still produce every element.
func TestSecHonestLargeSliceRoundTrips(t *testing.T) {
	const n = 1_500_000 // > sliceAllocLimit (1e6)

	in := make([]int64, n)
	for i := range in {
		in[i] = int64(i)
	}

	b, err := Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out []int64
	if err := Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(out) != n {
		t.Fatalf("length mismatch: got %d, want %d", len(out), n)
	}
	if out[0] != 0 || out[n-1] != int64(n-1) {
		t.Fatalf("value mismatch: out[0]=%d out[%d]=%d", out[0], n-1, out[n-1])
	}
}
