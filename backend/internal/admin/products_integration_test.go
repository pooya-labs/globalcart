package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
	"github.com/pooya16pg-spec/globalcart/backend/internal/dbgen"
)

func createProductManagementAdmin(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) (string, auth.IssuedSession) {
	t.Helper()
	userID := createOverviewUser(t, ctx, pool, suffix)
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key, granted_by_user_id)
		VALUES($1, 'admin', NULL)
	`, userID); err != nil {
		t.Fatalf("grant product-management admin fixture role: %v", err)
	}
	session := createOverviewSession(t, ctx, pool, userID)
	return userID, session
}

func TestManagedProductEditorialOverridesArchiveAndRestore(t *testing.T) {
	pool, ctx := openOverviewIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	_, adminSession := createProductManagementAdmin(t, ctx, pool, suffix)

	productID := time.Now().UnixNano()
	category := "admin-product-" + suffix
	original := catalog.Product{
		ID: productID, Title: "Original Supplier Camera",
		Description: "Supplier description", Category: category,
		Price: 45.50, DiscountPercentage: 5, Rating: 4.4, Stock: 8,
		Brand: "Supplier Brand", Thumbnail: "https://example.invalid/source.webp",
		Images: []string{"https://example.invalid/source.webp"},
		Source: "admin-product-integration", SourceID: suffix,
		ImportedAt: time.Now().UTC(),
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{original}); err != nil {
		t.Fatalf("upsert original supplier product: %v", err)
	}

	initial, err := ReadManagedProduct(ctx, pool, adminSession.SessionToken, productID)
	if err != nil {
		t.Fatalf("read original managed product: %v", err)
	}
	if initial.Title != original.Title || initial.OriginalTitle != original.Title ||
		initial.Brand != original.Brand || initial.VisibilityStatus != "active" ||
		initial.Revision != 0 || len(initial.OverriddenFields) != 0 {
		t.Fatalf("unexpected initial managed product: %+v", initial)
	}

	title := "  Curated Editorial Camera  "
	description := ""
	brand := "Studio Override"
	updated, err := UpdateManagedProduct(ctx, pool, adminSession.SessionToken, productID, ProductUpdateInput{
		ExpectedRevision: 0,
		Title:            &title, Description: &description, Brand: &brand,
	})
	if err != nil {
		t.Fatalf("update product editorial overrides: %v", err)
	}
	if updated.Title != "Curated Editorial Camera" ||
		updated.Description != "" || updated.Brand != brand ||
		updated.OriginalTitle != original.Title ||
		updated.OriginalDescription != original.Description ||
		updated.OriginalBrand != original.Brand {
		t.Fatalf("editorial view/source values are incorrect: %+v", updated)
	}
	if updated.Revision != 1 || len(updated.OverriddenFields) != 3 {
		t.Fatalf("revision/override fields = %d/%v, want 1 and all three", updated.Revision, updated.OverriddenFields)
	}
	staleTitle := "Stale administrator edit"
	if _, err := UpdateManagedProduct(ctx, pool, adminSession.SessionToken, productID, ProductUpdateInput{
		ExpectedRevision: 0, Title: &staleTitle,
	}); !errors.Is(err, ErrStaleProductRevision) {
		t.Fatalf("stale product update error = %v, want ErrStaleProductRevision", err)
	}

	// The public read-only catalog API applies the editorial view and brand
	// filtering, while source-price/stock and supplier provenance remain intact.
	queries := dbgen.New(pool)
	rows, err := queries.ListProducts(ctx, dbgen.ListProductsParams{
		Category: category, Brand: brand, PageLimit: 10,
	})
	if err != nil {
		t.Fatalf("list product by editorial brand: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != productID || rows[0].Title != updated.Title ||
		rows[0].Description != updated.Description || rows[0].Brand != updated.Brand {
		t.Fatalf("catalog list did not use editorial values: %+v", rows)
	}
	total, err := queries.CountProducts(ctx, dbgen.CountProductsParams{Category: category, Brand: brand})
	if err != nil || total != 1 {
		t.Fatalf("catalog count with editorial brand = %d, error %v; want 1", total, err)
	}
	searchRows, err := queries.ListProducts(ctx, dbgen.ListProductsParams{
		Query: "Curated Editorial Camera", PageLimit: 10,
	})
	if err != nil {
		t.Fatalf("search by editorial title: %v", err)
	}
	foundEditedProduct := false
	for _, row := range searchRows {
		if row.ID == productID {
			foundEditedProduct = row.Title == updated.Title
		}
	}
	if !foundEditedProduct {
		t.Fatalf("catalog search did not index the effective editorial title: %+v", searchRows)
	}

	// Re-import a fresh supplier snapshot. It updates only the source table;
	// editorial overrides survive and still win in the read model.
	original.Title = "Supplier Updated Title"
	original.Description = "Supplier refreshed description"
	original.Brand = "Supplier Updated Brand"
	original.ImportedAt = time.Now().UTC().Add(time.Second)
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{original}); err != nil {
		t.Fatalf("reimport supplier product after editorial change: %v", err)
	}
	refreshed, err := ReadManagedProduct(ctx, pool, adminSession.SessionToken, productID)
	if err != nil {
		t.Fatalf("read product after supplier refresh: %v", err)
	}
	if refreshed.Title != updated.Title || refreshed.Description != updated.Description ||
		refreshed.Brand != updated.Brand ||
		refreshed.OriginalTitle != original.Title ||
		refreshed.OriginalDescription != original.Description ||
		refreshed.OriginalBrand != original.Brand {
		t.Fatalf("supplier refresh clobbered or hid editorial/source values: %+v", refreshed)
	}

	archived, err := ArchiveManagedProduct(ctx, pool, adminSession.SessionToken, productID, 1)
	if err != nil {
		t.Fatalf("archive product: %v", err)
	}
	if archived.VisibilityStatus != "archived" || archived.Revision != 2 {
		t.Fatalf("archive state/revision = %s/%d, want archived/2", archived.VisibilityStatus, archived.Revision)
	}
	if _, err := RestoreManagedProduct(ctx, pool, adminSession.SessionToken, productID, 1); !errors.Is(err, ErrStaleProductRevision) {
		t.Fatalf("stale restore error = %v, want ErrStaleProductRevision", err)
	}
	archivedRows, err := queries.ListProducts(ctx, dbgen.ListProductsParams{
		Category: category, PageLimit: 10,
	})
	if err != nil || len(archivedRows) != 0 {
		t.Fatalf("archived product remains in catalog list: rows=%+v err=%v", archivedRows, err)
	}
	if _, err := queries.GetProduct(ctx, productID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("archived product detail error = %v, want pgx.ErrNoRows", err)
	}
	if _, err := ArchiveManagedProduct(ctx, pool, adminSession.SessionToken, productID, 2); !errors.Is(err, ErrProductAlreadyArchived) {
		t.Fatalf("repeated archive error = %v, want ErrProductAlreadyArchived", err)
	}

	restored, err := RestoreManagedProduct(ctx, pool, adminSession.SessionToken, productID, 2)
	if err != nil {
		t.Fatalf("restore product: %v", err)
	}
	if restored.VisibilityStatus != "active" || restored.Revision != 3 ||
		restored.Title != updated.Title || restored.Brand != updated.Brand {
		t.Fatalf("restored product lost editorial values: %+v", restored)
	}
	visibleAgain, err := queries.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("restored product detail: %v", err)
	}
	if visibleAgain.Title != updated.Title || visibleAgain.Brand != updated.Brand {
		t.Fatalf("restored catalog detail did not use editorial values: %+v", visibleAgain)
	}

	var revisionCount, updateEvents, archiveEvents, restoreEvents int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int FROM admin_product_revisions WHERE product_id = $1
	`, productID).Scan(&revisionCount); err != nil {
		t.Fatalf("count product revisions: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE event_type = 'admin.product.updated')::int,
			count(*) FILTER (WHERE event_type = 'admin.product.archived')::int,
			count(*) FILTER (WHERE event_type = 'admin.product.restored')::int
		FROM audit_events
		WHERE resource_type = 'product' AND resource_id = $1
	`, fmt.Sprint(productID)).Scan(&updateEvents, &archiveEvents, &restoreEvents); err != nil {
		t.Fatalf("count product audit events: %v", err)
	}
	if revisionCount != 3 || updateEvents != 1 || archiveEvents != 1 || restoreEvents != 1 {
		t.Fatalf("revision/audit counts = %d/%d/%d/%d, want 3/1/1/1",
			revisionCount, updateEvents, archiveEvents, restoreEvents)
	}
	history, err := ReadProductRevisionHistory(ctx, pool, adminSession.SessionToken, productID, 10, 0)
	if err != nil {
		t.Fatalf("read administrator product revision history: %v", err)
	}
	if len(history) != 3 || history[0].Revision != 3 || history[0].Action != "restored" ||
		history[1].Revision != 2 || history[1].Action != "archived" ||
		history[2].Revision != 1 || history[2].Action != "updated" {
		t.Fatalf("revision history ordering/content = %+v, want newest-first restored/archived/updated", history)
	}
	if !strings.Contains(string(history[2].After), "Curated Editorial Camera") ||
		!strings.Contains(string(history[2].Before), "Original Supplier Camera") {
		t.Fatalf("revision history did not preserve before/after snapshots: %+v", history[2])
	}
	if _, err := pool.Exec(ctx, `
		UPDATE admin_product_revisions SET action = action WHERE product_id = $1
	`, productID); err == nil {
		t.Fatal("product revision history accepted an update despite its append-only trigger")
	}

	var storedTitle, storedDescription, storedBrand string
	if err := pool.QueryRow(ctx, `
		SELECT title, description, brand FROM products WHERE id = $1
	`, productID).Scan(&storedTitle, &storedDescription, &storedBrand); err != nil {
		t.Fatalf("read final supplier source record: %v", err)
	}
	if storedTitle != original.Title || storedDescription != original.Description || storedBrand != original.Brand {
		t.Fatalf("admin changed supplier-owned source fields: %q / %q / %q", storedTitle, storedDescription, storedBrand)
	}
}

func TestManagedProductRequiresAdminAndValidInput(t *testing.T) {
	pool, ctx := openOverviewIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	customerID := createOverviewUser(t, ctx, pool, suffix+"-customer")
	customerSession := createOverviewSession(t, ctx, pool, customerID)

	productID := time.Now().UnixNano()
	product := catalog.Product{
		ID: productID, Title: "Permission Fixture", Description: "Fixture",
		Category: "admin-permission-" + suffix, Price: 10, Rating: 4,
		Stock: 1, Brand: "Fixture", Images: []string{},
		Source: "admin-product-permission", SourceID: suffix,
		ImportedAt: time.Now().UTC(),
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("upsert permission fixture: %v", err)
	}
	if _, err := ReadManagedProduct(ctx, pool, customerSession.SessionToken, productID); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer read managed product error = %v, want auth.ErrForbidden", err)
	}
	title := "Customer must not edit"
	if _, err := UpdateManagedProduct(ctx, pool, customerSession.SessionToken, productID, ProductUpdateInput{Title: &title}); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer edit managed product error = %v, want auth.ErrForbidden", err)
	}
	if _, err := ReadProductRevisionHistory(ctx, pool, customerSession.SessionToken, productID, 10, 0); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer read product revision history error = %v, want auth.ErrForbidden", err)
	}

	tooLong := strings.Repeat("x", 301)
	if _, _, err := normalizeProductUpdate(ProductUpdateInput{Title: &tooLong}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("overlong title validation error = %v, want ErrInvalidProductInput", err)
	}
	invalidClear := ProductUpdateInput{ClearFields: []string{"price"}}
	if _, _, err := normalizeProductUpdate(invalidClear); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("unknown clear field validation error = %v, want ErrInvalidProductInput", err)
	}
	conflict := ProductUpdateInput{Title: &title, ClearFields: []string{"title"}}
	if _, _, err := normalizeProductUpdate(conflict); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("set-and-clear validation error = %v, want ErrInvalidProductInput", err)
	}
	if _, _, err := normalizeProductUpdate(ProductUpdateInput{}); err != nil {
		t.Fatalf("empty input normalization should be allowed for explicit no-op detection: %v", err)
	}
}
