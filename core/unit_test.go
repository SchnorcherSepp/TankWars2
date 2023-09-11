package core

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewUnit(t *testing.T) {
	unit := NewUnit(RED, SOLDIER)
	assert.NotNil(t, unit)
	assert.Equal(t, uint8(RED), unit.Player)
	assert.Equal(t, uint8(SOLDIER), unit.Type)
	assert.NotEqual(t, 0, unit.ID)
	assert.Equal(t, 100, unit.Health)
	assert.Nil(t, unit.Activity)
	assert.Equal(t, 0, unit.View)
	assert.Equal(t, 0, unit.CloseView)
	assert.Equal(t, 0, unit.FireRange)
	assert.Equal(t, uint64(0), unit.Speed)
	assert.Equal(t, uint64(0), unit.FireSpeed)
	assert.False(t, unit.Hidden)
	assert.Equal(t, 0, unit.Armour)
	assert.False(t, unit.Demoralized)
	assert.Equal(t, float32(99), unit.Ammunition)
}
func TestCloneUnit(t *testing.T) {
	original := NewUnit(3, 4)
	original.Player = 1
	original.Type = SOLDIER
	original.ID = 123
	original.Health = 80
	original.Activity = &Activity{Name: "MOVE", From: [2]int{1, 2}, To: [2]int{3, 4}, Start: 5, End: 10}
	original.View = 5
	original.CloseView = 3
	original.FireRange = 4
	original.Speed = 2
	original.FireSpeed = 3
	original.Hidden = true
	original.Armour = 2
	original.Demoralized = true
	original.Ammunition = 50

	cloned := original.Clone()

	if cloned == original {
		t.Errorf("no cloning")
	}
	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Cloned unit does not match the original.\nOriginal: %+v\nCloned: %+v", original, cloned)
	}
}
