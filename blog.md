# Introduction
The two goals I had in mind when deciding to translate the from-scratch Python implementation of BTC that Andrej Karpathy made were: learning a new coding language, namely Go, and getting a technical familiarity with Bitcoin.

Please keep in mind that this code won't be pretty, and will evolve over the course of the blog post. I will be highlighting some of the challenges I faced when translating this code, and I will be sharing little nuggets of information discovered along the way. 

I hope that whoever reads this learns something new.

## Part 1: Elliptic Curve Cryptography

I followed along a blog post that Andrej Karpathy wrote while translating this code, so I will follow the order of his post in a similar fashion. 

The first thing we need in order to make a Bitcoin transaction is a set of private and public keys. Cryptography is quite a complex topic that I don't feel qualified explaining, but all we need to know is that a private key is a secret number that should NEVER be shared. This number is used to mathematically prove ownership over your Bitcoin. A public key is a set of points, that are on a given elliptic curve and related to the private key. You derive a public key from a private key but it is incredibly difficult to do the reverse operation. This is the heart of modern cryptography security.

So to begin, Karpathy creates a class called `Curve` which has three arguments, `p`, `a` and `b`. He also adds a decorator to the class declartion `@dataclass`. In Go, there are no classes, and the closest equivalent is a `struct`. Something else to keep in mind is you cannot give a default value to an argument in a `struct`. So we define the arguments of the struct in `main()`. Finally, Go is a strictly typed language by design where as Python is dynamically typed. Large integers (larger than 64 bits) are supported within the `int` type. In Go, we must use the `math/big` module which offers a `big.Int` type to serve our needs here. `big.Int` methods return whatever the result is to the object that called the method. `big.Int` methods normally return a pointer to a `big.Int` type, so the `p` attribute of our struct will be of `*big.int` type. I added a print statement at the end of main so the Go compiler wouldn't complain about declaring `btc_curve` and not using it. 

```
package main

import (
	"fmt"
	"math/big"
)

type Curve struct {
	p *big.Int
	a int64
	b int64
}

func main() {
	s := "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"
	i := new(big.Int)
	i.SetString(s, 16)
	btc_curve := Curve{
		p: i,
		a: 0x0000000000000000000000000000000000000000000000000000000000000000,
		b: 0x0000000000000000000000000000000000000000000000000000000000000007,
	}
	fmt.Printf("%v\n", btc_curve.p)
}
```
We then define a `Point` struct and instantiate the point `G` which is a point on the Elliptic Curve used in Bitcoin's cryptography. The `Point` struct will also be used for public keys later.
```
type Point struct {
	curve Curve
	x     *big.Int
	y     *big.Int
}
```
We will also define a method for the `Point` struct called `verify_on_curve` which takes a `*Curve` pointer and returns whether or not the point is on the curve.
```
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
```
 As Go is strictly typed, something worth mentioning about functions/methods is having to define the types for input arguments as well as all output values. To make a function a method of a specific struct, you must also include the equivalent of `self` before the method name. In the example above, the function becomes a method with `(p Point)` which tells Go "This funciton is a method of Point". 

Also, as I pointed out earlier, `big.Int` operations return the value to the `big.Int` that called the method unlike normal math operations, where the result can be stored in a new variable via assignment (ex: `a := 4+7`).
```
func main() {
	s := "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"
	i := new(big.Int)
	i.SetString(s, 16)
	btc_curve := Curve{
		p: i,
		a: 0x0000000000000000000000000000000000000000000000000000000000000000,
		b: 0x0000000000000000000000000000000000000000000000000000000000000007,
	}
	//fmt.Printf("%v\n", btc_curve.p)
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
}
```
A `Generator` struct is also defined. This data type will be used to create a valid private key.
```
type Generator struct {
	G *Point
	n *big.Int
}
```
```
func main() {
	s := "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"
	i := new(big.Int)
	i.SetString(s, 16)
	btc_curve := Curve{
		p: i,
		a: 0x0000000000000000000000000000000000000000000000000000000000000000,
		b: 0x0000000000000000000000000000000000000000000000000000000000000007,
	}
	//fmt.Printf("%v\n", btc_curve.p)
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
	s_n := "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"
	n := new(big.Int)
	n.SetString(s_n, 16)
	btc_gen := Generator{
		G: &G,
		n: n,
	}
	priv_key := new(big.Int)
	priv_key.SetBytes([]byte("btc is the future"))
	if priv_key.Cmp(big.NewInt(1)) == 0 || priv_key.Cmp(big.NewInt(1)) == 1 {
		if priv_key.Cmp(btc_gen.n) == -1 {
		fmt.Println("Valid key")
		}
	}
	fmt.Println(priv_key)
}
```
And with that we've created our private key. Notice that we can't use standard Go comparison operators between `big.Int` types. Instead, the `.Cmp()` method is provided. Also notice that no matter how many times we run this script, the private key doesn't change. They are deterministically derived  (in this case from the phrase "btc is the future"). This would not be a strong phrase for a private key. Most well built Bitcoin wallets generate a list of 12-24 words from 1024 possible choices which are then used to derive the private key. This is a very secure method of deriving a private key as each combination of 12-24 words are one in 12^1014 - 24^1024 possibilites (well not quite as there should not be any repeating words and the last word is a checksum). But this will suit our purposes just fine.

Now, to create the public key we have to add the Generator point to itself private key number of times. But this isn't a simple multiplication as a point has an x and y value, the private key is just a large integer. It also isn't like multiplying a vector by a scalar. It wouldn't be much of a secret key if you could just divide the public key by the generator point and get a private key. We now create a set of functions to perform this special operation.

We define a point called `INF` which is meant to be like a point at infinity. In Python, this point is defined using the `None` type to which there is no equivalent in Go. I first tried defining `INF` as `new(Point)`. From what I read on instantiating a variable but not assigning it a value in Go, the variable automatically assumes a default value of `0` or `""` or `false`. I'm not sure if this is the case for `big.Int` and so, I instead manually defined zero values for our `INF` point.
```
var INF = Point{
	curve: Curve{
		p: big.NewInt(0),
		a: 0,
		b: 0,
	},
	x: big.NewInt(0),
	y: big.NewInt(0),
}
```
Here are the functions necessary for adding a scalar to a point in the realm of cryptography. I won't pretend to understand what's going on here so instead I'll mention the difficulties I encountered when implementing it in Go. I created a `Compare` method for `Point` structs because in Python, by default, you're able to compare objects of the same class using standard operators. This is thanks to Dunder methods like `__eq__()`. Go does not have Dunder methods, so I created a `Compare` method which compares all attributes of point `p` to the attributes of `other_p`. If there are any differences, the method returns `false`.   
```
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
	_, x, _ := extended_euclidean_algorithm(n, p)
	mod = new(big.Int)
	mod.Mod(x, p)
	return mod
}

func (p Point) elliptic_curve_addition(other_p Point) Point {
	if p.Compare(INF) {
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
	if p.curve.p.Cmp(other_p.curve.p) != 0 {
		return false
	} else if p.curve.a != other_p.curve.a {
		return false
	} else if p.curve.b != other_p.curve.b {
		return false
	} else if p.x.Cmp(other_p.x) != 0 || p.y.Cmp(other_p.y) != 0 {
		return false
	}
	return true
}
```
As I mentioned above, there are no Dunder method equivalents in Go so we can't change the definition of `+` for `Point` objects like Karpathy did in Python. Instead, I made `elliptic_curve_addition()` a `Point` method. The `elliptic_curve_addition()` method returns a `Point` which is the public key. 

To test if this was properly implemented, we can try adding private keys with values of `1`, `2` and `3` to `G` using `elliptic_curve_addition()` and then testing if the resulting public keys are on the curve by adding the following to `main()`
```
	pk := G
	if pk.verify_on_curve(&btc_curve) {
		fmt.Println("pk valid")
	} else {
		fmt.Println("pk invalid")
	}
	fmt.Printf("G: %v\n", G)
	fmt.Printf("pk: %v\n", pk)
	pk_two := G.elliptic_curve_addition(G)
	if pk_two.verify_on_curve(&btc_curve) {
		fmt.Println("pk_two valid")
	} else {
		fmt.Println("pk_two invalid")
	}
	fmt.Printf("G: %v\n", G)
	fmt.Printf("pk_two: %v\n", pk_two)
	pk_three := G.elliptic_curve_addition(G).elliptic_curve_addition(G)
	if pk_two.verify_on_curve(&btc_curve) {
		fmt.Println("pk_three valid")
	} else {
		fmt.Println("pk_three invalid")
	}
	fmt.Printf("G: %v\n", G)
	fmt.Printf("pk_three: %v\n", pk_three)
```

Now, as you may have noticed, private keys are normally very large integers and manually writing out `elliptic_curve_addition()` private key number of times could be long if not impossible. So we will write a short cut that in Python would override the `__mul__()` Dunder method but in Go will just be another `Point` method.

```
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
```
And finally we can test if the public key created from our private key is on the curve by adding the following to `main()`
```
	pk_copy := new(big.Int).Set(priv_key)
	pub_key := G.double_and_add(pk_copy)
	fmt.Printf("x: %v\ny: %v\n", pub_key.x, pub_key.y)
	fmt.Printf("Pub_key is on curve? %v\n", pub_key.verify_on_curve(&btc_curve))
```
To use `double_and_add()` I first create a copy of the private key. The copy is necessary because `double_and_add()` would otherwise modify the value of the input argument provided with this line of code: `k.Rsh(k, 1)`. This was something that I missed and it would later haunt me while trying to use the private key to sign the Bitcoin transaction.

Now you'll notice that a public key does not at all look like a Bitcoin address. For that we need a couple more cryptographic tools. Namely, the SHA256 algorithm and the RIPEMD160 algorithm.

```
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
```
Here is where I'd say things started really clicking for me about how to properly use `math/big`. By that I mean creating temporary variables to store operation results in large equations as well as making sure that functions don't modify input arguments directly. One noteable mistake I had made when translating this algorithm was when I initially assigned values to `a`, `b`, `c`, `d`, `e`, `f`, `g` and `h`. I had written this line of code: `a, b, c, d, e, f, g, h := new(big.Int).Set(H[0]), new(big.Int).Set(H[1]), new(big.Int).Set(H[2]), new(big.Int).Set(H[3]), new(big.Int).Set(H[4]), new(big.Int).Set(H[5]), new(big.Int).Set(H[6]), new(big.Int).Set(H[7])` as `a, b, c, d, e, f, g, h := H[0], H[1], H[2], H[3], H[4], H[5], H[6], H[7]`. This basically made `a b c d e f g h` pointers to the values stored in the `H` array instead of copying the values and creating separate addresses in memory. This mistake caused the hashes to be completely wrong. When you actually look at the `genH()` function and see it returns `[8]*big.Int` which is an array of pointers, this makes sense. 

I found that the best way to debug in Go, especially with `big.Int` data types was to just add print statements throughout my code instead of using my IDE's debugger tool. Adding `big.Int` variables to the watch list won't display the actual value of the variable but instead only it's size and what memory addresses it occupies. At least that is the case in VSCode.  

I added `"bytes"` and `"encoding/hex"` packages to the import statement at the top of the file and the following lines to `main()` to test if I had successfully implemented `sha256()`
```
	mt_hash := sha256([]byte(""))
	encodedStr := hex.EncodeToString(mt_hash)
	fmt.Printf("%s\n", encodedStr)
	test_hash := sha256([]byte("here is a random bytes message, cool right?"))
	enc_str := hex.EncodeToString(test_hash)
	fmt.Printf("%s\n", enc_str)
```
Alongside SHA256, we also need RIPEMD160 which is implemented below
```
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
```
With `ripemd160()`, I decided to implement the helper functions as private functions within `ripemd160()`. While implementing this function, I learned alot about static typing. I had initially just used whatever integer type came to mind, but I encountered many errors while going from bytes to ints and vice versa. I also learned that Python and Go have implemented a different version of the modulus operation by default, which lead to some errors. I added the `modLikePython()` function based on [this article](https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go) to compensate for this discrepancy.

Here are the tests I added to `main()` to test `ripemd160()`. Don't forget to add `"encoding/binary"` and `"math"` to the import statement.
```
	test := ripemd160([]byte("hello this is a test"))
	test_hex := hex.EncodeToString(test)
	fmt.Printf("%s\n", test_hex)
	fmt.Printf("number of bytes in a RIPEMD-160 digest: %v\n", len(ripemd160([]byte(""))))
```
The next thing on the to-do list was to create the `PublicKey` struct. In Python, this is a child of the `Point` class but there is no concept of inheritance in Go. Instead, Go has composition where structs can extend the functionality of other structs. This is what I did for the `PublicKey` struct.

```
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
```
At this point, I think it's worth mentionning that Go has two different array-like data structures. One is an array, which acts as you would expect; the other is a slice which acts more like a list in Python. It's a lightweight, extensible array-like structure. This is something that I could've been more strict about using when translating this code. There are a few instances where I either knew or could've known what the expected size of the array structure was and could've used a proper array in Go as opposed to a slice. This would make the code more robust and resiliant to errors. But as this is a purely educational implementation of Bitcoin, I won't dwell on this too much.

Finally, the `b58encode()` and `address` methods were implemented
```
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
```
In Python, `[::-1]` is used to reverse the order of the indices in an array or list. There is no such function by default in Go so I created a `reverse()` function (which came in handy when creating a txid). Here, I also learned about the Go equivalent of a Python dictionary, or at least the closest thing to one: the `map` type.

And finally here is the additions to `main()` to at last get our bitcoin address. Add `"reflect"` and `"strings"` packages to the import statement as well.
```
	PubKey := PublicKey{
		Point: pub_key,
	}
	address := PubKey.address("test", true)
	fmt.Println(address)
```
## Part 2: Building a Bitcoin Transaction

[Here](https://blockstream.info/testnet/tx/65d9d108407bc588ac0c6ee17029f261cd7f9c66edacb579bac86a04f8d3cb4a) we can see our first transaction to our newly created Bitcoin address: requesting tBTC from a Testnet faucet. We can finally start doing some more interesting things in our from-scratch Bitcoin build. First we will add a few lines to `main()` to create a second Bitcoin address which we will send funds to.
```
	priv_key2 := new(big.Int)
	priv_key2.SetBytes([]byte("eth is a shitcoin"))
	pk2_copy := new(big.Int).Set(priv_key2)
	pub_key2 := G.double_and_add(pk2_copy)
	PubKey2 := PublicKey{
		Point: pub_key2,
	}
	address2 := PubKey2.address("test", true)
	fmt.Println(address2)
```
We can check if either Bitcoin addresses are actually valid by typing them into a testnet bitcoin block explorer and seeing if a page appears. The goal is to create a Bitcoin transaction which sends funds from the first bitcoin wallet we created to the second. For this we start by creating a `TxIn` struct and a `TxOut` struct
```
type TxIn struct {
	prev_tx               []byte
	prev_index            int
	script_sig            Script
	sequence              int64
	prev_tx_script_pubkey Script
}

type TxOut struct {
	amount        int32
	script_pubkey Script
}
```
To stop the Go compiler from yelling: "Hey what is this Script type you speak of?", we will need to create it. In Python, the `Script` class has an attribute called `cmds` which is a union of both `int` and `bytes` types. I learned that in Go, there are no native `union` types. As a substitute, Go has an `interface` type. Interfaces are defined as any data type which have all of the methods listed in the `interface` definition statement. For example, the `Script` interface below
```
type Script interface {
	ScriptEncode() []byte
}
```
is any data type that has a method called `ScriptEncode()`, which takes no input arguments and returns a byte slice. If, for example, a struct had a method called `ScriptEncode()` but also had another method called `foo()`, that struct would still be a member of the `Script` type. I created `ByteScript` and `IntScript` structs and created implementations of `ScriptEncode()` for both of these structs which use a version of `cmds` that is a byte slice and another that is an int slice.
```
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
```
So I used the `interface` data structure to circumvent the lack of `union` in Go. I also created a `NewTxIn()` constructor method for the `TxIn` struct since the `sequence` attribute of `TxIn` has a default value.
```
func NewTxIn(prev_tx []byte, prev_index int) (tx TxIn) {
	return TxIn{
		prev_tx:    prev_tx,
		prev_index: prev_index,
		sequence:   0xffffffff,
	}
}
```
additions to `main()`
```
	prev_tx, _ := hex.DecodeString("02db4cde61cbeb96640ff8d6a12c2dd9800127e7705b60204ca61ad02f95ca80")
	tx_in := NewTxIn(prev_tx, 1)
	//fmt.Printf("tx_in prev_tx bytes: %v\n", len(tx_in.prev_tx))
	tx_out1 := TxOut{
		amount: 50000,
	}
	tx_out2 := TxOut{
		amount: 954070,
	}
```
Any attribute not explicitely assigned a value is given a value of zero by default. What we just defined above are the input and outputs of the transaction we are building. The `TxIn` references the previous transaction, which, in this case, is the transaction from the tBTC faucet to our first address. It also specifies the index of the output of this transaction that belongs to us. By checking the transaction information on the block explorer, we can see that our unspent transaction output (UTXO) was in the 1st index. The outputs we created only specify the amount. If you sum both amounts and compare this with the amount that the `TxIn` references, the sum of the outputs are less than the amount referenced as input. The difference between both amounts is the transaction fee that the miner claims when your transaction is successfully included in a block. Another curious thing about the Bitcoin protocol is that we don't specify an input amount but instead only reference outputs from previous transactions. A great analogy to understand this design choice is via a comparison with physical cash:

	Suppose we went to a store and bought a 20$ item with a 50$ bill. The cashier will take our 50$ and give us 30$ in change. We will assume there are no 30$ bills and instead assume we received a 20$ bill first and then a 10$ bill. This would be like having a transaction input be the 50$ bill with three outputs, 20$ to the cashier, 20$ and 10$ back to us. Now we go to a different store and we wish to purchase a 5$ item. Using the lens of the Bitcoin protocol, to send 5$ to the shop clerk, we must first reference the past transaction we made at the first store, the one where we received 20$ and 10$. We either reference the 20$ output (index of 1) or the 10$ output (index of 2) or both. In any case, we don't try and cut the bill in quarters or in halves so that the merchant receives 5$ exactly. UTXOs, like cash, cannot be partially consumed. But we can specify what denominations of change we would like to receive. For example, I may use the 10$ bill and tell the shop clerk that I would like to receive two 2$ bills and a 1$ bill as change. In Bitcoin terms, this would mean we would have three `TxOuts`, two with an amount of 2$ and one with an amount of 1$. The great thing about Bitcoin is that a single Bitcoin can be subdivided into 100 million parts and a `TxOut` amount can be any amount we desire, so long as the sum total of these amounts does not exceed the referenced amount in the inputs. The `TxOut` amounts are denominated in units of 1/100 million of a Bitcoin, aka a Satoshi.

Above, we have two `TxOut` because the first `TxOut` is the specified amount (50 000 satoshis) being sent to our second bitcoin address. The second `TxOut` is the remaining change we are sending back to ourselves (954 070 satoshis). If we so desired, we could split our second `TxOut` into two `TxOut` each with an amount that together sums up to 954 070. If the recipient so desired, we could also split 50 000 satoshis into any combination of outputs that sum to 50 000.   

To broadcast a Bitcoin transaction to other nodes in the network, we must follow a standard format for encoding the transactions. Each component of a transaction has it's own specified way of being encoded. The scripts in the transactions are encoded as per the `ScriptEncode()` methods. `TxIn` and `TxOut` are encoded via the `txin_encode()` and `txout_encode()` methods, implemented below
```
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

func (t TxOut) txout_encode() []byte {
	var out [][]byte
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, uint64(t.amount))
	out = append(out, tmp)
	out = append(out, t.script_pubkey.ScriptEncode())
	return bytes.Join(out, []byte(""))
}
```
You may be thinking "What is a script in Bitcoin?" or "How is forgery prevented in Bitcoin if a transaction is just a reference to the output of a previous transaction being sent elsewhere? Couldn't I find any new transaction, reference it's outputs and send them to myself?". Scripts in Bitcoin are what prevent this unwanted spending and the specific script that does such a thing is called the locking script. It is effectively a cryptographic math challenge that must be succesfully completed in order to unlock the UTXO and spend it in a new transaction. The locking script uses the public key of the recipient to lock the outputs and the only way to solve the puzzle is to use the corresponding private key of the public key used to create the challenge. Below, we are going to create the locking scripts for our `TxOut`.
```
func main() {
	...
	out1_pkb_hash := PubKey2.encode(true, true)
	fmt.Printf("out1_pkb_hash: %v\n", out1_pkb_hash)
	out1_cmds := []byte{118, 169}
	out1_cmds = append(out1_cmds, []byte{byte(len(out1_pkb_hash))}...)
	out1_cmds = append(out1_cmds, out1_pkb_hash...)
	out1_cmds = append(out1_cmds, []byte{136, 172}...)
	out1_script := ByteScript{
		cmds: out1_cmds,
	}
	tx_out1.script_pubkey = out1_script
	out2_pkb_hash := PubKey.encode(true, true)
	fmt.Printf("out2_pkb_hash hexed: %v\n", hex.EncodeToString(out2_pkb_hash))
	out2_cmds := []byte{118, 169}
	out2_cmds = append(out2_cmds, []byte{byte(len(out2_pkb_hash))}...)
	out2_cmds = append(out2_cmds, out2_pkb_hash...)
	out2_cmds = append(out2_cmds, []byte{136, 172}...)
	out2_script := ByteScript{
		cmds: out2_cmds,
	}
	tx_out2.script_pubkey = out2_script
}
```
So here `out1_pkb_hash` is the hash of the public key to our second wallet. Which means our first `TxOut` will be locked to our second wallet. `out2_pkb_hash` is the hash of the public key of our first wallet, which means the second `TxOut` is our change. If we look at what's output at the line `fmt.Printf("out2_pkb_hash hexed: %v\n", hex.EncodeToString(out2_pkb_hash))` and compare that to the locking script of our output on the testnet faucet transaction found [here](https://blockstream.info/testnet/tx/65d9d108407bc588ac0c6ee17029f261cd7f9c66edacb579bac86a04f8d3cb4a?expand) we should see the exact same value.

Before we can unlock the referenced output and spend it in this transaction, we are going to create a struct called `Tx` which will act as a container for our `TxIn` and two `TxOut`. It will also have it's own encoding method which will encode all elements of the transaction in the proper fashion.
```
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
```
To unlock the previous transaction output, we must create the unlock script. The unlock script has three parts:
- A copy of the locking script we are unlocking, stored as the `prev_tx_script_pubkey` attribute of our `TxIn`
- Proof that we own the private key corresponding to the specified public key referenced in the locking script
- Approval over the transaction we are currently constructing

First, let's recreate the locking script of the previous transaction by adding the following lines to `main()`
```
	out2_copy := make([]byte, len(out2_cmds))
	copy(out2_copy, out2_cmds)
	source_script := ByteScript{
		cmds: out2_copy,
	}
	tx_in.prev_tx_script_pubkey = source_script
```
It's essentially a copy and paste of the locking script for our second `TxOut` (makes sense since this is the change we are sending back to ourselves). Finally, we will instantiate `Tx` in `main()`
```
	tx := Tx{
		version:  1,
		tx_ins:   []TxIn{tx_in},
		tx_outs:  []TxOut{tx_out1, tx_out2},
		locktime: 0,
	}
```
Next, we must encode the transaction. Since we haven't yet specified the `script_sig` attribute of `TxIn` we can't pass `-1` as our input argument to `TxEncode()`. Instead we will pass the index of the `TxOut` that belongs to the address that is authorizing this transaction, our first address. We will store the result in a variable called `message`
```
	message := tx.TxEncode(0)
```
So at this point, we've copied the locking script of the transaction we are referencing in our input. Now we must prove ownership over the referenced public key in the locking script and authorize the transaction we are constructing. To do this, we will create a cryptographic signature using our private key. Signing a known message (in this case, our encoded transaction) with a private key proves to the network that we are the owner of the associated public key without revealing what the private key is. Nodes that receive our broadcasted transaction will find the signature and our public key. Then, they calculate the validity of this signature all without needing our private key to do so. Pretty neat :)

To create a cryptographic signature, we will create a `Signature` struct and a `sign()` function which takes a message, the Bitcoin generator and our private key as inputs and returns a Signature. Additionally, we will define the `sig_encode()` method for encoding our signature 
```
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
```
A difficulty I experienced when translating this part of the code was in the `sign()` function where the variable `sk` is defined and assigned a value. At first, I didn't seed the random number generator which sets the value of `sk` and this would always produce this error when broadcasting my transacion `{"code":-26,"message":"mandatory-script-verify-flag-failed (Signature must be zero for failed CHECK(MULTI)SIG operation)"}`. When I seeded the random number generator with the sha256 hash of the message, it worked. My hunch is that `sk` is meant to be a pseudo-random value derived from our private key. Since `message` includes our public key hash, `sk` is then part of a set of random numbers associated to our private key. I know that in the `btcutil` package (an actual safe to use implementation of the Bitcoin protocol), `sk` is derived deterministically from the private key.

We have completed the unlock script and we are now ready to include our signature in the transaction to perform the final encoding. To include the encoded signature in our transaction, we will create another `Script` struct and assign it as our `script_sig` attribute of `TxIn`.
```
	sig := sign(priv_key, btc_gen, message)
	//fmt.Printf("Signature(r=%v, s=%v)\n", sig.r, sig.s)
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
```
Finally, we will encode our completed transaction,
```
	tx_bytes := tx.TxEncode(-1)
	fmt.Printf("%s\n", hex.EncodeToString(tx_bytes))
	tx_id := sha256(sha256(tx_bytes))
	reverse(tx_id)
	fmt.Printf("tx_id: %s\n", hex.EncodeToString(tx_id))
```
copy the output of `fmt.Printf("%s\n", hex.EncodeToString(tx_bytes))` and paste it [here](https://blockstream.info/testnet/tx/push). And voila, we have succesfully created and broadcasted a Bitcoin transaction from scratch ([Proof](https://blockstream.info/testnet/tx/d1f770cdfe980eca99c18c52598fad6a1f68b8a59444e539722198914694b73e))!!! I was really happy when this finally worked. 

When I was encountering difficulties with the `sign()` function, I ran the equivalent Python code and broadcasted a couple of transactions using the hex encoded output in the Python implementation. Third time's the charm they say. Since I had three 50 000 satoshi UTXOs in the second address I created, I figured it would be interesting to try and craft a consolidation transaction. Just kidding, I did this accidentally when trying to send all the tBTC from both addresses back to the faucet. I will show you this happy little accident in the next part.

## Part 3: Consolidation Transaction

With three 50 000 Satoshi UTXOs in the second wallet and the remainder in the first, my intention was to create a transaction using all UTXOs and sending it to a tBTC faucet. So to accomplish this, I created four `TxIn` and one `TxOut`. 
```
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
```
I only created one `TxOut` as I didn't want any change and wanted to spend the entire amount minus the transaction fee. Next I created the locking script for `TxOut`. To do this, I navigated the block explorer to find the address page of the tBTC faucet address and copied it's pubkey hash. On proper Bitcoin wallet software, all you need is the recipient address, your wallet will decode the address into a pubkey hash.
```
	out_pkb_hash, _ := hex.DecodeString("344a0f48ca150ec2b903817660b9b68b13a67026")
	out_cmds := []byte{118, 169}
	out_cmds = append(out_cmds, []byte{byte(len(out_pkb_hash))}...)
	out_cmds = append(out_cmds, out_pkb_hash...)
	out_cmds = append(out_cmds, []byte{136, 172}...)
	out_script := ByteScript{
		cmds: out1_cmds, //Ahhhh a happy little accident, I've created a consolidation tx
	}
	tx_out.script_pubkey = out_script
```
As you can see, I accidentally referenced `out1_cmds` and not `out_cmds` which means the UTXO of this transaction will be locked to our second address and not the tBTC faucet address. Oh well, this now becomes a consolidation transaction where we destroy many UTXOs and create one larger UTXO. After creating the locking script, I then created the unlock scripts for all the inputs.
```
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
```
Then I assembled all these `TxIn` and `TxOut` into a `Tx` struct
```
	new_tx := Tx{
		version:  1,
		tx_ins:   []TxIn{tx_in1, tx_in2, tx_in3, tx_in4},
		tx_outs:  []TxOut{tx_out},
		locktime: 0,
	}
```
Here is where things start to differ. Because we have multiple `TxIn` which are unlocked by two different addresses, we must create four different messages to sign four different times. Otherwise we will not solve the unlock script for each `TxIn` referenced output.
```
	msg := new_tx.TxEncode(0)
	new_sig := sign(priv_key, btc_gen, msg)
	new_sig_bytes := new_sig.sig_encode()
	new_sig_bytes = append(new_sig_bytes, byte('\x01'))
	var new_script_sig_cmds []byte
	new_script_sig_cmds = append(new_script_sig_cmds, []byte{byte(len(new_sig_bytes))}...)
	new_script_sig_cmds = append(new_script_sig_cmds, new_sig_bytes...)
	new_script_sig_cmds = append(new_script_sig_cmds, []byte{byte(len(pubkey_bytes))}...)
	new_script_sig_cmds = append(new_script_sig_cmds, pubkey_bytes...)
	new_tx.tx_ins[0].script_sig = ByteScript{
		cmds: new_script_sig_cmds,
	}
```
So the first address signs the `TxIn` at the 0th index. Then the second address will sign the rest of them.
```
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
```
And then we encode our transaction and broadcast it.
```
	new_tx_bytes := new_tx.TxEncode(-1)
	fmt.Printf("%s\n", hex.EncodeToString(new_tx_bytes))
	new_tx_id := sha256(sha256(new_tx_bytes))
	reverse(new_tx_id)
	fmt.Printf("tx_id: %s\n", hex.EncodeToString(new_tx_id))
```
You can see what this consolidation transaction looks like [here](https://blockstream.info/testnet/tx/ac461b593b6b825117c33421947ede73d4f196f8250b5d82d279ef7918741ee5)

## Part 4: Returning Funds to the Faucet

A criticism I have towards Karpathy's blog post, is not encouraging those following along to return the tBTC funds to the faucet. This is important if we wish to continue having accessible funds from these faucets. So I will show how I constructed the transaction that did just that.

Again we start by defining `TxIn` and `TxOut`.
```
	prev_tx, _ = hex.DecodeString("ac461b593b6b825117c33421947ede73d4f196f8250b5d82d279ef7918741ee5")
	new_tx_in := NewTxIn(prev_tx, 0)
	new_tx_out := TxOut{
		amount: 1101850,
	} 
```
Since we just consolidated all our UTXOs into one output, we will have one `TxIn` and one `TxOut`. Pretty simple. Then we create the locking script for `TxOut`. Remember to lock it to the pubkey hash of the faucet and not one of your own pubkey hashes.
```
	new_out_script := ByteScript{
		cmds: out_cmds, //now we use the out script we created earlier
	}
	new_tx_out.script_pubkey = new_out_script
```
Then define the unlock script and assemble it all into a `Tx` struct.
```
	new_tx_in.prev_tx_script_pubkey = ByteScript{
		cmds: in2_cmds,
	}
	final_tx := Tx{
		version:  1,
		tx_ins:   []TxIn{new_tx_in},
		tx_outs:  []TxOut{new_tx_out},
		locktime: 0,
	}
```
This time, we will only create one encoded message and sign it once with our second address.
```
	final_msg := final_tx.TxEncode(0)
	final_sig := sign(priv_key2, btc_gen, final_msg)
	final_sig_bytes := final_sig.sig_encode()
	final_sig_bytes = append(final_sig_bytes, byte('\x01'))
```
Added the encoded signature to `Tx` and encoded it once more for broadcasting
```
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
```
And here is [that transaction](https://blockstream.info/testnet/tx/9bc54a7672a39fcf7e599c8474a01918ae1aa65aa5738a070b7a8dd77e52f503).

I hope you all learned something by reading through this blog post.