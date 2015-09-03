# fibvec

Package fibvec implements a vector that can store unsigned integers by first
converting them to their fibonacci encoded values before saving to a bit array.
This can save memory space (especially for small values) in exchange for slower
operations.

This implements *Random access to Fibonacci coded sequences* described in
[Simple Random Access Compression][1] by Kimmo Fredriksson and Fedor Nikitin
with auxilliary structure to support fast select operations from [Fast, Small,
Simple Rank/Select on Bitmaps][2]. The encoding algorithm was taken from [Fast
Fibonacci Encoding Algorithm][3] and the decoding algorithm was taken
from [Fast decoding algorithm for variable-lengths codes][4] and [The Fast
Fibonacci Decompression Algorithm][5].

[1]: http://cs.uef.fi/~fredriks/pub/papers/fi09.pdf
[2]: http://dcc.uchile.cl/~gnavarro/ps/sea12.1.pdf
[3]: http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.232.682&rep=rep1&type=pdf#page=78
[4]: http://www.researchgate.net/profile/Vaclav_Snasel/publication/220311886_Fast_decoding_algorithms_for_variable-lengths_codes/links/00b7d52cb1363228a1000000.pdf
[5]: http://arxiv.org/pdf/0712.0811

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
BenchmarkFibEnc      1000000          1417 ns/op
BenchmarkFibDec      3000000           517 ns/op
BenchmarkAdd         1000000          1521 ns/op
BenchmarkGet         2000000           650 ns/op
```
