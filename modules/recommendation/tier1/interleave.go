package tier1

import "container/heap"

// candidateEntry holds a candidate with its distance for heap ordering.
type candidateEntry struct {
	userIDBytes [16]byte
	distKm      float64
	source      string
}

// candidateHeap is a min-heap of candidateEntry ordered by distKm.
type candidateHeap []candidateEntry

func (h candidateHeap) Len() int            { return len(h) }
func (h candidateHeap) Less(i, j int) bool { return h[i].distKm < h[j].distKm }
func (h candidateHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *candidateHeap) Push(x any) {
	*h = append(*h, x.(candidateEntry))
}

func (h *candidateHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// mergeByDistance takes multiple slices of (userIDBytes, distKm, source)
// and returns them merged in ascending distance order, deduplicated.
func mergeByDistance(batches [][]candidateEntry, limit int) []candidateEntry {
	h := &candidateHeap{}
	heap.Init(h)

	for _, batch := range batches {
		for _, e := range batch {
			heap.Push(h, e)
		}
	}

	seen := make(map[[16]byte]bool)
	result := make([]candidateEntry, 0, limit)

	for h.Len() > 0 && len(result) < limit {
		e := heap.Pop(h).(candidateEntry)
		if !seen[e.userIDBytes] {
			seen[e.userIDBytes] = true
			result = append(result, e)
		}
	}

	return result
}
