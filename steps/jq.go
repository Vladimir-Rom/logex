package steps

import (
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/vladimir-rom/logex/pipeline"
)

func FilterByJq(opts pipeline.PipelineOptions, filter string) (pipeline.Step[JSON, JSON], error) {
	if len(filter) == 0 {
		return Noop[JSON](), nil
	}

	expression, err := gojq.Parse(filter)
	if err != nil {
		return nil, fmt.Errorf("filter parsing error: %w", err)
	}

	return pipeline.NewStep[JSON, JSON](opts, func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
		if obj.Metadata.Removed {
			return yield(obj, nil)
		}

		iter := expression.Run(map[string]any(obj.Value))
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			switch item := v.(type) {
			case error:
				if !yield(obj.WithValue(nil), item) {
					return false
				}
			case bool:
				obj.Metadata.Removed = !item
				if !yield(obj, nil) {
					return false
				}
			case map[string]any:
				if !yield(obj.WithValue(item), nil) {
					return false
				}
			default:
				if !yield(obj.WithValue(JSON{"item": item}), nil) {
					return false
				}
			}
		}

		return true
	}), nil
}
