package services

import (
	"gibraltar/config"
	"math"
)

func onSuccess(old float64) (new float64) {
	new = old + (config.Gain * math.Pow((1-old/config.MAX), config.P))
	if new > 100 {
		new = 100
	}
	return new
}

func onFailure(old float64) (new float64) {
	new = old - (config.Decay*math.Pow(old/config.MAX, config.Q)*old + config.MinDrop)
	if new < 0 {
		new = 0
	}
	return new
}
