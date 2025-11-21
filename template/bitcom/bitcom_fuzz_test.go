package bitcom

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
)

// FuzzDecode tests the Decode function with random script bytes.
// The decoder should never panic regardless of input.
// Run with: go test -fuzz=FuzzDecode -fuzztime=10s
func FuzzDecode(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus with meaningful test cases

	// Empty script
	f.Add([]byte{})

	// Just OP_RETURN
	f.Add([]byte{script.OpRETURN})

	// OP_FALSE + OP_RETURN (common prefix)
	f.Add([]byte{script.OpFALSE, script.OpRETURN})

	// OP_RETURN with some data
	f.Add([]byte{script.OpRETURN, 0x04, 't', 'e', 's', 't'})

	// Script with pipe separator
	f.Add([]byte{script.OpRETURN, 0x03, 'a', 'b', 'c', 0x01, '|', 0x03, 'd', 'e', 'f'})

	// MAP protocol prefix
	f.Add([]byte{script.OpRETURN, 0x22, '1', 'P', 'u', 'Q', 'a', '7', 'K', '6', '2', 'M', 'i', 'K', 'C', 't', 's', 's', 'S', 'L', 'K', 'y', '1', 'k', 'h', '5', '6', 'W', 'W', 'U', '7', 'M', 't', 'U', 'R', '5'})

	// Various edge cases
	f.Add([]byte{0xff}) // Invalid opcode
	// Truncated PUSHDATA inputs are intentionally excluded (e.g., OP_PUSHDATA1 with
	// non-zero length but missing data, or OP_PUSHDATA4 with incomplete length bytes).
	// The go-sdk script parser has a bug that causes infinite loops on truncated input.
	// Once fixed upstream, we can add tests like: []byte{0x4c, 0xff} for truncated pushes.
	f.Add([]byte{0x4c, 0x00})                   // OP_PUSHDATA1 with zero length
	f.Add([]byte{0x4d, 0x00, 0x00})             // OP_PUSHDATA2 with zero length
	f.Add([]byte{0x4e, 0x00, 0x00, 0x00, 0x00}) // OP_PUSHDATA4 with zero length (needs 4 bytes for length)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Reset global state
		ZERO = 0

		// Create script from bytes - should never panic
		scr := script.NewFromBytes(data)

		// Decode should never panic, regardless of input
		_ = Decode(scr)
	})
}

// FuzzDecodeMap tests the DecodeMap function with random script bytes.
// The decoder should handle malformed MAP protocol data gracefully.
// Run with: go test -fuzz=FuzzDecodeMap -fuzztime=10s
func FuzzDecodeMap(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Seed corpus with meaningful test cases

	// Empty script
	f.Add([]byte{})

	// Valid SET command structure
	setCmd := &script.Script{}
	_ = setCmd.AppendPushData([]byte("SET"))
	_ = setCmd.AppendPushData([]byte("key"))
	_ = setCmd.AppendPushData([]byte("value"))
	f.Add(setCmd.Bytes())

	// SET with null bytes in value
	setNull := &script.Script{}
	_ = setNull.AppendPushData([]byte("SET"))
	_ = setNull.AppendPushData([]byte("key"))
	_ = setNull.AppendPushData([]byte{0x00, 0x00, 0x00})
	f.Add(setNull.Bytes())

	// SET with missing value (odd number of fields)
	setMissing := &script.Script{}
	_ = setMissing.AppendPushData([]byte("SET"))
	_ = setMissing.AppendPushData([]byte("key"))
	f.Add(setMissing.Bytes())

	// Just a command with no data
	f.Add([]byte{0x03, 'S', 'E', 'T'})

	// Very short scripts
	f.Add([]byte{0x01, 'S'})
	f.Add([]byte{0x00})

	// Script with UTF-8 replacement scenarios
	setUTF8 := &script.Script{}
	_ = setUTF8.AppendPushData([]byte("SET"))
	_ = setUTF8.AppendPushData([]byte("key"))
	_ = setUTF8.AppendPushData([]byte("value\\u0000test"))
	f.Add(setUTF8.Bytes())

	f.Fuzz(func(t *testing.T, data []byte) {
		// Reset global state
		ZERO = 0

		// DecodeMap accepts any type, test with bytes directly
		_ = DecodeMap(data)

		// Also test with script pointer
		scr := script.NewFromBytes(data)
		_ = DecodeMap(scr)
	})
}

// FuzzDecodeBAP tests the DecodeBAP function with random Bitcom data.
// The decoder should handle all BAP message types gracefully.
// Run with: go test -fuzz=FuzzDecodeBAP -fuzztime=10s
func FuzzDecodeBAP(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping fuzz test in short mode")
	}

	// Helper to create a Bitcom with BAP protocol
	createBAPBitcom := func(scriptData []byte) *Bitcom {
		return &Bitcom{
			Protocols: []*BitcomProtocol{
				{
					Protocol: BAPPrefix,
					Script:   scriptData,
				},
			},
		}
	}

	// Seed with various BAP message structures

	// ID type
	idScript := &script.Script{}
	_ = idScript.AppendPushData([]byte("ID"))
	_ = idScript.AppendPushData([]byte("identity_key"))
	_ = idScript.AppendPushData([]byte("1Address"))
	f.Add(idScript.Bytes())

	// ATTEST type
	attestScript := &script.Script{}
	_ = attestScript.AppendPushData([]byte("ATTEST"))
	_ = attestScript.AppendPushData([]byte("txid_hash"))
	_ = attestScript.AppendPushData([]byte("1"))
	f.Add(attestScript.Bytes())

	// REVOKE type
	revokeScript := &script.Script{}
	_ = revokeScript.AppendPushData([]byte("REVOKE"))
	_ = revokeScript.AppendPushData([]byte("txid_hash"))
	_ = revokeScript.AppendPushData([]byte("1"))
	f.Add(revokeScript.Bytes())

	// ALIAS type
	aliasScript := &script.Script{}
	_ = aliasScript.AppendPushData([]byte("ALIAS"))
	_ = aliasScript.AppendPushData([]byte("my_alias"))
	_ = aliasScript.AppendPushData([]byte(`{"profile":"data"}`))
	f.Add(aliasScript.Bytes())

	// Empty script
	f.Add([]byte{})

	// Just type with no data
	f.Add([]byte{0x02, 'I', 'D'})

	// Malformed types
	f.Add([]byte{0x07, 'U', 'N', 'K', 'N', 'O', 'W', 'N'})

	// Script with pipe separator for AIP signature
	idWithAIP := &script.Script{}
	_ = idWithAIP.AppendPushData([]byte("ID"))
	_ = idWithAIP.AppendPushData([]byte("identity_key"))
	_ = idWithAIP.AppendPushData([]byte("1Address"))
	_ = idWithAIP.AppendPushData([]byte("|"))
	_ = idWithAIP.AppendPushData([]byte("AIP_PREFIX"))
	_ = idWithAIP.AppendPushData([]byte("BITCOIN_ECDSA"))
	_ = idWithAIP.AppendPushData([]byte("1SignerAddress"))
	_ = idWithAIP.AppendPushData([]byte("signature_base64"))
	f.Add(idWithAIP.Bytes())

	f.Fuzz(func(t *testing.T, data []byte) {
		// Reset global state
		ZERO = 0

		// Test DecodeBAP with fuzzed script data
		bitcom := createBAPBitcom(data)
		_ = DecodeBAP(bitcom)

		// Also test with nil and empty Bitcom
		_ = DecodeBAP(nil)
		_ = DecodeBAP(&Bitcom{})
		_ = DecodeBAP(&Bitcom{Protocols: []*BitcomProtocol{}})
	})
}
