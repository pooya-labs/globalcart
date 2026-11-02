package auth

import (
	"errors"
	"slices"
	"strings"
)

type AuthorizationAction string

const (
	AuthorizationCustomerReadOwnResource  AuthorizationAction = "customer.resource.read"
	AuthorizationCustomerWriteOwnResource AuthorizationAction = "customer.resource.write"
	AuthorizationSellerReadVendor         AuthorizationAction = "seller.vendor.read"
	AuthorizationSellerWriteVendor        AuthorizationAction = "seller.vendor.write"
	AuthorizationAdminReadOverview        AuthorizationAction = "admin.overview.read"
	AuthorizationAdminManageProducts      AuthorizationAction = "admin.products.manage"
	AuthorizationAdminManageUsers         AuthorizationAction = "admin.users.manage"
	AuthorizationAdminReviewMembership    AuthorizationAction = "admin.seller_membership.review"
	AuthorizationAdminManageRoles         AuthorizationAction = "admin.roles.manage"
)

type AuthorizationResource struct {
	OwnerUserID string
	VendorID    int64
}

var (
	ErrUnauthenticated = errors.New("authenticated session required")
	ErrForbidden       = errors.New("authorization denied")
)

// Authorize evaluates an explicit action against a principal resolved by
// ResolveSession. Callers must never construct the principal from request
// JSON, cookies other than the opaque session token, or browser role claims.
// Unknown actions are denied by default.
func Authorize(principal SessionPrincipal, action AuthorizationAction, resource AuthorizationResource) error {
	if strings.TrimSpace(principal.ID) == "" || strings.TrimSpace(principal.UserID) == "" {
		return ErrUnauthenticated
	}

	switch action {
	case AuthorizationCustomerReadOwnResource, AuthorizationCustomerWriteOwnResource:
		if !slices.Contains(principal.Roles, "customer") ||
			resource.OwnerUserID == "" ||
			resource.OwnerUserID != principal.UserID {
			return ErrForbidden
		}
		return nil

	case AuthorizationSellerReadVendor, AuthorizationSellerWriteVendor:
		if resource.VendorID <= 0 ||
			!slices.Contains(principal.Roles, "seller") ||
			!slices.Contains(principal.SellerVendorIDs, resource.VendorID) {
			return ErrForbidden
		}
		return nil

	case AuthorizationAdminReadOverview,
		AuthorizationAdminManageProducts,
		AuthorizationAdminManageUsers,
		AuthorizationAdminReviewMembership,
		AuthorizationAdminManageRoles:
		if !slices.Contains(principal.Roles, "admin") {
			return ErrForbidden
		}
		return nil

	default:
		return ErrForbidden
	}
}
