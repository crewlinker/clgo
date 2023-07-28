package cldynamo_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/crewlinker/clgo/cldynamo"
	testdatav1 "github.com/crewlinker/clgo/cldynamo/testdata/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCldynamo(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "cldynamo")
}

var _ = DescribeTable("marshal/unmarshal", func(inp proto.Message, exp map[string]types.AttributeValue) {
	act, err := cldynamo.MarshalMap(inp)
	Expect(err).ToNot(HaveOccurred())

	diff := cmp.Diff(act, exp, ignoreAttrUnexported())
	Expect(diff).To(BeEmpty())
},
	Entry("some scalar fields",
		&testdatav1.Kitchen{
			KitchenId: "id1",
		},
		map[string]types.AttributeValue{
			"1": &types.AttributeValueMemberS{Value: "id1"},
		}),
	Entry("enum field",
		&testdatav1.Kitchen{
			FridgeBrand: testdatav1.FridgeBrand_FRIDGE_BRAND_SIEMENS,
		},
		map[string]types.AttributeValue{
			"2": &types.AttributeValueMemberN{Value: "1"},
		}),
	Entry("oneof set",
		&testdatav1.Kitchen{TilingStyle: &testdatav1.Kitchen_Terracotta{Terracotta: "foo"}},
		map[string]types.AttributeValue{
			"6": &types.AttributeValueMemberS{Value: "foo"},
		}),
	Entry("oneof kitchen (recurse)",
		&testdatav1.Kitchen{TilingStyle: &testdatav1.Kitchen_YetAnother{
			YetAnother: &testdatav1.Kitchen{KitchenId: "nestedid"},
		}},
		map[string]types.AttributeValue{
			"8": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"1": &types.AttributeValueMemberS{Value: "nestedid"},
			}},
		}),
	Entry("repeated kitchens (recurse)",
		&testdatav1.Kitchen{Others: []*testdatav1.Kitchen{
			{KitchenId: "l1"},
			{KitchenId: "l2"},
		}},
		map[string]types.AttributeValue{
			"5": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"1": &types.AttributeValueMemberS{Value: "l1"},
					}},
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"1": &types.AttributeValueMemberS{Value: "l2"},
					}},
				},
			},
		}),
	Entry("map of kitchens (recurse)",
		&testdatav1.Kitchen{MapOthers: map[int64]*testdatav1.Kitchen{
			100: {KitchenId: "m1"},
			200: {KitchenId: "m2"},
		}},
		map[string]types.AttributeValue{
			"9": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"100": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
					"1": &types.AttributeValueMemberS{Value: "m1"},
				}},
				"200": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
					"1": &types.AttributeValueMemberS{Value: "m2"},
				}},
			}},
		}),
	Entry("timestamp",
		&testdatav1.Kitchen{BuildAt: timestamppb.New(time.Unix(1690528463, 100))},
		map[string]types.AttributeValue{
			"10": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"1": &types.AttributeValueMemberN{Value: "1690528463"},
				"2": &types.AttributeValueMemberN{Value: "100"},
			}},
		}),
)

// option to ignore exported fields of these types.
func ignoreAttrUnexported() cmp.Option {
	return cmpopts.IgnoreUnexported(
		types.AttributeValueMemberS{},
		types.AttributeValueMemberSS{},
		types.AttributeValueMemberNULL{},
		types.AttributeValueMemberM{},
		types.AttributeValueMemberN{},
		types.AttributeValueMemberL{},
		types.AttributeValueMemberBOOL{},
		types.AttributeValueMemberBS{},
	)
}
