package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDb(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "pass")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "tili")

	db := NewDb()

	if assert.NotNil(t, db) {
		assert.NotNil(t, db.DB)
		_ = db.DB.Close()
	}
}
