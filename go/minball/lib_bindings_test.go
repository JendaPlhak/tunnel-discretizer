package minball

import (
	"testing"
)

func TestComputeMinball2D(t *testing.T) {
	cmpBalls := func(t *testing.T, expected, actuall Ball2D) {
		if expected.Center != actuall.Center {
			t.Errorf("Centers differ: %f vs %f (expected vs actuall)",
				expected.Center, actuall.Center)
		}
		if expected.Radius != actuall.Radius {
			t.Errorf("Radiuses differ: %f vs %f (expected vs actuall)",
				expected.Radius, actuall.Radius)
		}
	}
	t.Run("Single circle", func(t *testing.T) {
		ball := Ball2D{
			Center: [2]float64{0, 0},
			Radius: 1,
		}
		minball := ComputeMinball2D([]Ball2D{ball})
		cmpBalls(t, ball, minball)
	})
	t.Run("Two symetric circles", func(t *testing.T) {
		balls := []Ball2D{
			Ball2D{
				Center: [2]float64{-1, 0},
				Radius: 1,
			},
			Ball2D{
				Center: [2]float64{1, 0},
				Radius: 1,
			},
		}
		expBall := Ball2D{
			Center: [2]float64{0, 0},
			Radius: 2,
		}
		minball := ComputeMinball2D(balls)
		cmpBalls(t, expBall, minball)
	})
	t.Run("Two asymetric circles", func(t *testing.T) {
		balls := []Ball2D{
			Ball2D{
				Center: [2]float64{-1, 0},
				Radius: 2,
			},
			Ball2D{
				Center: [2]float64{1, 0},
				Radius: 1,
			},
		}
		expBall := Ball2D{
			Center: [2]float64{-0.5, 0},
			Radius: 2.5,
		}
		minball := ComputeMinball2D(balls)
		cmpBalls(t, expBall, minball)
	})
}
