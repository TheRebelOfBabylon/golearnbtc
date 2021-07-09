package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	//"io"
	//"crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
)

type Curve struct {
	p *big.Int
	a int64
	b int64
}

type Point struct {
	curve Curve
	x     *big.Int
	y     *big.Int
}

// var INF *Point
var INF = Point{
	curve: Curve{
		p: big.NewInt(0),
		a: 0,
		b: 0,
	},
	x: big.NewInt(0),
	y: big.NewInt(0),
}

type Generator struct {
	G *Point
	n *big.Int
}

func extended_euclidean_algorithm(a *big.Int, b *big.Int) (old_r *big.Int, old_s *big.Int, old_t *big.Int) {
	old_r, r := new(big.Int), new(big.Int)
	old_r.Set(a)
	r.Set(b)
	old_s, s := big.NewInt(1), big.NewInt(0)
	old_t, t := big.NewInt(0), big.NewInt(1)
	quotient, tmp := new(big.Int), new(big.Int)
	for r.Cmp(big.NewInt(0)) == 1 {
		quotient.Div(old_r, r)
		tmp.Mul(quotient, r)
		old_r, r = r, old_r.Sub(old_r, tmp)
		tmp.Mul(quotient, s)
		old_s, s = s, old_s.Sub(old_s, tmp)
		tmp.Mul(quotient, t)
		old_t, t = t, old_t.Sub(old_t, tmp)
	}
	return old_r, old_s, old_t
}

func inv(n *big.Int, p *big.Int) (mod *big.Int) {
	// tmp_p, tmp_n := new(big.Int), new(big.Int)
	// tmp_p.Set(p)
	// tmp_n.Set(n)
	_, x, _ := extended_euclidean_algorithm(n, p)
	//fmt.Printf("x: %v\np: %v\n", x, p)
	mod = new(big.Int)
	mod.Mod(x, p)
	return mod
}

func (p Point) elliptic_curve_addition(other_p Point) Point {
	if p.Compare(INF) {
		//fmt.Println("Yes")
		return other_p
	}
	if other_p.Compare(INF) {
		return p
	}
	if p.x.Cmp(other_p.x) == 0 && p.y.Cmp(other_p.y) != 0 {
		return INF
	}
	m := new(big.Int)
	if p.x == other_p.x {
		a := big.NewInt(2)
		a.Mul(a, p.y)
		i := inv(a, p.curve.p)
		tmp := new(big.Int)
		tmp.Set(p.x)
		tmp.Exp(tmp, big.NewInt(2), big.NewInt(0))
		b := big.NewInt(3)
		b.Mul(b, tmp)
		m.Mul(b, i)
	} else {
		a := new(big.Int)
		a.Sub(p.y, other_p.y)
		i := new(big.Int)
		i.Sub(p.x, other_p.x)
		tmp := inv(i, p.curve.p)
		m.Mul(a, tmp)
	}
	rx, m_prime := new(big.Int), new(big.Int)
	m_prime.Set(m)
	m_prime.Exp(m_prime, big.NewInt(2), big.NewInt(0))
	m_prime.Sub(m_prime, p.x)
	m_prime.Sub(m_prime, other_p.x)
	rx.Mod(m_prime, p.curve.p)
	ry, rx_prime := new(big.Int), new(big.Int)
	rx_prime.Set(rx)
	rx_prime.Sub(rx_prime, p.x)
	rx_prime.Mul(m, rx_prime)
	rx_prime.Add(rx_prime, p.y)
	rx_prime.Mul(rx_prime, big.NewInt(-1))
	ry.Mod(rx_prime, p.curve.p)
	return Point{
		curve: p.curve,
		x:     rx,
		y:     ry,
	}
}

func (p Point) Compare(other_p Point) bool {
	//fmt.Println("Here 0.1")
	if p.curve.p.Cmp(other_p.curve.p) != 0 {
		//fmt.Println("Here 0.2")
		return false
	} else if p.curve.a != other_p.curve.a {
		//fmt.Println("Here 0.3")
		return false
	} else if p.curve.b != other_p.curve.b {
		//fmt.Println("Here 0.4")
		return false
	} else if p.x.Cmp(other_p.x) != 0 || p.y.Cmp(other_p.y) != 0 {
		//fmt.Println("Here 0.5")
		return false
	}
	//fmt.Println("Here 0.6")
	return true
}

func (p Point) verify_on_curve(curve *Curve) bool {
	a := new(big.Int)
	a.Exp(p.y, big.NewInt(2), big.NewInt(0))
	b := new(big.Int)
	b.Exp(p.x, big.NewInt(3), big.NewInt(0))
	sub := new(big.Int)
	sub.Sub(a, b)
	sub.Sub(sub, big.NewInt(7))
	mod := new(big.Int)
	mod.Mod(sub, curve.p)
	return mod.Cmp(big.NewInt(0)) == 0
}

func (p Point) double_and_add(k *big.Int) Point {
	if k.Cmp(big.NewInt(0)) == -1 {
		panic(fmt.Sprintf("%v is smaller than 0", k))
	}
	result := INF
	append := p
	for k.Cmp(big.NewInt(0)) == 1 {
		//fmt.Printf("result: %v\nappend: %v\n\n", result, append)
		z := new(big.Int)
		z.Set(k)
		if z.And(z, big.NewInt(1)).Cmp(big.NewInt(1)) == 0 {
			result = result.elliptic_curve_addition(append)
			//fmt.Printf("1/2\nresult: %v\nappend: %v\n\n", result, append)
		}
		append = append.elliptic_curve_addition(append)
		k.Rsh(k, 1)
	}
	return result
}

func rotr(x, n, size *big.Int) *big.Int {
	one, two, three, tmp := new(big.Int), new(big.Int), new(big.Int), new(big.Int)
	var tmp_two64 uint64
	var tmp_two uint
	tmp.Set(n)
	tmp_two64 = tmp.Uint64()
	tmp_two = uint(tmp_two64)
	one.Rsh(x, tmp_two)
	two.Sub(size, n)
	tmp.Set(two)
	tmp_two64 = tmp.Uint64()
	tmp_two = uint(tmp_two64)
	two.Lsh(x, tmp_two)
	three.Exp(big.NewInt(2), size, big.NewInt(0)).Sub(three, big.NewInt(1))
	two.And(two, three)
	return one.Or(one, two)
}

func sig0(x *big.Int) *big.Int {
	one, two, three := new(big.Int), new(big.Int), new(big.Int)
	one = rotr(x, big.NewInt(7), big.NewInt(32))
	two = rotr(x, big.NewInt(18), big.NewInt(32))
	three.Rsh(x, 3)
	one.Xor(one, two)
	return one.Xor(one, three)
}

func sig1(x *big.Int) *big.Int {
	one, two, three := new(big.Int), new(big.Int), new(big.Int)
	one = rotr(x, big.NewInt(17), big.NewInt(32))
	two = rotr(x, big.NewInt(19), big.NewInt(32))
	three.Rsh(x, 10)
	one.Xor(one, two)
	return one.Xor(one, three)
}

func capsig0(x *big.Int) *big.Int {
	one, two, three := new(big.Int), new(big.Int), new(big.Int)
	one = rotr(x, big.NewInt(2), big.NewInt(32))
	two = rotr(x, big.NewInt(13), big.NewInt(32))
	three = rotr(x, big.NewInt(22), big.NewInt(32))
	one.Xor(one, two)
	return one.Xor(one, three)
}

func capsig1(x *big.Int) *big.Int {
	one, two, three := new(big.Int), new(big.Int), new(big.Int)
	one = rotr(x, big.NewInt(6), big.NewInt(32))
	two = rotr(x, big.NewInt(11), big.NewInt(32))
	three = rotr(x, big.NewInt(25), big.NewInt(32))
	one.Xor(one, two)
	return one.Xor(one, three)
}

func ch(x, y, z *big.Int) *big.Int {
	one, two := new(big.Int), new(big.Int)
	one.And(x, y)
	two.Not(x)
	two.And(two, z)
	return one.Xor(one, two)
}

func maj(x, y, z *big.Int) *big.Int {
	one, two, three := new(big.Int), new(big.Int), new(big.Int)
	one.And(x, y)
	two.And(x, z)
	three.And(y, z)
	one.Xor(one, two)
	return one.Xor(one, three)
}

func is_prime(n *big.Int) bool {
	l, t := new(big.Int), new(big.Int)
	l.Sqrt(n).Add(l, big.NewInt(1))
	for f := big.NewInt(2); f.Cmp(l) == -1; f.Add(f, big.NewInt(1)) {
		if t.Mod(n, f).Cmp(big.NewInt(0)) == 0 {
			return false
		}
	}
	return true
}

func first_n_primes(n int) []*big.Int {
	var primes = make([]*big.Int, n)
	j := big.NewInt(2)
	for i := 0; i < n; i++ {
		for k := new(big.Int).Set(j); k.Cmp(big.NewInt(350)) == -1; k.Add(k, big.NewInt(1)) {
			if is_prime(k) {
				tmp := new(big.Int).Set(k)
				primes[i] = tmp
				k.Add(k, big.NewInt(1))
				j.Set(k)
				break
			}
		}
	}
	return primes
}

func frac_bin(f float64) *big.Int {
	n := new(big.Float).SetInt64(int64(math.Pow(2, 32)))
	tmp := int64(f)
	y := new(big.Float).SetInt64(tmp)
	x := big.NewFloat(f)
	x.Sub(x, y)
	x.Mul(x, n)
	ans, _ := x.Int64()
	return big.NewInt(ans)
}

func genK() [64]*big.Int {
	p := first_n_primes(64)
	var ans [64]*big.Int
	for i := 0; i < len(p); i++ {
		tmp := float64(p[i].Int64())
		ans[i] = frac_bin(math.Pow(tmp, 1/3.0))
	}
	return ans
}

func genH() [8]*big.Int {
	p := first_n_primes(8)
	var ans [8]*big.Int
	for i := 0; i < len(p); i++ {
		tmp := float64(p[i].Int64())
		ans[i] = frac_bin(math.Pow(tmp, 1/2.0))
	}
	return ans
}

func i2b(num *big.Int) []byte {
	return num.Bytes()
}

func b2i(arr []byte) *big.Int {
	return new(big.Int).SetBytes(arr)
}

func pad(b []byte) []byte {
	c := b[:]
	l := len(c) * 8
	c = append(c, 0b10000000)
	for (len(c)*8)%512 != 448 {
		c = append(c, 0x00)
	}
	ext := big.NewInt(int64(l))
	c = append(c, ext.FillBytes(make([]byte, 8))...)
	return c
}

func sha256(b []byte) []byte {
	bit := big.NewInt(2)
	bit.Exp(bit, big.NewInt(32), big.NewInt(0))
	K := genK()
	c := pad(b)
	blocks := make([][]byte, len(c)/64)
	for i := 0; i < len(c)/64; i++ {
		blocks[i] = c[i*64 : i*64+64]
	}
	H := genH()
	for m := 0; m < len(blocks); m++ {
		var W [64][]byte
		for t := 0; t < 64; t++ {
			if t <= 15 {
				W[t] = blocks[m][t*4 : t*4+4]
				//fmt.Printf("t: %v\nW[t]: %q\n\n", t, W[t])
			} else {
				term1_copy := new(big.Int)
				term1 := sig1(b2i(W[t-2]))
				term1_copy.Set(term1)
				term2 := b2i(W[t-7])
				term3 := sig0(b2i(W[t-15]))
				term4 := b2i(W[t-16])
				term1.Add(term1, term2)
				term1.Add(term1, term3)
				term1.Add(term1, term4)
				total := new(big.Int)
				total.Mod(term1, bit)
				//fmt.Printf("t: %v\nterm1: %v\nterm2: %v\nterm3: %v\nterm4: %v\ntotal: %v\n\n", t, term1_copy, term2, term3, term4, total)
				W[t] = i2b(total)
			}
		}
		//fmt.Printf("%q\n", W)
		a, b, c, d, e, f, g, h := new(big.Int).Set(H[0]), new(big.Int).Set(H[1]), new(big.Int).Set(H[2]), new(big.Int).Set(H[3]), new(big.Int).Set(H[4]), new(big.Int).Set(H[5]), new(big.Int).Set(H[6]), new(big.Int).Set(H[7])
		for t := 0; t < 64; t++ {
			tmp1 := new(big.Int).Set(h)
			tmp1.Add(h, capsig1(e)).Add(tmp1, ch(e, f, g)).Add(tmp1, K[t]).Add(tmp1, b2i(W[t])).Mod(tmp1, bit)
			tmp2 := new(big.Int)
			tmp2.Add(capsig0(a), maj(a, b, c)).Mod(tmp2, bit)
			h.Set(g)
			g.Set(f)
			f.Set(e)
			e.Add(d, tmp1).Mod(e, bit)
			d.Set(c)
			c.Set(b)
			b.Set(a)
			a.Add(tmp1, tmp2).Mod(a, bit)
		}
		delta := [8]*big.Int{a, b, c, d, e, f, g, h}
		for i := 0; i < len(H); i++ {
			tmp := new(big.Int)
			tmp.Add(H[i], delta[i])
			tmp.Mod(tmp, bit)
			H[i] = new(big.Int).Set(tmp)
		}
	}
	var data [][]byte
	for i := 0; i < len(H); i++ {
		data = append(data, i2b(H[i]))
	}
	sep := []byte("")
	res := bytes.Join(data, sep)
	return res
}

func modLikePython(d, m int64) int64 {
	var res int64 = d % m
	if (res < 0 && m > 0) || (res > 0 && m < 0) {
		return res + m
	}
	return res
}

func ripemd160(b []byte) []byte {

	PADDING := make([]byte, 64)
	PADDING[0] = 0x80
	var K0 int64 = 0x00000000
	var K1 int64 = 0x5A827999
	var K2 int64 = 0x6ED9EBA1
	var K3 int64 = 0x8F1BBCDC
	var K4 int64 = 0xA953FD4E
	var KK0 int64 = 0x50A28BE6
	var KK1 int64 = 0x5C4DD124
	var KK2 int64 = 0x6D703EF3
	var KK3 int64 = 0x7A6D76E9
	var KK4 int64 = 0x00000000

	type RMDContext struct {
		state  *[5]int64
		count  *int
		buffer []byte
	}

	NewContext := func() RMDContext {
		cnt := 0
		st := [5]int64{0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0}
		return RMDContext{
			state:  &st,
			count:  &cnt,
			buffer: make([]byte, 64),
		}
	}

	ROL := func(n, x int64) int64 {
		return ((x << n) & 0xffffffff) | (x >> (32 - n))
	}

	F0 := func(x, y, z int64) int64 {
		return x ^ y ^ z
	}

	F1 := func(x, y, z int64) int64 {
		return (x & y) | (modLikePython(^x, 0x100000000) & z)
	}

	F2 := func(x, y, z int64) int64 {
		return (x | modLikePython(^y, 0x100000000)) ^ z
	}

	F3 := func(x, y, z int64) int64 {
		return (x & z) | (modLikePython(^z, 0x100000000) & y)
	}

	F4 := func(x, y, z int64) int64 {
		return x ^ (y | modLikePython(^z, 0x100000000))
	}

	R := func(a, b, c, d, e int64, Fj func(int64, int64, int64) int64, Kj, sj, rj int64, X []int64) (int64, int64) {
		tmp := Fj(b, c, d)
		//fmt.Printf("1 tmp: %v\n", tmp)
		tmp += a + X[rj] + Kj
		//fmt.Printf("2 tmp: %v\n", tmp)
		//fmt.Printf("X[rj]: %v\nX: %v\nrj: %v\n", X[rj], X, rj)
		a = ROL(sj, modLikePython(tmp, 0x100000000)) + e
		c = ROL(10, c)
		return modLikePython(a, 0x100000000), c
	}

	RMD160Transform := func(state *[5]int64, block []byte) {
		x := make([]int64, 16)
		// buf := bytes.NewReader(block[0:64])
		// fmt.Printf("buf: %q\nbuf: %v\nlen(x): %v\n", buf, buf, len(x))
		// err := binary.Read(buf, binary.LittleEndian, x)
		// fmt.Printf("buf after: %q\n", buf)
		// if err != nil {
		// 	fmt.Println("binary.Read failed:", err)
		// }
		//fmt.Println(x)
		for i := 0; i < len(x); i++ {
			x[i] = int64(binary.LittleEndian.Uint32(block[i*4:]))
			//fmt.Println(x)
			//fmt.Println(i)
		}
		//fmt.Printf("x: %v\n", x)
		a := state[0]
		b := state[1]
		c := state[2]
		d := state[3]
		e := state[4]

		/* Round 1 */
		a, c = R(a, b, c, d, e, F0, K0, 11, 0, x)
		//fmt.Printf("ROUND 1\na: %v\nc: %v\n", a, c)
		//os.Exit(0)
		e, b = R(e, a, b, c, d, F0, K0, 14, 1, x)
		d, a = R(d, e, a, b, c, F0, K0, 15, 2, x)
		c, e = R(c, d, e, a, b, F0, K0, 12, 3, x)
		b, d = R(b, c, d, e, a, F0, K0, 5, 4, x)
		a, c = R(a, b, c, d, e, F0, K0, 8, 5, x)
		e, b = R(e, a, b, c, d, F0, K0, 7, 6, x)
		d, a = R(d, e, a, b, c, F0, K0, 9, 7, x)
		c, e = R(c, d, e, a, b, F0, K0, 11, 8, x)
		b, d = R(b, c, d, e, a, F0, K0, 13, 9, x)
		a, c = R(a, b, c, d, e, F0, K0, 14, 10, x)
		e, b = R(e, a, b, c, d, F0, K0, 15, 11, x)
		d, a = R(d, e, a, b, c, F0, K0, 6, 12, x)
		c, e = R(c, d, e, a, b, F0, K0, 7, 13, x)
		b, d = R(b, c, d, e, a, F0, K0, 9, 14, x)
		a, c = R(a, b, c, d, e, F0, K0, 8, 15, x)
		/* Round 2 */
		e, b = R(e, a, b, c, d, F1, K1, 7, 7, x)
		d, a = R(d, e, a, b, c, F1, K1, 6, 4, x)
		c, e = R(c, d, e, a, b, F1, K1, 8, 13, x)
		b, d = R(b, c, d, e, a, F1, K1, 13, 1, x)
		a, c = R(a, b, c, d, e, F1, K1, 11, 10, x)
		e, b = R(e, a, b, c, d, F1, K1, 9, 6, x)
		d, a = R(d, e, a, b, c, F1, K1, 7, 15, x)
		c, e = R(c, d, e, a, b, F1, K1, 15, 3, x)
		b, d = R(b, c, d, e, a, F1, K1, 7, 12, x)
		a, c = R(a, b, c, d, e, F1, K1, 12, 0, x)
		e, b = R(e, a, b, c, d, F1, K1, 15, 9, x)
		d, a = R(d, e, a, b, c, F1, K1, 9, 5, x)
		c, e = R(c, d, e, a, b, F1, K1, 11, 2, x)
		b, d = R(b, c, d, e, a, F1, K1, 7, 14, x)
		a, c = R(a, b, c, d, e, F1, K1, 13, 11, x)
		e, b = R(e, a, b, c, d, F1, K1, 12, 8, x)
		/* Round 3 */
		d, a = R(d, e, a, b, c, F2, K2, 11, 3, x)
		c, e = R(c, d, e, a, b, F2, K2, 13, 10, x)
		b, d = R(b, c, d, e, a, F2, K2, 6, 14, x)
		a, c = R(a, b, c, d, e, F2, K2, 7, 4, x)
		e, b = R(e, a, b, c, d, F2, K2, 14, 9, x)
		d, a = R(d, e, a, b, c, F2, K2, 9, 15, x)
		c, e = R(c, d, e, a, b, F2, K2, 13, 8, x)
		b, d = R(b, c, d, e, a, F2, K2, 15, 1, x)
		a, c = R(a, b, c, d, e, F2, K2, 14, 2, x)
		e, b = R(e, a, b, c, d, F2, K2, 8, 7, x)
		d, a = R(d, e, a, b, c, F2, K2, 13, 0, x)
		c, e = R(c, d, e, a, b, F2, K2, 6, 6, x)
		b, d = R(b, c, d, e, a, F2, K2, 5, 13, x)
		a, c = R(a, b, c, d, e, F2, K2, 12, 11, x)
		e, b = R(e, a, b, c, d, F2, K2, 7, 5, x)
		d, a = R(d, e, a, b, c, F2, K2, 5, 12, x)
		/* Round 4 */
		c, e = R(c, d, e, a, b, F3, K3, 11, 1, x)
		b, d = R(b, c, d, e, a, F3, K3, 12, 9, x)
		a, c = R(a, b, c, d, e, F3, K3, 14, 11, x)
		e, b = R(e, a, b, c, d, F3, K3, 15, 10, x)
		d, a = R(d, e, a, b, c, F3, K3, 14, 0, x)
		c, e = R(c, d, e, a, b, F3, K3, 15, 8, x)
		b, d = R(b, c, d, e, a, F3, K3, 9, 12, x)
		a, c = R(a, b, c, d, e, F3, K3, 8, 4, x)
		e, b = R(e, a, b, c, d, F3, K3, 9, 13, x)
		d, a = R(d, e, a, b, c, F3, K3, 14, 3, x)
		c, e = R(c, d, e, a, b, F3, K3, 5, 7, x)
		b, d = R(b, c, d, e, a, F3, K3, 6, 15, x)
		a, c = R(a, b, c, d, e, F3, K3, 8, 14, x)
		e, b = R(e, a, b, c, d, F3, K3, 6, 5, x)
		d, a = R(d, e, a, b, c, F3, K3, 5, 6, x)
		c, e = R(c, d, e, a, b, F3, K3, 12, 2, x)
		/* Round 5 */
		b, d = R(b, c, d, e, a, F4, K4, 9, 4, x)
		a, c = R(a, b, c, d, e, F4, K4, 15, 0, x)
		e, b = R(e, a, b, c, d, F4, K4, 5, 5, x)
		d, a = R(d, e, a, b, c, F4, K4, 11, 9, x)
		c, e = R(c, d, e, a, b, F4, K4, 6, 7, x)
		b, d = R(b, c, d, e, a, F4, K4, 8, 12, x)
		a, c = R(a, b, c, d, e, F4, K4, 13, 2, x)
		e, b = R(e, a, b, c, d, F4, K4, 12, 10, x)
		d, a = R(d, e, a, b, c, F4, K4, 5, 14, x)
		c, e = R(c, d, e, a, b, F4, K4, 12, 1, x)
		b, d = R(b, c, d, e, a, F4, K4, 13, 3, x)
		a, c = R(a, b, c, d, e, F4, K4, 14, 8, x)
		e, b = R(e, a, b, c, d, F4, K4, 11, 11, x)
		d, a = R(d, e, a, b, c, F4, K4, 8, 6, x)
		c, e = R(c, d, e, a, b, F4, K4, 5, 15, x)
		b, d = R(b, c, d, e, a, F4, K4, 6, 13, x)

		//fmt.Printf("AFTER FIRST 5 ROUNDS\na: %v\nb: %v\nc: %v\nd: %v\ne: %v\n", a, b, c, d, e)

		aa := a
		bb := b
		cc := c
		dd := d
		ee := e

		a = state[0]
		b = state[1]
		c = state[2]
		d = state[3]
		e = state[4]

		/* Parallel round 1 */
		a, c = R(a, b, c, d, e, F4, KK0, 8, 5, x)
		e, b = R(e, a, b, c, d, F4, KK0, 9, 14, x)
		d, a = R(d, e, a, b, c, F4, KK0, 9, 7, x)
		c, e = R(c, d, e, a, b, F4, KK0, 11, 0, x)
		b, d = R(b, c, d, e, a, F4, KK0, 13, 9, x)
		a, c = R(a, b, c, d, e, F4, KK0, 15, 2, x)
		e, b = R(e, a, b, c, d, F4, KK0, 15, 11, x)
		d, a = R(d, e, a, b, c, F4, KK0, 5, 4, x)
		c, e = R(c, d, e, a, b, F4, KK0, 7, 13, x)
		b, d = R(b, c, d, e, a, F4, KK0, 7, 6, x)
		a, c = R(a, b, c, d, e, F4, KK0, 8, 15, x)
		e, b = R(e, a, b, c, d, F4, KK0, 11, 8, x)
		d, a = R(d, e, a, b, c, F4, KK0, 14, 1, x)
		c, e = R(c, d, e, a, b, F4, KK0, 14, 10, x)
		b, d = R(b, c, d, e, a, F4, KK0, 12, 3, x)
		a, c = R(a, b, c, d, e, F4, KK0, 6, 12, x)
		/* Parallel round 2 */
		e, b = R(e, a, b, c, d, F3, KK1, 9, 6, x)
		d, a = R(d, e, a, b, c, F3, KK1, 13, 11, x)
		c, e = R(c, d, e, a, b, F3, KK1, 15, 3, x)
		b, d = R(b, c, d, e, a, F3, KK1, 7, 7, x)
		a, c = R(a, b, c, d, e, F3, KK1, 12, 0, x)
		e, b = R(e, a, b, c, d, F3, KK1, 8, 13, x)
		d, a = R(d, e, a, b, c, F3, KK1, 9, 5, x)
		c, e = R(c, d, e, a, b, F3, KK1, 11, 10, x)
		b, d = R(b, c, d, e, a, F3, KK1, 7, 14, x)
		a, c = R(a, b, c, d, e, F3, KK1, 7, 15, x)
		e, b = R(e, a, b, c, d, F3, KK1, 12, 8, x)
		d, a = R(d, e, a, b, c, F3, KK1, 7, 12, x)
		c, e = R(c, d, e, a, b, F3, KK1, 6, 4, x)
		b, d = R(b, c, d, e, a, F3, KK1, 15, 9, x)
		a, c = R(a, b, c, d, e, F3, KK1, 13, 1, x)
		e, b = R(e, a, b, c, d, F3, KK1, 11, 2, x)
		/* Parallel round 3 */
		d, a = R(d, e, a, b, c, F2, KK2, 9, 15, x)
		c, e = R(c, d, e, a, b, F2, KK2, 7, 5, x)
		b, d = R(b, c, d, e, a, F2, KK2, 15, 1, x)
		a, c = R(a, b, c, d, e, F2, KK2, 11, 3, x)
		e, b = R(e, a, b, c, d, F2, KK2, 8, 7, x)
		d, a = R(d, e, a, b, c, F2, KK2, 6, 14, x)
		c, e = R(c, d, e, a, b, F2, KK2, 6, 6, x)
		b, d = R(b, c, d, e, a, F2, KK2, 14, 9, x)
		a, c = R(a, b, c, d, e, F2, KK2, 12, 11, x)
		e, b = R(e, a, b, c, d, F2, KK2, 13, 8, x)
		d, a = R(d, e, a, b, c, F2, KK2, 5, 12, x)
		c, e = R(c, d, e, a, b, F2, KK2, 14, 2, x)
		b, d = R(b, c, d, e, a, F2, KK2, 13, 10, x)
		a, c = R(a, b, c, d, e, F2, KK2, 13, 0, x)
		e, b = R(e, a, b, c, d, F2, KK2, 7, 4, x)
		d, a = R(d, e, a, b, c, F2, KK2, 5, 13, x)
		/* Parallel round 4 */
		c, e = R(c, d, e, a, b, F1, KK3, 15, 8, x)
		b, d = R(b, c, d, e, a, F1, KK3, 5, 6, x)
		a, c = R(a, b, c, d, e, F1, KK3, 8, 4, x)
		e, b = R(e, a, b, c, d, F1, KK3, 11, 1, x)
		d, a = R(d, e, a, b, c, F1, KK3, 14, 3, x)
		c, e = R(c, d, e, a, b, F1, KK3, 14, 11, x)
		b, d = R(b, c, d, e, a, F1, KK3, 6, 15, x)
		a, c = R(a, b, c, d, e, F1, KK3, 14, 0, x)
		e, b = R(e, a, b, c, d, F1, KK3, 6, 5, x)
		d, a = R(d, e, a, b, c, F1, KK3, 9, 12, x)
		c, e = R(c, d, e, a, b, F1, KK3, 12, 2, x)
		b, d = R(b, c, d, e, a, F1, KK3, 9, 13, x)
		a, c = R(a, b, c, d, e, F1, KK3, 12, 9, x)
		e, b = R(e, a, b, c, d, F1, KK3, 5, 7, x)
		d, a = R(d, e, a, b, c, F1, KK3, 15, 10, x)
		c, e = R(c, d, e, a, b, F1, KK3, 8, 14, x)
		/* Parallel round 5 */
		b, d = R(b, c, d, e, a, F0, KK4, 8, 12, x)
		a, c = R(a, b, c, d, e, F0, KK4, 5, 15, x)
		e, b = R(e, a, b, c, d, F0, KK4, 12, 10, x)
		d, a = R(d, e, a, b, c, F0, KK4, 9, 4, x)
		c, e = R(c, d, e, a, b, F0, KK4, 12, 1, x)
		b, d = R(b, c, d, e, a, F0, KK4, 5, 5, x)
		a, c = R(a, b, c, d, e, F0, KK4, 14, 8, x)
		e, b = R(e, a, b, c, d, F0, KK4, 6, 7, x)
		d, a = R(d, e, a, b, c, F0, KK4, 8, 6, x)
		c, e = R(c, d, e, a, b, F0, KK4, 13, 2, x)
		b, d = R(b, c, d, e, a, F0, KK4, 6, 13, x)
		a, c = R(a, b, c, d, e, F0, KK4, 5, 14, x)
		e, b = R(e, a, b, c, d, F0, KK4, 15, 0, x)
		d, a = R(d, e, a, b, c, F0, KK4, 13, 3, x)
		c, e = R(c, d, e, a, b, F0, KK4, 11, 9, x)
		b, d = R(b, c, d, e, a, F0, KK4, 11, 11, x)

		t := modLikePython((state[1] + cc + d), 0x100000000)
		state[1] = modLikePython((state[2] + dd + e), 0x100000000)
		state[2] = modLikePython((state[3] + ee + a), 0x100000000)
		state[3] = modLikePython((state[4] + aa + b), 0x100000000)
		state[4] = modLikePython((state[0] + bb + c), 0x100000000)
		state[0] = modLikePython(t, 0x100000000)
	}

	RMD160Update := func(ctx RMDContext, inp []byte, inplen int) {
		have := *ctx.count / 8 % 64
		need := 64 - have
		//fmt.Printf("ctx.count before: %v\n", *ctx.count)
		*ctx.count += 8 * inplen
		//fmt.Printf("ctx.count after: %v\n", *ctx.count)
		off := 0
		if inplen >= need {
			if have != 0 {
				for i := 0; i < need; i++ {
					ctx.buffer[have+i] = inp[i]
				}
				//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
				RMD160Transform(ctx.state, ctx.buffer)
				//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
				off = need
				have = 0
			}
			//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
			for off+64 <= inplen {
				RMD160Transform(ctx.state, inp[off:])
				//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
				off += 64
			}
		}
		if off < inplen {
			for i := 0; i < inplen-off; i++ {
				ctx.buffer[have+i] = inp[off+i]
			}
		}
	}

	RMD160Final := func(ctx RMDContext) []byte {
		size := make([]byte, 8)
		binary.LittleEndian.PutUint64(size, uint64(*ctx.count))
		padlen := 64 - ((*ctx.count / 8) % 64)
		if padlen < 1+8 {
			padlen += 64
		}
		//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
		//fmt.Printf("Second RMD160Update\n")
		RMD160Update(ctx, PADDING, padlen-8)
		//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
		//fmt.Printf("Third RMD160Update\n")
		RMD160Update(ctx, size, 8)
		// fmt.Println("Before binary write")
		// fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
		// Probably best to manually write
		// buf := new(bytes.Buffer)
		// err := binary.Write(buf, binary.LittleEndian, *ctx.state)
		// if err != nil {
		// 	fmt.Println("binary.Write failed:", err)
		// }
		// return buf.Bytes()
		var buf []byte
		for i := 0; i < len(*ctx.state); i++ {
			tmp := make([]byte, 4)
			binary.LittleEndian.PutUint32(tmp, uint32(ctx.state[i]))
			buf = append(buf, tmp...)
		}
		return buf
	}

	in_ripemd160 := func(b []byte) []byte {
		ctx := NewContext()
		//fmt.Printf("First RMD160Update\n")
		RMD160Update(ctx, b, len(b))
		//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
		digest := RMD160Final(ctx)
		//fmt.Printf("ctx state: %v\ncount: %v\nbuffer: %v\n", *ctx.state, *ctx.count, ctx.buffer)
		//fmt.Printf("%q\n", digest)
		return digest
	}

	return in_ripemd160(b)
}

type PublicKey struct {
	Point
}

func (pub PublicKey) encode(compressed, hash160 bool) []byte {
	var pkb []byte
	if compressed {
		pkb = make([]byte, 0, 33)
		var prefix byte
		if pub.y.Bit(0) == 0 {
			prefix = byte('\x02')
		} else {
			prefix = byte('\x03')
		}
		pkb = append(pkb, prefix)
		for i := 0; i < int(32)-len(pub.x.Bytes()); i++ {
			pkb = append(pkb, 0)
		}
		pkb = append(pkb, pub.x.Bytes()...)
		// fmt.Printf("bytes: %q\n", pub.x.Bytes())
		// fmt.Printf("bytes of self.x: %q\nlength of bytes: %v\n", pub.x.FillBytes(make([]byte, 32)), len(pub.x.FillBytes(make([]byte, 32))))
		// pkb = append(append(pkb, prefix), pub.x.FillBytes(make([]byte, 32))...)
	} else {
		pkb = make([]byte, 0, 65)
		pkb = append(append(append(pkb, byte('\x04')), pub.x.FillBytes(make([]byte, 32))...), pub.y.FillBytes(make([]byte, 32))...)
	}
	if hash160 {
		return ripemd160(sha256(pkb))
	}
	return pkb
}

func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}

}

func b58encode(b []byte) string {
	alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	if len(b) != 25 {
		fmt.Printf("Address must be 25 bytes long: %v\n", len(b))
	}
	n := b2i(b)
	var chars []string
	i := new(big.Int)
	test := new(big.Int).Set(n)
	for test.Abs(n).Cmp(big.NewInt(0)) == 1 {
		n, i = n.DivMod(n, big.NewInt(58), i)
		//fmt.Printf("n: %v\ni: %v\n", n, i)
		chars = append(chars, string(alphabet[i.Int64()]))
	}
	num_leading_zeros := len(b) - len(bytes.TrimLeft(b, "\x00"))
	//fmt.Println(num_leading_zeros)
	reverse(chars)
	//fmt.Println(chars)
	res := strings.Repeat(string(alphabet[0]), num_leading_zeros) + strings.Join(chars, "")
	return res
}

func (pub PublicKey) address(net string, compressed bool) string {
	pkb_hash := pub.encode(compressed, true)
	version := map[string]byte{
		"main": byte('\x00'),
		"test": byte('\x6f'),
	}
	var ver_pkb_hash, byte_address []byte
	ver_pkb_hash = append(append(ver_pkb_hash, version[net]), pkb_hash...)
	checksum := sha256(sha256(ver_pkb_hash))[:4]
	byte_address = append(append(byte_address, ver_pkb_hash...), checksum...)
	return b58encode(byte_address)
}

type Script interface {
	ScriptEncode() []byte
}

type IntScript struct {
	cmds []int
}

func (s IntScript) ScriptEncode() []byte {
	var out [][]byte
	for cmd := range s.cmds {
		out = append(out, []byte{byte(cmd)})
	}
	ret := []byte{byte(len(bytes.Join(out, []byte(""))))}
	ret = append(ret, bytes.Join(out, []byte(""))...)
	return ret
}

type ByteScript struct {
	cmds []byte
}

func (s ByteScript) ScriptEncode() []byte {
	var out [][]byte
	for _, cmd := range s.cmds {
		out = append(out, []byte{cmd})
	}
	joined := bytes.Join(out, []byte(""))
	ret := []byte{byte(len(joined))}
	ret = append(ret, joined...)
	//fmt.Printf("%v\n%v\n", len(ret), len(s.cmds))
	return ret
}

type TxIn struct {
	prev_tx               []byte
	prev_index            int
	script_sig            Script
	sequence              int64
	prev_tx_script_pubkey Script
}

func (t TxIn) txin_encode(script_override string) []byte {
	//fmt.Println(script_override)
	var out [][]byte
	prev_tx := make([]byte, len(t.prev_tx))
	copy(prev_tx, t.prev_tx)
	reverse(prev_tx)
	out = append(out, prev_tx)
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(t.prev_index))
	out = append(out, tmp)
	if script_override == "none" {
		//fmt.Printf("with prev_tx: %v\n", out)
		script_sig_tmp := t.script_sig.ScriptEncode()
		//fmt.Printf("script sig encoded bytes: %v\n", script_sig_tmp)
		out = append(out, script_sig_tmp)

	} else if script_override == "true" {
		out = append(out, t.prev_tx_script_pubkey.ScriptEncode())
	} else if script_override == "false" {
		//fmt.Println("We do end up here")
		dummy_script := ByteScript{
			cmds: []byte{},
		}
		tmp_two := dummy_script.ScriptEncode()
		//fmt.Printf("Empty script encoded: %v\n", tmp_two)
		out = append(out, tmp_two)
	} else {
		fmt.Println("script_override myst be 'none'|'true'|'false'")
	}
	//fmt.Printf("tx_in out: %v\n", out)
	tmp = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(t.sequence))
	out = append(out, tmp)
	return bytes.Join(out, []byte(""))
}

type TxOut struct {
	amount        int32
	script_pubkey Script
}

func (t TxOut) txout_encode() []byte {
	var out [][]byte
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, uint64(t.amount))
	out = append(out, tmp)
	out = append(out, t.script_pubkey.ScriptEncode())
	return bytes.Join(out, []byte(""))
}

func NewTxIn(prev_tx []byte, prev_index int) (tx TxIn) {
	return TxIn{
		prev_tx:    prev_tx,
		prev_index: prev_index,
		sequence:   0xffffffff,
	}
}

type Tx struct {
	version  uint32
	tx_ins   []TxIn
	tx_outs  []TxOut
	locktime int
}

func (t Tx) TxEncode(sig_index int) []byte {
	var out [][]byte
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, t.version)
	out = append(out, tmp)
	out = append(out, []byte{byte(len(t.tx_ins))})
	//fmt.Printf("version and tx in length: %v\n", out)
	if sig_index == -1 {
		//fmt.Println("YEET")
		for _, tx_in := range t.tx_ins {
			//fmt.Println("Encoding tx_in with none...")
			out = append(out, tx_in.txin_encode("none"))
		}
	} else {
		for i, tx_in := range t.tx_ins {
			//fmt.Printf("%v\n%v\n", i, tx_in)
			if sig_index == i {
				out = append(out, tx_in.txin_encode("true"))
			} else {
				out = append(out, tx_in.txin_encode("false"))
			}
			//fmt.Printf("out with new tx_in encoding: %v\n", out)
		}
		//fmt.Printf("out: %v\n", out)
	}
	//fmt.Printf("with tx_in all encoded: %v\n", out)
	out = append(out, []byte{byte(len(t.tx_outs))})
	for _, tx_out := range t.tx_outs {
		out = append(out, tx_out.txout_encode())
	}
	//fmt.Printf("with tx_outs encoded: %v\n", out)
	tmp = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(t.locktime))
	out = append(out, tmp)
	if sig_index != -1 {
		tmp = make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, uint32(1))
		out = append(out, tmp)
	} else {
		out = append(out, []byte(""))
	}
	//fmt.Printf("full out before join: %v\n", out)
	return bytes.Join(out, []byte(""))
}

type Signature struct {
	r *big.Int
	s *big.Int
}

func sign(secret_key *big.Int, gen Generator, message []byte) Signature {
	fmt.Printf("secret_key: %v\n", secret_key)
	z := new(big.Int).SetBytes(sha256(sha256(message)))
	seed := new(big.Int)
	seed.SetBytes(sha256(message))
	ran := rand.New(rand.NewSource(seed.Int64()))
	sk := new(big.Int)
	sk.Rand(ran, gen.n)
	fmt.Printf("sk: %v\n", sk)
	sk_copy := new(big.Int).Set(sk)
	P := gen.G.double_and_add(sk_copy)
	r := P.x
	inv := inv(sk, gen.n)
	fmt.Printf("inv: %v\n", inv)
	s := new(big.Int).Mul(secret_key, r)
	s.Add(s, z)
	fmt.Printf("s: %v\n", s)
	s.Mul(s, inv).Mod(s, gen.n)
	tmp := new(big.Int).Set(gen.n)
	tmp.Div(tmp, big.NewInt(2))
	if s.Cmp(tmp) == 1 {
		s.Sub(gen.n, s)
	}

	sig := Signature{
		r: r,
		s: s,
	}
	return sig
}

func (s Signature) sig_encode() []byte {
	dern := func(n *big.Int) []byte {
		tmp := make([]byte, 32)
		//fmt.Printf("n is a %v bit number\n", n.BitLen())
		nb := n.FillBytes(tmp)
		nb = bytes.TrimLeft(nb, "\x00")
		var b []byte
		if nb[0]&0x80 != 0 {
			b = append(b, byte('\x00'))
			nb = append(b, nb...)
		}
		return nb
	}
	rb := dern(s.r)
	sb := dern(s.s)
	content := bytes.Join([][]byte{{0x02, byte(len(rb))}, rb, {0x02, byte(len(sb))}, sb}, []byte(""))
	frame := bytes.Join([][]byte{{0x30, byte(len(content))}, content}, []byte(""))
	return frame
}

// func verify(public_key Point, message: []byte, sig Signature) bool {

// }

func main() {
	s := "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"
	i := new(big.Int)
	i.SetString(s, 16)
	btc_curve := Curve{
		p: i,
		a: 0x0000000000000000000000000000000000000000000000000000000000000000,
		b: 0x0000000000000000000000000000000000000000000000000000000000000007,
	}
	s_x := "79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"
	s_y := "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"
	i_x := new(big.Int)
	i_y := new(big.Int)
	i_x.SetString(s_x, 16)
	i_y.SetString(s_y, 16)
	G := Point{
		curve: btc_curve,
		x:     i_x,
		y:     i_y,
	}
	//Test if generator is on the curve
	if G.verify_on_curve(&btc_curve) {
		fmt.Println("TRUE")
	} else {
		fmt.Println("FALSE")
	}
	s_n := "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"
	n := new(big.Int)
	n.SetString(s_n, 16)
	btc_gen := Generator{
		G: &G,
		n: n,
	}
	priv_key := new(big.Int)
	priv_key.SetBytes([]byte("btc is the future"))
	//priv_key.SetBytes([]byte("Andrej is cool :P"))
	if priv_key.Cmp(big.NewInt(1)) == 0 || priv_key.Cmp(big.NewInt(1)) == 1 {
		if priv_key.Cmp(btc_gen.n) == -1 {
			fmt.Println("Valid key")
		}
	}
	fmt.Println(priv_key)
	// pk := G
	// if pk.verify_on_curve(&btc_curve) {
	// 	fmt.Println("pk valid")
	// } else {
	// 	fmt.Println("pk invalid")
	// }
	// fmt.Printf("G: %v\n", G)
	// fmt.Printf("pk: %v\n", pk)
	// pk_two := G.elliptic_curve_addition(G)
	// if pk_two.verify_on_curve(&btc_curve) {
	// 	fmt.Println("pk_two valid")
	// } else {
	// 	fmt.Println("pk_two invalid")
	// }
	// fmt.Printf("G: %v\n", G)
	// fmt.Printf("pk_two: %v\n", pk_two)
	// pk_three := G.elliptic_curve_addition(G).elliptic_curve_addition(G)
	// if pk_two.verify_on_curve(&btc_curve) {
	// 	fmt.Println("pk_three valid")
	// } else {
	// 	fmt.Println("pk_three invalid")
	// }
	// fmt.Printf("G: %v\n", G)
	// fmt.Printf("pk_three: %v\n", pk_three)
	// t_pk := G.double_and_add(big.NewInt(1))
	// fmt.Println(t_pk.verify_on_curve(&btc_curve))
	// t_pk_two := G.double_and_add(big.NewInt(2))
	// fmt.Println(t_pk_two.verify_on_curve(&btc_curve))
	pk_copy := new(big.Int).Set(priv_key)
	pub_key := G.double_and_add(pk_copy)
	fmt.Printf("x: %v\ny: %v\n", pub_key.x, pub_key.y)
	fmt.Printf("Pub_key is on curve? %v\n", pub_key.verify_on_curve(&btc_curve))
	mt_hash := sha256([]byte(""))
	encodedStr := hex.EncodeToString(mt_hash)
	fmt.Printf("%s\n", encodedStr)
	test_hash := sha256([]byte("here is a random bytes message, cool right?"))
	enc_str := hex.EncodeToString(test_hash)
	fmt.Printf("%s\n", enc_str)
	test := ripemd160([]byte("hello this is a test"))
	test_hex := hex.EncodeToString(test)
	fmt.Printf("%s\n", test_hex)
	fmt.Printf("number of bytes in a RIPEMD-160 digest: %v\n", len(ripemd160([]byte(""))))
	PubKey := PublicKey{
		Point: pub_key,
	}
	address := PubKey.address("test", true)
	fmt.Println(address)
	priv_key2 := new(big.Int)
	priv_key2.SetBytes([]byte("eth is a shitcoin"))
	pk2_copy := new(big.Int).Set(priv_key2)
	pub_key2 := G.double_and_add(pk2_copy)
	PubKey2 := PublicKey{
		Point: pub_key2,
	}
	address2 := PubKey2.address("test", true)
	fmt.Println(address2)
	prev_tx, _ := hex.DecodeString("02db4cde61cbeb96640ff8d6a12c2dd9800127e7705b60204ca61ad02f95ca80")
	// tx_in := TxIn{
	// 	prev_tx:    prev_tx,
	// 	prev_index: 1,
	// }
	tx_in := NewTxIn(prev_tx, 1)
	//fmt.Printf("tx_in prev_tx bytes: %v\n", len(tx_in.prev_tx))
	tx_out1 := TxOut{
		amount: 50000,
	}
	tx_out2 := TxOut{
		amount: 954070,
	}
	out1_pkb_hash := PubKey2.encode(true, true)
	//fmt.Printf("out1_pkb_hash: %v\n", out1_pkb_hash)
	out1_cmds := []byte{118, 169}
	out1_cmds = append(out1_cmds, []byte{byte(len(out1_pkb_hash))}...)
	out1_cmds = append(out1_cmds, out1_pkb_hash...)
	out1_cmds = append(out1_cmds, []byte{136, 172}...)
	out1_script := ByteScript{
		cmds: out1_cmds,
	}
	tx_out1.script_pubkey = out1_script
	out2_pkb_hash := PubKey.encode(true, true)
	//fmt.Printf("out2_pkb_hash hexed: %v\n", hex.EncodeToString(out2_pkb_hash))
	out2_cmds := []byte{118, 169}
	out2_cmds = append(out2_cmds, []byte{byte(len(out2_pkb_hash))}...)
	out2_cmds = append(out2_cmds, out2_pkb_hash...)
	out2_cmds = append(out2_cmds, []byte{136, 172}...)
	// out2_script := ByteScript{
	// 	cmds: []byte{118, 169, bytes.Join([][]byte{out2_pkb_hash}, []byte(""))[0], 136, 172},
	// }
	out2_script := ByteScript{
		cmds: out2_cmds,
	}
	tx_out2.script_pubkey = out2_script
	out2_copy := make([]byte, len(out2_cmds))
	copy(out2_copy, out2_cmds)
	source_script := ByteScript{
		cmds: out2_copy,
	}
	tx_in.prev_tx_script_pubkey = source_script
	enc1 := out1_script.ScriptEncode()
	enc2 := out2_script.ScriptEncode()
	//fmt.Printf("script2 encoded: %v\ntx_in prev_tx_script_pubkey: %v\n", enc2)
	fmt.Printf("%s\n", hex.EncodeToString(enc1))
	fmt.Printf("%s\n", hex.EncodeToString(enc2))
	tx := Tx{
		version:  1,
		tx_ins:   []TxIn{tx_in},
		tx_outs:  []TxOut{tx_out1, tx_out2},
		locktime: 0,
	}
	message := tx.TxEncode(0)
	fmt.Printf("%s\n", hex.EncodeToString(message))
	sig := sign(priv_key, btc_gen, message)
	fmt.Printf("Signature(r=%v, s=%v)\n", sig.r, sig.s)
	sig_bytes := sig.sig_encode()
	sig_bytes = append(sig_bytes, byte('\x01'))
	pubkey_bytes := PubKey.encode(true, false)
	var script_sig_cmds []byte
	script_sig_cmds = append(script_sig_cmds, []byte{byte(len(sig_bytes))}...)
	script_sig_cmds = append(script_sig_cmds, sig_bytes...)
	script_sig_cmds = append(script_sig_cmds, []byte{byte(len(pubkey_bytes))}...)
	script_sig_cmds = append(script_sig_cmds, pubkey_bytes...)
	tx.tx_ins[0].script_sig = ByteScript{
		cmds: script_sig_cmds,
	}
	tx_bytes := tx.TxEncode(-1)
	fmt.Printf("%s\n", hex.EncodeToString(tx_bytes))
	tx_id := sha256(sha256(tx_bytes))
	reverse(tx_id)
	fmt.Printf("tx_id: %s\n", hex.EncodeToString(tx_id))

	//Returning tBTC to testnet address mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt
	prev_tx, _ = hex.DecodeString("d1f770cdfe980eca99c18c52598fad6a1f68b8a59444e539722198914694b73e")
	tx_in1 := NewTxIn(prev_tx, 1) // first wallet
	// Second wallet
	tx_in2 := NewTxIn(prev_tx, 0)
	prev_tx, _ = hex.DecodeString("02db4cde61cbeb96640ff8d6a12c2dd9800127e7705b60204ca61ad02f95ca80")
	tx_in3 := NewTxIn(prev_tx, 0)
	prev_tx, _ = hex.DecodeString("72d3a1fbbc09ce0fe740d42afa356fd60353967578a3d8657e0e433d2039726e")
	tx_in4 := NewTxIn(prev_tx, 0)
	// TxOut to faucet address
	tx_out := TxOut{
		amount: 1102960,
	}
	out_pkb_hash, _ := hex.DecodeString("344a0f48ca150ec2b903817660b9b68b13a67026")
	out_cmds := []byte{118, 169}
	out_cmds = append(out_cmds, []byte{byte(len(out_pkb_hash))}...)
	out_cmds = append(out_cmds, out_pkb_hash...)
	out_cmds = append(out_cmds, []byte{136, 172}...)
	out_script := ByteScript{
		cmds: out1_cmds, //Ahhhh a happy little accident, I've created a consolidation tx
	}
	tx_out.script_pubkey = out_script

	in1_pkb_hash := PubKey.encode(true, true)
	in1_cmds := []byte{118, 169}
	in1_cmds = append(in1_cmds, []byte{byte(len(in1_pkb_hash))}...)
	in1_cmds = append(in1_cmds, in1_pkb_hash...)
	in1_cmds = append(in1_cmds, []byte{136, 172}...)
	in1_source_script := ByteScript{
		cmds: in1_cmds,
	}
	tx_in1.prev_tx_script_pubkey = in1_source_script

	in2_pkb_hash := PubKey2.encode(true, true)
	in2_cmds := []byte{118, 169}
	in2_cmds = append(in2_cmds, []byte{byte(len(in2_pkb_hash))}...)
	in2_cmds = append(in2_cmds, in2_pkb_hash...)
	in2_cmds = append(in2_cmds, []byte{136, 172}...)
	in2_source_script := ByteScript{
		cmds: in2_cmds,
	}
	tx_in2.prev_tx_script_pubkey = in2_source_script
	in3_cmds, in4_cmds := make([]byte, len(in2_cmds)), make([]byte, len(in2_cmds))
	copy(in3_cmds, in2_cmds)
	copy(in4_cmds, in2_cmds)
	in3_source_script := ByteScript{
		cmds: in3_cmds,
	}
	in4_source_script := ByteScript{
		cmds: in4_cmds,
	}
	tx_in3.prev_tx_script_pubkey = in3_source_script
	tx_in4.prev_tx_script_pubkey = in4_source_script

	new_tx := Tx{
		version:  1,
		tx_ins:   []TxIn{tx_in1, tx_in2, tx_in3, tx_in4},
		tx_outs:  []TxOut{tx_out},
		locktime: 0,
	}
	msg := new_tx.TxEncode(0)
	new_sig := sign(priv_key, btc_gen, msg)
	// fmt.Printf("Signature(r=%v, s=%v)\n", sig.r, sig.s)
	new_sig_bytes := new_sig.sig_encode()
	new_sig_bytes = append(new_sig_bytes, byte('\x01'))
	//pubkey_bytes := PubKey.encode(true, false)
	var new_script_sig_cmds []byte
	new_script_sig_cmds = append(new_script_sig_cmds, []byte{byte(len(new_sig_bytes))}...)
	new_script_sig_cmds = append(new_script_sig_cmds, new_sig_bytes...)
	new_script_sig_cmds = append(new_script_sig_cmds, []byte{byte(len(pubkey_bytes))}...)
	new_script_sig_cmds = append(new_script_sig_cmds, pubkey_bytes...)
	new_tx.tx_ins[0].script_sig = ByteScript{
		cmds: new_script_sig_cmds,
	}
	msg = new_tx.TxEncode(1)
	new_sig2 := sign(priv_key2, btc_gen, msg)
	new_sig_bytes2 := new_sig2.sig_encode()
	new_sig_bytes2 = append(new_sig_bytes2, byte('\x01'))
	pubkey2_bytes := PubKey2.encode(true, false)
	var new_script2_sig_cmds []byte
	new_script2_sig_cmds = append(new_script2_sig_cmds, []byte{byte(len(new_sig_bytes2))}...)
	new_script2_sig_cmds = append(new_script2_sig_cmds, new_sig_bytes2...)
	new_script2_sig_cmds = append(new_script2_sig_cmds, []byte{byte(len(pubkey2_bytes))}...)
	new_script2_sig_cmds = append(new_script2_sig_cmds, pubkey2_bytes...)

	msg = new_tx.TxEncode(2)
	new_sig3 := sign(priv_key2, btc_gen, msg)
	new_sig_bytes3 := new_sig3.sig_encode()
	new_sig_bytes3 = append(new_sig_bytes3, byte('\x01'))
	var new_script3_sig_cmds []byte
	new_script3_sig_cmds = append(new_script3_sig_cmds, []byte{byte(len(new_sig_bytes3))}...)
	new_script3_sig_cmds = append(new_script3_sig_cmds, new_sig_bytes3...)
	new_script3_sig_cmds = append(new_script3_sig_cmds, []byte{byte(len(pubkey2_bytes))}...)
	new_script3_sig_cmds = append(new_script3_sig_cmds, pubkey2_bytes...)

	msg = new_tx.TxEncode(3)
	new_sig4 := sign(priv_key2, btc_gen, msg)
	new_sig_bytes4 := new_sig4.sig_encode()
	new_sig_bytes4 = append(new_sig_bytes4, byte('\x01'))
	var new_script4_sig_cmds []byte
	new_script4_sig_cmds = append(new_script4_sig_cmds, []byte{byte(len(new_sig_bytes4))}...)
	new_script4_sig_cmds = append(new_script4_sig_cmds, new_sig_bytes4...)
	new_script4_sig_cmds = append(new_script4_sig_cmds, []byte{byte(len(pubkey2_bytes))}...)
	new_script4_sig_cmds = append(new_script4_sig_cmds, pubkey2_bytes...)
	new_tx.tx_ins[1].script_sig = ByteScript{
		cmds: new_script2_sig_cmds,
	}
	new_tx.tx_ins[2].script_sig = ByteScript{
		cmds: new_script3_sig_cmds,
	}
	new_tx.tx_ins[3].script_sig = ByteScript{
		cmds: new_script4_sig_cmds,
	}
	new_tx_bytes := new_tx.TxEncode(-1)
	fmt.Printf("%s\n", hex.EncodeToString(new_tx_bytes))
	new_tx_id := sha256(sha256(new_tx_bytes))
	reverse(new_tx_id)
	fmt.Printf("tx_id: %s\n", hex.EncodeToString(new_tx_id))

	//create the final tx to faucet
	prev_tx, _ = hex.DecodeString("ac461b593b6b825117c33421947ede73d4f196f8250b5d82d279ef7918741ee5")
	new_tx_in := NewTxIn(prev_tx, 0)
	new_tx_out := TxOut{
		amount: 1101850,
	}
	new_out_script := ByteScript{
		cmds: out_cmds, //now we use the out script from earlier
	}
	new_tx_out.script_pubkey = new_out_script
	new_tx_in.prev_tx_script_pubkey = ByteScript{
		cmds: in2_cmds,
	}
	final_tx := Tx{
		version:  1,
		tx_ins:   []TxIn{new_tx_in},
		tx_outs:  []TxOut{new_tx_out},
		locktime: 0,
	}
	final_msg := final_tx.TxEncode(0)
	final_sig := sign(priv_key2, btc_gen, final_msg)
	final_sig_bytes := final_sig.sig_encode()
	final_sig_bytes = append(final_sig_bytes, byte('\x01'))
	//pubkey2_bytes := PubKey2.encode(true, false)
	var final_script_sig_cmds []byte
	final_script_sig_cmds = append(final_script_sig_cmds, []byte{byte(len(final_sig_bytes))}...)
	final_script_sig_cmds = append(final_script_sig_cmds, final_sig_bytes...)
	final_script_sig_cmds = append(final_script_sig_cmds, []byte{byte(len(pubkey2_bytes))}...)
	final_script_sig_cmds = append(final_script_sig_cmds, pubkey2_bytes...)
	final_tx.tx_ins[0].script_sig = ByteScript{
		cmds: final_script_sig_cmds,
	}
	final_tx_bytes := final_tx.TxEncode(-1)
	fmt.Printf("%s\n", hex.EncodeToString(final_tx_bytes))
	final_tx_id := sha256(sha256(final_tx_bytes))
	reverse(final_tx_id)
	fmt.Printf("tx_id: %s\n", hex.EncodeToString(final_tx_id))
}
