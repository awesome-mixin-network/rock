package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash/fnv"

	uuid "github.com/satori/go.uuid"
)

const (
	Scissor int = iota
	Rock
	Paper
	base
)

func gesture(v int) int {
	v = v % base
	if v < 0 {
		v += base
	}

	return v
}

func PK(a, b int) int {
	a, b = gesture(a), gesture(b)

	x := a - b
	if x == -2 || x == 2 {
		x = x / -2
	}

	return x
}

func NewGestureFromUUID(id uuid.UUID, idx int) int {
	secret := []byte(fmt.Sprintf("%d", idx))
	h := hmac.New(sha256.New, secret)
	h.Write(id.Bytes())
	h32 := fnv.New32a()
	h32.Write(h.Sum(nil))
	v := h32.Sum32()
	return int(v)
}

func NewGesturesFromString(text string) []int {
	gestures := make([]int, len(text))
	for idx, r := range text {
		gestures[idx] = gesture(int(r - '0'))
	}

	return gestures
}

func GestureString(gestures []int) string {
	b := make([]byte, len(gestures))
	for idx, g := range gestures {
		b[idx] = '0' + byte(gesture(g))
	}

	return string(b)
}
