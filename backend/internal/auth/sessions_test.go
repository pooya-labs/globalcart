package auth

import "testing"

func TestSessionRevocationReasonAllowList(t *testing.T) {
	valid := []SessionRevocationReason{
		SessionRevocationPasswordChanged,
		SessionRevocationAccountDisabled,
		SessionRevocationPrivilegeChanged,
		SessionRevocationAdminRevocation,
		SessionRevocationSecurityRotation,
	}
	for _, reason := range valid {
		if !validSessionRevocationReason(reason) {
			t.Errorf("validSessionRevocationReason(%q) = false, want true", reason)
		}
	}

	invalid := []SessionRevocationReason{
		"",
		"logout", // Single-session logout uses RevokeSession.
		"email_changed",
		"password-change",
		"PASSWORD_CHANGED",
		"unknown",
	}
	for _, reason := range invalid {
		if validSessionRevocationReason(reason) {
			t.Errorf("validSessionRevocationReason(%q) = true, want false", reason)
		}
	}
}
