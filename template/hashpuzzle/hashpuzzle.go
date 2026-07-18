// Package hashpuzzle implements a hash puzzle script template.
//
// A hash puzzle locks satoshis to the SHA-256 hash of a secret preimage;
// whoever reveals the preimage can spend the output. The script pattern is:
//
//	Locking script:   OP_SHA256 <32-byte hash> OP_EQUAL
//	Unlocking script: <secret>
//
// Spending requires no signature, which makes unlocking cheap to construct
// (no ECDSA signing, no key management) and keeps unlocking scripts small
// (~33 bytes for a 32-byte secret vs ~107 bytes for P2PKH). The trade-off is
// that the unlocking data does not commit to the spending transaction: once a
// preimage is revealed, anyone who sees it can claim the output. Use only
// where that is acceptable — small-value outputs, atomic-swap legs, or
// high-throughput fee-fuel pools.
//
// This is a Go port of the HashPuzzle ScriptTemplate used in the
// truth-machine demo (github.com/bsv-blockchain-demos/truth-machine).
package hashpuzzle

import (
	"crypto/rand"
	"errors"

	primitives "github.com/bsv-blockchain/go-sdk/primitives/hash"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

// SecretLength is the preimage length produced by GenerateSecretPair.
const SecretLength = 32

// HashLength is the required length of the SHA-256 hash in the locking script.
const HashLength = 32

var (
	ErrInvalidHashLength = errors.New("hash must be 32 bytes")
	ErrNoSecret          = errors.New("secret (preimage) not supplied")
)

// HashPuzzle holds the decoded form of a hash puzzle locking script.
type HashPuzzle struct {
	Hash []byte `json:"hash"`
}

// Lock creates a locking script that can only be spent by revealing a
// preimage whose SHA-256 digest equals hash32: OP_SHA256 <hash32> OP_EQUAL.
func Lock(hash32 []byte) (*script.Script, error) {
	if len(hash32) != HashLength {
		return nil, ErrInvalidHashLength
	}
	scr := script.Script(make([]byte, 0, HashLength+3))
	s := &scr
	_ = s.AppendOpcodes(script.OpSHA256)
	if err := s.AppendPushData(hash32); err != nil {
		return nil, err
	}
	_ = s.AppendOpcodes(script.OpEQUAL)
	return s, nil
}

// Decode returns the HashPuzzle encoded by s, or nil when s is not a
// canonical hash puzzle locking script.
func Decode(s *script.Script) *HashPuzzle {
	chunks, err := s.Chunks()
	if err != nil || len(chunks) != 3 {
		return nil
	}
	if chunks[0].Op == script.OpSHA256 &&
		len(chunks[1].Data) == HashLength &&
		chunks[2].Op == script.OpEQUAL {
		return &HashPuzzle{Hash: chunks[1].Data}
	}
	return nil
}

// Unlock returns an unlocking template that reveals the secret preimage.
func Unlock(secret []byte) (*HashPuzzleUnlocker, error) {
	if len(secret) == 0 {
		return nil, ErrNoSecret
	}
	return &HashPuzzleUnlocker{Secret: secret}, nil
}

// HashPuzzleUnlocker satisfies a hash puzzle by pushing the secret preimage.
// It implements the go-sdk transaction.UnlockingScriptTemplate interface.
type HashPuzzleUnlocker struct {
	Secret []byte
}

// Sign returns the unlocking script. No signature is produced — the script is
// a single push of the secret and does not depend on the transaction.
func (h *HashPuzzleUnlocker) Sign(_ *transaction.Transaction, _ uint32) (*script.Script, error) {
	if len(h.Secret) == 0 {
		return nil, ErrNoSecret
	}
	s := &script.Script{}
	if err := s.AppendPushData(h.Secret); err != nil {
		return nil, err
	}
	return s, nil
}

// EstimateLength returns the exact unlocking script length for the secret.
func (h *HashPuzzleUnlocker) EstimateLength(_ *transaction.Transaction, _ uint32) uint32 {
	return pushDataLength(len(h.Secret))
}

// pushDataLength returns the serialized length of a single minimal push of n
// bytes of data.
func pushDataLength(n int) uint32 {
	un := uint32(n) //nolint:gosec // n is a script push length, far below MaxUint32
	switch {
	case n <= 75:
		return un + 1
	case n <= 0xff:
		return un + 2 // OP_PUSHDATA1 <len>
	case n <= 0xffff:
		return un + 3 // OP_PUSHDATA2 <len:2>
	default:
		return un + 5 // OP_PUSHDATA4 <len:4>
	}
}

// GenerateSecretPair returns a fresh random 32-byte secret and its SHA-256
// hash, mirroring the truth-machine HashPuzzle.generateSecretPair helper.
// Use the hash in Lock and keep the secret for Unlock.
func GenerateSecretPair() (secret, hash32 []byte, err error) {
	secret = make([]byte, SecretLength)
	if _, err = rand.Read(secret); err != nil {
		return nil, nil, err
	}
	return secret, primitives.Sha256(secret), nil
}
