# fibvec

Package fibvec implements a vector that can store unsigned integers by first
converting them to their fibonacci encoded values before saving to a bit array.
This can save memory space (especially for small values) in exchange for slower
operations.

This implements *Random access to Fibonacci coded sequences* described in
[Simple Random Access Compression][1] by Kimmo Fredriksson and Fedor Nikitin
with auxilliary structure to support fast select operations from [Fast, Small,
Simple Rank/Select on Bitmaps][2]. The fast fibonacci decoding algorithm was
taken from [Fast decoding algorithm for variable-lengths codes][3].

[1]: http://cs.uef.fi/~fredriks/pub/papers/fi09.pdf
[2]: http://dcc.uchile.cl/~gnavarro/ps/sea12.1.pdf
[3]: http://www.researchgate.net/profile/Vaclav_Snasel/publication/220311886_Fast_decoding_algorithms_for_variable-lengths_codes/links/00b7d52cb1363228a1000000.pdf

## Installation
```sh
go get github.com/robskie/fibvec
```

## API Reference

Godoc documentation can be found
[here](https://godoc.org/github.com/robskie/fibvec).

## Benchmarks

These benchmarks are done on a Core i5 at 2.3GHz. You can run these benchmarks
by typing ```go test github.com/robskie/fibvec -bench=.*``` from terminal.

```
BenchmarkFibDecode      5000000        306 ns/op
BenchmarkAdd            3000000        423 ns/op
BenchmarkGet            2000000        697 ns/op
```
