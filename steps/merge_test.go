package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vladimir-rom/logex/pipeline"
)

func TestMergeMainPath(t *testing.T) {
	assert.Equal(t,
		[]JSON{
			{"ts": "1", "prop1": "val_1_1"},
			{"ts": "2", "prop1": "val_2_1"},
			{"ts": "3", "prop1": "val_1_2"},
			{"ts": "4", "prop1": "val_2_2"},
			{"ts": "5", "prop1": "val_2_3"},
			{"ts": "6", "prop1": "val_1_3"},
		},
		seqToSlice(Merge(
			pipeline.PipelineOptions{},
			[]string{"ts"},
			[]pipeline.Seq[JSON]{
				sliceToSeq([]JSON{
					{"ts": "1", "prop1": "val_1_1"},
					{"ts": "3", "prop1": "val_1_2"},
					{"ts": "6", "prop1": "val_1_3"},
				}),
				sliceToSeq([]JSON{
					{"ts": "2", "prop1": "val_2_1"},
					{"ts": "4", "prop1": "val_2_2"},
					{"ts": "5", "prop1": "val_2_3"},
				}),
			},
		)))
}

func TestMergeByManyProps(t *testing.T) {
	assert.Equal(t,
		[]JSON{
			{"ts1": "1", "prop1": "val_1_1"},
			{"ts2": "2", "prop1": "val_2_1"},
			{"ts1": "3", "prop1": "val_1_2"},
			{"ts2": "4", "prop1": "val_2_2"},
			{"ts2": "5", "prop1": "val_2_3"},
			{"ts1": "6", "prop1": "val_1_3"},
		},
		seqToSlice(Merge(
			pipeline.PipelineOptions{},
			[]string{"ts1", "ts2"},
			[]pipeline.Seq[JSON]{
				sliceToSeq([]JSON{
					{"ts1": "1", "prop1": "val_1_1"},
					{"ts1": "3", "prop1": "val_1_2"},
					{"ts1": "6", "prop1": "val_1_3"},
				}),
				sliceToSeq([]JSON{
					{"ts2": "2", "prop1": "val_2_1"},
					{"ts2": "4", "prop1": "val_2_2"},
					{"ts2": "5", "prop1": "val_2_3"},
				}),
			},
		)))
}
func TestMergeSingleSeq(t *testing.T) {
	assert.Equal(t,
		[]JSON{
			{"ts": "1", "prop1": "val_1_1"},
			{"ts": "3", "prop1": "val_1_2"},
			{"ts": "6", "prop1": "val_1_3"},
		},
		seqToSlice(Merge(
			pipeline.PipelineOptions{},
			[]string{"ts"},
			[]pipeline.Seq[JSON]{
				sliceToSeq([]JSON{
					{"ts": "1", "prop1": "val_1_1"},
					{"ts": "3", "prop1": "val_1_2"},
					{"ts": "6", "prop1": "val_1_3"},
				}),
			},
		)))
}

func TestMergeWithEmptySeq(t *testing.T) {
	assert.Equal(t,
		[]JSON{
			{"ts": "1", "prop1": "val_1_1"},
			{"ts": "3", "prop1": "val_1_2"},
			{"ts": "6", "prop1": "val_1_3"},
		},
		seqToSlice(Merge(
			pipeline.PipelineOptions{},
			[]string{"ts"},
			[]pipeline.Seq[JSON]{
				sliceToSeq([]JSON{
					{"ts": "1", "prop1": "val_1_1"},
					{"ts": "3", "prop1": "val_1_2"},
					{"ts": "6", "prop1": "val_1_3"},
				}),
				sliceToSeq([]JSON{}),
			},
		)))
}

func sliceToSeq(in []JSON) pipeline.Seq[JSON] {
	return func(yield pipeline.Yield[JSON]) {
		for _, i := range in {
			if !yield(pipeline.Item[JSON]{Value: i}, nil) {
				break
			}
		}
	}
}

func seqToSlice(in pipeline.Seq[JSON]) []JSON {
	res := make([]JSON, 0)
	for json, _ := range in {
		res = append(res, json.Value)
	}
	return res
}
