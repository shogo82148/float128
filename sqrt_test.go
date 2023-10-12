package float128

import "testing"

func TestSqrt(t *testing.T) {
	tests := []struct {
		x, want Float128
	}{
		// special cases
		{Float128{}, Float128{}},                               // zero
		{Float128{signMask128H, 0}, Float128{signMask128H, 0}}, // negative zero
		{Inf(1), Inf(1)},                                       // +Inf
		{NaN(), NaN()},

		{Float128{0x3fff000000000000, 0}, Float128{0x3fff000000000000, 0}},                  // 1
		{Float128{0x4000000000000000, 0}, Float128{0x3fff6a09e667f3bc, 0xc908b2fb1366ea95}}, // 2
		{Float128{0x4001000000000000, 0}, Float128{0x4000000000000000, 0}},                  // 4
	}

	for _, test := range tests {
		got := test.x.Sqrt()
		if got != test.want {
			t.Errorf("Sqrt(%#v) = %#v; want %#v", test.x, got, test.want)
		}
	}
}

func BenchmarkSqrt(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		x, _ := r.Float128Pair()
		x.Sqrt()
	}
}
