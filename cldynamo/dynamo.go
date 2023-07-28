// Package cldynamo provides re-usable DynamoDB utilities.
package cldynamo

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// withJSONTagKey configures the encoding to just the jsong tag key.
func withJSONTagKey(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" }

// MarshalMap will marshal a protobuf message 'msg' into a map of DynamoDB attributes such that it is
// stable over the field numbers of a message.
func MarshalMap(msg proto.Message) (map[string]types.AttributeValue, error) {
	mav, err := attributevalue.MarshalMapWithOptions(msg, withJSONTagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to perform initial marshal: %w", err)
	}

	remapToStable(mav, msg.ProtoReflect())

	return mav, nil
}

// remapToStable uses protobuf reflection into to rerusively change the key used in the attribute map
// to be the field number. This is more stable as the protobuf schema changes over time.
func remapToStable(mav map[string]types.AttributeValue, rmsg protoreflect.Message) {
	rmsg.Range(func(fld protoreflect.FieldDescriptor, prv protoreflect.Value) bool {
		// step 1: we need to retrieve the actual AttributeValue that we want to set
		// with the stable field number as a key. Oneof map values need to be unwrapped
		// specifically as default marshal behaviour keeps the actual value in a container.
		var val types.AttributeValue
		if oneof := fld.ContainingOneof(); oneof != nil {
			oomv := mav[strcase.ToCamel(string(oneof.Name()))]
			oomav := mustBeAttrMap(oneof.Name(), "oneof", oomv)

			val = oomav.Value[strcase.ToCamel(string(fld.Name()))]
		} else {
			val = mav[string(fld.Name())]
		}

		// step 2: in case the field is a message kind we need to remap recursively.
		// repeated fields need to be iterated before they can be recursed.
		if fld.Message() != nil {
			switch {
			case fld.IsMap():
				avMap := mustBeAttrMap(fld.Name(), "field", val)
				prv.Map().Range(func(mk protoreflect.MapKey, elVal protoreflect.Value) bool {
					el, ok := avMap.Value[mk.String()]
					if !ok {
						panic("failed to find map element for key: " + mk.String())
					}

					elMap := mustBeAttrMap(fld.Name(), "map element", el)

					remapToStable(elMap.Value, elVal.Message())

					return true
				})

			case fld.IsList():
				avList := mustBeAttrList(fld.Name(), "field", val)
				for i, el := range avList.Value {
					elMap := mustBeAttrMap(fld.Name(), "list element", el)
					remapToStable(elMap.Value, prv.List().Get(i).Message())
				}

			default:
				avMap := mustBeAttrMap(fld.Name(), "field", val)
				remapToStable(avMap.Value, prv.Message())
			}
		}

		// step3: now set the stable key, and unset the unstable key.
		mav[strconv.FormatInt(int64(fld.Number()), 10)] = val
		delete(mav, string(fld.Name()))

		return true
	})

	// step 4: oneofs will always leave a unstable key, we remove them here.
	desc := rmsg.Descriptor()
	for i := 0; i < desc.Oneofs().Len(); i++ {
		delete(mav, strcase.ToCamel(string(desc.Oneofs().Get(i).Name())))
	}
}

// mustBeAttrMap asserts that the AttributeValue is a map type or it panics.
func mustBeAttrMap(fieldName protoreflect.Name, sel string, v types.AttributeValue) *types.AttributeValueMemberM {
	mv, ok := v.(*types.AttributeValueMemberM)
	if !ok {
		panic(fmt.Sprintf(`field '%s' (%s): did not marshal to AttributeValueMemberM, got: %T`, fieldName, sel, v))
	}

	return mv
}

// mustBeAttrList asserts that the AttributeValue is a map type or it panics.
func mustBeAttrList(fieldName protoreflect.Name, sel string, v types.AttributeValue) *types.AttributeValueMemberL {
	mv, ok := v.(*types.AttributeValueMemberL)
	if !ok {
		panic(fmt.Sprintf(`field '%s' (%s): did not marshal to AttributeValueMemberL, got: %T`, fieldName, sel, v))
	}

	return mv
}
