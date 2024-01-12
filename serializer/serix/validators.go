package serix

import (
	"context"
	"reflect"
	"sync"

	"github.com/iotaledger/hive.go/ierrors"
)

type validators struct {
	bytesValidator     reflect.Value
	syntacticValidator reflect.Value
}

func parseValidatorFunc(validatorFn interface{}) (reflect.Value, error) {
	if validatorFn == nil {
		return reflect.Value{}, nil
	}
	funcValue := reflect.ValueOf(validatorFn)
	if !funcValue.IsValid() || funcValue.IsZero() {
		return reflect.Value{}, nil
	}
	if funcValue.Kind() != reflect.Func {
		return reflect.Value{}, ierrors.Errorf(
			"validator must be a function, got %T(%s)", validatorFn, funcValue.Kind(),
		)
	}
	funcType := funcValue.Type()
	if funcType.NumIn() != 2 {
		return reflect.Value{}, ierrors.New("validator func must have two arguments")
	}
	firstArgType := funcType.In(0)
	if firstArgType != ctxType {
		return reflect.Value{}, ierrors.New("validator func's first argument must be context")
	}
	if funcType.NumOut() != 1 {
		return reflect.Value{}, ierrors.Errorf("validator func must have one return value, got %d", funcType.NumOut())
	}
	returnType := funcType.Out(0)
	if returnType != errorType {
		return reflect.Value{}, ierrors.Errorf("validator func must have 'error' return type, got %s", returnType)
	}

	return funcValue, nil
}

func checkBytesValidatorSignature(funcValue reflect.Value) error {
	funcType := funcValue.Type()
	argumentType := funcType.In(1)
	if argumentType != bytesType {
		return ierrors.Errorf("bytesValidatorFn's argument must be bytes, got %s", argumentType)
	}

	return nil
}

func checkSyntacticValidatorSignature(objectType reflect.Type, funcValue reflect.Value) error {
	funcType := funcValue.Type()
	argumentType := funcType.In(1)
	if argumentType != objectType {
		return ierrors.Errorf(
			"syntacticValidatorFn's argument must have the same type as the object it was registered for, "+
				"objectType=%s, argumentType=%s",
			objectType, argumentType,
		)
	}

	return nil
}

type validatorsRegistry struct {
	// the registered validators for the known objects
	validatorsRegistryMutex sync.RWMutex
	validatorsRegistry      map[reflect.Type]validators
}

func newValidatorsRegistry() *validatorsRegistry {
	return &validatorsRegistry{
		validatorsRegistry: make(map[reflect.Type]validators),
	}
}

func (r *validatorsRegistry) Get(objType reflect.Type) (validators, bool) {
	r.validatorsRegistryMutex.RLock()
	defer r.validatorsRegistryMutex.RUnlock()

	vldtrs, exists := r.validatorsRegistry[objType]

	return vldtrs, exists
}

func (r *validatorsRegistry) AddValidators(objType reflect.Type, vldtrs validators) {
	r.validatorsRegistryMutex.Lock()
	defer r.validatorsRegistryMutex.Unlock()

	r.validatorsRegistry[objType] = vldtrs
}

func (r *validatorsRegistry) RegisterValidators(obj any, bytesValidatorFn func(context.Context, []byte) error, syntacticValidatorFn interface{}) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return ierrors.New("'obj' is a nil interface, it needs to be a valid type")
	}

	bytesValidatorValue, err := parseValidatorFunc(bytesValidatorFn)
	if err != nil {
		return ierrors.Wrap(err, "failed to parse bytesValidatorFn")
	}

	syntacticValidatorValue, err := parseValidatorFunc(syntacticValidatorFn)
	if err != nil {
		return ierrors.Wrap(err, "failed to parse syntacticValidatorFn")
	}

	vldtrs := validators{}
	if bytesValidatorValue.IsValid() {
		if err := checkBytesValidatorSignature(bytesValidatorValue); err != nil {
			return ierrors.WithStack(err)
		}
		vldtrs.bytesValidator = bytesValidatorValue
	}

	if syntacticValidatorValue.IsValid() {
		if err := checkSyntacticValidatorSignature(objType, syntacticValidatorValue); err != nil {
			return ierrors.WithStack(err)
		}
		vldtrs.syntacticValidator = syntacticValidatorValue
	}

	r.AddValidators(objType, vldtrs)

	return nil
}
