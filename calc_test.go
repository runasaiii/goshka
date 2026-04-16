package goshka

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	assert.Equal(t, 5, Add(2, 3))
}

func TestAddTableDriven(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "both positive", a: 2, b: 3, want: 5},
		{name: "positive plus zero", a: 5, b: 0, want: 5},
		{name: "negative plus positive", a: -1, b: 4, want: 3},
		{name: "both negative", a: -2, b: -3, want: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Add(tt.a, tt.b))
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "both positive", a: 10, b: 2, want: 5},
		{name: "negative divided by positive", a: -10, b: 2, want: -5},
		{name: "integer truncation", a: 7, b: 2, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Divide(tt.a, tt.b)

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}

	t.Run("division by zero", func(t *testing.T) {
		result, err := Divide(10, 0)

		require.Error(t, err)
		assert.Equal(t, 0, result)
		assert.EqualError(t, err, "division by zero")
	})
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "both positive numbers", a: 8, b: 3, want: 5},
		{name: "positive minus zero", a: 5, b: 0, want: 5},
		{name: "negative minus positive", a: -4, b: 6, want: -10},
		{name: "both negative", a: -4, b: -6, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Subtract(tt.a, tt.b))
		})
	}
}
