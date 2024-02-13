package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vladimir-rom/logex/steps"
)

func TestNoModification(t *testing.T) {
	in := []steps.JSON{{"field1": "value1", "field2": "value2"}}
	testPipelineJson(t, &filterParams{}, in, in)
}

func TestSelect(t *testing.T) {
	testPipelineJson(t,
		&filterParams{selectProps: []string{"field1"}},
		[]steps.JSON{{"field1": "value1", "field2": "value2"}},
		[]steps.JSON{{"field1": "value1"}})
}

func TestFirst(t *testing.T) {
	testPipelineJson(t,
		&filterParams{first: 1},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value1"}})
}

func TestLast(t *testing.T) {
	testPipelineJson(t,
		&filterParams{last: 1},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value2"}})
}

func TestKQL(t *testing.T) {
	testPipelineJson(t,
		&filterParams{kqlFilter: "field:value2"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value2"}})
}

func testPipelineJson(t *testing.T, params *filterParams, in []steps.JSON, expectedOut []steps.JSON) {
	t.Helper()
	inBuffer := bytes.Buffer{}
	for _, j := range in {
		b, err := json.Marshal(j)
		assert.NoError(t, err)
		inBuffer.Write(b)
		inBuffer.WriteByte('\n')
	}

	outBuffer := bytes.Buffer{}

	err := runPipeline(params, &inBuffer, &outBuffer)
	assert.NoError(t, err)

	out := make([]steps.JSON, 0, len(in))
	decoder := json.NewDecoder(&outBuffer)
	for outBuffer.Len() > 0 {
		j := make(steps.JSON)
		err = decoder.Decode(&j)
		assert.NoError(t, err)
		out = append(out, j)
	}
	assert.Equal(t, expectedOut, out)
}
