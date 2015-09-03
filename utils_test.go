package fibvec

import (
	"math/rand"
	"testing"

	"github.com/robskie/bit"
	"github.com/stretchr/testify/assert"
)

func TestFibEncDec(t *testing.T) {
	array := bit.NewArray(0)

	num := int(1e5)
	values := make([]uint, num)
	for i := range values {
		v := uint(rand.Int63())
		values[i] = v

		fc, lfc := fibencode(v)
		for _, f := range fc[:len(fc)-1] {
			array.Add(f, 64)
			lfc -= 64
		}
		array.Add(fc[len(fc)-1], lfc)
	}
	array.Add(0x3, 16)

	bytes := byteSliceFromUint64Slice(array.Bits())
	result := fibdecode(bytes, num)
	for i, v := range values {
		if !assert.Equal(t, v, result[i]) {
			break
		}
	}
}

func BenchmarkFibEnc(b *testing.B) {
	val := make([]uint, b.N)
	for i := range val {
		val[i] = uint(rand.Int63())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fibencode(val[i])
	}
}

func BenchmarkFibDec(b *testing.B) {
	enc := make([][]byte, 1e5)
	for i := range enc {
		v := uint(rand.Int63())
		fc, lfc := fibencode(v)

		array := bit.NewArray(0)
		for _, f := range fc[:len(fc)-1] {
			array.Add(f, 64)
			lfc -= 64
		}
		array.Add(fc[len(fc)-1], lfc)
		array.Add(0x3, 16)

		enc[i] = byteSliceFromUint64Slice(array.Bits())
	}

	idx := make([]int, b.N)
	for i := range idx {
		idx[i] = rand.Intn(len(enc))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fibdecode(enc[idx[i]], 1)
	}
}
