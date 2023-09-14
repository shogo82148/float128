package float128

import (
	"github.com/shogo82148/int128"
)

func squash64(x uint64) uint64 {
	if x == 0 {
		return 0
	}
	return 1
}

func squash128(x int128.Uint128) uint64 {
	return squash64(x.H | x.L)
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
	if a.isZero() || b.isZero() {
		// ±0 * b = ±0
		// a * ±0 = ±0
		return Float128{sign, 0}
	}

	// add guard and round bits
	fracA = fracA.Lsh(9)
	fracB = fracB.Lsh(9)

	exp := expA + expB
	frac256 := mul128(fracA, fracB)
	frac := int128.Uint128{H: frac256.a, L: frac256.b | squash64(frac256.c) | squash64(frac256.d)}
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
		return Float128{sign | frac.H, frac.L}
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

func (a Float128) Quo(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		return propagateNaN(a, b)
	}
	if b.isZero() {
		if a.isZero() {
			return nan
		}
		// ±a / ±0 = ±Inf
		return Float128{((a.h ^ b.h) & signMask128H) | inf.h, inf.l}
	}
	if a.isZero() {
		// ±0 / b = ±0
		return Float128{((a.h ^ b.h) & signMask128H), 0}
	}

	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()

	if expA == mask128-bias128 {
		// NaN check is done above; a is ±Inf
		if expB == mask128-bias128 {
			// +Inf / ±Inf = NaN
			// -Inf / ±Inf = NaN
			return nan
		} else {
			// ±Inf / finite = ±Inf
			return Float128{a.h ^ signB, a.l}
		}
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±Inf
		// NaN and Inf checks are done above; a is finite
		// finite / ±Inf = ±0
		return Float128{signA ^ signB, 0}
	}

	sign := signA ^ signB
	exp := expA - expB
	if fracA.Cmp(fracB) < 0 {
		exp--
		fracA = fracA.Lsh(1)
	}

	fracA256 := uint256{a: fracA.H, b: fracA.L, c: 0, d: 0}
	frac256, mod := fracA256.divMod128(fracB)
	frac256.d |= squash128(mod)

	// log.Printf("fracA256: %#v", fracA256)
	// log.Printf("   fracB: %#v", fracB)
	// log.Printf(" frac256: %#v", frac256)
	// log.Printf("     mod: %#v", mod)

	if exp < -(bias128 + shift128) {
		// underflow
		return Float128{sign, 0}
	} else if exp <= -bias128 {
		// the result is subnormal
		shift := uint(1 - (exp + bias128))
		ff := uint256{d: 1}.lsh(uint(shift) + 16 - 1).sub(uint256{d: 1})
		ff = ff.add(frac256.rsh(uint(shift) + 16).and(uint256{d: 1}))
		frac256 = frac256.add(ff) // round to nearest even
		frac256 = frac256.rsh(uint(shift) + 16)
		// log.Printf("%#v", frac256)
		return Float128{sign | frac256.c, frac256.d}
	} else if exp >= mask128-bias128 {
		// overflow
		return Float128{sign | inf.h, inf.l}
	}

	frac256 = frac256.add(uint256{d: 0x7fff}).add(frac256.rsh(16).and(uint256{d: 1})) // round to nearest even
	frac256 = frac256.rsh(16)
	// log.Printf("%#v", frac256)
	exp += bias128
	return Float128{sign | uint64(exp)<<(shift128-64) | frac256.c&fracMask128H, frac256.d}
}

func (a Float128) Add(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		return propagateNaN(a, b)
	}
	if a.isZero() {
		if b.isZero() {
			return Float128{a.h & b.h, 0}
		}
		// ±0 + b = b
		return b
	}
	if b.isZero() {
		// a + ±0 = a
		return a
	}

	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()

	// handle special cases
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±Inf
		if expB == mask128-bias128 {
			// NaN check is done above; b is ±Inf
			if signA == signB {
				// ±Inf + ±Inf = ±Inf
				return Float128{signA | inf.h, inf.l}
			} else {
				// ±Inf + ∓Inf = NaN
				return nan
			}
		} else {
			// b is finite, the result is ±Inf
			return a
		}
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±Inf
		// NaN and Inf checks are done above; a is finite
		return b
	}

	if expA < expB {
		signA, signB = signB, signA
		expA, expB = expB, expA
		fracA, fracB = fracB, fracA
	}
	exp := expA

	// align the fraction
	const margin = 4
	fracA256 := uint256{a: fracA.H, b: fracA.L, c: 0, d: 0}.rsh(margin)
	fracB256 := uint256{a: fracB.H, b: fracB.L, c: 0, d: 0}.rsh(margin)
	fracB256 = fracB256.rsh(uint(expA - expB))

	// log.Printf("fracA256 = %#v", fracA256)
	// log.Printf("fracB256 = %#v", fracB256)

	// do addition
	ifracA256 := fracA256.int256().setSign(signA != 0)
	ifracB256 := fracB256.int256().setSign(signB != 0)
	ifrac256 := ifracA256.add(ifracB256)

	// split into sign and absolute value
	sign := uint64(ifrac256.a) & signMask128H
	frac256 := ifrac256.abs()

	// normalize
	var shift int32
	shift = int32(frac256.leadingZeros() - 19)

	if frac256.isZero() || exp-shift < -(bias128+shift128) {
		// underflow
		return Float128{sign, 0}
	} else if exp-shift <= -bias128 {
		// the result is subnormal
		shift = 1 - (exp + bias128)
		// offset := one.Lsh(uint(shift) + 1).Sub(one).Add(frac256.rsh(uint(shift) + 2).and(one))
		// frac256 = frac256.add(offset) // round to nearest even
		// log.Printf(" frac256 = %#v >> %d", frac256, shift-margin+64)
		frac256 = frac256.rsh(uint(shift - margin + 64))
		// log.Printf(" frac256 = %#v", frac256)
		return Float128{sign | frac256.b, frac256.c}
	}

	one := uint256{a: 0, b: 0, c: 0, d: 1}
	ff := uint256{0, 0, 0x7fff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}
	// log.Printf("           %#v", frac256.rsh(128-uint(shift+margin)))
	ff = ff.add(frac256.rsh(128 - uint(shift+margin)).and(one)) // round to nearest even
	frac256 = frac256.add(ff.rsh(uint(shift + margin)))
	shift = int32(frac256.leadingZeros() - 19)
	// log.Printf(" frac256 = %#v << %d", frac256, shift+4)
	frac256 = frac256.lsh(uint(shift + margin))
	// log.Printf(" frac256 = %#v", frac256)

	exp -= shift
	exp += bias128
	if exp >= mask128 {
		// overflow
		return Float128{sign | inf.h, inf.l}
	}

	return Float128{sign | uint64(exp)<<(shift128-64) | (frac256.a & fracMask128H), frac256.b}
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

	ia := a.comparable()
	ib := b.comparable()
	return ia.Cmp(ib)
}

// comparable returns a comparable format for a.
func (a Float128) comparable() int128.Int128 {
	i := int128.Int128{H: int64(a.h), L: a.l}
	i = i.Xor(int128.Int128{H: int64(a.h) >> 63 & 0x7fff_ffff_ffff_ffff, L: uint64(int64(a.h) >> 63)})
	i = i.Add(int128.Int128{L: a.h >> 63})
	return i
}

// Eq returns a == b.
func (a Float128) Eq(b Float128) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	if a == b {
		// a and b has same bit pattern
		return true
	}

	// check -0 == 0
	return (((a.h | b.h) &^ signMask128H) | (a.l | b.l)) == 0
}

// Ne returns a != b.
func (a Float128) Ne(b Float128) bool {
	return !a.Eq(b)
}

// Lt returns a < b.
func (a Float128) Lt(b Float128) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}

	ia := a.comparable()
	ib := b.comparable()
	return ia.Cmp(ib) < 0
}

// Gt returns a > b.
func (a Float128) Gt(b Float128) bool {
	return b.Lt(a)
}

// Le returns a <= b.
func (a Float128) Le(b Float128) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}

	ia := a.comparable()
	ib := b.comparable()
	return ia.Cmp(ib) <= 0
}

// Ge returns a >= b.
func (a Float128) Ge(b Float128) bool {
	return b.Le(a)
}

// Abs returns the absolute value of f.
func (f Float128) Abs() Float128 {
	return Float128{f.h &^ signMask128H, f.l}
}

// Neg returns the negated value of f.
func (f Float128) Neg() Float128 {
	return Float128{f.h ^ signMask128H, f.l}
}

// FMA returns x * y + z, computed with only one rounding.
// (That is, FMA returns the fused multiply-add of x, y, and z.)
func FMA(x, y, z Float128) Float128 {
	// handling NaN
	if x.IsNaN() || y.IsNaN() || z.IsNaN() {
		return nan
	}

	// Inf involved. At most one rounding will occur.
	if x.IsInf(0) || y.IsInf(0) {
		return x.Mul(y).Add(z)
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if z.IsInf(0) {
		return z
	}

	signA, expA, fracA := x.split()
	signB, expB, fracB := y.split()
	signC, expC, fracC := z.split()
	sign := signA ^ signB

	// add guard and round bits
	fracA = fracA.Lsh(4)
	fracB = fracB.Lsh(4)
	fracC256 := uint256{a: fracC.H, b: fracC.L, c: 0, d: 0}.rsh(8)
	// log.Printf(" fracC256 = %#v", fracC256)

	// calculate a * b
	exp := expA + expB
	frac256 := mul128(fracA, fracB)
	// log.Printf("  frac256 = %#v", frac256)
	if frac256.isZero() {
		// +0 + ±0 = +0
		// -0 + ±0 = ±0
		if z.isZero() {
			return Float128{sign & z.h, 0}
		}

		// ±0 + z = z
		return z
	}

	// add c
	if expC <= exp {
		one := uint256{a: 0, b: 0, c: 0, d: 1}
		ff := one.lsh(uint(exp - expC)).sub(one)
		// log.Printf("        ff: %#v", ff)
		ff = fracC256.and(ff)
		fracC256 = fracC256.rsh(uint(exp - expC))
		fracC256.d |= squash64(ff.a) | squash64(ff.b) | squash64(ff.c) | squash64(ff.d)
	} else {
		one := uint256{a: 0, b: 0, c: 0, d: 1}
		ff := one.lsh(uint(expC - exp)).sub(one)
		ff = frac256.and(ff)
		frac256 = frac256.rsh(uint(expC - exp))
		frac256.d |= squash64(ff.a) | squash64(ff.b) | squash64(ff.c) | squash64(ff.d)
		exp = expC
	}
	// log.Printf("  fracC256: %#v", fracC256)
	// log.Printf("+  frac256: %#v", frac256)
	ifracC256 := fracC256.int256().setSign(signC != 0)
	ifrac256 := frac256.int256().setSign(sign != 0)
	ifrac256 = ifrac256.add(ifracC256)
	sign = uint64(ifrac256.a) & signMask128H
	frac256 = ifrac256.abs()
	// log.Printf("=  frac256: %#v", frac256)

	// normalize
	// log.Println("leading zero:", frac256.leadingZeros())
	shift := int32(23 - frac256.leadingZeros())
	expTmp := exp + shift
	// log.Println("exp:", exp)
	if frac256.isZero() || expTmp < -(bias128+shift128) {
		// underflow
		return Float128{sign, 0}
	} else if expTmp <= -bias128 {
		shift := uint(128 - 7 - (expTmp - shift + bias128))
		one := uint256{a: 0, b: 0, c: 0, d: 1}
		ff := one.lsh(shift - 1).sub(one)
		ff = ff.add(frac256.rsh(shift).and(one)) // round to nearest even
		frac256 = frac256.add(ff)
		// log.Printf(" frac256 = %#v >> %d", frac256, shift)
		frac256 = frac256.rsh(shift)
		// log.Printf(" frac256 = %#v", frac256)

		return Float128{sign | frac256.c, frac256.d}
	}

	if 128-8+shift >= 0 {
		one := uint256{a: 0, b: 0, c: 0, d: 1}
		ff := one.lsh(uint(128-8+shift) - 1).sub(one)
		ff = ff.add(frac256.rsh(uint(128 - 8 + shift)).and(one)) // round to nearest even
		frac256 = frac256.add(ff)
		shift = int32(23 - frac256.leadingZeros())
		exp += shift
		frac256 = frac256.rsh(uint(128 - 8 + shift))
		// log.Printf(" frac256 = %#v", frac256)
	} else {
		shift = int32(23 - frac256.leadingZeros())
		exp += shift
		frac256 = frac256.lsh(uint(-128 + 8 - shift))
	}
	exp += bias128
	if exp >= mask128 {
		// overflow
		return Float128{sign | inf.h, inf.l}
	}
	_ = signC
	_ = expC
	// log.Println()

	return Float128{sign | uint64(exp)<<(shift128-64) | (frac256.c & fracMask128H), frac256.d}
}
