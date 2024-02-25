package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMeta(t *testing.T) {
	checkMetaParsing(t, " rnum:recnum", &metaConfig{rnumName: "recnum"}, false)
	checkMetaParsing(t, " rnum  file", &metaConfig{rnumName: "rnum", file: "file"}, false)
	checkMetaParsing(t, "rnum file:f1 ", &metaConfig{rnumName: "rnum", file: "f1"}, false)
	checkMetaParsing(t, "rnum file ", &metaConfig{rnumName: "rnum", file: "file"}, false)
	checkMetaParsing(t, "file", &metaConfig{rnumName: "", file: "file"}, false)
	checkMetaParsing(t, "file:f2", &metaConfig{rnumName: "", file: "f2"}, false)
	checkMetaParsing(t, "file:f2 foo:bar", nil, true)
}

func checkMetaParsing(t *testing.T, metaCfg string, expected *metaConfig, expectError bool) {
	t.Helper()
	cfg, err := parseMetaConfig(metaCfg)
	assert.Truef(t,
		(err == nil && !expectError) || (err != nil && expectError),
		"error expected: %t, error: %#v", expectError, err)
	assert.Equal(t, expected, cfg)
}
