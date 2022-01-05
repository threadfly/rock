package base

import (
	"hash/crc32"
)

func StringHashCode(s string) int {
	v := int(crc32.ChecksumIEEE(StringToBytes(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}

func StringHashCodeV2(s string) int {
	byts := StringToBytes(s)
	offset := 0
	if len(byts) > 32 {
		offset = len(byts) - 32
	}

	v := int(crc32.ChecksumIEEE(byts[offset:]))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}
