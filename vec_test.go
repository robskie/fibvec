package fibvec

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestAddGet(t *testing.T) {
	vec := NewVector()
	values := make([]uint, 1e6)
	for i := range values {
		v := uint(rand.Intn(MaxValue))
		values[i] = v

		vec.Add(v)
	}

	for i, v := range values {
		if !assert.Equal(t, v, vec.Get(i)) {
			break
		}
	}
}

func TestGetValues(t *testing.T) {
	vec := NewVector()
	values := make([]uint, 1e3)
	for i := range values {
		v := uint(rand.Intn(MaxValue))
		values[i] = v

		vec.Add(v)
	}

	for i := range values {
		if !assert.Equal(t, values[0:i+1], vec.GetValues(0, i+1)) {
			break
		}
	}

}

// Tests the overhead of the rank
// and select auxilliary arrays.
func TestAuxOverhead(t *testing.T) {
	vec := NewVector()
	for i := 0; i < 1e6; i++ {
		v := uint(rand.Intn(MaxValue))
		vec.Add(v)
	}

	rawsize := float64(vec.bits.Size())
	overhead := float64(vec.Size()) - rawsize
	percentage := (overhead / rawsize) * 100

	fmt.Printf("=== OVERHEAD: %.2f%%\n", percentage)
}

func TestCompression32(t *testing.T) {
	vec := NewVector()
	for i := 0; i < 1e6; i++ {
		v := uint(rand.Intn(math.MaxUint32))
		vec.Add(v)
	}

	sizeofUint := int(unsafe.Sizeof(uint(0)))

	rawsize := float64(sizeofUint * 1e6)
	vecsize := float64(vec.Size())

	percentage := (rawsize - vecsize) / rawsize * 100
	fmt.Printf("=== COMPRESSION: %.2f%%\n", percentage)
}

func TestCompression(t *testing.T) {
	vec := NewVector()
	for i := 0; i < 1e6; i++ {
		v := uint(rand.Intn(MaxValue))
		vec.Add(v)
	}

	sizeofUint := int(unsafe.Sizeof(uint(0)))

	rawsize := float64(sizeofUint * 1e6)
	vecsize := float64(vec.Size())

	percentage := (rawsize - vecsize) / rawsize * 100
	fmt.Printf("=== COMPRESSION: %.2f%%\n", percentage)
}

func BenchmarkAdd(b *testing.B) {
	vec := NewVector()
	values := make([]uint, b.N)
	for i := range values {
		values[i] = uint(rand.Intn(MaxValue))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vec.Add(values[i])
	}
}

func BenchmarkGet(b *testing.B) {
	vec := NewVector()
	for i := 0; i < 1e6; i++ {
		v := uint(rand.Intn(MaxValue))
		vec.Add(v)
	}

	idx := make([]int, b.N)
	for i := range idx {
		idx[i] = rand.Intn(vec.length)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vec.Get(idx[i])
	}
}
