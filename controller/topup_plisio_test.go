package controller

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signPlisioBody mirrors Plisio's documented Node signing for json=true
// callbacks: JSON.stringify(body without verify_hash), HMAC-SHA1 with the
// secret key, hex-encoded.
func signPlisioBody(t *testing.T, bodyWithoutHash string, secret string) string {
	t.Helper()
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(bodyWithoutHash))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyPlisioSignature_AcceptsGenuineCallback(t *testing.T) {
	const secret = "plisio-secret-key"
	// Canonical (signing) payload: compact JSON, verify_hash omitted, fields in
	// the same order Plisio sends them. Plisio echoes order_number back so we
	// can match the local order.
	canonical := `{"txn_id":"abc123","ipn_type":"invoice","order_number":"PLSUSR1NOxyz1","order_name":"Top-up #5","status":"completed","amount":"0.0001","currency":"BTC","source_currency":"USD","source_amount":"5.00"}`
	sig := signPlisioBody(t, canonical, secret)

	// The delivered body is the canonical payload plus verify_hash appended.
	delivered := canonical[:len(canonical)-1] + `,"verify_hash":"` + sig + `"}`

	require.True(t, verifyPlisioSignature([]byte(delivered), sig, secret))
}

func TestVerifyPlisioSignature_RejectsTamperedAmount(t *testing.T) {
	const secret = "plisio-secret-key"
	canonical := `{"order_number":"PLSUSR1NOxyz1","status":"completed","source_amount":"5.00"}`
	sig := signPlisioBody(t, canonical, secret)

	tampered := `{"order_number":"PLSUSR1NOxyz1","status":"completed","source_amount":"5000.00","verify_hash":"` + sig + `"}`
	require.False(t, verifyPlisioSignature([]byte(tampered), sig, secret))
}

func TestVerifyPlisioSignature_RejectsBadSecret(t *testing.T) {
	canonical := `{"order_number":"PLSUSR1NOxyz1","status":"completed"}`
	sig := signPlisioBody(t, canonical, "real-secret")
	delivered := canonical[:len(canonical)-1] + `,"verify_hash":"` + sig + `"}`

	require.False(t, verifyPlisioSignature([]byte(delivered), sig, "wrong-secret"))
}

// The provider mismatch guard is enforced in PlisioWebhook; here we only verify
// the signature primitive handles an empty secret/hash defensively.
func TestVerifyPlisioSignature_RejectsEmptyInputs(t *testing.T) {
	require.False(t, verifyPlisioSignature([]byte(`{}`), "hash", ""))
	require.False(t, verifyPlisioSignature([]byte(`{}`), "", "secret"))
	require.False(t, verifyPlisioSignature([]byte(`{}`), "", ""))
}

func TestPlisioCanonicalBody_DropsVerifyHashAnyPosition(t *testing.T) {
	// verify_hash in the middle of the object must still be removed while every
	// other value keeps its original bytes and order.
	input := `{"a":"1","verify_hash":"zzz","b":"2","c":"3"}`
	got, err := plisioCanonicalBody([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, `{"a":"1","b":"2","c":"3"}`, got)

	// verify_hash last (the common case).
	inputLast := `{"a":"1","b":"2","verify_hash":"zzz"}`
	got, err = plisioCanonicalBody([]byte(inputLast))
	require.NoError(t, err)
	assert.Equal(t, `{"a":"1","b":"2"}`, got)
}

func TestPlisioCanonicalBody_PreservesValueBytes(t *testing.T) {
	// Numbers and nested structures are preserved byte-for-byte (matches
	// JSON.stringify semantics on the receiver side).
	input := `{"n":42,"f":3.14,"arr":[1,2,3],"obj":{"k":"v"}}`
	got, err := plisioCanonicalBody([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, input, got)
}
