package msgpack_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

// zeroByMethod has a value-receiver IsZero, exercising the omitzero path that
// defers to a type's own IsZero rather than reflect's structural zero check.
type zeroByMethod struct{ n int }

func (z zeroByMethod) IsZero() bool { return z.n == 0 }

// TestOmitZero checks that the `omitzero` struct-tag option is honored, matching
// encoding/json (Go 1.24+): a field is omitted when it holds the zero value for
// its type, where a type's own IsZero method takes precedence. This is distinct
// from `omitempty`, which also drops non-nil empty slices, maps, and strings.
func TestOmitZero(t *testing.T) {
	type S struct {
		Keep     int          `msgpack:"keep"`               // no option: always encoded
		OmitInt  int          `msgpack:"omit_int,omitzero"`  // dropped when 0
		OmitStr  string       `msgpack:"omit_str,omitzero"`  // dropped when ""
		OmitBool bool         `msgpack:"omit_bool,omitzero"` // dropped when false
		OmitObj  zeroByMethod `msgpack:"omit_obj,omitzero"`  // dropped when IsZero()
		OmitPtr  *time.Time   `msgpack:"omit_ptr,omitzero"`  // dropped when nil
	}

	encodeToMap := func(t *testing.T, v any) map[string]any {
		t.Helper()
		b, err := msgpack.Marshal(v)
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, msgpack.Unmarshal(b, &m))
		return m
	}

	omitzeroKeys := []string{"omit_int", "omit_str", "omit_bool", "omit_obj", "omit_ptr"}

	// Zero value: every omitzero field is dropped; the untagged field stays.
	// The nil *time.Time also exercises the nil-pointer guard: its element type
	// has an IsZero method, which must not be invoked on the nil pointer.
	m := encodeToMap(t, S{})
	require.Contains(t, m, "keep", "field without omitzero must always be encoded")
	for _, k := range omitzeroKeys {
		require.NotContains(t, m, k, "omitzero field %q must be omitted at the zero value", k)
	}

	// Non-zero value: every omitzero field is encoded.
	now := time.Unix(1, 0)
	m = encodeToMap(t, S{
		OmitInt:  7,
		OmitStr:  "x",
		OmitBool: true,
		OmitObj:  zeroByMethod{n: 1},
		OmitPtr:  &now,
	})
	for _, k := range omitzeroKeys {
		require.Contains(t, m, k, "omitzero field %q must be encoded when non-zero", k)
	}
}
