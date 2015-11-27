package fibvec

import (
	"fmt"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestAddGet(t *testing.T) {
	vec := NewVector()
	values := make([]int, 1e5)
	for i := range values {
		v := rand.Intn(MaxValue)

		values[i] = v
		vec.Add(v)
	}

	for i, v := range values {
		if !assert.Equal(t, v, vec.Get(i)) {
			break
		}
	}
}

func TestAddGetNegative(t *testing.T) {
	vec := NewVector()
	values := []int{MinValue, -3, -2, -1, 0, 1, 2, 3, MaxValue}
	for _, v := range values {
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
	values := make([]int, 1e3)
	for i := range values {
		v := rand.Intn(MaxValue)

		values[i] = v
		vec.Add(v)
	}

	for i := range values {
		if !assert.Equal(t, values[0:i+1], vec.GetValues(0, i+1)) {
			break
		}
	}

}

func TestEncodeDecode(t *testing.T) {
	vec := NewVector()
	values := make([]int, 1e5)
	for i := range values {
		v := rand.Intn(MaxValue)

		values[i] = v
		vec.Add(v)
	}

	data, _ := vec.GobEncode()
	nvec := NewVector()
	nvec.GobDecode(data)

	for i, v := range values {
		if !assert.Equal(t, v, nvec.Get(i)) {
			break
		}
	}
}

// TestAuxOverhead calculates the
// overhead of the rank and select
// auxilliary arrays for uint32 values.
func TestAuxOverhead(t *testing.T) {
	vec := NewVector()
	for i := 0; i < 1e5; i++ {
		v := int(rand.Uint32())
		vec.Add(v)
	}

	rawsize := float64(vec.bits.Size())
	overhead := float64(vec.Size()) - rawsize
	percentage := (overhead / rawsize) * 100

	fmt.Printf("=== OVERHEAD: %.2f%%\n", percentage)
}

// TestCompression calculates the
// space saved with respect to the
// raw size for random uint32 values.
func TestCompression(t *testing.T) {
	vec := NewVector()
	for i := 0; i < 1e5; i++ {
		v := int(rand.Uint32())
		vec.Add(v)
	}

	sizeofUint := int(unsafe.Sizeof(uint(0)))

	rawsize := float64(sizeofUint * 1e5)
	vecsize := float64(vec.Size())

	percentage := ((rawsize - vecsize) / rawsize) * 100
	fmt.Printf("=== COMPRESSION: %.2f%%\n", percentage)
}

func BenchmarkAdd(b *testing.B) {
	vec := NewVector()
	values := make([]int, b.N)
	for i := range values {
		values[i] = rand.Intn(MaxValue)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vec.Add(values[i])
	}
}

func BenchmarkGet(b *testing.B) {
	vec := NewVector()
	for i := 0; i < 1e5; i++ {
		vec.Add(rand.Intn(MaxValue))
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
