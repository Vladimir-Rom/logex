package steps

import (
	"fmt"
	"slices"

	"github.com/vladimir-rom/logex/pipeline"
)

func Merge(
	opts pipeline.PipelineOptions,
	properties []string,
	in []pipeline.Seq[JSON]) pipeline.Seq[JSON] {
	if len(in) == 1 {
		return in[0]
	}

	return func(yield pipeline.Yield[JSON]) {
		out := make([]chan jsonWithErr, len(in))
		cancels := make([]chan struct{}, len(in))
		defer func() {
			for _, c := range cancels {
				if c != nil {
					c <- struct{}{}
				}
			}
		}()
		for i, input := range in {
			out[i] = make(chan jsonWithErr, 1000)
			cancels[i] = make(chan struct{}, 1)
			go func() {
				defer close(out[i])
				for item, err := range input {
					select {
					case <-cancels[i]:
						return
					case out[i] <- jsonWithErr{
						item:          item,
						err:           err,
						propertyNames: properties,
					}:
					}
				}
			}()
		}

		iterators := make([]*jsonChanIterator, len(in))
		for i := range iterators {
			iterators[i] = newJsonChanIterator(out[i])
		}

		for item := range merge(iterators, func(i, j jsonWithErr) bool {
			return less(i.GetValue(), j.GetValue())
		}) {
			if !yield(item.item, item.err) {
				break
			}
		}
	}
}

type jsonWithErr struct {
	item                     pipeline.Item[JSON]
	err                      error
	propertyValue            any
	propertyValueInitialized bool
	propertyNames            []string
}

func (j *jsonWithErr) GetValue() any {
	if j.propertyValueInitialized {
		return j.propertyValue
	}

	for _, p := range j.propertyNames {
		if v, ok := j.item.Value[p]; ok {
			j.propertyValue = v
			break
		}
	}

	j.propertyValueInitialized = true
	return j.propertyValue
}

type jsonChanIterator struct {
	value jsonWithErr
	ch    <-chan jsonWithErr
}

func newJsonChanIterator(ch <-chan jsonWithErr) *jsonChanIterator {
	return &jsonChanIterator{
		ch: ch,
	}
}

func (i *jsonChanIterator) GetValue() jsonWithErr {
	return i.value
}

func (i *jsonChanIterator) Next() bool {
	if r, ok := <-i.ch; ok {
		if r.err != nil {
			return i.Next()
		}

		i.value = r
		return true
	}

	return false
}

func merge(iterators []*jsonChanIterator, less func(i, k jsonWithErr) bool) func(yield func(item jsonWithErr) bool) {
	return func(yield func(item jsonWithErr) bool) {
		its := make([]*jsonChanIterator, 0, len(iterators))
		for _, it := range iterators {
			if it.Next() {
				its = append(its, it)
			}
		}

		for len(its) > 1 {
			minIndex := 0
			for i := 1; i < len(its); i++ {
				if !less(its[minIndex].GetValue(), its[i].GetValue()) {
					minIndex = i
				}
			}

			if !yield(its[minIndex].GetValue()) {
				return
			}

			if !its[minIndex].Next() {
				its = slices.Delete(its, minIndex, minIndex+1)
			}
		}

		if len(its) == 1 {
			for {
				if !yield(its[0].GetValue()) {
					return
				}
				if !its[0].Next() {
					break
				}
			}
		}
	}
}

func less(i, j any) bool {
	switch v := i.(type) {
	case string:
		return lessString(v, j)
	case bool:
		return lessBool(v, j)
	case float64:
		return lessFloat64(v, j)
	default:
		return lessString(fmt.Sprint(i), fmt.Sprint(j))
	}
}

func lessString(i string, j any) bool {
	switch v := j.(type) {
	case string:
		return i < v
	default:
		return i < fmt.Sprint(j)
	}
}

func lessBool(i bool, j any) bool {
	switch v := j.(type) {
	case bool:
		return !i && v
	default:
		return !i && (len(fmt.Sprint(j)) > 0)
	}
}

func lessFloat64(i float64, j any) bool {
	switch v := j.(type) {
	case float64:
		return i < v
	default:
		return fmt.Sprint(i) < fmt.Sprint(j)
	}
}
