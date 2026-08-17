package hll

import (
	"hash/maphash"
	"math"
	"math/bits"
)

// precision p is the number of register-index bits; m = 2^p registers.
// p = 14 gives m = 16384 registers (~1.6% standard error on the raw estimate).
const (
	precision = 14
	m         = 1 << precision
)

// alpha is the bias-correction constant for the raw HLL estimate
// (Flajolet et al. 2007).
var alpha = 0.7213 / (1 + 1.079/float64(m))

// HLL is a HyperLogLog cardinality estimator.
type HLL struct {
	seed      maphash.Seed
	registers [m]uint8
}

// New returns an empty estimator. Each estimator gets its own hash seed;
// the seed only relabels registers, so estimates are unbiased and stable
// across runs.
func New() *HLL {
	return &HLL{seed: maphash.MakeSeed()}
}

// Add records s in the estimator.
func (h *HLL) Add(s string) {
	x := hash64(h.seed, s)
	// The top p bits of the hash select the register.
	idx := x >> (64 - precision)
	// On the remaining (64 - p) bits, rho is the position of the leftmost
	// 1-bit counted from the most-significant end, plus one.
	w := x << precision
	var rho uint8
	if w == 0 {
		// All remaining bits were zero.
		rho = uint8(64-precision) + 1
	} else {
		rho = uint8(bits.LeadingZeros64(w) + 1)
	}
	if rho > h.registers[idx] {
		h.registers[idx] = rho
	}
}

// Estimate returns the approximate number of distinct elements added.
func (h *HLL) Estimate() float64 {
	var sum float64
	var zeros int
	for _, r := range h.registers {
		sum += math.Pow(2, -float64(r))
		if r == 0 {
			zeros++
		}
	}

	// Raw HLL estimate.
	E := alpha * float64(m) * float64(m) / sum

	// Small-range correction (linear counting). When the true cardinality
	// is small relative to m, most registers stay zero and the raw estimate
	// is badly biased (it collapses toward alpha*m regardless of n). In that
	// regime the count of zero registers is a far better signal:
	//   E = m * ln(m / zeros)
	// This is the "小范围线性计数修正" the structure was missing.
	if E <= 2.5*float64(m) && zeros > 0 {
		E = float64(m) * math.Log(float64(m)/float64(zeros))
	}

	// Large-range correction is unnecessary with a 64-bit hash: the
	// collision threshold (2^32) is never approached in practice.
	return E
}

// hash64 maps s to a uniformly distributed 64-bit value. maphash is the
// standard-library hash designed for hashing keys into buckets; unlike
// FNV it avalanches well even for very short inputs, which is essential
// for the small-cardinality regime tested here.
func hash64(seed maphash.Seed, s string) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	h.WriteString(s)
	return h.Sum64()
}
