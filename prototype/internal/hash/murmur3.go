package hash

import (
	"encoding/binary"
)

func rotl32(x uint32, r byte) uint32 {
	return (x << r) | (x >> (32 - r))
}

func getblock32(bytes []byte) uint32 {
	return binary.LittleEndian.Uint32(bytes)
}

func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func Murmur3(input []byte) (digest []byte) {

	var seed uint32 = 0x3242 // whatever
	length := uint32(len(input))
	nblocks := length / 16

	var h1 uint32 = seed
	var h2 uint32 = seed
	var h3 uint32 = seed
	var h4 uint32 = seed

	// the following are magic numbers used as constants for MurmurHash3 taken from:
	// https://github.com/aappleby/smhasher/blob/0ff96f7835817a27d0487325b6c16033e2992eb5/src/MurmurHash3.h
	const c1 uint32 = 0x239b961b
	const c2 uint32 = 0xab0e9789
	const c3 uint32 = 0x38b34ae5
	const c4 uint32 = 0xa1e38b93

	// body
	body := input[:nblocks*16]
	currIndex := nblocks*16 - 1
	for currIndex > 0 && nblocks != 0 {
		var k1 uint32 = getblock32(body[currIndex-4:])
		var k2 uint32 = getblock32(body[currIndex-8:])
		var k3 uint32 = getblock32(body[currIndex-12:])
		var k4 uint32 = getblock32(body[currIndex-16:])

		// k1 operations
		k1 *= c1
		k1 = rotl32(k1, 15)
		k1 *= c2
		h1 ^= k1
		h1 = rotl32(h1, 19)
		h1 += h2
		h1 = h1*5 + 0x561ccd1b

		// k2 operations
		k2 *= c2
		k2 = rotl32(k2, 16)
		k2 *= c3
		h2 ^= k2
		h2 = rotl32(h2, 17)
		h2 += h3
		h2 = h2*5 + 0x0bcaa747

		// k3 operations
		k3 *= c3
		k3 = rotl32(k3, 17)
		k3 *= c4
		h3 ^= k3
		h3 = rotl32(h3, 15)
		h3 += h4
		h3 = h3*5 + 0x96cd1c35

		// k4 operations
		k4 *= c4
		k4 = rotl32(k4, 18)
		k4 *= c1
		h4 ^= k4
		h4 = rotl32(h4, 13)
		h4 += h1
		h4 = h4*5 + 0x32ac3b17

		currIndex -= 32
	}

	// tail
	var tail []byte
	if length%16 != 0 {
		tail = input[nblocks*16:]
	}

	var k1 uint32
	var k2 uint32
	var k3 uint32
	var k4 uint32

	switch length & 15 {
	case 15:
		k4 ^= uint32(tail[14]) << 16
		fallthrough
	case 14:
		k4 ^= uint32(tail[13]) << 8
		fallthrough
	case 13:
		k4 ^= uint32(tail[12]) << 0
		k4 *= c4
		k4 = rotl32(k4, 18)
		k4 *= c1
		h4 ^= k4
		fallthrough
	case 12:
		k3 ^= uint32(tail[11]) << 24
		fallthrough
	case 11:
		k3 ^= uint32(tail[10]) << 16
		fallthrough
	case 10:
		k3 ^= uint32(tail[9]) << 8
		fallthrough
	case 9:
		k3 ^= uint32(tail[8]) << 0
		k3 *= c3
		k3 = rotl32(k3, 17)
		k3 *= c4
		h3 ^= k3
		fallthrough
	case 8:
		k2 ^= uint32(tail[7]) << 24
		fallthrough
	case 7:
		k2 ^= uint32(tail[6]) << 16
		fallthrough
	case 6:
		k2 ^= uint32(tail[5]) << 8
		fallthrough
	case 5:
		k2 ^= uint32(tail[4]) << 0
		k2 *= c2
		k2 = rotl32(k2, 16)
		k2 *= c3
		h2 ^= k2
		fallthrough
	case 4:
		k1 ^= uint32(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0]) << 0
		k1 *= c1
		k1 = rotl32(k1, 15)
		k1 *= c2
		h1 ^= k1
	}

	// finalization

	h1 ^= length
	h2 ^= length
	h3 ^= length
	h4 ^= length

	h1 += h2
	h1 += h3
	h1 += h4
	h2 += h1
	h3 += h1
	h4 += h1

	h1 = fmix32(h1)
	h2 = fmix32(h2)
	h3 = fmix32(h3)
	h4 = fmix32(h4)

	h1 += h2
	h1 += h3
	h1 += h4
	h2 += h1
	h3 += h1
	h4 += h1

	output := make([]byte, 16)
	binary.LittleEndian.PutUint32(output, h1)
	binary.LittleEndian.PutUint32(output[4:], h2)
	binary.LittleEndian.PutUint32(output[8:], h3)
	binary.LittleEndian.PutUint32(output[12:], h4)
	return output
}
