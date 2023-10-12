package float128

import "github.com/shogo82148/int128"

func (x Float128) Sqrt() Float128 {
	// special cases
	switch {
	case x.IsNaN() || x.IsInf(1) || x.isZero():
		return x
	case x.h&signMask128H != 0:
		return NaN()
	}

	// normalize x
	_, exp, frac := x.split()

	if exp%2 != 0 {
		frac = frac.Lsh(1)
	}
	// exponent of sqrt(x)
	exp >>= 1

	// generate sqrt(frac) bit by bit
	frac = frac.Lsh(1)
	var q, s int128.Uint128
	r := int128.Uint128{H: 1 << (shift128 - 64 + 1)}
	for r.H != 0 || r.L != 0 {
		t := s.Add(r)
		if t.Cmp(frac) <= 0 {
			s = t.Add(r)
			frac = frac.Sub(t)
			q = q.Add(r)
		}
		frac = frac.Lsh(1)
		r = r.Rsh(1)
	}

	// final rounding
	if frac.H != 0 || frac.L != 0 {
		q = q.Add(int128.Uint128{H: 0, L: 1})
	}
	q = q.Rsh(1)
	q = q.Add(int128.Uint128{H: uint64(exp-1+bias128) << (shift128 - 64)})
	return Float128{q.H, q.L}
}
