package rng

import (
	"math/rand"
	"sync"
)

var (
	randSrc = rand.New(rand.NewSource(rand.Int63()))
	lock    = sync.Mutex{}
)

// Get a normally distributed integer in the closed interval [min, max]
func ScaledNormalDistribution(min, max int32) int32 {
	lock.Lock()
	defer lock.Unlock()
	return ScaledNormalDistributionFromRand(randSrc, min, max)
}

// Get a normally distributed integer in the closed interval [min, max]
// using an underlying rng.
func ScaledNormalDistributionFromRand(r *rand.Rand, min, max int32) int32 {
	// We use 3 standard deviations and exclude everything else
	// mean is the midpoint
	if max < min {
		min, max = max, min
	}
	stdDev := float64((max - min) / 6)
	mean := float64(min/2 + max/2 + (min & max & 1))

	for {
		sample := r.NormFloat64()*stdDev + mean
		sampleInt := int32(sample)
		if sampleInt >= min && sampleInt <= max {
			return sampleInt
		}
	}
}
