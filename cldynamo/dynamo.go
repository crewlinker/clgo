// Package cldynamo provides re-usable DynamoDB utilities.
package cldynamo

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// withEncodingOpts configures the encoding to just the jsong tag key.
func withEncodingOpts(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" }

// withDecodingOpts configures the Decoding to just the jsong tag key.
func withDecodingOpts(eo *attributevalue.DecoderOptions) { eo.TagKey = "json" }

// UnmarshalMap will unmarshal a map of attribute values into a protobuf message. It will use the stable
// field numbers instead of field names.
func UnmarshalMap(mav map[string]types.AttributeValue, msg proto.Message) error {
	remapFromStable(mav, msg.ProtoReflect(), msg.ProtoReflect().Descriptor())

	if err := attributevalue.UnmarshalMapWithOptions(mav, msg, withDecodingOpts); err != nil {
		return fmt.Errorf("failed to attributevalue unmarshal: %w", err)
	}

	return nil
}

// MarshalMap will marshal a protobuf message 'msg' into a map of DynamoDB attributes such that it is
// stable over the field numbers of a message.
func MarshalMap(msg proto.Message) (map[string]types.AttributeValue, error) {
	mav, err := attributevalue.MarshalMapWithOptions(msg, withEncodingOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to attributevalue marshal: %w", err)
	}

	remapToStable(mav, msg.ProtoReflect())

	return mav, nil
}

// remapFromStable uses protobuf descriptor info to turn field number keys into unstable field name
// keys such that regular AttributeValue unmarshalling can fill the protobuf message.
func remapFromStable(mav map[string]types.AttributeValue,
	_ protoreflect.Message,
	dmsg protoreflect.MessageDescriptor,
) {
	for key, val := range mav {
		fdesc, ok := fieldDescByKey(key, dmsg)
		if !ok {
			continue // if no field exists, skip it, maybe it got removed in the schema
		}

		if oneof := fdesc.ContainingOneof(); oneof != nil {
			// FAIL: it looks like there is no way we can create an attribute value that will
			// cause attributevalue.UnmarshalMap to populate the oneOf field of the proto.Message.
			// this would require custom marshalling logic.
		} else {
			mav[string(fdesc.Name())] = val
		}

		delete(mav, key)
	}
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

// fieldDescByKey will get the field descriptor from a field number in string format. Or false if the
// message doesn't describe a field with that number.
func fieldDescByKey(key string, rmsg protoreflect.MessageDescriptor) (protoreflect.FieldDescriptor, bool) {
	nr, err := strconv.ParseInt(key, 10, 32)
	if err != nil {
		panic("non-numeric key in attribute map: " + key)
	}

	fdesc := rmsg.Fields().ByNumber(protowire.Number(nr))
	if fdesc == nil {
		return nil, false
	}

	return fdesc, true
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
