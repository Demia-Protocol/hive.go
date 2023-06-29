package valuerange

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/hive.go/stringify"
)

// EndPoint contains information about where ValueRanges start and end. It combines a threshold value with a BoundType.
type EndPoint struct {
	value     Value
	boundType BoundType
}

// NewEndPoint create a new EndPoint from the given details.
func NewEndPoint(value Value, boundType BoundType) *EndPoint {
	return &EndPoint{
		value:     value,
		boundType: boundType,
	}
}

// EndPointFromBytes unmarshals an EndPoint from a sequence of bytes.
func EndPointFromBytes(endPointBytes []byte) (endPoint *EndPoint, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(endPointBytes)
	if endPoint, err = EndPointFromMarshalUtil(marshalUtil); err != nil {
		err = ierrors.Wrap(err, "failed to parse EndPoint from MarshalUtil")

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// EndPointFromMarshalUtil unmarshals an EndPoint using a MarshalUtil (for easier unmarshalling).
func EndPointFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (endPoint *EndPoint, err error) {
	endPoint = &EndPoint{}
	if endPoint.value, err = ValueFromMarshalUtil(marshalUtil); err != nil {
		err = ierrors.Wrap(err, "failed to parse Value from MarshalUtil")

		return
	}
	if endPoint.boundType, err = BoundTypeFromMarshalUtil(marshalUtil); err != nil {
		err = ierrors.Wrap(err, "failed to parse BoundType from MarshalUtil")

		return
	}

	return
}

// Value returns the Value of the EndPoint.
func (e *EndPoint) Value() Value {
	return e.value
}

// BoundType returns the BoundType of the EndPoint.
func (e *EndPoint) BoundType() BoundType {
	return e.boundType
}

// Bytes returns a marshaled version of the EndPoint.
func (e *EndPoint) Bytes() []byte {
	return marshalutil.New().
		Write(e.value).
		Write(e.boundType).
		Bytes()
}

// String returns a human-readable version of the EndPoint.
func (e *EndPoint) String() string {
	return stringify.Struct("EndPoint",
		stringify.NewStructField("value", e.value),
		stringify.NewStructField("boundType", e.boundType),
	)
}
