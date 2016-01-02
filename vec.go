// Package fibvec provides a vector that can store unsigned integers by first
// converting them to their fibonacci encoded values before saving to a bit
// array. This can save memory space (especially for small values) in exchange
// for slower operations.
package fibvec

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"unsafe"

	"github.com/robskie/bit"
)

const (
	// These variables affects the size and
	// speed of the vector. Lower values means
	// larger size but faster Gets and vice versa.

	// sr is the rank sampling block size.
	// This represents the number of bits in
	// each rank sampling block.
	sr = 512

	// ss is the number of 1s in each select
	// sampling block. Note that the number of
	// bits in each block varies.
	ss = 640
)

// Vector represents a container for unsigned integers.
type Vector struct {
	bits *bit.Array

	// ranks[i] is the number of 11s
	// from 0 to index (i*sr)-1
	ranks []int

	// indices[i] points to the
	// beginning of the uint64 (LSB)
	// that contains the (i*ss)+1th
	// pair of bits.
	indices []int

	popcount int

	length      int
	initialized bool
}

// Initialize vector
func (v *Vector) init() {
	v.bits = bit.NewArray(0)
	v.ranks = make([]int, 1)
	v.indices = make([]int, 1)

	// Add terminating bits
	v.bits.Add(0x3, 3)

	v.initialized = true
}

// NewVector creates a new vector.
func NewVector() *Vector {
	vec := &Vector{}
	vec.init()
	return vec
}

// Add adds an integer to the vector.
func (v *Vector) Add(n int) {
	if n > MaxValue || n < MinValue {
		panic("fibvec: input is not in the range of encodable values")
	} else if !v.initialized {
		v.init()
	}

	// Convert to sign-magnitude representation
	// so that "small" negative numbers such as
	// -1, -2, -3... can be encoded
	nn := toSignMagnitude(n)

	v.length++
	idx := v.bits.Len() - 3
	fc, lfc := fibencode(nn)
	size := lfc

	if lfc > 64 {
		v.bits.Insert(idx, fc[0], 64)
		lfc -= 64

		for _, f := range fc[1 : len(fc)-1] {
			v.bits.Add(f, 64)
			lfc -= 64
		}
		v.bits.Add(fc[len(fc)-1], lfc)
	} else {
		v.bits.Insert(idx, fc[0], lfc)
	}

	// Add bit padding so that pairs
	// of 1 (11s) don't get separated
	// by array boundaries.
	if (v.bits.Len()-1)&63 == 62 {
		v.bits.Add(0x3, 2)
	}

	v.popcount++
	vlen := v.bits.Len()

	lenranks := len(v.ranks)
	overflow := vlen - (lenranks * sr)
	if overflow > 0 {
		v.ranks = append(v.ranks, 0)
		v.ranks[lenranks] = v.popcount
		if size <= overflow {
			v.ranks[lenranks]--
		}
	}

	lenidx := len(v.indices)
	if v.popcount-(lenidx*ss) > 0 {
		v.indices = append(v.indices, 0)
		v.indices[lenidx] = idx ^ 0x3F
	}

	// Add terminating bits so that
	// the last value can be decoded
	v.bits.Add(0x3, 3)
}

// Get returns the value at index i.
func (v *Vector) Get(i int) int {
	if i >= v.length {
		panic("fibvec: index out of bounds")
	} else if i < 0 {
		panic("fibvec: invalid index")
	}

	idx := v.select11(i + 1)
	bits := v.bits.Bits()

	// Temporary store and
	// zero out extra bits
	aidx := idx >> 6
	bidx := idx & 63
	temp := bits[aidx]
	bits[aidx] &= ^((1 << uint(bidx)) - 1)

	// Transform to bytes
	bytes := byteSliceFromUint64Slice(bits)
	bytes = bytes[idx>>3:]

	// This makes sure that the last number is decoded
	if len(bytes) < 16 {
		bytes = append(bytes, []byte{0, 0}...)
	}
	result := fibdecode(bytes, 1)

	// Restore bits
	bits[aidx] = temp

	return result[0]
}

// GetValues returns the values from start to end-1.
func (v *Vector) GetValues(start, end int) []int {
	if end-start <= 0 {
		panic("fibvec: end must be greater than start")
	} else if start < 0 || end < 0 {
		panic("fibvec: invalid index")
	} else if end > v.length {
		panic("fibvec: index out of bounds")
	}

	idx := v.select11(start + 1)
	bits := v.bits.Bits()

	// Temporary store and
	// zero out extra bits
	aidx := idx >> 6
	bidx := idx & 63
	temp := bits[aidx]
	bits[aidx] &= ^((1 << uint(bidx)) - 1)

	// Transform to bytes
	bytes := byteSliceFromUint64Slice(bits)
	bytes = bytes[idx>>3:]

	// This makes sure that the last number is decoded
	if len(bytes) < 16 {
		bytes = append(bytes, []byte{0, 0}...)
	}
	results := fibdecode(bytes, end-start)

	// Restore bits
	bits[aidx] = temp

	return results
}

// Size returns the vector size in bytes.
func (v *Vector) Size() int {
	sizeofInt := int(unsafe.Sizeof(int(0)))

	size := v.bits.Size()
	size += len(v.ranks) * sizeofInt
	size += len(v.indices) * sizeofInt

	return size
}

// Len returns the number of values stored.
func (v *Vector) Len() int {
	return v.length
}

func checkErr(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}

	return nil
}

// GobEncode encodes this vector into gob streams.
func (v *Vector) GobEncode() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	err := checkErr(
		enc.Encode(v.bits),
		enc.Encode(v.ranks),
		enc.Encode(v.indices),
		enc.Encode(v.popcount),
		enc.Encode(v.length),
		enc.Encode(v.initialized),
	)

	if err != nil {
		err = fmt.Errorf("fibvec: encode failed (%v)", err)
	}

	return buf.Bytes(), err
}

// GobDecode populates this vector from gob streams.
func (v *Vector) GobDecode(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)

	v.bits = bit.NewArray(0)
	err := checkErr(
		dec.Decode(v.bits),
		dec.Decode(&v.ranks),
		dec.Decode(&v.indices),
		dec.Decode(&v.popcount),
		dec.Decode(&v.length),
		dec.Decode(&v.initialized),
	)

	if err != nil {
		err = fmt.Errorf("fibvec: decode failed (%v)", err)
	}

	return err
}

// select11 selects the ith 11 pair.
//
// Taken from "Fast, Small, Simple Rank/Select
// on Bitmaps" by Navarro et al., with some minor
// modifications.
func (v *Vector) select11(i int) int {
	const m = 0xC000000000000000

	j := (i - 1) / ss
	q := v.indices[j] / sr

	k := 0
	r := 0
	rq := v.ranks[q:]
	for k, r = range rq {
		if r >= i {
			k--
			break
		}
	}

	idx := 0
	rank := rq[k]
	vbits := v.bits.Bits()
	aidx := ((q + k) * sr) >> 6

	vbits = vbits[aidx:]
	for ii, b := range vbits {
		rank += popcount11_64(b)

		// If b ends with 11 and the next bits
		// starts with 1, then the 11 in b is
		// not the beginning of an encoded value,
		// but popcount11_64 has already counted
		// it so we need to subtract 1 to rank
		if b&m == m && vbits[ii+1]&1 == 1 {
			rank--
		}

		if rank >= i {
			idx = (aidx + ii) << 6
			overflow := rank - i
			popcnt := popcount11_64(b)
			if b&m == m && vbits[ii+1]&1 == 1 {
				popcnt--
			}

			idx += select11_64(b, popcnt-overflow)

			break
		}
	}

	return idx
}

// popcount11 counts the number of 11 pairs
// in v. This assumes that v doesn't contain
// more than 3 consecutive 1s. This assumption
// is satisfied since the minimum encoded value
// is 011.
func popcount11_64(v uint64) int {
	// Reduce cluster of 1s by 1.
	// This makes 11 to 01, 111 to 011,
	// and unsets all 1s.
	v &= v >> 1

	// Reduces all 11s to 10s
	// while maintaining all lone 1s.
	v &= ^(v >> 1)

	// Proceed to regular bit counting
	return bit.PopCount(v)
}

// select11 returns the index of the ith 11 pair.
func select11_64(v uint64, i int) int {
	// Same with popcount11
	v &= v >> 1
	v &= ^(v >> 1)

	// Perform regular select
	return bit.Select(v, i)
}
