package hll

type HLL struct {
	seen map[string]struct{}
}

func New() *HLL { return &HLL{seen: map[string]struct{}{}} }

func (h *HLL) Add(s string) { h.seen[s] = struct{}{} }

func (h *HLL) Estimate() float64 {
	// BUG: underestimate by half (no linear counting / wrong scale)
	return float64(len(h.seen)) / 2
}
