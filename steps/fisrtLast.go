package steps

import (
	"github.com/vladimir-rom/logex/pipeline"
)

func Last(count int) pipeline.Step[JSON, JSON] {
	if count <= 0 {
		return Noop[JSON]()
	}

	type rec struct {
		json pipeline.Item[JSON]
		err  error
	}
	buffer := newRingBuffer[rec](count)

	return pipeline.NewStepWithFin(
		func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
			buffer.Add(rec{obj, nil})
			return true
		},
		func(yield pipeline.Yield[JSON]) {
			for r := range buffer.All {
				if !yield(r.json, r.err) {
					return
				}
			}
		},
	)
}

func First(count int) pipeline.Step[JSON, JSON] {
	if count <= 0 {
		return Noop[JSON]()
	}

	returned := 0
	return pipeline.NewStep[JSON, JSON](func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
		returned++
		if returned > count {
			return false
		}

		return yield(obj, nil)
	})
}
