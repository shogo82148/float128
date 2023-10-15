package float128

import "testing"

func TestText(t *testing.T) {
	tests := []struct {
		x    Float128
		fmt  byte
		prec int
		s    string
	}{
		/****** binary format ******/
		{FromFloat64(0), 'b', 0, "0p-16494"},
		{FromFloat64(1), 'b', 0, "5192296858534827628530496329220096p-112"},

		/****** hexadecimal format ******/
		{FromFloat64(0), 'x', -1, "0x0p+00"},
		{FromFloat64(1), 'x', -1, "0x1p+00"},
		{FromFloat64(0.1), 'x', -1, "0x1.999999999999ap-04"},
		{FromFloat64(1).Quo(FromFloat64(10)), 'x', -1, "0x1.999999999999999999999999999ap-04"},

		{FromFloat64(0.1), 'x', 0, "0x1p-03"},
		{FromFloat64(0.1), 'x', 2, "0x1.9ap-04"},
		{FromFloat64(1).Quo(FromFloat64(10)), 'x', 28, "0x1.999999999999999999999999999ap-04"},
		{FromFloat64(1).Quo(FromFloat64(10)), 'x', 30, "0x1.999999999999999999999999999a00p-04"},
		{Float128{0, 1}, 'x', 1, "0x1.0p-16494"},
	}

	for _, tt := range tests {
		got := tt.x.Text(tt.fmt, tt.prec)
		if got != tt.s {
			t.Errorf("%#v: expected %s, got %s", tt, tt.s, got)
		}
	}
}
