package opns

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
)

// FuzzDecode tests the OpNS Decode function with random script bytes.
// The decoder should never panic regardless of input.
// Run with: go test -fuzz=FuzzDecode -fuzztime=10s
func FuzzDecode(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus with meaningful test cases

	// Empty script
	f.Add([]byte{})

	// Just the contract prefix (truncated)
	if len(contract) > 10 {
		f.Add(contract[:10])
	}

	// Contract prefix with some extra data
	prefixWithData := make([]byte, len(contract)+10)
	copy(prefixWithData, contract)
	f.Add(prefixWithData)

	// Contract prefix + state section marker
	prefixWithState := make([]byte, len(contract)+50)
	copy(prefixWithState, contract)
	// Add OP_RETURN OP_FALSE after contract
	prefixWithState[len(contract)] = script.OpRETURN
	prefixWithState[len(contract)+1] = script.OpFALSE
	f.Add(prefixWithState)

	// Random bytes that don't match contract prefix
	f.Add([]byte{0x00, 0x01, 0x02, 0x03, 0x04})
	f.Add([]byte{0xff, 0xff, 0xff})

	// Script with valid genesis outpoint structure
	genesisBytes := GENESIS().TxBytes()
	validState := make([]byte, len(contract)+100)
	copy(validState, contract)
	validState[len(contract)] = script.OpRETURN
	validState[len(contract)+1] = script.OpFALSE
	// Push genesis (36 bytes)
	validState[len(contract)+2] = 0x24 // push 36 bytes
	copy(validState[len(contract)+3:], genesisBytes)
	f.Add(validState)

	// Edge cases with various opcode patterns
	f.Add([]byte{0x4c, 0x00})             // OP_PUSHDATA1 with zero length
	f.Add([]byte{0x4d, 0x00, 0x00})       // OP_PUSHDATA2 with zero length
	f.Add([]byte{0x4e, 0x00, 0x00, 0x00}) // OP_PUSHDATA4 with zero length

	f.Fuzz(func(t *testing.T, data []byte) {
		// Create script from bytes - should never panic
		scr := script.NewFromBytes(data)

		// Decode should never panic, regardless of input
		_ = Decode(scr)
	})
}

// FuzzTestSolution tests the TestSolution proof-of-work validator.
// It should handle various nonce lengths and byte patterns without panicking.
// Run with: go test -fuzz=FuzzTestSolution -fuzztime=10s
func FuzzTestSolution(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus with various test cases

	// Standard 32-byte nonce
	f.Add(byte('a'), make([]byte, 32))

	// Empty nonce
	f.Add(byte('b'), []byte{})

	// Single byte nonce
	f.Add(byte('c'), []byte{0x00})

	// Very long nonce (longer than expected)
	longNonce := make([]byte, 100)
	for i := range longNonce {
		longNonce[i] = byte(i)
	}
	f.Add(byte('d'), longNonce)

	// All zeros
	f.Add(byte(0x00), make([]byte, 32))

	// All 0xff
	allFF := make([]byte, 32)
	for i := range allFF {
		allFF[i] = 0xff
	}
	f.Add(byte(0xff), allFF)

	// Various character values
	f.Add(byte('0'), []byte{0x01, 0x02, 0x03})
	f.Add(byte('z'), []byte{0xfe, 0xfd, 0xfc})

	f.Fuzz(func(t *testing.T, char byte, nonce []byte) {
		// Create an OpNS with a fixed Pow field
		opns := &OpNS{
			Pow:     make([]byte, 32),
			Claimed: []byte{0x01},
			Domain:  "test",
		}

		// TestSolution should never panic
		_ = opns.TestSolution(char, nonce)
	})
}

// FuzzLock tests the Lock function with various inputs.
// It should handle edge cases in domain and pow values.
// Run with: go test -fuzz=FuzzLock -fuzztime=10s
func FuzzLock(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus
	f.Add([]byte{0x01}, "test", make([]byte, 32))
	f.Add([]byte{}, "", []byte{})
	f.Add([]byte{0x00, 0x00}, "a", []byte{0xff})
	f.Add(make([]byte, 100), "verylongdomainname", make([]byte, 64))

	f.Fuzz(func(t *testing.T, claimed []byte, domain string, pow []byte) {
		// Lock should never panic
		result := Lock(claimed, domain, pow)

		// Result should always be a valid script
		if result == nil {
			t.Error("Lock returned nil script")
		}
	})
}
