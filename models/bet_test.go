package models

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestPK(t *testing.T) {
	var tuples = [][3]int{
		{Scissor, Scissor, 0},
		{Scissor, Rock, -1},
		{Scissor, Paper, 1},
		{Rock, Rock, 0},
		{Rock, Paper, -1},
		{Paper, Paper, 0},
	}

	for _, tuple := range tuples {
		assert.Equal(t, tuple[2], PK(tuple[0], tuple[1]))
		assert.Equal(t, -1*tuple[2], PK(tuple[1], tuple[0]))
	}
}

func TestBetFromUUID(t *testing.T) {
	id, _ := uuid.NewV4()
	times := map[int]int{}
	for idx := 1; idx < 100000; idx++ {
		g := NewGestureFromUUID(id, idx)
		g = gesture(g)
		assert.True(t, g >= 0 && g < base)
		times[g] = times[g] + 1
	}

	assert.Equal(t, times[Scissor], times[Rock])
	assert.Equal(t, times[Scissor], times[Paper])
}

func TestGestures(t *testing.T) {
	text := "012012012012"
	gestures := NewGesturesFromString(text)
	assert.Equal(t, gestures[:3], []int{Scissor, Rock, Paper})
}
