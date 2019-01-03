package server

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/soooooooon/rock/models"
	"github.com/stretchr/testify/assert"
)

func TestMarshalArena(t *testing.T) {
	a := models.Arena{
		ID: 120,
	}

	view := arenaView{
		Arena: &a,
		ID:    models.HashIdEncode(a.ID),
	}

	data, err := jsoniter.MarshalToString(view)
	assert.Nil(t, err)
	assert.Empty(t, data)
}
