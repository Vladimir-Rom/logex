package pipeline

import "fmt"

type PipelineOptions struct {
	ContextEnabled bool
}

type Metadata struct {
	Removed  bool
	RecNum   int
	FileName string
}

type Item[Value any] struct {
	Value    Value
	Metadata Metadata
}

func (i Item[Value]) WithValue(value Value) Item[Value] {
	return Item[Value]{
		Value:    value,
		Metadata: i.Metadata,
	}
}

func ToItem[Value1, Value2 any](item Item[Value1], value Value2) Item[Value2] {
	return Item[Value2]{
		Value:    value,
		Metadata: item.Metadata,
	}
}

type Yield[T any] func(Item[T], error) bool
type Seq[T any] func(Yield[T])
type Step[In, Out any] func(Seq[In]) Seq[Out]

func NewStep[In, Out any](opts PipelineOptions, sink func(item Item[In], yield Yield[Out]) bool) Step[In, Out] {
	return NewStepWithFin(opts, sink, func(internalYield Yield[Out]) {})
}

func NewStepWithFin[In, Out any](
	opts PipelineOptions,
	sink func(item Item[In], yield Yield[Out]) bool,
	finalize func(yield Yield[Out])) Step[In, Out] {
	return func(in Seq[In]) Seq[Out] {
		return func(internalYield Yield[Out]) {
			for v, err := range in {
				if err != nil {
					var def Item[Out]
					if !internalYield(def, err) {
						return
					}
				} else {
					if opts.ContextEnabled || !v.Metadata.Removed {
						if !sink(v, internalYield) {
							return
						}
					}
				}
			}
			finalize(internalYield)
		}
	}
}

func Combine[Item any](steps ...Step[Item, Item]) Step[Item, Item] {
	if len(steps) < 2 {
		panic(fmt.Sprintf("cannt combine %d steps", len(steps)))
	}
	result := func(in Seq[Item]) Seq[Item] {
		return steps[1](steps[0](in))
	}

	for i := 2; i < len(steps); i++ {
		r := result
		result = func(in Seq[Item]) Seq[Item] {
			return steps[i](r(in))
		}
	}
	return result
}
