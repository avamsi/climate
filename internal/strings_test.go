package internal

import "testing"

func TestNormalizeToKebabCase(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{
			in:   "",
			want: "",
		},
		{
			in:   "quick",
			want: "quick",
		},
		{
			in:   "quick-brown-fox",
			want: "quick-brown-fox",
		},
		{
			in:   "quickBrownFox",
			want: "quick-brown-fox",
		},
		{
			in:   "QuickBrownFox",
			want: "quick-brown-fox",
		},
		{
			in:   "quick_brown_fox",
			want: "quick-brown-fox",
		},
		{
			in:   "QUICK_BROWN_FOX",
			want: "quick-brown-fox",
		},
		{
			in:   "qu42ck",
			want: "qu42ck",
		},
		{
			in:   "Quick42Brown",
			want: "quick42-brown",
		},
		{
			in:   "quickBrownFOX42",
			want: "quick-brown-fox42",
		},
		{
			in:   "q̀úîc̃k̄",
			want: "quick",
		},
		{
			in:   "--quick",
			want: "quick",
		},
		{
			in:   "q̀úîβc̃k̄BrownFOX_JUMPSOver-the",
			want: "qui-ck-brown-fox-jumps-over-the",
		},
	}
	for _, tc := range tests {
		if got := NormalizeToKebabCase(tc.in); got != tc.want {
			t.Errorf("normalizeToKebabCase(%#v) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}
