package hashpuzzle

import (
	"testing"

	primitives "github.com/bsv-blockchain/go-sdk/primitives/hash"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/script/interpreter"
	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecretPair(t *testing.T) {
	secret, hash32, err := GenerateSecretPair()
	require.NoError(t, err)
	require.Len(t, secret, SecretLength)
	require.Len(t, hash32, HashLength)
	require.Equal(t, primitives.Sha256(secret), hash32)

	secret2, _, err := GenerateSecretPair()
	require.NoError(t, err)
	require.NotEqual(t, secret, secret2, "secrets must be random")
}

func TestLockAndDecode_RoundTrip(t *testing.T) {
	_, hash32, err := GenerateSecretPair()
	require.NoError(t, err)

	lockingScript, err := Lock(hash32)
	require.NoError(t, err)
	require.Len(t, []byte(*lockingScript), HashLength+3) // OP_SHA256 + push(1+32) + OP_EQUAL

	decoded := Decode(lockingScript)
	require.NotNil(t, decoded)
	require.Equal(t, hash32, decoded.Hash)
}

func TestLock_InvalidHashLength(t *testing.T) {
	_, err := Lock([]byte{0x01, 0x02})
	require.ErrorIs(t, err, ErrInvalidHashLength)

	_, err = Lock(make([]byte, 33))
	require.ErrorIs(t, err, ErrInvalidHashLength)
}

func TestDecode_InvalidScripts(t *testing.T) {
	// Not a script with 3 chunks
	invalid := script.Script([]byte{0x00, 0x01})
	require.Nil(t, Decode(&invalid))

	// Right shape, wrong opcodes (P2PKH-ish)
	s := &script.Script{}
	_ = s.AppendOpcodes(script.OpDUP)
	_ = s.AppendPushData(make([]byte, 32))
	_ = s.AppendOpcodes(script.OpEQUAL)
	require.Nil(t, Decode(s))

	// Wrong hash length
	s = &script.Script{}
	_ = s.AppendOpcodes(script.OpSHA256)
	_ = s.AppendPushData(make([]byte, 20))
	_ = s.AppendOpcodes(script.OpEQUAL)
	require.Nil(t, Decode(s))
}

func TestUnlock_Validation(t *testing.T) {
	_, err := Unlock(nil)
	require.ErrorIs(t, err, ErrNoSecret)

	unlocker, err := Unlock([]byte("secret"))
	require.NoError(t, err)
	require.NotNil(t, unlocker)
}

func TestUnlock_EstimateLengthMatchesActual(t *testing.T) {
	for _, n := range []int{1, 32, 75, 76, 255, 256} {
		secret := make([]byte, n)
		unlocker, err := Unlock(secret)
		require.NoError(t, err)

		unlockingScript, err := unlocker.Sign(nil, 0)
		require.NoError(t, err)
		require.Equal(t, int(unlocker.EstimateLength(nil, 0)), len(*unlockingScript),
			"estimate must match actual for secret length %d", n)
	}
}

// TestSpend_InterpreterValidates locks an output with a hash puzzle, spends it
// revealing the preimage, and validates the pair with the go-sdk interpreter.
func TestSpend_InterpreterValidates(t *testing.T) {
	secret, hash32, err := GenerateSecretPair()
	require.NoError(t, err)

	lockingScript, err := Lock(hash32)
	require.NoError(t, err)

	sourceTx := transaction.NewTransaction()
	sourceTx.AddOutput(&transaction.TransactionOutput{
		Satoshis:      240,
		LockingScript: lockingScript,
	})

	spendTx := transaction.NewTransaction()
	unlocker, err := Unlock(secret)
	require.NoError(t, err)
	spendTx.AddInputFromTx(sourceTx, 0, unlocker)

	require.NoError(t, spendTx.Sign())

	err = interpreter.NewEngine().Execute(
		interpreter.WithTx(spendTx, 0, sourceTx.Outputs[0]),
		interpreter.WithForkID(),
		interpreter.WithAfterGenesis(),
	)
	require.NoError(t, err, "interpreter must accept the revealed preimage")
}

// TestSpend_WrongSecretFails proves a wrong preimage is rejected.
func TestSpend_WrongSecretFails(t *testing.T) {
	_, hash32, err := GenerateSecretPair()
	require.NoError(t, err)
	wrongSecret, _, err := GenerateSecretPair()
	require.NoError(t, err)

	lockingScript, err := Lock(hash32)
	require.NoError(t, err)

	sourceTx := transaction.NewTransaction()
	sourceTx.AddOutput(&transaction.TransactionOutput{
		Satoshis:      240,
		LockingScript: lockingScript,
	})

	spendTx := transaction.NewTransaction()
	unlocker, err := Unlock(wrongSecret)
	require.NoError(t, err)
	spendTx.AddInputFromTx(sourceTx, 0, unlocker)

	require.NoError(t, spendTx.Sign())

	err = interpreter.NewEngine().Execute(
		interpreter.WithTx(spendTx, 0, sourceTx.Outputs[0]),
		interpreter.WithForkID(),
		interpreter.WithAfterGenesis(),
	)
	require.Error(t, err, "interpreter must reject a wrong preimage")
}
