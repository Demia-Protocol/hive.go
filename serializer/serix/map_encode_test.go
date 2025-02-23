package serix_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type Identifier [blake2b.Size256]byte

type serializableStruct struct {
	bytes Identifier `serix:""`
	index uint64     `serix:""`
}

func (s serializableStruct) EncodeJSON() (any, error) {
	return fmt.Sprintf("%s:%d", base58.Encode(s.bytes[:]), s.index), nil
}

func (s *serializableStruct) DecodeJSON(val any) error {
	serialized, ok := val.(string)
	if !ok {
		return ierrors.New("incorrect type")
	}

	parts := strings.Split(serialized, ":")
	bytes, err := base58.Decode(parts[0])
	if err != nil {
		return err
	}
	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	copy(s.bytes[:], bytes)
	s.index = uint64(idx)

	return nil
}

func TestMapEncodeDecode(t *testing.T) {
	type paras struct {
		api *serix.API
		in  any
	}

	type test struct {
		name     string
		paras    paras
		expected string
	}

	tests := []test{
		{
			name: "basic types",
			paras: func() paras {
				type example struct {
					Uint64    uint64  `serix:""`
					Uint32    uint32  `serix:""`
					Uint16    uint16  `serix:""`
					Uint8     uint8   `serix:""`
					Int64     int64   `serix:""`
					Int32     int32   `serix:""`
					Int16     int16   `serix:""`
					Int8      int8    `serix:""`
					ZeroInt32 int32   `serix:",omitempty"`
					Float32   float32 `serix:""`
					Float64   float64 `serix:""`
					String    string  `serix:""`
					Bool      bool    `serix:""`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(42))))

				return paras{
					api: api,
					in: &example{
						Uint64:    64,
						Uint32:    32,
						Uint16:    16,
						Uint8:     8,
						Int64:     -64,
						Int32:     -32,
						Int16:     -16,
						Int8:      -8,
						ZeroInt32: 0,
						Float32:   0.33,
						Float64:   0.44,
						String:    "abcd",
						Bool:      true,
					},
				}
			}(),
			expected: `{
				"type": 42,
				"uint64": "64",
				"uint32": 32,
				"uint16": 16,
				"uint8": 8,
				"int64": "-64",
				"int32": -32,
				"int16": -16,
				"int8": -8,
				"float32": "0.33000001311302185",
				"float64": "0.44",
				"string": "abcd",
				"bool": true
			}`,
		},
		{
			name: "big int",
			paras: func() paras {
				type example struct {
					BigInt *big.Int `serix:""`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(66))))

				return paras{
					api: api,
					in: &example{
						BigInt: big.NewInt(1337),
					},
				}
			}(),
			expected: `{
				"type": 66,
 				"bigInt": "0x539"
			}`,
		},
		{
			name: "map",
			paras: func() paras {
				type example struct {
					Map map[string]string `serix:""`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(99))))

				return paras{
					api: api,
					in: &example{
						Map: map[string]string{
							"alice": "123",
						},
					},
				}
			}(),
			expected: `{
				"type": 99,
 				"map": {
					"alice": "123"
				}
			}`,
		},
		{
			name: "byte slices/arrays",
			paras: func() paras {

				type example struct {
					ByteSlice         []byte    `serix:""`
					Array             [5]byte   `serix:""`
					SliceOfByteSlices [][]byte  `serix:""`
					SliceOfByteArrays [][3]byte `serix:""`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(5))))

				return paras{
					api: api,
					in: &example{
						ByteSlice: []byte{1, 2, 3, 4, 5},
						Array:     [5]byte{5, 4, 3, 2, 1},
						SliceOfByteSlices: [][]byte{
							{1, 2, 3},
							{3, 2, 1},
						},
						SliceOfByteArrays: [][3]byte{
							{5, 6, 7},
							{7, 6, 5},
						},
					},
				}
			}(),
			expected: `{
				"type": 5,
 				"byteSlice": "0x0102030405",
				"array": "0x0504030201",
				"sliceOfByteSlices": [
					"0x010203",
					"0x030201"
				],
				"sliceOfByteArrays": [
					"0x050607",
					"0x070605"
				]
			}`,
		},
		{
			name: "inner struct",
			paras: func() paras {
				type (
					inner struct {
						String string `serix:""`
					}

					example struct {
						inner `serix:""`
					}
				)

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(22))))

				return paras{
					api: api,
					in: &example{
						inner{String: "abcd"},
					},
				}
			}(),
			expected: `{
				"type": 22,
 				"string": "abcd"
			}`,
		},
		{
			name: "interface & direct pointer",
			paras: func() paras {
				type (
					InterfaceType      interface{}
					InterfaceTypeImpl1 [4]byte
					OtherObj           [2]byte

					example struct {
						Interface InterfaceType `serix:""`
						Other     *OtherObj     `serix:""`
					}
				)

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(33))))
				must(api.RegisterTypeSettings(InterfaceTypeImpl1{},
					serix.TypeSettings{}.WithObjectType(uint8(5)).WithFieldKey("customInnerKey")),
				)
				must(api.RegisterInterfaceObjects((*InterfaceType)(nil), (*InterfaceTypeImpl1)(nil)))
				must(api.RegisterTypeSettings(OtherObj{},
					serix.TypeSettings{}.WithObjectType(uint8(2)).WithFieldKey("otherObjKey")),
				)

				return paras{
					api: api,
					in: &example{
						Interface: &InterfaceTypeImpl1{1, 2, 3, 4},
						Other:     &OtherObj{1, 2},
					},
				}
			}(),
			expected: `{
				"type": 33,
 				"interface": {
					"type": 5,
					"customInnerKey": "0x01020304"
				},
				"other": {
					"type": 2,
					"otherObjKey": "0x0102"
				}
			}`,
		},
		{
			name: "slice of interface",
			paras: func() paras {
				type (
					Interface interface{}
					Impl1     struct {
						String string `serix:""`
					}
					Impl2 struct {
						Uint16 uint16 `serix:""`
					}

					example struct {
						Slice []Interface `serix:""`
					}
				)

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(11))))
				must(api.RegisterTypeSettings(Impl1{}, serix.TypeSettings{}.WithObjectType(uint8(0))))
				must(api.RegisterTypeSettings(Impl2{}, serix.TypeSettings{}.WithObjectType(uint8(1))))
				must(api.RegisterInterfaceObjects((*Interface)(nil), (*Impl1)(nil), (*Impl2)(nil)))

				return paras{
					api: api,
					in: &example{
						Slice: []Interface{
							&Impl1{String: "impl1"},
							&Impl2{Uint16: 1337},
						},
					},
				}
			}(),
			expected: `{
				"type": 11,
 				"slice": [
					{
						"type": 0,
						"string": "impl1"
					},
					{
						"type": 1,
						"uint16": 1337
					}
				]
			}`,
		},
		{
			name: "no map key",
			paras: func() paras {
				type example struct {
					CaptainHook string `serix:""`
					LiquidSoul  int64  `serix:""`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(23))))

				return paras{
					api: api,
					in: &example{
						CaptainHook: "jump",
						LiquidSoul:  30,
					},
				}
			}(),
			expected: `{
				"type": 23,
 				"captainHook": "jump",
				"liquidSoul": "30"
			}`,
		},
		{
			name: "time",
			paras: func() paras {
				type example struct {
					CreationDate time.Time `serix:""`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(23))))

				uint64Time, err := serix.DecodeUint64("1660301478120072000")
				require.NoError(t, err)
				exampleTime := serializer.Uint64ToTime(uint64Time)

				return paras{
					api: api,
					in: &example{
						CreationDate: exampleTime,
					},
				}
			}(),
			expected: `{
				"type": 23,
 				"creationDate": "1660301478120072000"
			}`,
		},

		{
			name: "serializable",
			paras: func() paras {
				type example struct {
					Entries map[serializableStruct]struct{} `serix:""`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(23))))

				return paras{
					api: api,
					in: &example{
						Entries: map[serializableStruct]struct{}{
							{
								bytes: blake2b.Sum256([]byte("test")),
								index: 1,
							}: {},
						},
					},
				}
			}(),
			expected: `{
				"type": 23,
				"entries": {
					"As3ZuwnL9LpoW3wz8HoDpHtZqJ4dhPFFnv87GYrnCYKj:1": {}
				}
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// encode input to a map
			out, err := test.paras.api.MapEncode(context.Background(), test.paras.in, serix.WithValidation())
			require.NoError(t, err)
			jsonOut, err := json.MarshalIndent(out, "", "\t")
			require.NoError(t, err)

			// re-arrange expected json output to conform to same indentation
			aux := orderedmap.New()
			require.NoError(t, json.Unmarshal([]byte(test.expected), aux))
			expectedJSON, err := json.MarshalIndent(aux, "", "\t")
			require.NoError(t, err)
			require.EqualValues(t, string(expectedJSON), string(jsonOut))

			mapTarget := map[string]any{}
			require.NoError(t, json.Unmarshal(expectedJSON, &mapTarget))

			dest := reflect.New(reflect.TypeOf(test.paras.in).Elem()).Interface()
			require.NoError(t, test.paras.api.MapDecode(context.Background(), mapTarget, dest))
			require.EqualValues(t, test.paras.in, dest)
		})
	}
}
