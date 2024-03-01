package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func TestInclude(t *testing.T) {
	testPipelineJson(t,
		&filterParams{include: []string{"Value2"}},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value2"}})
}

func TestIncludeRegexp(t *testing.T) {
	testPipelineJson(t,
		&filterParams{includeRegexp: []string{"value[1-2]"}},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}})
}

func TestExcludeRegexp(t *testing.T) {
	testPipelineJson(t,
		&filterParams{excludeRegexp: []string{"value[1-2]"}},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value3"}})
}

func TestContext(t *testing.T) {
	input := make([]steps.JSON, 0)
	for i := range 5 {
		input = append(input, steps.JSON{"field": fmt.Sprintf("value%d", i+1)})
	}

	testPipelineJson(t,
		&filterParams{include: []string{"Value3"}, context: 1},
		input,
		[]steps.JSON{{"field": "value2"}, {"field": "value3"}, {"field": "value4"}})

	testPipelineJson(t,
		&filterParams{include: []string{"Value1"}, context: 1},
		input,
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}})

	testPipelineJson(t,
		&filterParams{include: []string{"Value5"}, context: 1},
		input,
		[]steps.JSON{{"field": "value4"}, {"field": "value5"}})

	testPipelineJson(t,
		&filterParams{include: []string{"Value3"}, context: 5},
		input,
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}, {"field": "value4"}, {"field": "value5"}})

	testPipelineJson(t,
		&filterParams{include: []string{"Value1111"}, context: 1},
		input,
		[]steps.JSON{})

}

func TestMetadata(t *testing.T) {
	testPipelineJson(t,
		&filterParams{metadata: "rnum"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value1", "rnum": 0.0}, {"field": "value2", "rnum": 1.0}})

	testPipelineJson(t,
		&filterParams{metadata: "rnum:r file:f"},
		[]steps.JSON{{"field": "value1"}},
		[]steps.JSON{{"field": "value1", "r": 0.0, "f": "test"}})

	testPipelineJson(t,
		&filterParams{metadata: "rnum", kqlFilter: "rnum > 0"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value2", "rnum": 1.0}})
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

	err := runPipeline(params, "test", &inBuffer, &outBuffer)
	assert.NoError(t, err)

	out := make([]steps.JSON, 0, len(in))
	decoder := json.NewDecoder(&outBuffer)
	for {
		j := make(steps.JSON)
		if err := decoder.Decode(&j); errors.Is(err, io.EOF) {
			break
		}
		out = append(out, j)
	}

	assert.Equal(t, expectedOut, out)
}
