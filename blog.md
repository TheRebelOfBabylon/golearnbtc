# Introduction
The two goals I had in mind when deciding to translate the from-scratch, Python implementation of BTC that Andrej Karpathy made are: to learn a new coding language, namely Go, and
get a technical familiarity with Bitcoin.

With these goals established, please keep in mind this code won't be pretty, and will evolve over the course of this blog post. This blog post will highlight some of the challenges I faced
when translating this code that I encountered and little nuggets of information I discovered along the way. I hope that whoever reads this learns something new.

## Part 1: Elliptic Curve Cryptography

I followed along the blog post that Andrej Karpathy put together to translate this code so I will follow the order of his blog post in a similar fashion. This first thing we need
to make a Bitcoin transaction is a set of private and public keys. Cryptography is quite a complex topic that I don't feel qualified to explain; all we need to know is that a private key is 
a secret number that is should NEVER be shared as this number is used to mathematically prove ownership over your Bitcoin. A public key is a set of points, that are on a given elliptic curve
and related to the private key. You derive a public key from a private key but it is incredibly difficult to do the reverse operation. This is the heart of security with modern cryptography.

So to begin, Karpathy creates a class called `Curve` which has three arguments, `p`, `a` and `b`. He also adds a decorator to the class delartion `@dataclass`. In Go, there are no classes,
the closest equivalent in Go is a `struct`. Something else to keep in mind is you cannot give a default value to a an argument in a `struct`. So we define the arguments of the struct in `main()`.
Finally, go is a strictly typed language by design where as Python is dynamically typed. Large integers (larger than 64 bits) are supported within the `int` type. In Go, we must use the `math/big` module which offers
a `big.Int` type to serve our needs here. `big.Int` methods return whatever the result is to the object that called the method itself and normally returns a point to a `big.Int` type, so our struct will
want `*big.int` for the `p` argument. I added a print statement at the end of main just so go wouldn't complain about declaring `btc_curve` and not using it. 

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
Something worth mentionning about functions/methods in Go. As it is strictly typed, you have to define the types for any input arguments as well as all output values. To make a function a method of a specific struct, you have to include the equivalent of `self` in Python before the method name. In the example above, the function becomes a method with `(p Point)` which tells Go "This funciton is a method of Point". 

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
And with that we've created our private key. Notice that we can't use standard go comparison operators between `big.Int` types. Instead, the `.Cmp()` method is provided. Also notice that no matter how many times we run this script, the private key doesn't change. They are deterministically derived, in this case from the phrase "btc is the future". This would not be a strong phrase for a private key. Most well built Bitcoin wallets generate a list of 12-24 words from 1024 possible choices which are then used to derive the private key. This is a very secure method of deriving a private key as each combination of 12-24 words are one in 12^1014 - 24^1024 possibilites (well not quite as there should not be any repeating words and the last worm is a checksum and not random). But this will suit our purposes just fine.

Now to create the public key, we have to add the Generator point to itself private key number of times. But this isn't a simple multiplication as a point has an x and y value, the private key is just a large integer. It also isn't like multiplying a vector by a scalar, atleast that was my gut. It wouldn't be much of a secret key if you could just divide the public key by the generator point and get a private key. So we now create a set of functions to perform this special operation.

We define a point called `INF` which is meant to be like a point at infinity. In Python, this point is defined using the `None` type to which there is no equivalent in Golang. I first tried defining `INF` as `new(Point)`. From what I read on instantiating a variable but not assigning it a value in Go, the variable automatically assumes a default value of `0` or `""` or `false`. I'm not sure if this is the case for `big.Int` and so I manually defined zero values for our `INF` point.
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


