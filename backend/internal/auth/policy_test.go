package auth

import (
	"errors"
	"testing"
)

func TestAuthorizeDenyByDefaultPolicyMatrix(t *testing.T) {
	customer := SessionPrincipal{
		ID:     "session-customer",
		UserID: "customer-user",
		Roles:  []string{"customer"},
	}
	seller := SessionPrincipal{
		ID:              "session-seller",
		UserID:          "seller-user",
		Roles:           []string{"customer", "seller"},
		SellerVendorIDs: []int64{10, 20},
	}
	sellerWithoutScope := SessionPrincipal{
		ID:     "session-seller-no-scope",
		UserID: "seller-user-no-scope",
		Roles:  []string{"customer", "seller"},
	}
	admin := SessionPrincipal{
		ID:     "session-admin",
		UserID: "admin-user",
		Roles:  []string{"customer", "admin"},
	}

	tests := []struct {
		name      string
		principal SessionPrincipal
		action    AuthorizationAction
		resource  AuthorizationResource
		wantErr   error
	}{
		{
			name:      "customer reads own resource",
			principal: customer,
			action:    AuthorizationCustomerReadOwnResource,
			resource:  AuthorizationResource{OwnerUserID: customer.UserID},
		},
		{
			name:      "customer updates own resource",
			principal: customer,
			action:    AuthorizationCustomerWriteOwnResource,
			resource:  AuthorizationResource{OwnerUserID: customer.UserID},
		},
		{
			name:      "customer cannot read another user's resource",
			principal: customer,
			action:    AuthorizationCustomerReadOwnResource,
			resource:  AuthorizationResource{OwnerUserID: "another-user"},
			wantErr:   ErrForbidden,
		},
		{
			name:      "customer cannot omit resource owner",
			principal: customer,
			action:    AuthorizationCustomerReadOwnResource,
			wantErr:   ErrForbidden,
		},
		{
			name:      "seller reads assigned vendor",
			principal: seller,
			action:    AuthorizationSellerReadVendor,
			resource:  AuthorizationResource{VendorID: 10},
		},
		{
			name:      "seller writes assigned vendor",
			principal: seller,
			action:    AuthorizationSellerWriteVendor,
			resource:  AuthorizationResource{VendorID: 20},
		},
		{
			name:      "seller cannot access unassigned vendor",
			principal: seller,
			action:    AuthorizationSellerReadVendor,
			resource:  AuthorizationResource{VendorID: 30},
			wantErr:   ErrForbidden,
		},
		{
			name:      "seller cannot omit vendor scope",
			principal: seller,
			action:    AuthorizationSellerWriteVendor,
			wantErr:   ErrForbidden,
		},
		{
			name:      "seller role without active vendor scope is denied",
			principal: sellerWithoutScope,
			action:    AuthorizationSellerReadVendor,
			resource:  AuthorizationResource{VendorID: 10},
			wantErr:   ErrForbidden,
		},
		{
			name:      "customer cannot manage users",
			principal: customer,
			action:    AuthorizationAdminManageUsers,
			wantErr:   ErrForbidden,
		},
		{
			name:      "customer cannot read admin overview",
			principal: customer,
			action:    AuthorizationAdminReadOverview,
			wantErr:   ErrForbidden,
		},
		{
			name:      "admin can read admin overview",
			principal: admin,
			action:    AuthorizationAdminReadOverview,
		},
		{
			name:      "admin can manage users",
			principal: admin,
			action:    AuthorizationAdminManageUsers,
		},
		{
			name:      "admin can review seller membership",
			principal: admin,
			action:    AuthorizationAdminReviewMembership,
		},
		{
			name:      "admin can manage roles",
			principal: admin,
			action:    AuthorizationAdminManageRoles,
		},
		{
			name:      "admin still needs explicit seller vendor scope",
			principal: admin,
			action:    AuthorizationSellerWriteVendor,
			resource:  AuthorizationResource{VendorID: 10},
			wantErr:   ErrForbidden,
		},
		{
			name:      "unknown action is denied",
			principal: admin,
			action:    AuthorizationAction("admin.everything"),
			wantErr:   ErrForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Authorize(tc.principal, tc.action, tc.resource)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Authorize() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestAuthorizeRequiresResolvedSessionIdentity(t *testing.T) {
	principals := []SessionPrincipal{
		{UserID: "user-without-session", Roles: []string{"customer"}},
		{ID: "session-without-user", Roles: []string{"customer"}},
		{},
	}
	for i, principal := range principals {
		if err := Authorize(
			principal,
			AuthorizationCustomerReadOwnResource,
			AuthorizationResource{OwnerUserID: principal.UserID},
		); !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("principal %d authorization error = %v, want ErrUnauthenticated", i, err)
		}
	}
}

func TestAuthorizeDoesNotTrustRoleOrVendorFromResource(t *testing.T) {
	customer := SessionPrincipal{
		ID:     "session-customer",
		UserID: "customer-user",
		Roles:  []string{"customer"},
	}
	// Resource claims cannot add a role or scope that is absent from the
	// database-resolved principal.
	err := Authorize(
		customer,
		AuthorizationSellerWriteVendor,
		AuthorizationResource{
			OwnerUserID: customer.UserID,
			VendorID:    10,
		},
	)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("customer-supplied vendor resource authorization error = %v, want ErrForbidden", err)
	}
}
