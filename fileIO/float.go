package fileIO


import (
	"math"
)

func GetFloat32BE(b []byte) float32 {
	return math.Float32frombits(GetUint32BE(b))
}

func PutFloat32BE(b []byte, v float32) {
	PutUint32BE(b, math.Float32bits(v))
}

func GetFloat32LE(b []byte) float32 {
	return math.Float32frombits(GetUint32LE(b))
}

func PutFloat32LE(b []byte, v float32) {
	PutUint32LE(b, math.Float32bits(v))
}

func GetFloat64BE(b []byte) float64 {
	return math.Float64frombits(GetUint64BE(b))
}

func PutFloat64BE(b []byte, v float64) {
	PutUint64BE(b, math.Float64bits(v))
}

func GetFloat64LE(b []byte) float64 {
	return math.Float64frombits(GetUint64LE(b))
}

func PutFloat64LE(b []byte, v float64) {
	PutUint64LE(b, math.Float64bits(v))
}
