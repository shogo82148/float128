package float128

import (
	"fmt"
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

func (x uint256) mul128(y int128.Uint128) uint256 {
	h1, l1 := bits.Mul64(x.d, y.L)
	h2, l2 := bits.Mul64(x.c, y.L)
	h3, l3 := bits.Mul64(x.b, y.L)
	_, l4 := bits.Mul64(x.a, y.L)

	h5, l5 := bits.Mul64(x.d, y.H)
	h6, l6 := bits.Mul64(x.c, y.H)
	_, l7 := bits.Mul64(x.b, y.H)

	//     x.a  x.b  x.c  x.d
	//               y.H  y.L
	//     ------------------
	//                h1   l1
	//           h2   l2
	//      h3   l3
	//      l4
	//           h5   l5
	//      h6   l6
	//      l7
	//-----------------------
	//       a    b    c    d

	a := h3
	b := h2
	c := h1
	d := l1

	var carry uint64
	c, carry = bits.Add64(c, l2, 0)
	b, carry = bits.Add64(b, l3, carry)
	a, _ = bits.Add64(a, l4, carry)

	c, carry = bits.Add64(c, l5, 0)
	b, carry = bits.Add64(b, l6, carry)
	a, _ = bits.Add64(a, l7, carry)

	b, carry = bits.Add64(b, h5, 0)
	a, _ = bits.Add64(a, h6, carry)
	return uint256{a, b, c, d}
}

func (x uint256) add(y uint256) uint256 {
	var carry uint64
	x.d, carry = bits.Add64(x.d, y.d, 0)
	x.c, carry = bits.Add64(x.c, y.c, carry)
	x.b, carry = bits.Add64(x.b, y.b, carry)
	x.a, _ = bits.Add64(x.a, y.a, carry)
	return x
}

func (x uint256) sub(y uint256) uint256 {
	var borrow uint64
	x.d, borrow = bits.Sub64(x.d, y.d, 0)
	x.c, borrow = bits.Sub64(x.c, y.c, borrow)
	x.b, borrow = bits.Sub64(x.b, y.b, borrow)
	x.a, _ = bits.Sub64(x.a, y.a, borrow)
	return x
}

// rsh returns x >> n.
func (x uint256) rsh(n uint) uint256 {
	a, b, c, d := x.a, x.b, x.c, x.d
	switch {
	case n >= 256:
		return uint256{}
	case n >= 192:
		a, b, c, d = 0, 0, 0, a
		n -= 192
	case n >= 128:
		a, b, c, d = 0, 0, a, b
		n -= 128
	case n >= 64:
		a, b, c, d = 0, a, b, c
		n -= 64
	}
	return uint256{
		a >> n,
		b>>n | a<<(64-n),
		c>>n | b<<(64-n),
		d>>n | c<<(64-n),
	}
}

func (x uint256) lsh(n uint) uint256 {
	a, b, c, d := x.a, x.b, x.c, x.d
	switch {
	case n >= 256:
		return uint256{}
	case n >= 192:
		a, b, c, d = d, 0, 0, 0
		n -= 192
	case n >= 128:
		a, b, c, d = c, d, 0, 0
		n -= 128
	case n >= 64:
		a, b, c, d = b, c, d, 0
		n -= 64
	}
	return uint256{
		a<<n | b>>(64-n),
		b<<n | c>>(64-n),
		c<<n | d>>(64-n),
		d << n,
	}
}

func (x uint256) and(y uint256) uint256 {
	return uint256{x.a & y.a, x.b & y.b, x.c & y.c, x.d & y.d}
}

func (x uint256) leadingZeros() int {
	n := bits.LeadingZeros64(x.a)
	if n == 64 {
		n += bits.LeadingZeros64(x.b)
	}
	if n == 128 {
		n += bits.LeadingZeros64(x.c)
	}
	if n == 192 {
		n += bits.LeadingZeros64(x.d)
	}
	return n
}

func (x uint256) isZero() bool {
	return (x.a | x.b | x.c | x.d) == 0
}

func (x uint256) divMod128(y int128.Uint128) (div uint256, mod int128.Uint128) {
	if (y.H | y.L) == 0 {
		panic("division by zero")
	}
	if y.H == 0 {
		// fast path for uint256 / uint64
		div.a = x.a / y.L
		q, r := bits.Div64(x.a%y.L, x.b, y.L)
		div.b = q
		q, r = bits.Div64(r, x.c, y.L)
		div.c = q
		q, r = bits.Div64(r, x.d, y.L)
		div.d = q
		mod.L = r
		return
	}

	n := bits.LeadingZeros64(y.H)
	x2 := x.rsh(1)
	y2 := y.Lsh(uint(n))

	div.a = x2.a / y2.H
	q, r := bits.Div64(x2.a%y2.H, x2.b, y2.H)
	div.b = q
	q, r = bits.Div64(r, x2.c, y2.H)
	div.c = q
	q, _ = bits.Div64(r, x2.d, y2.H)
	div.d = q
	div = div.rsh(uint(63 - n))
	if !div.isZero() {
		div = div.sub(uint256{0, 0, 0, 1})
	}

	z := div.mul128(y)
	mod256 := x.sub(z)
	mod = int128.Uint128{H: mod256.c, L: mod256.d}
	if mod.Cmp(y) >= 0 {
		div = div.add(uint256{0, 0, 0, 1})
		mod = mod.Sub(y)
	}
	return
}

func (x uint256) GoString() string {
	return fmt.Sprintf("0x%016x_%016x_%016x_%016x", x.a, x.b, x.c, x.d)
}

func (x uint256) int256() int256 {
	return int256{int64(x.a), x.b, x.c, x.d}
}

type int256 struct {
	a       int64
	b, c, d uint64
}

func (x int256) GoString() string {
	sign := '+'
	a := x.a
	if a < 0 {
		sign = '-'
		a = -a
	}
	return fmt.Sprintf("%c0x%016x_%016x_%016x_%016x", sign, a, x.b, x.c, x.d)
}

func (x int256) add(y int256) int256 {
	d, carry := bits.Add64(x.d, y.d, 0)
	c, carry := bits.Add64(x.c, y.c, carry)
	b, carry := bits.Add64(x.b, y.b, carry)
	a, _ := bits.Add64(uint64(x.a), uint64(y.a), carry)
	return int256{int64(a), b, c, d}
}

func (x int256) setSign(b bool) int256 {
	if b && x.a >= 0 {
		return x.neg()
	}
	return x
}

func (x int256) neg() int256 {
	d, borrow := bits.Sub64(0, x.d, 0)
	c, borrow := bits.Sub64(0, x.c, borrow)
	b, borrow := bits.Sub64(0, x.b, borrow)
	a, _ := bits.Sub64(0, uint64(x.a), borrow)
	return int256{int64(a), b, c, d}
}

func (x int256) uint256() uint256 {
	return uint256{uint64(x.a), x.b, x.c, x.d}
}

func (x int256) abs() uint256 {
	if x.a < 0 {
		return x.neg().uint256()
	}
	return x.uint256()
}
