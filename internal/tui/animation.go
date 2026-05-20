package tui

import "math"

const smoothFactor = 4.0

type AnimFloat struct {
	Current float64
	Target  float64
}

func (a *AnimFloat) Update(dt float64) {
	diff := a.Target - a.Current
	if math.Abs(diff) < 0.001 {
		a.Current = a.Target
		return
	}
	factor := 1.0 - math.Exp(-smoothFactor*dt)
	a.Current += diff * factor
}

func EaseOutCubic(t float64) float64 {
	if t >= 1 {
		return 1
	}
	return 1 - math.Pow(1-t, 3)
}
