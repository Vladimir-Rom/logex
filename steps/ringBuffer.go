package steps

type ringBuffer[Item any] struct {
	items     []Item
	head      int
	overflown bool
}

func newRingBuffer[Item any](size int) *ringBuffer[Item] {
	return &ringBuffer[Item]{
		items: make([]Item, size),
	}
}

func (rb *ringBuffer[Item]) Add(item Item) {
	rb.items[rb.head] = item
	rb.head++
	if rb.head >= len(rb.items) {
		rb.head = 0
		rb.overflown = true
	}
}

func (rb *ringBuffer[Item]) All(yield func(item Item) bool) {
	if rb.overflown {
		for i := rb.head; i < len(rb.items); i++ {
			if !yield(rb.items[i]) {
				return
			}
		}
	}

	for i := range rb.head {
		if !yield(rb.items[i]) {
			return
		}
	}
}
