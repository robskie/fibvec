package fibvec

import (
	"math"
	"reflect"
	"unsafe"

	"github.com/robskie/bit"
)

// MaxValue is the maximum
// value that can be stored.
const MaxValue = math.MaxUint64 - 3

type decRecord struct {
	// shift contains the size of
	// the partially decoded value
	shift uint8

	// numbers contains the fully
	// decoded values from a byte
	numbers []uint8

	// incomplete contains
	// partially decoded value
	incomplete uint8
}

type encRecord struct {
	code   uint8
	length uint8
	nmin   uint
	nmax   uint
}

// rfibshift8 returns the right fibonacci shift
// of n., ie., V(F(n) >>f k) where n is V(F(n))
// as long as k is a multiple of 8.
func rfibshift8(n uint, shift int) uint {
	const phi = 1.618033989

	est := float64(n) / math.Pow(phi, float64(shift))
	res := uint(est + 0.5)
	res--

	rec := fencTable[shift/8][res]
	if n > rec.nmax {
		res++
	}

	return res

}

// fibencode encodes n to
// its fibonacci coded value.
//
// See Fast Fibonnaci Encoding Algorithm
// by Platos et al. for more details.
func fibencode(n uint) ([]uint64, int) {
	res := bit.NewArray(128)
	res.Add(1, 1)

	// Add 2 to n so that the minimum encoded
	// value would be 011 to make sure that there
	// is no more than 3 consecutive 1s in the bit
	// array which makes it easier to count pairs of 1s.
	n += 2

	k := 8
	f := fib[1:]
	var code uint8
	var rec encRecord
	if n < f[k] {
		rec = fencTable[0][n]
		code = rec.code >> uint(8-rec.length)

		res.Add(uint64(code), int(rec.length))
		return res.Bits(), res.Len()
	}

	for n >= f[k+8] {
		k += 8
	}

	i := rfibshift8(n, k)
	rec = fencTable[k/8][i]
	code = rec.code >> uint(8-rec.length)

	res.Add(uint64(code), int(rec.length))
	n -= rec.nmin

	for k > 8 {
		k -= 8
		if n >= f[k] {
			i = rfibshift8(n, k)
			rec = fencTable[k/8][i]

			res.Add(uint64(rec.code), 8)
			n -= rec.nmin
		} else {
			res.Add(0, 8)
		}
	}

	rec = fencTable[0][n]
	res.Add(uint64(rec.code), 8)

	return res.Bits(), res.Len()
}

// fibdecode decodes the input bytes given the
// number of decoded values to return.
//
// See Fast decoding algorithms for variable-length codes
// and Fast Fibonacci Decompression Algorithm by Platos et al.
func fibdecode(input []byte, count int) []uint {
	prevIn := input[0]
	fbuffer := make([]byte, 0, 16)
	prevRec := fdecTable[0][prevIn]
	result := make([]uint, 0, count)

	for _, in := range input[1:] {
		startWithOne := false
		endWithOne := prevIn&0x80 != 0

		rec := fdecTable[0][in]
		if in&1 == 1 && rec.shift > 0 {
			startWithOne = true
			prevRec = fdecTable[1][prevIn]
		}
		prevIn = in

		shift := int(prevRec.shift)
		if shift > 0 {
			fbuffer = append(fbuffer, prevRec.incomplete)
		}

		dec := uint(0)
		for _, num := range prevRec.numbers {
			if shift == 0 {
				dec = decodeBuffer(fbuffer, 8)
			} else {
				dec = decodeBuffer(fbuffer, shift)
			}
			fbuffer = fbuffer[:0]

			if dec > 1 {
				// Subtract 2 to cancel out
				// what is added during encoding
				result = append(result, dec-2)
				if len(result) == count {
					return result
				}
			}

			shift = 0
			fbuffer = append(fbuffer, num)
		}

		if startWithOne && endWithOne {
			dec = decodeBuffer(fbuffer, 7)
			fbuffer = fbuffer[:0]

			if dec > 1 {
				result = append(result, dec-2)
				if len(result) == count {
					return result
				}
			}
		}

		prevRec = rec
	}

	return result
}

func decodeBuffer(fbuffer []byte, shift int) uint {
	n := len(fbuffer)
	if n == 0 {
		return 0
	}

	sum := uint(fbuffer[n-1])
	for i := n - 2; i >= 0; i-- {
		fb := fbuffer[i]
		sum += lfibshift(uint(fb), shift)
		shift += 8
	}

	return sum
}

// lfibshift performs fibonacci left shift
// on n, ie., V(F(n) <<f k) where n is V(F(n))
func lfibshift(n uint, shift int) uint {
	return (fib[shift] * n) + (fib[shift-1] * uint(vf1[n]))
}

func byteSliceFromUint64Slice(bits []uint64) []byte {
	sh := &reflect.SliceHeader{}
	sh.Cap = cap(bits) * 8
	sh.Len = len(bits) * 8
	sh.Data = (uintptr)(unsafe.Pointer(&bits[0]))
	bytes := *(*[]uint8)(unsafe.Pointer(sh))

	return bytes
}
