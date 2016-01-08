package fileIO

import (
	"encoding/binary"
)

func GetUint16LE(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b)
}

func PutUint16LE(b []byte, v uint16) {
	binary.LittleEndian.PutUint16(b, v)
}

func GetUint16BE(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func PutUint16BE(b []byte, v uint16) {
	binary.BigEndian.PutUint16(b, v)
}

func GetUint24LE(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

func PutUint24LE(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
}

func GetUint24BE(b []byte) uint32 {
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
}

func PutUint24BE(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func GetUint32LE(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func PutUint32LE(b []byte, v uint32) {
	binary.LittleEndian.PutUint32(b, v)
}

func GetUint32BE(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func PutUint32BE(b []byte, v uint32) {
	binary.BigEndian.PutUint32(b, v)
}

func GetUint40LE(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32
}

func PutUint40LE(b []byte, v uint64) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
}

func GetUint40BE(b []byte) uint64 {
	return uint64(b[4]) | uint64(b[3])<<8 | uint64(b[2])<<16 | uint64(b[1])<<24 | uint64(b[0])<<32
}

func PutUint40BE(b []byte, v uint64) {
	b[0] = byte(v >> 32)
	b[1] = byte(v >> 24)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 8)
	b[4] = byte(v)
}

func GetUint48LE(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32 | uint64(b[5])<<40
}

func PutUint48LE(b []byte, v uint64) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
}

func GetUint48BE(b []byte) uint64 {
	return uint64(b[5]) | uint64(b[4])<<8 | uint64(b[3])<<16 | uint64(b[2])<<24 | uint64(b[1])<<32 | uint64(b[0])<<40
}

func PutUint48BE(b []byte, v uint64) {
	b[0] = byte(v >> 40)
	b[1] = byte(v >> 32)
	b[2] = byte(v >> 24)
	b[3] = byte(v >> 16)
	b[4] = byte(v >> 8)
	b[5] = byte(v)
}

func GetUint56LE(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48
}

func PutUint56LE(b []byte, v uint64) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
}

func GetUint56BE(b []byte) uint64 {
	return uint64(b[6]) | uint64(b[5])<<8 | uint64(b[4])<<16 | uint64(b[3])<<24 | uint64(b[2])<<32 | uint64(b[1])<<40 | uint64(b[0])<<48
}

func PutUint56BE(b []byte, v uint64) {
	b[0] = byte(v >> 48)
	b[1] = byte(v >> 40)
	b[2] = byte(v >> 32)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 16)
	b[5] = byte(v >> 8)
	b[6] = byte(v)
}

func GetUint64LE(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

func PutUint64LE(b []byte, v uint64) {
	binary.LittleEndian.PutUint64(b, v)
}

func GetUint64BE(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func PutUint64BE(b []byte, v uint64) {
	binary.BigEndian.PutUint64(b, v)
}
