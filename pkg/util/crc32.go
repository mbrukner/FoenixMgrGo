package util

// CalculateCRC32 calculates a CRC32 checksum using the ZIP polynomial
// This matches the mycrc() function from the Python implementation
func CalculateCRC32(data []byte) uint32 {
	const poly = 0xEDB88320
	crc := uint32(0)

	for _, b := range data {
		crc ^= uint32(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ poly
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}
