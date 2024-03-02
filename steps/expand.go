package steps

import (
	"encoding/json"

	"github.com/vladimir-rom/logex/pipeline"
)

func Expand(opts pipeline.PipelineOptions, properties []string) pipeline.Step[JSON, JSON] {
	if len(properties) == 0 {
		return Noop[JSON]()
	}

	return pipeline.NewStep(opts, func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
		if obj.Metadata.Removed {
			return yield(obj, nil)
		}

		for _, property := range properties {
			if value, ok := obj.Value[property]; ok {
				if valueStr, ok := value.(string); ok {
					var expanded map[string]any
					if json.Unmarshal([]byte(valueStr), &expanded) == nil {
						obj.Value[property] = expanded
					}
				}
			}
		}

		return yield(obj, nil)
	})
}
