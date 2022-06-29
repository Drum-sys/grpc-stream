package client

type Weighted struct {
	Server          string
	Weight          int
	CurrentWeight   int
	EffectiveWeight int
}

func nextWeighted(servers []*Weighted) (best *Weighted) {
	total := 0

	for i := 0; i < len(servers); i++ {
		w := servers[i]

		if w == nil {
			continue
		}

		w.CurrentWeight += w.EffectiveWeight
		total += w.EffectiveWeight

		if w.EffectiveWeight < w.Weight {
			w.EffectiveWeight++
		}

		if best == nil || best.CurrentWeight < w.CurrentWeight {
			best = w
		}
	}

	if best == nil {
		return nil
	}
	best.CurrentWeight -= total
	return best
}
