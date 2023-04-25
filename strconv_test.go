package climate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSliceParser(t *testing.T) {
	{
		var (
			in   = "true"
			want = []bool{true}
		)
		if got := sliceParser(parseBool)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseBool)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "false,true" // no space
			want = []bool{false, true}
		)
		if got := sliceParser(parseBool)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseBool)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "1, 2, 3, 4, 5" // with space
			want = []int64{1, 2, 3, 4, 5}
		)
		if got := sliceParser(parseInt64)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseInt64)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "4398046511104" // 2^42
			want = []int64{4398046511104}
		)
		if got := sliceParser(parseInt64)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseInt64)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "18446744073709551615" // 2^64 - 1
			want = []uint64{18446744073709551615}
		)
		if got := sliceParser(parseUint64)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseUint64)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "3.14"
			want = []float64{3.14}
		)
		if got := sliceParser(parseFloat64)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseFloat64)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "1.7976931348623157e+308"
			want = []float64{1.7976931348623157e+308}
		)
		if got := sliceParser(parseFloat64)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseFloat64)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "a,b,c," // trailing comma
			want = []string{"a", "b", "c"}
		)
		if got := sliceParser(parseString)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseString)(%#v) = %#v, want %#v", in, got, want)
		}
	}
	{
		var (
			in   = "a,b,c,\"d,e\"," // "d,e" is quoted
			want = []string{"a", "b", "c", "d,e"}
		)
		if got := sliceParser(parseString)(in); !cmp.Equal(got, want) {
			t.Errorf("sliceParser(parseString)(%#v) = %#v, want %#v", in, got, want)
		}
	}
}
