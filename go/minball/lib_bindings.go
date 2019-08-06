package minball

// #cgo CFLAGS: -DNDEBUG -Wall -O3 -I${SRCDIR}/../../minball
// #cgo LDFLAGS: -L${SRCDIR}/../../minball -lCGAL -lgmp -lminball
// #include "minball_C_interface.h"
import "C"

type Ball2D struct {
	Center [2]float64
	Radius float64
}

func ComputeMinball2D(balls []Ball2D) Ball2D {
	cBalls := []C.struct_Ball2D{}
	for _, ball := range balls {
		cCenter := [2]C.double{
			C.double(ball.Center[0]), C.double(ball.Center[1]),
		}
		cBalls = append(cBalls, C.struct_Ball2D{
			cCenter, C.double(ball.Radius),
		})
	}
	cBallsPtr := (*C.struct_Ball2D)(&cBalls[0])
	cResult := C.compute_minball2D(cBallsPtr, C.int(len(balls)))

	goCenter := [2]float64{
		float64(cResult.center[0]), float64(cResult.center[1]),
	}
	return Ball2D{
		Center: goCenter,
		Radius: float64(cResult.radius),
	}
}
