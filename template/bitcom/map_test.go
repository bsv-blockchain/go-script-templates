package bitcom

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/require"
)

// resetTestState resets any global state that affects test outcomes.
// Call this function at the beginning of each test function and subtest
// to ensure tests don't interfere with each other through shared state.
func resetTestState() {
	// Reset the global script position counter
	ZERO = 0
}

func TestDecodeMap(t *testing.T) {
	// Reset test state before each test
	resetTestState()

	t.Run("empty script", func(t *testing.T) {
		// Reset test state before each subtest
		resetTestState()

		emptyScript := script.Script{}
		result := DecodeMap(emptyScript)
		require.Nil(t, result, "Expected nil result for empty script")
	})

	t.Run("SET command with bsocial post data", func(t *testing.T) {
		// Reset test state before each subtest
		resetTestState()

		s := &script.Script{}

		t.Logf("Adding MapCmdSet: %s", MapCmdSet)
		_ = s.AppendPushData([]byte(MapCmdSet))
		t.Logf("Adding key 'app'")
		_ = s.AppendPushData([]byte("app"))
		t.Logf("Adding value 'bsocial'")
		_ = s.AppendPushData([]byte("bsocial"))
		t.Logf("Adding key 'type'")
		_ = s.AppendPushData([]byte("type"))
		t.Logf("Adding value 'post'")
		_ = s.AppendPushData([]byte("post"))

		// Debug printouts
		t.Logf("Script: %+v", s)
		t.Logf("Script bytes (hex): %x", s.Bytes())

		// Try direct script
		result := DecodeMap(s)
		t.Logf("Result with script pointer: %+v", result)
		if result != nil {
			t.Logf("Result data: %+v", result.Data)
		}

		// The test expects this to succeed
		require.NotNil(t, result, "Expected non-nil result")
		if result != nil {
			require.Equal(t, MapCmdSet, result.Cmd)
			require.Equal(t, "bsocial", result.Data["app"])
			require.Equal(t, "post", result.Data["type"])
		}
	})

	t.Run("SET command with null values", func(t *testing.T) {
		// Reset test state before each subtest
		resetTestState()

		s := &script.Script{}
		t.Logf("Adding MapCmdSet: %s", MapCmdSet)
		_ = s.AppendPushData([]byte(MapCmdSet))
		t.Logf("Adding key 'key1'")
		_ = s.AppendPushData([]byte("key1"))
		t.Logf("Adding null value")
		_ = s.AppendPushData([]byte{0x00})

		// Debug printouts
		t.Logf("Script: %+v", s)
		t.Logf("Script bytes (hex): %x", s.Bytes())

		result := DecodeMap(s)
		t.Logf("Result for null values test: %+v", result)
		if result != nil {
			t.Logf("Result data: %+v", result.Data)
		}

		require.NotNil(t, result, "Expected non-nil result")
		if result != nil {
			require.Equal(t, MapCmdSet, result.Cmd)
			require.Equal(t, " ", result.Data["key1"])
		}
	})

	t.Run("SET command with missing value", func(t *testing.T) {
		// Reset test state before each subtest
		resetTestState()

		s := &script.Script{}
		t.Logf("Adding MapCmdSet: %s", MapCmdSet)
		_ = s.AppendPushData([]byte(MapCmdSet))
		t.Logf("Adding key 'key2'")
		_ = s.AppendPushData([]byte("key2"))
		// Intentionally missing value

		// Debug printouts
		t.Logf("Script: %+v", s)
		t.Logf("Script bytes (hex): %x", s.Bytes())

		result := DecodeMap(s)
		t.Logf("Result for missing value test: %+v", result)
		if result != nil {
			t.Logf("Result data: %+v", result.Data)
			t.Logf("Data keys length: %d", len(result.Data))
		}

		require.NotNil(t, result, "Expected non-nil result")
		if result != nil {
			require.Equal(t, MapCmdSet, result.Cmd)
			require.Empty(t, result.Data["key2"])
			require.Equal(t, 0, len(result.Data))
		}
	})
}

// TestDecodeMap_Bytes tests that DecodeMap can handle raw bytes input
func TestDecodeMap_Bytes(t *testing.T) {
	// Reset test state
	resetTestState()

	// Test nil input
	result := DecodeMap(nil)
	require.Nil(t, result, "Expected nil result for nil input")

	// Create a valid MAP protocol script
	s := &script.Script{}
	t.Logf("Adding MapCmdSet: %s", MapCmdSet)
	_ = s.AppendPushData([]byte(MapCmdSet))
	t.Logf("Adding key 'app'")
	_ = s.AppendPushData([]byte("app"))
	t.Logf("Adding value 'bsocial'")
	_ = s.AppendPushData([]byte("bsocial"))
	t.Logf("Adding key 'type'")
	_ = s.AppendPushData([]byte("type"))
	t.Logf("Adding value 'post'")
	_ = s.AppendPushData([]byte("post"))

	// Debug prints
	t.Logf("Script for bytes test: %+v", s)

	// Get the bytes
	scriptBytes := s.Bytes()
	t.Logf("Script bytes (hex): %x", scriptBytes)

	// Try with new script from bytes
	newScript := script.NewFromBytes(scriptBytes)
	t.Logf("New script from bytes: %+v", newScript)

	resultFromNewScript := DecodeMap(newScript)
	t.Logf("Result with newScript: %+v", resultFromNewScript)
	if resultFromNewScript != nil {
		t.Logf("Result data: %+v", resultFromNewScript.Data)
	}

	// Reset test state before the next test
	resetTestState()

	// Now try with raw bytes
	result = DecodeMap(scriptBytes)
	t.Logf("Result with raw bytes: %+v", result)
	if result != nil {
		t.Logf("Result data: %+v", result.Data)
	}

	// Reset test state before the next test
	resetTestState()

	// Try using a different approach to create the script bytes
	manualScript := &script.Script{}
	_ = manualScript.AppendPushData([]byte(MapCmdSet))
	_ = manualScript.AppendPushData([]byte("app"))
	_ = manualScript.AppendPushData([]byte("bsocial"))
	_ = manualScript.AppendPushData([]byte("type"))
	_ = manualScript.AppendPushData([]byte("post"))
	manualBytes := manualScript.Bytes()
	t.Logf("Manual script bytes (hex): %x", manualBytes)
	resultManual := DecodeMap(manualBytes)
	t.Logf("Result with manual bytes: %+v", resultManual)

	// The test expects this to work
	require.NotNil(t, result, "Expected non-nil result for valid script bytes")
	if result != nil {
		require.Equal(t, MapCmdSet, result.Cmd, "Expected correct command")
		require.Equal(t, "bsocial", result.Data["app"], "Expected correct app value")
		require.Equal(t, "post", result.Data["type"], "Expected correct type value")
	}

	// Reset test state before the next test
	resetTestState()

	// Test invalid script bytes
	invalidBytes := []byte{0x00, 0x01} // Just some random bytes
	result = DecodeMap(invalidBytes)
	// Invalid bytes should return nil since they don't have a proper prefix
	require.Nil(t, result, "Expected nil result for invalid script bytes")
}

// TestToScript tests the ToScript helper function directly
func TestToScript(t *testing.T) {
	// Reset test state
	resetTestState()

	// Create a valid MAP protocol script
	s := &script.Script{}
	_ = s.AppendPushData([]byte(MapPrefix))
	_ = s.AppendPushData([]byte(MapCmdSet))
	_ = s.AppendPushData([]byte("app"))
	_ = s.AppendPushData([]byte("bsocial"))

	// Test converting script to script
	scriptPtr := ToScript(s)
	require.NotNil(t, scriptPtr, "ToScript should handle *script.Script")
	require.Equal(t, s, scriptPtr, "ToScript should return the same script pointer")

	// Reset test state
	resetTestState()

	// Test converting script value to script
	scriptVal := *s
	scriptFromVal := ToScript(scriptVal)
	require.NotNil(t, scriptFromVal, "ToScript should handle script.Script")
	require.Equal(t, s.Bytes(), scriptFromVal.Bytes(), "Bytes should match")

	// Reset test state
	resetTestState()

	// Test converting bytes to script
	bytes := s.Bytes()
	t.Logf("Original script bytes: %x", bytes)
	scriptFromBytes := ToScript(bytes)
	require.NotNil(t, scriptFromBytes, "ToScript should handle []byte")
	t.Logf("scriptFromBytes: %+v", scriptFromBytes)
	t.Logf("scriptFromBytes bytes: %x", scriptFromBytes.Bytes())

	// Reset test state
	resetTestState()

	// A script that still carries the MAP prefix is not what DecodeMap is
	// handed in practice: bitcom.Decode strips the prefix and passes only the
	// protocol body. Feeding the prefixed script in reads the prefix itself as
	// the command, which is not a MAP command, so it decodes to nil.
	resetTestState()
	require.Nil(t, DecodeMap(s), "prefix is not a MAP command, so it does not decode")

	// DecodeMap operates on the protocol body, prefix already stripped.
	body := &script.Script{}
	_ = body.AppendPushData([]byte(MapCmdSet))
	_ = body.AppendPushData([]byte("app"))
	_ = body.AppendPushData([]byte("bsocial"))
	bodyBytes := body.Bytes()

	// Decode Map from different sources
	resetTestState()
	mapFromScript := DecodeMap(body)

	// Reset test state
	resetTestState()

	mapFromBytes := DecodeMap(bodyBytes)

	t.Logf("mapFromScript: %+v", mapFromScript)
	t.Logf("mapFromBytes: %+v", mapFromBytes)

	// Check if both are non-nil
	require.NotNil(t, mapFromScript, "DecodeMap should work with script")
	require.NotNil(t, mapFromBytes, "DecodeMap should work with bytes")
	require.Equal(t, "bsocial", mapFromScript.Data["app"])
	require.Equal(t, "bsocial", mapFromBytes.Data["app"])
}

// mapScript builds a bare MAP protocol body: the command followed by its
// pushes, with no protocol prefix. This is what bitcom.Decode hands DecodeMap.
func mapScript(t *testing.T, cmd MapCmd, pushes ...string) *script.Script {
	t.Helper()
	s := &script.Script{}
	require.NoError(t, s.AppendPushData([]byte(cmd)))
	for _, p := range pushes {
		require.NoError(t, s.AppendPushData([]byte(p)))
	}
	return s
}

// TestDecodeMapRemove covers the REMOVE command, which clears single-value
// keys. Every push after the command is a key.
func TestDecodeMapRemove(t *testing.T) {
	resetTestState()

	t.Run("reads every key it names", func(t *testing.T) {
		resetTestState()

		result := DecodeMap(mapScript(t, MapCmdRemove, "profile.name", "profile.text"))

		require.NotNil(t, result)
		require.Equal(t, MapCmdRemove, result.Cmd)
		require.Len(t, result.Data, 2)
		require.Contains(t, result.Data, "profile.name")
		require.Contains(t, result.Data, "profile.text")
		require.Equal(t, "", result.Data["profile.name"])
		require.Equal(t, "", result.Data["profile.text"])
		require.Empty(t, result.Adds)
		require.Empty(t, result.Deletes)
	})

	t.Run("single key", func(t *testing.T) {
		resetTestState()

		result := DecodeMap(mapScript(t, MapCmdRemove, "profile.name"))

		require.NotNil(t, result)
		require.Equal(t, MapCmdRemove, result.Cmd)
		require.Len(t, result.Data, 1)
		require.Contains(t, result.Data, "profile.name")
	})

	t.Run("no keys returns nil", func(t *testing.T) {
		resetTestState()

		require.Nil(t, DecodeMap(mapScript(t, MapCmdRemove)))
	})
}

// TestDecodeMapAdd covers the ADD command, which appends values to one
// list-valued key. The key is read first, then every remaining push is a value.
func TestDecodeMapAdd(t *testing.T) {
	resetTestState()

	t.Run("reads one key then all remaining values", func(t *testing.T) {
		resetTestState()

		result := DecodeMap(mapScript(t, MapCmdAdd, "interests", "cars", "science", "boats"))

		require.NotNil(t, result)
		require.Equal(t, MapCmdAdd, result.Cmd)
		// The key is a key, never one of the values.
		require.Equal(t, []string{"cars", "science", "boats"}, result.Adds)
		require.NotContains(t, result.Adds, "interests")
		require.Equal(t, "cars science boats", result.Data["interests"])
		require.Len(t, result.Data, 1)
		require.Empty(t, result.Deletes)
	})

	t.Run("single value", func(t *testing.T) {
		resetTestState()

		result := DecodeMap(mapScript(t, MapCmdAdd, "interests", "cars"))

		require.NotNil(t, result)
		require.Equal(t, []string{"cars"}, result.Adds)
		require.Equal(t, "cars", result.Data["interests"])
	})

	t.Run("key with no values still names the key", func(t *testing.T) {
		resetTestState()

		result := DecodeMap(mapScript(t, MapCmdAdd, "interests"))

		require.NotNil(t, result)
		require.Equal(t, MapCmdAdd, result.Cmd)
		require.Contains(t, result.Data, "interests")
		require.Equal(t, "", result.Data["interests"])
		require.Empty(t, result.Adds)
	})

	t.Run("no key returns nil", func(t *testing.T) {
		resetTestState()

		require.Nil(t, DecodeMap(mapScript(t, MapCmdAdd)))
	})
}

// TestDecodeMapDelete covers the DELETE command. DELETE names its key first and
// only then the values to strike, so that values are removed from the intended
// list and no other.
func TestDecodeMapDelete(t *testing.T) {
	resetTestState()

	t.Run("names the key before the values to strike", func(t *testing.T) {
		resetTestState()

		result := DecodeMap(mapScript(t, MapCmdDelete, "interests", "cars", "boats"))

		require.NotNil(t, result)
		require.Equal(t, MapCmdDelete, result.Cmd)

		// Ordering guarantee: the first push after the command is the key, and
		// it must not be mistaken for one of the values to strike.
		require.Equal(t, []string{"cars", "boats"}, result.Deletes)
		require.NotContains(t, result.Deletes, "interests")
		require.Len(t, result.Data, 1)
		require.Contains(t, result.Data, "interests")
		require.Equal(t, "cars boats", result.Data["interests"])
		require.Empty(t, result.Adds)
	})

	t.Run("values are struck only from the named list", func(t *testing.T) {
		resetTestState()

		// Two DELETEs naming different keys but the same value must not be
		// conflated: each records the value against its own key.
		first := DecodeMap(mapScript(t, MapCmdDelete, "interests", "cars"))
		resetTestState()
		second := DecodeMap(mapScript(t, MapCmdDelete, "dislikes", "cars"))

		require.NotNil(t, first)
		require.NotNil(t, second)
		require.Contains(t, first.Data, "interests")
		require.NotContains(t, first.Data, "dislikes")
		require.Contains(t, second.Data, "dislikes")
		require.NotContains(t, second.Data, "interests")
		require.Equal(t, []string{"cars"}, first.Deletes)
		require.Equal(t, []string{"cars"}, second.Deletes)
	})

	t.Run("key with no values still names the key", func(t *testing.T) {
		resetTestState()

		result := DecodeMap(mapScript(t, MapCmdDelete, "interests"))

		require.NotNil(t, result)
		require.Contains(t, result.Data, "interests")
		require.Empty(t, result.Deletes)
	})

	t.Run("no key returns nil", func(t *testing.T) {
		resetTestState()

		require.Nil(t, DecodeMap(mapScript(t, MapCmdDelete)))
	})
}

// TestDecodeMapUnhandledCommands checks that a command DecodeMap cannot turn
// into a key/value record returns nil rather than a non-nil record with an
// empty Data map, which callers would read as "decoded, but empty".
func TestDecodeMapUnhandledCommands(t *testing.T) {
	resetTestState()

	t.Run("unrecognized command returns nil", func(t *testing.T) {
		resetTestState()

		// DEL was never a MAP command; nothing on chain was written with it.
		require.Nil(t, DecodeMap(mapScript(t, "DEL", "app", "bsocial")))
		resetTestState()
		require.Nil(t, DecodeMap(mapScript(t, "NOTACOMMAND", "app", "bsocial")))
		resetTestState()
		require.Nil(t, DecodeMap(mapScript(t, "set", "app", "bsocial")), "commands are case sensitive")
	})

	t.Run("SELECT is recognized but not decoded here", func(t *testing.T) {
		resetTestState()

		// SELECT designates a txid as context for a following command, which
		// needs ::: instruction-set support to represent.
		txid := "b14113a50b2d1c2a3644346662de921a60e1ed63bee9962dd4c7f7ee8a1f3ffb"
		require.Nil(t, DecodeMap(mapScript(t, MapCmdSelect, txid)))
	})

	t.Run("CLEAR is recognized but not decoded here", func(t *testing.T) {
		resetTestState()

		txid := "b14113a50b2d1c2a3644346662de921a60e1ed63bee9962dd4c7f7ee8a1f3ffb"
		require.Nil(t, DecodeMap(mapScript(t, MapCmdClear, txid)))
	})
}

// TestMapCommandSet pins the command vocabulary to version 2 of the MAP
// specification: https://github.com/opldotdev/MAP
func TestMapCommandSet(t *testing.T) {
	resetTestState()

	require.Equal(t, MapCmd("SET"), MapCmdSet)
	require.Equal(t, MapCmd("REMOVE"), MapCmdRemove)
	require.Equal(t, MapCmd("ADD"), MapCmdAdd)
	require.Equal(t, MapCmd("DELETE"), MapCmdDelete)
	require.Equal(t, MapCmd("SELECT"), MapCmdSelect)
	require.Equal(t, MapCmd("CLEAR"), MapCmdClear)
}

// TestDecodeMapSetUnchanged pins the SET behaviour that existing callers rely
// on, so the command-set correction does not disturb it.
func TestDecodeMapSetUnchanged(t *testing.T) {
	resetTestState()

	result := DecodeMap(mapScript(t, MapCmdSet, "app", "bsocial", "type", "post"))

	require.NotNil(t, result)
	require.Equal(t, MapCmdSet, result.Cmd)
	require.Equal(t, "bsocial", result.Data["app"])
	require.Equal(t, "post", result.Data["type"])
	require.Empty(t, result.Adds)
	require.Empty(t, result.Deletes)
}
