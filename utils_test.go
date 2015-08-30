package fibvec

import (
	"math/rand"
	"testing"

	"github.com/robskie/bit"
	"github.com/stretchr/testify/assert"
)

func TestFibDecode(t *testing.T) {
	array := bit.NewArray(0)

	// Populate array
	num := int(1e6)
	values := make([]uint, num)
	for i := 0; i < num; i++ {
		v := uint(rand.Intn(MaxValue) + 2)
		values[i] = v - 2

		fc, lfc := fencode(v)
		array.Add(fc, lfc)

		// Add padding at array boundary
		if (array.Len()-1)&63 == 62 {
			array.Add(0x3, 2)
		}
	}

	// Add terminating bits plus padding.
	array.Add(0x3, 16)

	bits := array.Bits()
	bytes := byteSliceFromUint64Slice(bits)

	result := fibdecode(bytes, num)
	for i, v := range values {
		if !assert.Equal(t, v, result[i]) {
			break
		}
	}
}

func BenchmarkFibDecode(b *testing.B) {
	values := make([][]byte, 1e3)
	for i := range values {
		v := uint(rand.Intn(MaxValue) + 2)
		fc, lfc := fencode(v)

		array := bit.NewArray(0)
		array.Add(fc, lfc)
		array.Add(0x3, 16)

		values[i] = byteSliceFromUint64Slice(array.Bits())
	}

	idx := make([]int, b.N)
	for i := range idx {
		idx[i] = rand.Intn(len(values))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fibdecode(values[idx[i]], 1)
	}
}
