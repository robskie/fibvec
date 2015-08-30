package fibvec

import (
	"reflect"
	"unsafe"
)

// fib contains the ith fibonacci number.
// ie., fib[0]=1, fib[1]=1, fib[2]=2 ...
var fib [64]uint

// vf1 is the right shifted value of i,
// ie. V(F(n) >>f 1) where i is V(F(n)).
// This has only 55 values because the
// maximum encoded value that can fit in
// a byte is 54.
var vf1 [55]uint8

var fibTable1 [256]*fibRecord
var fibTable2 [256]*fibRecord

// MaxValue is the maximum
// value that can be stored.
const MaxValue = 10610209857720

type fibRecord struct {
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

func init() {
	fib[0] = 1
	fib[1] = 1
	for i := 2; i < 64; i++ {
		fib[i] = fib[i-1] + fib[i-2]
	}

	// Create decoding table
	for i := 0; i < 256; i++ {
		fibTable1[i] = createRecord(uint8(i), 8)
		fibTable2[i] = fibTable1[i]

		if i&0x80 != 0 {
			fibTable2[i] = createRecord(uint8(i<<1), 7)
		}
	}

	// Create V(F(n) >>f 1) table
	for i := uint(1); i < 55; i++ {
		fc, lfc := fencode(i)
		fc = fc << uint(64-lfc)

		bytes := *(*[8]byte)(unsafe.Pointer(&fc))
		v := lfibshift1(bytes[7])

		vf1[i] = v
	}
}

func createRecord(bits, length uint8) *fibRecord {
	rec := &fibRecord{}

	var dec uint8
	var shift uint8
	var complete bool
	var processed uint8
	for bits != 0 {
		dec, shift, complete = fdecode(bits)

		rec.shift = shift
		processed += shift
		rec.numbers = append(rec.numbers, dec)

		bits <<= shift
	}

	if processed < length {
		if complete {
			rec.shift = length - processed
			rec.numbers = append(rec.numbers, 0)
		} else {
			rec.shift += length - processed
		}
	} else if complete { // processed >= length && complete
		rec.shift = 0
	}

	// Reverse numbers
	for i, j := 0, len(rec.numbers)-1; i < j; i++ {
		rec.numbers[i], rec.numbers[j] = rec.numbers[j], rec.numbers[i]
		j--
	}

	if rec.shift > 0 {
		if len(rec.numbers) > 0 {
			rec.incomplete = rec.numbers[0]
			rec.numbers = rec.numbers[1:]
		} else {
			rec.incomplete = 0
		}
	}

	return rec
}

// fdecode decodes the first fibonacci code in bits
// starting from the most significant bit and returns
// the decoded value, its size, and if the decoded value
// is complete.
func fdecode(bits uint8) (uint8, uint8, bool) {
	n := uint8(0)
	pbit := uint8(0)
	shift := uint8(0)

	for bits != 0 {
		shift++
		bit := bits & 0x80

		if bit&pbit != 0 {
			return n, shift, true
		}

		if bit != 0 {
			n += uint8(fib[shift])
		}

		pbit = bit
		bits <<= 1
	}

	return n, shift, false
}

// lfibshift1 computes V(F(n) >>f 1) where bits is F(n).
func lfibshift1(bits uint8) uint8 {
	n := uint8(0)
	pbit := uint8(0)
	shift := uint8(0)

	for bits != 0 {
		bit := bits & 0x80

		if bit&pbit != 0 {
			return n
		}

		if bit != 0 {
			n += uint8(fib[shift])
		}

		shift++
		pbit = bit
		bits <<= 1
	}

	return n
}

// fbuffer is the storage for
// incomplete undecoded values.
var fbuffer = make([]byte, 0, 16)

// fibencode returns the fibonacci encoded
// value of n and its size.
func fibencode(n uint) (uint64, int) {
	// Add 2 to n so that the minimum encoded
	// value would be 011. This is to make sure
	// that there is no more than 3 consecutive
	// 1s in the bit array which makes it easier
	// to count pair of 1s.
	n += 2

	return fencode(n)
}

// fibdecode decodes the input bytes given the
// number of decoded values to return.
//
// See Fast decoding algorithms for
// variable-length codes by Platos et al.
func fibdecode(input []byte, count int) []uint {
	prevIn := input[0]
	fbuffer = fbuffer[:0]
	prevRec := fibTable1[prevIn]
	result := make([]uint, 0, count)

	for _, in := range input[1:] {
		startWithOne := false
		endWithOne := prevIn&0x80 != 0

		rec := fibTable1[in]
		if in&1 == 1 && rec.shift > 0 {
			startWithOne = true
			prevRec = fibTable2[prevIn]
		}
		prevIn = in

		shift := int(prevRec.shift)
		if shift > 0 {
			fbuffer = append(fbuffer, prevRec.incomplete)
		}

		dec := uint(0)
		for _, num := range prevRec.numbers {
			if shift == 0 {
				dec = decodeBuffer(8)
			} else {
				dec = decodeBuffer(shift)
			}

			if dec > 1 {
				// Subtract 2 to cancel out the
				// previously added during encoding
				result = append(result, dec-2)
				if len(result) == count {
					return result
				}
			}

			shift = 0
			fbuffer = append(fbuffer, num)
		}

		if startWithOne && endWithOne {
			dec = decodeBuffer(7)

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

func decodeBuffer(shift int) uint {
	n := len(fbuffer)
	if n == 0 {
		return 0
	}

	sum := uint(fbuffer[n-1])
	for i := n - 2; i >= 0; i-- {
		fb := fbuffer[i]
		sum += fibshift(uint(fb), shift)
		shift += 8
	}

	fbuffer = fbuffer[:0]
	return sum
}

// fibshift performs fibonacci left shift on vn.
// ie. V(F(n) <<f k) where vn is V(F(n))
func fibshift(vn uint, shift int) uint {
	return (fib[shift] * vn) + (fib[shift-1] * uint(vf1[vn]))
}

// fencode returns the fibonacci coded value of n
// and its size. Note that the returned value is
// reversed which means the two LSB is always 11.
func fencode(n uint) (uint64, int) {
	p := 1
	for fib[p] <= n {
		p++
	}
	m := uint64(1 << uint(p))
	p--

	fn := m
	lfn := 0
	for p >= 0 {
		fn >>= 1
		if fib[p] <= n {
			fn |= m
			n = n - fib[p]
		}

		lfn++
		p--
	}

	return fn, lfn
}

func byteSliceFromUint64Slice(bits []uint64) []byte {
	sh := &reflect.SliceHeader{}
	sh.Cap = cap(bits) * 8
	sh.Len = len(bits) * 8
	sh.Data = (uintptr)(unsafe.Pointer(&bits[0]))
	bytes := *(*[]uint8)(unsafe.Pointer(sh))

	return bytes
}
