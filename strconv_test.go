package float128

import "testing"

func TestText(t *testing.T) {
	tests := []struct {
		x    Float128
		fmt  byte
		prec int
		s    string
	}{
		{FromFloat64(0), 'b', 0, "0p-16494"},
		{FromFloat64(1), 'b', 0, "5192296858534827628530496329220096p-112"},
	}

	for _, tt := range tests {
		got := tt.x.Text(tt.fmt, tt.prec)
		if got != tt.s {
			t.Errorf("%#v: expected %s, got %s", tt, tt.s, got)
		}
	}
}
