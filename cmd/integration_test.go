package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vladimir-rom/logex/steps"
)

func TestNoModification(t *testing.T) {
	in := []steps.JSON{{"field1": "value1", "field2": "value2"}}
	testCmd(t, nil, in, in)
}

func TestSelect(t *testing.T) {
	testCmd(t,
		[]string{"--select", "field1"},
		[]steps.JSON{{"field1": "value1", "field2": "value2"}},
		[]steps.JSON{{"field1": "value1"}})
}

func TestFirst(t *testing.T) {
	testCmd(t,
		[]string{"--first", "1"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value1"}})
}

func TestLast(t *testing.T) {
	testCmd(t,
		[]string{"--last", "1"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value2"}})
}

func TestKQL(t *testing.T) {
	testCmd(t,
		[]string{"-f", "field:value2"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value2"}})
}

func TestHide(t *testing.T) {
	testCmd(t,
		[]string{"--hide", "field2"},
		[]steps.JSON{{"field1": "value1", "field2": "value2"}},
		[]steps.JSON{{"field1": "value1"}})
}

func TestExpand(t *testing.T) {
	testCmd(t,
		[]string{"-f", "inner.foo:bar", "--expand", "inner"},
		[]steps.JSON{
			{"field": "value1"},
			{"field": "value2", "inner": `{"foo":"bar"}`},
			{"field": "value3"}},
		[]steps.JSON{{"field": "value2", "inner": map[string]any{"foo": "bar"}}})

	testCmd(t,
		[]string{"-f", "inner:{prop:val2}", "--expand", "inner"},
		[]steps.JSON{
			{"field": "value1"},
			{"field": "value2", "inner": `[{"prop":"val2"}]`},
			{"field": "value3", "inner": `[{"prop":"val3"}]`},
		},
		[]steps.JSON{
			{"field": "value2", "inner": []any{map[string]any{"prop": "val2"}}},
		})

}

func TestJq(t *testing.T) {
	testCmd(t,
		[]string{"--jq", ".field==\"value2\""},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value2"}})

	testCmd(t,
		[]string{"--jq", `. + {"foo":"bar"}`},
		[]steps.JSON{{"field": "value1"}},
		[]steps.JSON{{"field": "value1", "foo": "bar"}})
}

func TestInclude(t *testing.T) {
	testCmd(t,
		[]string{"--include", "Value2"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value2"}})
}

func TestIncludeRegexp(t *testing.T) {
	testCmd(t,
		[]string{"--include-regexp", "value[1-2]"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}})
}

func TestExcludeRegexp(t *testing.T) {
	testCmd(t,
		[]string{"--exclude-regexp", "value[1-2]"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}},
		[]steps.JSON{{"field": "value3"}})
}

func TestContext(t *testing.T) {
	input := make([]steps.JSON, 0)
	for i := range 5 {
		input = append(input, steps.JSON{"field": fmt.Sprintf("value%d", i+1)})
	}

	testCmd(t,
		[]string{"--include", "Value3", "--context", "1"},
		input,
		[]steps.JSON{{"field": "value2"}, {"field": "value3"}, {"field": "value4"}})

	testCmd(t,
		[]string{"--include", "Value1", "--context", "1"},
		input,
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}})

	testCmd(t,
		[]string{"--include", "Value5", "--context", "1"},
		input,
		[]steps.JSON{{"field": "value4"}, {"field": "value5"}})

	testCmd(t,
		[]string{"--include", "Value3", "--context", "5"},
		input,
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}, {"field": "value3"}, {"field": "value4"}, {"field": "value5"}})

	testCmd(t,
		[]string{"--include", "Value1111", "--context", "1"},
		input,
		[]steps.JSON{})
}

func TestMetadata(t *testing.T) {
	testCmd(t,
		[]string{"-m", "rnum"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value1", "rnum": 0.0}, {"field": "value2", "rnum": 1.0}})

	testCmd(t,
		[]string{"-m", "rnum:r file:f"},
		[]steps.JSON{{"field": "value1"}},
		[]steps.JSON{{"field": "value1", "r": 0.0, "f": "stdin"}})

	testCmd(t,
		[]string{"-m", "rnum", "-f", "rnum > 0"},
		[]steps.JSON{{"field": "value1"}, {"field": "value2"}},
		[]steps.JSON{{"field": "value2", "rnum": 1.0}})
}

func testCmd(t *testing.T, args []string, in []steps.JSON, expectedOut []steps.JSON) {
	t.Helper()
	cmd := createRootCmd()
	args = append(args, "--show-errors", "-", "--format", "json")
	if !slices.Contains(args, "--metadata") && !slices.Contains(args, "-m") {
		args = append(args, "--metadata", "")
	}
	cmd.SetArgs(args)

	inBuffer := marshalJson(t, in)
	outBuffer := bytes.Buffer{}
	cmd.SetIn(inBuffer)
	cmd.SetOut(&outBuffer)
	err := cmd.Execute()
	t.Log(outBuffer.String())
	require.NoError(t, err)
	checkOutput(t, expectedOut, &outBuffer)
}

func checkOutput(t *testing.T, expectedOut []steps.JSON, output io.Reader) {
	t.Helper()
	out := make([]steps.JSON, 0)

	decoder := json.NewDecoder(output)
	for {
		j := make(steps.JSON)
		err := decoder.Decode(&j)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		out = append(out, j)
	}

	assert.Equal(t, expectedOut, out)
}

func marshalJson(t *testing.T, in []steps.JSON) io.Reader {
	t.Helper()
	inBuffer := bytes.Buffer{}
	for _, j := range in {
		b, err := json.Marshal(j)
		assert.NoError(t, err)
		inBuffer.Write(b)
		inBuffer.WriteByte('\n')
	}
	return &inBuffer
}
