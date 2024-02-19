package pipeline

type Metadata struct {
	removed bool
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

func NewItem1[Value any](value Value) Item[Value] {
	return Item[Value]{
		Value: value,
	}
}

type Yield[T any] func(Item[T], error) bool
type Seq[T any] func(Yield[T])
type Step[In, Out any] func(Seq[In]) Seq[Out]

func NewStep[In, Out any](sink func(item Item[In], yield Yield[Out]) bool) Step[In, Out] {
	return NewStepWithFin(sink, func(internalYield Yield[Out]) {})
}

func NewStepWithFin[In, Out any](
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
					if !sink(v, internalYield) {
						return
					}
				}
			}
			finalize(internalYield)
		}
	}
}