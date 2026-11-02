package auth

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"
)

func grantAdminRoleFixture(t *testing.T, userID string) {
	t.Helper()
	_, err := testAdminRolePool(t).Exec(testAdminRoleContext(t), `
		INSERT INTO user_roles(user_id, role_key)
		VALUES ($1, 'admin')
	`, userID)
	if err != nil {
		t.Fatalf("grant administrator fixture role: %v", err)
	}
}
