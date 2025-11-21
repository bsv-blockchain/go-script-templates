package pow20

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
)

// FuzzDecode tests the POW20 Decode function with random script bytes.
// The decoder supports both JSON-based BSV21 inscriptions and traditional
// script-based POW20 tokens. It should never panic regardless of input.
// Run with: go test -fuzz=FuzzDecode -fuzztime=10s
func FuzzDecode(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus with meaningful test cases

	// Empty script
	f.Add([]byte{})

	// Nil-like inputs
	f.Add([]byte{0x00})

	// Valid JSON inscription structure (BSV21 style)
	// OP_FALSE OP_IF "ord" OP_1 "application/bsv-20" OP_0 <json> OP_ENDIF
	jsonInscription := &script.Script{}
	_ = jsonInscription.AppendOpcodes(script.OpFALSE, script.OpIF)
	_ = jsonInscription.AppendPushData([]byte("ord"))
	_ = jsonInscription.AppendOpcodes(script.Op1)
	_ = jsonInscription.AppendPushData([]byte("application/bsv-20"))
	_ = jsonInscription.AppendOpcodes(script.Op0)
	_ = jsonInscription.AppendPushData([]byte(`{"p":"bsv-20","op":"deploy","contract":"pow-20","sym":"TEST","maxSupply":"1000000","difficulty":"20","startingReward":"100"}`))
	_ = jsonInscription.AppendOpcodes(script.OpENDIF)
	f.Add(jsonInscription.Bytes())

	// Malformed JSON inscription
	malformedJSON := &script.Script{}
	_ = malformedJSON.AppendOpcodes(script.OpFALSE, script.OpIF)
	_ = malformedJSON.AppendPushData([]byte("ord"))
	_ = malformedJSON.AppendOpcodes(script.Op1)
	_ = malformedJSON.AppendPushData([]byte("application/bsv-20"))
	_ = malformedJSON.AppendOpcodes(script.Op0)
	_ = malformedJSON.AppendPushData([]byte(`{invalid json`))
	_ = malformedJSON.AppendOpcodes(script.OpENDIF)
	f.Add(malformedJSON.Bytes())

	// JSON with wrong contract type
	wrongContract := &script.Script{}
	_ = wrongContract.AppendOpcodes(script.OpFALSE, script.OpIF)
	_ = wrongContract.AppendPushData([]byte("ord"))
	_ = wrongContract.AppendOpcodes(script.Op1)
	_ = wrongContract.AppendPushData([]byte("application/bsv-20"))
	_ = wrongContract.AppendOpcodes(script.Op0)
	_ = wrongContract.AppendPushData([]byte(`{"p":"bsv-20","contract":"other"}`))
	_ = wrongContract.AppendOpcodes(script.OpENDIF)
	f.Add(wrongContract.Bytes())

	// Edge cases with various opcode patterns
	f.Add([]byte{0x4c, 0x00})             // OP_PUSHDATA1 with zero length
	f.Add([]byte{0x4d, 0x00, 0x00})       // OP_PUSHDATA2 with zero length
	f.Add([]byte{0x4e, 0x00, 0x00, 0x00}) // OP_PUSHDATA4 with zero length

	// Random bytes that might trigger edge cases
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Add([]byte{script.OpRETURN, 0x00})

	// Script with pow20 prefix pattern (if it exists in constants)
	// Just add some patterns that might partially match
	f.Add([]byte{0x00, 0x63}) // OP_FALSE OP_IF start

	f.Fuzz(func(t *testing.T, data []byte) {
		// Create script from bytes - should never panic
		scr := script.NewFromBytes(data)

		// Decode should never panic, regardless of input
		_ = Decode(scr)
	})
}

// FuzzBuildInscription tests the BuildInscription function.
// It should handle various ID strings and amounts without panicking.
// Run with: go test -fuzz=FuzzBuildInscription -fuzztime=10s
func FuzzBuildInscription(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus
	f.Add("abc123_0", uint64(1000))
	f.Add("", uint64(0))
	f.Add("very_long_transaction_id_that_exceeds_normal_length_0", uint64(18446744073709551615))
	f.Add("special\x00chars\n\t", uint64(1))
	f.Add(`"quoted"`, uint64(100))

	f.Fuzz(func(t *testing.T, id string, amt uint64) {
		// BuildInscription should never panic
		result := BuildInscription(id, amt)

		// Result should always be a valid script
		if result == nil {
			t.Error("BuildInscription returned nil script")
		}
	})
}

// FuzzUint64ToBytes tests the uint64ToBytes utility function.
// It should handle all uint64 values correctly.
// Run with: go test -fuzz=FuzzUint64ToBytes -fuzztime=10s
func FuzzUint64ToBytes(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus with edge cases
	f.Add(uint64(0))
	f.Add(uint64(1))
	f.Add(uint64(127))
	f.Add(uint64(128))
	f.Add(uint64(255))
	f.Add(uint64(256))
	f.Add(uint64(65535))
	f.Add(uint64(65536))
	f.Add(uint64(16777215))
	f.Add(uint64(16777216))
	f.Add(uint64(4294967295))
	f.Add(uint64(4294967296))
	f.Add(uint64(18446744073709551615)) // Max uint64

	f.Fuzz(func(t *testing.T, v uint64) {
		// uint64ToBytes should never panic
		result := uint64ToBytes(v)

		// Result should be non-nil (even for 0, which returns empty slice)
		if result == nil {
			t.Error("uint64ToBytes returned nil")
		}

		// Result length should be reasonable (0-8 bytes for uint64)
		if len(result) > 8 {
			t.Errorf("uint64ToBytes returned %d bytes, expected <= 8", len(result))
		}
	})
}
