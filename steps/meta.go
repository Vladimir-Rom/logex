package steps

import (
	"fmt"
	"strings"

	"github.com/vladimir-rom/logex/pipeline"
)

type metaConfig struct {
	rnumName string
	file     string
}

func AddMeta(opts pipeline.PipelineOptions, metaCfg string) (pipeline.Step[JSON, JSON], error) {
	mc, err := parseMetaConfig(metaCfg)
	if err != nil {
		return nil, err
	}

	if len(mc.rnumName) == 0 && len(mc.file) == 0 {
		return Noop[JSON](), nil
	}

	return pipeline.NewStep[JSON, JSON](opts, func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
		if len(mc.rnumName) != 0 {
			obj.Value[mc.rnumName] = obj.Metadata.RecNum
		}
		if len(mc.file) != 0 {
			obj.Value[mc.file] = obj.Metadata.FileName
		}

		return yield(obj, nil)
	}), nil
}

func parseMetaConfig(metaCfg string) (*metaConfig, error) {
	result := &metaConfig{}
	metaCfg = strings.TrimSpace(metaCfg)
	for _, part := range strings.Split(metaCfg, " ") {
		if len(part) == 0 {
			continue
		}
		nameVal := strings.Split(part, ":")
		getVal := func(dflt string) string {
			if len(nameVal) < 2 {
				return dflt
			}
			return nameVal[1]
		}

		switch nameVal[0] {
		case "rnum":
			result.rnumName = getVal("rnum")
		case "file":
			result.file = getVal("file")
		default:
			return nil, fmt.Errorf("unknown metadata field: %s", nameVal[0])
		}
	}

	return result, nil
}
