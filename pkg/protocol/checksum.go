package protocol

// calculateLRC computes the Longitudinal Redundancy Check (LRC) checksum
// LRC is calculated as the XOR of all bytes in the data
func calculateLRC(data []byte) byte {
	lrc := byte(0)
	for _, b := range data {
		lrc ^= b
	}
	return lrc
}

// verifyLRC checks if the LRC checksum is correct for the given data
// The last byte of data should be the LRC checksum
func verifyLRC(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	expectedLRC := data[len(data)-1]
	actualLRC := calculateLRC(data[:len(data)-1])

	return expectedLRC == actualLRC
}
