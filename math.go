package float128

import (
	"math/bits"

	"github.com/shogo82148/int128"
)

type uint256 struct {
	a, b, c, d uint64
}

func mul128(x, y int128.Uint128) uint256 {
	h1, l1 := bits.Mul64(x.L, y.L)
	h2, l2 := bits.Mul64(x.H, y.L)
	h3, l3 := bits.Mul64(x.L, y.H)
	h4, l4 := bits.Mul64(x.H, y.H)

	//             x.H  x.L
	//             y.H  y.L
	//        -------------
	//              h1  l1
	//          h2  l2
	//          h3  l3
	//      h4  l4
	//---------------------
	//       a   b   c   d

	d := l1
	c := h1
	b := l4
	a := h4

	var carry uint64
	c, carry = bits.Add64(c, l2, 0)
	b, carry = bits.Add64(b, h2, carry)
	a, _ = bits.Add64(a, 0, carry)
	c, carry = bits.Add64(c, l3, 0)
	b, carry = bits.Add64(b, h3, carry)
	a, _ = bits.Add64(a, 0, carry)
	return uint256{a, b, c, d}
}

func squash(x uint64) uint64 {
	if x == 0 {
		return 0
	}
	return 1
}

func (f Float128) split() (sign uint64, exp int32, frac int128.Uint128) {
	sign = f.h & signMask128H
	exp = int32((f.h>>(shift128-64))&mask128) - bias128
	if exp == -bias128 {
		frac = int128.Uint128{H: f.h & fracMask128H, L: f.l}
		l := frac.Len()
		frac = frac.Lsh(uint(shift128-l) + 1)
		exp = int32(l) - (bias128 + shift128)
	} else {
		frac = int128.Uint128{H: (f.h & fracMask128H) | (1 << (shift128 - 64)), L: f.l}
	}
	return
}

func (a Float128) Mul(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		return propagateNaN(a, b)
	}

	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()

	sign := signA ^ signB

	// handle special cases
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±Inf
		if b.isZero() {
			// ±Inf * ±0 = NaN
			return nan
		} else {
			// b is finite, the result is ±Inf
			return Float128{sign | inf.h, inf.l}
		}
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±Inf
		if a.isZero() {
			// ±0 * ±Inf = NaN
			return nan
		} else {
			// a is finite, the result is ±Inf
			return Float128{sign | inf.h, inf.l}
		}
	}

	// add guard and round bits
	fracA = fracA.Lsh(9)
	fracB = fracB.Lsh(9)

	exp := expA + expB
	frac256 := mul128(fracA, fracB)
	frac := int128.Uint128{H: frac256.a, L: frac256.b | squash(frac256.c) | squash(frac256.d)}
	shift := int32(frac.Len()) - (shift128 + 1 + 2)
	exp += shift
	if exp < -(bias128 + shift128) {
		// underflow
		return Float128{sign, 0}
	} else if exp <= -bias128 {
		// the result is subnormal
		shift = 1 - (expA + expB + bias128)
		offset := one.Lsh(uint(shift) + 1).Sub(one).Add(frac.Rsh(uint(shift) + 2).And(one))
		frac = frac.Add(offset) // round to nearest even
		frac = frac.Rsh(uint(shift) + 2)
		return Float128{sign | (frac.H & fracMask128H), frac.L}
	}

	frac = frac.Add(int128.Uint128{H: 0, L: (1<<(shift+1) - 1) + (frac.L>>(shift+2))&1}) // round to nearest even
	shift = int32(frac.Len()) - (shift128 + 1 + 2)
	exp = expA + expB + shift

	if exp >= mask128-bias128 {
		// overflow
		return Float128{sign | inf.h, inf.l}
	}

	frac = frac.Rsh(uint(shift) + 2)
	exp += bias128
	return Float128{sign | uint64(exp)<<(shift128-64) | (frac.H & fracMask128H), frac.L}
}

func (f Float128) isZero() bool {
	return (f.h&^signMask128H | f.l) == 0
}

func (a Float128) Add(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		return propagateNaN(a, b)
	}
	if ((a.h &^ signMask128H) | a.l) == 0 {
		// ±0 + b = b
		return b
	}
	if ((b.h &^ signMask128H) | b.l) == 0 {
		// a + ±0 = a
		return a
	}

	signA := a.h & signMask128H
	expA := int((a.h>>(shift128-64))&mask128) - bias128
	signB := b.h & signMask128H
	expB := int((b.h>>(shift128-64))&mask128) - bias128

	if signA == signB {
		return add(signA, expA, expB, a, b)
	}

	absA := a.Abs()
	absB := b.Abs()
	switch absA.Compare(absB) {
	case 0:
		return Float128{0, 0}
	case 1:
		return sub(signA, expA, expB, absA, absB)
	case -1:
		return sub(signB, expB, expA, absB, absA)
	}
	panic("never reach here")
}

// add returns a + b.
// a and b must have same sign.
func add(sign uint64, expA, expB int, a, b Float128) Float128 {
	if a.Compare(b) < 0 {
		expA, expB = expB, expA
		a, b = b, a
	}

	var fracA int128.Uint128
	if expA == -bias128 {
		fracA = int128.Uint128{H: a.h & fracMask128H, L: a.l}
		l := fracA.Len()
		fracA = fracA.Lsh(uint(shift128-l) + 1)
		expA = l - (bias128 + shift128)
	} else {
		fracA = int128.Uint128{H: (a.h & fracMask128H) | (1 << (shift128 - 64)), L: a.l}
	}
	fracA = fracA.Lsh(2) // add guard and round bits

	var fracB int128.Uint128
	if expB == -bias128 {
		fracB = int128.Uint128{H: b.h & fracMask128H, L: b.l}
		l := fracB.Len()
		fracB = fracB.Lsh(uint(shift128-l) + 1)
		expB = l - (bias128 + shift128)
	} else {
		fracB = int128.Uint128{H: (b.h & fracMask128H) | (1 << (shift128 - 64)), L: b.l}
	}
	fracB = fracB.Lsh(2) // add guard and round bits

	fracB = fracB.Add(one.Lsh(uint(expA - expB)).Sub(one))
	fracB = fracB.Rsh(uint(expA - expB))
	frac := fracA.Add(fracB)
	exp := expA

	shift := frac.Len() - (shift128 + 1 + 2)
	frac = frac.Add(int128.Uint128{H: 0, L: (1<<(shift+1) - 1) + (frac.L>>(shift+2))&1}) // round to nearest even
	shift = frac.Len() - (shift128 + 1 + 2)
	frac = frac.Rsh(uint(shift) + 2)
	exp += shift

	if exp >= mask128-bias128 {
		// overflow
		return Float128{sign | inf.h, inf.l}
	}
	if exp <= -bias128 {
		// the result is subnormal
		shift = -(exp + bias128) - 1
		offset := one.Lsh(uint(shift) + 1).Sub(one).Add(frac.Rsh(uint(shift) + 2).And(one))
		frac = frac.Add(offset) // round to nearest even
		frac = frac.Rsh(uint(shift) + 2)
		return Float128{sign | (frac.H & fracMask128H), frac.L}
	}

	exp += bias128
	return Float128{sign | uint64(exp<<(shift128-64)) | (frac.H & fracMask128H), frac.L}
}

// sub returns a - b and set sign to the sign of the result.
// a - b be must be positive.
func sub(sign uint64, expA, expB int, a, b Float128) Float128 {
	var fracA int128.Uint128
	if expA == -bias128 {
		fracA = int128.Uint128{H: a.h & fracMask128H, L: a.l}
		l := fracA.Len()
		fracA = fracA.Lsh(uint(shift128-l) + 1)
		expA = l - (bias128 + shift128)
	} else {
		fracA = int128.Uint128{H: (a.h & fracMask128H) | (1 << (shift128 - 64)), L: a.l}
	}
	fracA = fracA.Lsh(2) // add guard and round bits

	var fracB int128.Uint128
	if expB == -bias128 {
		fracB = int128.Uint128{H: b.h & fracMask128H, L: b.l}
		l := fracB.Len()
		fracB = fracB.Lsh(uint(shift128-l) + 1)
		expB = l - (bias128 + shift128)
	} else {
		fracB = int128.Uint128{H: (b.h & fracMask128H) | (1 << (shift128 - 64)), L: b.l}
	}
	fracB = fracB.Lsh(2) // add guard and round bits

	fracB = fracB.Add(one.Lsh(uint(expA - expB)).Sub(one))
	fracB = fracB.Rsh(uint(expA - expB))
	frac := fracA.Sub(fracB)
	exp := expA

	shift := frac.Len() - (shift128 + 1 + 2)
	frac = frac.Add(int128.Uint128{H: 0, L: (1<<(shift+1) - 1) + (frac.L>>(shift+2))&1}) // round to nearest even
	shift = frac.Len() - (shift128 + 1 + 2)
	frac = frac.Rsh(uint(shift) + 2)
	exp += shift

	if exp >= mask128-bias128 {
		// overflow
		return Float128{sign | inf.h, inf.l}
	}
	if exp <= -bias128 {
		// the result is subnormal
		shift = -(exp + bias128) - 1
		offset := one.Lsh(uint(shift) + 1).Sub(one).Add(frac.Rsh(uint(shift) + 2).And(one))
		frac = frac.Add(offset) // round to nearest even
		frac = frac.Rsh(uint(shift) + 2)
		return Float128{sign | (frac.H & fracMask128H), frac.L}
	}

	exp += bias128
	return Float128{sign | uint64(exp<<(shift128-64)) | (frac.H & fracMask128H), frac.L}
}

func (a Float128) Sub(b Float128) Float128 {
	return a.Add(b.Neg())
}

// Compare compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//	+1 if x >  y
//
// a NaN is considered less than any non-NaN, and two NaNs are equal.
func (a Float128) Compare(b Float128) int {
	aNaN := a.IsNaN()
	bNaN := b.IsNaN()
	if aNaN && bNaN {
		return 0
	}
	if aNaN {
		return -1
	}
	if bNaN {
		return 1
	}

	ia := int128.Int128{H: int64(a.h), L: a.l}
	ia = ia.Xor(int128.Int128{H: int64(a.h) >> 63 & 0x7fff_ffff_ffff_ffff, L: uint64(int64(a.h) >> 63)})
	ia = ia.Add(int128.Int128{L: a.h >> 63})
	ib := int128.Int128{H: int64(b.h), L: b.l}
	ib = ib.Xor(int128.Int128{H: int64(b.h) >> 63 & 0x7fff_ffff_ffff_ffff, L: uint64(int64(b.h) >> 63)})
	ib = ib.Add(int128.Int128{L: b.h >> 63})
	return ia.Cmp(ib)
}

// Abs returns the absolute value of f.
func (f Float128) Abs() Float128 {
	return Float128{f.h &^ signMask128H, f.l}
}

// Neg returns the negated value of f.
func (f Float128) Neg() Float128 {
	return Float128{f.h ^ signMask128H, f.l}
}
