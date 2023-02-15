package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/cerrors"
	"github.com/iotaledger/hive.go/core/crypto/ed25519"
	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

// IDLength defines the length of an ID.
const IDLength = sha256.Size

// ID is a unique identifier for each peer.
type ID [IDLength]byte

// NewID computes the ID corresponding to the given public key.
func NewID(key ed25519.PublicKey) ID {
	return sha256.Sum256(lo.PanicOnErr(key.Bytes()))
}

// IDFromMarshalUtil unmarshals an ID using a MarshalUtil (for easier unmarshaling).
func IDFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (id ID, err error) {
	idBytes, err := marshalUtil.ReadBytes(IDLength)
	if err != nil {
		err = xerrors.Errorf("failed to parse ID (%v): %w", err, cerrors.ErrParseBytesFailed)

		return
	}

	copy(id[:], idBytes)

	return
}

// Bytes returns the byte slice representation of the ID.
func (id ID) Bytes() ([]byte, error) {
	return id[:], nil
}

// FromBytes decodes ID from bytes.
func (id *ID) FromBytes(bytes []byte) (consumedBytes int, err error) {
	if consumedBytes, err = serix.DefaultAPI.Decode(context.Background(), bytes, id); err != nil {
		return consumedBytes, errors.Errorf("failed to decode node identity from bytes: %w", err)
	}

	return
}

// String returns a shortened version of the ID as a base58 encoded string.
func (id ID) String() string {
	if idAlias, exists := idAliases[id]; exists {
		return "ID(" + idAlias + ")"
	}

	return id.EncodeBase58()[:8]
}

// EncodeBase58 returns a full version of the ID as a base58 encoded string.
func (id ID) EncodeBase58() string {
	return base58.Encode(id[:])
}

// DecodeIDBase58 decodes a base58 encoded ID.
func DecodeIDBase58(s string) (ID, error) {
	b, err := base58.Decode(s)
	if err != nil {
		return ID{}, errors.Wrap(err, "failed to decode ID from base58 string")
	}
	var id ID
	copy(id[:], b)

	return id, nil
}

// ParseID parses a hex encoded ID.
func ParseID(s string) (ID, error) {
	var id ID
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return id, err
	}
	if len(b) != len(ID{}) {
		return id, fmt.Errorf("invalid length: need %d hex chars", hex.EncodedLen(len(ID{})))
	}
	copy(id[:], b)

	return id, nil
}

// RandomIDInsecure creates a random id which can for example be used in unit tests.
// The result is not cryptographically secure.
func RandomIDInsecure() (id ID, err error) {
	// generate a random sequence of bytes
	idBytes := make([]byte, sha256.Size)
	//nolint:gosec // we do not care about weak random numbers here
	if _, err = rand.Read(idBytes); err != nil {
		return
	}

	// copy the generated bytes into the result
	copy(id[:], idBytes)

	return
}

// idAliases contains a list of aliases registered for a set of IDs.
var idAliases = make(map[ID]string)

// RegisterIDAlias registers an alias that will modify the String() output of the ID to show a human
// readable string instead of the base58 encoded version of itself.
func RegisterIDAlias(id ID, alias string) {
	idAliases[id] = alias
}

// UnregisterIDAliases removes all aliases registered through the RegisterIDAlias function.
func UnregisterIDAliases() {
	idAliases = make(map[ID]string)
}
