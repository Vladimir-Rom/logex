package steps

import "sort"

type JSON map[string]any

func (j JSON) SortedByKey(yield func(key string, val any) bool) {
	type pair struct {
		k string
		v any
	}

	pairs := make([]pair, 0, len(j))
	for k, v := range j {
		pairs = append(pairs, pair{k, v})
	}

	sort.Slice(pairs, func(i, k int) bool {
		return pairs[i].k < pairs[k].k
	})

	for i := range pairs {
		if !yield(pairs[i].k, pairs[i].v) {
			break
		}
	}
}
