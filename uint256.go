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
		// we can directly calculate uint256 / uint64
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

	// calculate the high 128-bits of div.
	q, r := int128.Uint128{H: x.a, L: x.b}.DivMod(y)
	div.a, div.b = q.H, q.L
	x.a, x.b = r.H, r.L

	n := uint(y.LeadingZeros())
	y = y.Lsh(n)

	// next 64-bits
	one := int128.Uint128{L: 1}
	two64 := int128.Uint128{H: 1}

	yn1 := int128.Uint128{L: y.H}
	yn0 := int128.Uint128{L: y.L}
	xn := x.lsh(n)
	un64 := int128.Uint128{H: xn.a, L: xn.b}
	un1 := int128.Uint128{L: xn.c}
	un0 := int128.Uint128{L: xn.d}

	q1 := un64.Div(yn1)
	r = un64.Sub(q1.Mul(yn1))

	for q1.Cmp(two64) >= 0 || q1.Mul(yn0).Cmp(r.Mul(two64).Add(un1)) > 0 {
		q1 = q1.Sub(one)
		r = r.Add(yn1)
		if r.Cmp(two64) >= 0 {
			break
		}
	}

	// last 64-bits
	un21 := un64.Mul(two64).Add(un1).Sub(q1.Mul(y))
	q0 := un21.Div(yn1)
	r = un21.Sub(q0.Mul(yn1))
	for q0.Cmp(two64) >= 0 || q0.Mul(yn0).Cmp(r.Mul(two64).Add(un0)) > 0 {
		q0 = q0.Sub(one)
		r = r.Add(yn1)
		if r.Cmp(two64) >= 0 {
			break
		}
	}

	q = q1.Mul(two64).Add(q0)
	div.c, div.d = q.H, q.L
	mod = un21.Mul(two64).Add(un0).Sub(q0.Mul(y)).Rsh(n)
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
