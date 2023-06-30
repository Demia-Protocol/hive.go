package ed25519

import (
	"crypto/ed25519"
	"encoding/json"

	"github.com/mr-tron/base58"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

// PublicKey is the type of Ed25519 public keys.
type PublicKey [PublicKeySize]byte

// PublicKeyFromString parses the given string with base58 encoding and returns a PublicKey.
func PublicKeyFromString(s string) (publicKey PublicKey, err error) {
	b, err := base58.Decode(s)
	if err != nil {
		return publicKey, ierrors.Wrapf(err, "failed to parse public key %s from base58 string", s)
	}
	publicKey, _, err = PublicKeyFromBytes(b)

	return publicKey, err
}

// PublicKeyFromBytes creates a PublicKey from the given bytes.
func PublicKeyFromBytes(bytes []byte) (result PublicKey, consumedBytes int, err error) {
	consumedBytes, err = (&result).FromBytes(bytes)
	return
}

// NativeToPublicKeys converts crypto/ed25519 native public keys into a []PublicKey.
func NativeToPublicKeys(nativePubKeys []ed25519.PublicKey) (result []PublicKey) {
	return lo.Map(nativePubKeys, func(key ed25519.PublicKey) PublicKey {
		return PublicKey(key)
	})
}

// RecoverKey makes sure that key and signature have the correct length
// and verifies whether sig is a valid signature of data by pub.
func RecoverKey(key, data, sig []byte) (result PublicKey, err error) {
	if l := len(key); l != PublicKeySize {
		err = ierrors.Errorf("invalid key length: %d, need %d", l, PublicKeySize)

		return
	}
	if l := len(sig); l != SignatureSize {
		err = ierrors.Errorf("invalid signature length: %d, need %d", l, SignatureSize)

		return
	}
	if !Verify(key, data, sig) {
		err = ierrors.New("invalid signature")

		return
	}

	copy(result[:], key)

	return
}

func ParsePublicKey(marshalUtil *marshalutil.MarshalUtil) (PublicKey, error) {
	id, err := marshalUtil.Parse(func(data []byte) (interface{}, int, error) { return PublicKeyFromBytes(data) })
	if err != nil {
		return PublicKey{}, err
	}

	return id.(PublicKey), nil
}

// VerifySignature reports whether signature is a valid signature of message by publicKey.
func (publicKey PublicKey) VerifySignature(data []byte, signature Signature) bool {
	return Verify(publicKey[:], data, signature[:])
}

// FromBytes initialized the PublicKey from the given bytes.
func (publicKey *PublicKey) FromBytes(bytes []byte) (int, error) {
	if len(bytes) < PublicKeySize {
		return 0, ErrNotEnoughBytes
	}

	copy(publicKey[:], bytes)

	return PublicKeySize, nil
}

// Bytes returns the publicKey in bytes.
func (publicKey PublicKey) Bytes() ([]byte, error) {
	return publicKey[:], nil
}

// String returns a human-readable version of the PublicKey (base58 encoded).
func (publicKey PublicKey) String() string {
	return base58.Encode(publicKey[:])
}

func (publicKey *PublicKey) UnmarshalBinary(bytes []byte) error {
	if _, err := publicKey.FromBytes(bytes); err != nil {
		return err
	}

	return nil
}

// MarshalJSON serializes public key to JSON as base58 encoded string.
func (publicKey PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(publicKey.String())
}

// UnmarshalJSON parses public key from JSON in base58 encoding.
func (publicKey *PublicKey) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	pk, err := PublicKeyFromString(s)
	if err != nil {
		return ierrors.Wrap(err, "failed to parse public key from JSON")
	}
	*publicKey = pk

	return nil
}

// ToEd25519 returns the public key as native crypto/ed25519.PublicKey.
func (publicKey PublicKey) ToEd25519() ed25519.PublicKey {
	nativePubKey := make(ed25519.PublicKey, PublicKeySize)
	copy(nativePubKey[:], publicKey[:])

	return nativePubKey
}
