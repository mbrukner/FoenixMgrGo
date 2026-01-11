package protocol

import "fmt"

// WriteBlock32 writes data to a machine requiring 32-bit alignment (68040/68060)
// If the address or data size is not 4-byte aligned, it performs a read-modify-write:
//  1. Align address down to 4-byte boundary
//  2. Read the aligned block from hardware memory
//  3. Modify the specific bytes within the aligned buffer
//  4. Write the entire aligned block back
func (dp *DebugPort) WriteBlock32(address uint32, data []byte) error {
	size := uint32(len(data))
	addressAlign := address % 4

	// If the block is already aligned, just write it directly
	if addressAlign == 0 && size%4 == 0 {
		_, err := dp.transfer(CMDWriteMem, address, data, 0)
		return err
	}

	// Otherwise, we need to perform read-modify-write for alignment
	adjustedAddress := address - addressAlign
	adjustedSize := size + addressAlign

	// Round size up to next multiple of 4
	sizeAlign := adjustedSize % 4
	if sizeAlign > 0 {
		adjustedSize += (4 - sizeAlign)
	}

	// Read the current contents from the machine's RAM
	block, err := dp.ReadBlock(adjustedAddress, uint16(adjustedSize))
	if err != nil {
		return fmt.Errorf("failed to read block for alignment: %w", err)
	}

	// Verify we got the expected amount of data
	if uint32(len(block)) != adjustedSize {
		return fmt.Errorf("read returned %d bytes, expected %d", len(block), adjustedSize)
	}

	// Copy the new data to the correct position within the buffer
	copy(block[addressAlign:], data)

	// Write the modified block back to the machine's RAM
	_, err = dp.transfer(CMDWriteMem, adjustedAddress, block, 0)
	if err != nil {
		return fmt.Errorf("failed to write aligned block: %w", err)
	}

	return nil
}
