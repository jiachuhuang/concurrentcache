package concurrentcache

func MurmurHash2(data string) uint32 {
	var h, k uint32
	l := uint32(len(data))
	h = 0 ^ l
	for l >= 4 {
		k = uint32(data[0])
		k |= uint32(data[1]) << 8
		k |= uint32(data[2]) << 16
		k |= uint32(data[3]) << 24

		k *= 0x5bd1e995
		k ^= k >> 24
		k *= 0x5bd1e995

		h *= 0x5bd1e995
		h ^= k

		data = data[4:]
		l = uint32(len(data))
	}
	switch l {
	case 3:
		h ^= uint32(data[2]) << 16
		h ^= uint32(data[1]) << 8
		h ^= uint32(data[0])
		h *= 0x5bd1e995
	case 2:
		h ^= uint32(data[1]) << 8
		h ^= uint32(data[0])
		h *= 0x5bd1e995
	case 1:
		h ^= uint32(data[0])
		h *= 0x5bd1e995
	}
	h ^= h >> 13
	h *= 0x5bd1e995
	h ^= h >> 15
	return h
}
