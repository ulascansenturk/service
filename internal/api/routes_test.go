//go:build tests_unit

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "ulascansenturk/service/internal/api/v1"
)

func TestNewRoutes(t *testing.T) {
	t.Run("NewRoutes", func(t *testing.T) {
		got := NewRoutes(&v1.API{})

		assert.NotNil(t, got)
		assert.NotNil(t, got.v1)
	})
}
