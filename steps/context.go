package steps

import "github.com/vladimir-rom/logex/pipeline"

func Context(opts pipeline.PipelineOptions, countBefore, countAfter int) pipeline.Step[JSON, JSON] {
	if !opts.ContextEnabled {
		return pipeline.NewStep(
			opts,
			func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
				if !obj.Metadata.Removed {
					return yield(obj, nil)
				}
				return true
			})
	}

	type rec struct {
		json pipeline.Item[JSON]
		err  error
	}
	buffer := newRingBuffer[rec](countBefore)
	remaindedToWrite := 0
	return pipeline.NewStep(
		opts,
		func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
			if !obj.Metadata.Removed {
				for item := range buffer.All {
					if !yield(item.json, item.err) {
						return false
					}
				}

				buffer = newRingBuffer[rec](countBefore)
				remaindedToWrite = countAfter
				return yield(obj, nil)
			}

			if remaindedToWrite > 0 {
				remaindedToWrite--
				return yield(obj, nil)
			}

			buffer.Add(rec{obj, nil})
			return true
		})
}
