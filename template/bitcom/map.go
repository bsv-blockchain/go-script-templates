package bitcom

import (
	"bytes"
	"strings"

	"github.com/bsv-blockchain/go-sdk/script"
)

const MapPrefix = "1PuQa7K62MiKCtssSLKy1kh56WWU7MtUR5"

type MapCmd string

var ZERO = 0

// The Magic Attribute Protocol command set, as defined by version 2 of the
// specification (https://github.com/opldotdev/MAP).
//
// SET and REMOVE operate on single-value keys. ADD and DELETE operate on
// list-valued keys; DELETE names its key first so that the values that follow
// are only struck from that one list. SELECT designates a txid as the context
// for a following command, and CLEAR erases every record a transaction wrote.
var (
	MapCmdSet    MapCmd = "SET"
	MapCmdRemove MapCmd = "REMOVE"
	MapCmdAdd    MapCmd = "ADD"
	MapCmdDelete MapCmd = "DELETE"
	MapCmdSelect MapCmd = "SELECT"
	MapCmdClear  MapCmd = "CLEAR"
)

type Map struct {
	Cmd  MapCmd            `json:"cmd"`
	Data map[string]string `json:"data"`
	// Adds holds the values appended by an ADD command, in script order.
	Adds []string `json:"adds,omitempty"`
	// Deletes holds the values struck by a DELETE command, in script order.
	Deletes []string `json:"deletes,omitempty"`
}

// DecodeMap decodes a single MAP command from the transaction script.
//
// A nil return means no MAP record could be decoded. That covers an unusable
// script, an unreadable or unrecognized command, and a data-carrying command
// that named no keys at all. SET is the one exception: for backwards
// compatibility it still returns a non-nil result when a trailing key has no
// value, matching its long-standing behaviour.
//
// SELECT and CLEAR are recognized but not decoded here. Both describe an
// instruction set rather than the key/value data this struct holds, so
// supporting them depends on the ::: separator, which is not yet handled.
func DecodeMap(data any) *Map {
	scr := ToScript(data)
	if scr == nil || len(*scr) == 0 {
		return nil
	}

	pos := ZERO
	var op *script.ScriptChunk
	var err error

	// If length is < minimum, return nil
	if len(*scr) < 6 {
		return nil
	}

	// Read command
	if op, err = scr.ReadOp(&pos); err != nil {
		return nil
	}
	cmd := MapCmd(op.Data)

	// Create map
	m := &Map{
		Cmd:  cmd,
		Data: make(map[string]string),
	}

	switch cmd {
	case MapCmdSet:
		decodeMapPairs(scr, &pos, m)
		// SET keeps its original lenient contract: a dangling key still
		// yields a non-nil result with an empty Data map.
		return m

	case MapCmdRemove:
		// REMOVE clears single-value keys. Every remaining push is a key.
		for {
			if op, err = scr.ReadOp(&pos); err != nil {
				break
			}
			m.Data[cleanMapString(op.Data)] = ""
		}

	case MapCmdAdd:
		// ADD names one list-valued key, then the values to append.
		m.Adds = decodeMapKeyValues(scr, &pos, m)

	case MapCmdDelete:
		// DELETE names one list-valued key, then the values to strike from
		// that list. The key comes first so values are never removed from a
		// list they were not meant for.
		m.Deletes = decodeMapKeyValues(scr, &pos, m)

	case MapCmdSelect, MapCmdClear:
		// Recognized, but not representable as a single key/value record.
		return nil

	default:
		// Unrecognized command. Returning nil rather than an empty record
		// keeps callers from mistaking undecoded data for absent data.
		return nil
	}

	if len(m.Data) == 0 {
		return nil
	}
	return m
}

// decodeMapPairs reads alternating key/value pushes into m.Data. A trailing
// key with no value is discarded and the read position rewound to it.
func decodeMapPairs(scr *script.Script, pos *int, m *Map) {
	for {
		// Save position to revert if needed
		keyPos := *pos

		// Try to read key
		op, err := scr.ReadOp(pos)
		if err != nil {
			return
		}
		opKey := cleanMapString(op.Data)

		// Try to read value
		if op, err = scr.ReadOp(pos); err != nil {
			// Couldn't read value, revert to position before key and break
			*pos = keyPos
			return
		}

		m.Data[opKey] = cleanMapString(op.Data)
	}
}

// decodeMapKeyValues reads one key push followed by every remaining push as a
// value. The values are recorded on m.Data under that key, space joined, and
// also returned in script order. It returns nil if no key is present.
func decodeMapKeyValues(scr *script.Script, pos *int, m *Map) []string {
	op, err := scr.ReadOp(pos)
	if err != nil {
		return nil
	}
	key := cleanMapString(op.Data)

	var values []string
	for {
		if op, err = scr.ReadOp(pos); err != nil {
			break
		}
		values = append(values, cleanMapString(op.Data))
	}

	m.Data[key] = strings.Join(values, " ")
	return values
}

// cleanMapString replaces null bytes with spaces rather than discarding the
// push, so that invalid sequences do not cost a whole key/value pair.
func cleanMapString(b []byte) string {
	return strings.ReplaceAll(string(bytes.ReplaceAll(b, []byte{0}, []byte{' '})), "\\u0000", " ")
}
