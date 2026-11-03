package admin

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pooya-labs/globalcart/backend/internal/auth"
	"github.com/pooya-labs/globalcart/backend/internal/catalog"
	"github.com/pooya-labs/globalcart/backend/internal/database"
)

func TestListManagedProductsRequiresAdminAndSupportsFiltersAndPagination(t *testing.T) {
	pool, ctx := openOverviewIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	_, adminSession := createProductManagementAdmin(t, ctx, pool, suffix+"-admin")

	customerID := createOverviewUser(t, ctx, pool, suffix+"-customer")
	customerSession := createOverviewSession(t, ctx, pool, customerID)
	if _, err := ListManagedProducts(ctx, pool, customerSession.SessionToken, ManagedProductListInput{}); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer product-list access error = %v, want auth.ErrForbidden", err)
	}

	baseID := time.Now().UnixNano()
	category := "managed-list-" + suffix
	now := time.Now().UTC()
	products := []catalog.Product{
		{
			ID: baseID + 101, Title: "Supplier Alpha", Description: "Alpha description",
			Category: category, Price: 12.50, DiscountPercentage: 5, Rating: 4.2,
			Stock: 9, Brand: "Source Brand", Thumbnail: "https://example.invalid/alpha.webp",
			Images: []string{"https://example.invalid/alpha.webp"},
			Source: "managed-list-fixture", SourceID: suffix + "-alpha", ImportedAt: now,
		},
		{
			ID: baseID + 102, Title: "Supplier Beta", Description: "Beta description",
			Category: category, Price: 22.50, DiscountPercentage: 0, Rating: 4.6,
			Stock: 4, Brand: "Source Brand", Thumbnail: "https://example.invalid/beta.webp",
			Images: []string{"https://example.invalid/beta.webp"},
			Source: "managed-list-fixture", SourceID: suffix + "-beta", ImportedAt: now.Add(time.Second),
		},
		{
			ID: baseID + 103, Title: "Supplier Gamma", Description: "Gamma description",
			Category: category, Price: 32.50, DiscountPercentage: 10, Rating: 4.8,
			Stock: 2, Brand: "Source Brand", Thumbnail: "https://example.invalid/gamma.webp",
			Images: []string{"https://example.invalid/gamma.webp"},
			Source: "managed-list-fixture", SourceID: suffix + "-gamma", ImportedAt: now.Add(2 * time.Second),
		},
	}
	if err := database.UpsertProducts(ctx, pool, products); err != nil {
		t.Fatalf("upsert managed-list fixtures: %v", err)
	}

	editorialTitle := "Curated Studio Alpha"
	editorialBrand := "Editorial Labs"
	updated, err := UpdateManagedProduct(ctx, pool, adminSession.SessionToken, products[0].ID, ProductUpdateInput{
		ExpectedRevision: 0, Title: &editorialTitle, Brand: &editorialBrand,
	})
	if err != nil {
		t.Fatalf("apply editorial values: %v", err)
	}
	if updated.Title != editorialTitle || updated.OriginalTitle != products[0].Title {
		t.Fatalf("editorial/source title pair = %q/%q", updated.Title, updated.OriginalTitle)
	}

	if _, err := ArchiveManagedProduct(ctx, pool, adminSession.SessionToken, products[2].ID, 0); err != nil {
		t.Fatalf("archive gamma fixture: %v", err)
	}

	activePage, err := ListManagedProducts(ctx, pool, adminSession.SessionToken, ManagedProductListInput{Category: category})
	if err != nil {
		t.Fatalf("list active category products: %v", err)
	}
	if activePage.Total != 2 || len(activePage.Items) != 2 || activePage.Limit != managedProductListDefaultLimit {
		t.Fatalf("default active page = total %d, items %d, limit %d; want 2, 2, %d",
			activePage.Total, len(activePage.Items), activePage.Limit, managedProductListDefaultLimit)
	}
	for _, item := range activePage.Items {
		if item.VisibilityStatus != "active" {
			t.Fatalf("default list included non-active product: %+v", item)
		}
	}

	searchPage, err := ListManagedProducts(ctx, pool, adminSession.SessionToken, ManagedProductListInput{
		Query: "  curated studio  ", Category: category,
	})
	if err != nil {
		t.Fatalf("search effective editorial title: %v", err)
	}
	if searchPage.Total != 1 || len(searchPage.Items) != 1 ||
		searchPage.Items[0].ID != products[0].ID ||
		searchPage.Items[0].Title != editorialTitle ||
		searchPage.Items[0].OriginalTitle != products[0].Title {
		t.Fatalf("effective-title search result = %+v", searchPage)
	}

	brandPage, err := ListManagedProducts(ctx, pool, adminSession.SessionToken, ManagedProductListInput{
		Brand: "editorial labs", Category: category,
	})
	if err != nil {
		t.Fatalf("filter effective editorial brand: %v", err)
	}
	if brandPage.Total != 1 || len(brandPage.Items) != 1 || brandPage.Items[0].ID != products[0].ID {
		t.Fatalf("effective-brand filter result = %+v", brandPage)
	}

	sourcePage, err := ListManagedProducts(ctx, pool, adminSession.SessionToken, ManagedProductListInput{
		Source: "MANAGED-LIST-FIXTURE", Category: category,
	})
	if err != nil {
		t.Fatalf("filter source provenance: %v", err)
	}
	if sourcePage.Total != 2 || len(sourcePage.Items) != 2 {
		t.Fatalf("source filter result total/items = %d/%d, want 2/2", sourcePage.Total, len(sourcePage.Items))
	}

	archivedPage, err := ListManagedProducts(ctx, pool, adminSession.SessionToken, ManagedProductListInput{
		Category: category, Status: "archived",
	})
	if err != nil {
		t.Fatalf("list archived products: %v", err)
	}
	if archivedPage.Total != 1 || len(archivedPage.Items) != 1 ||
		archivedPage.Items[0].ID != products[2].ID ||
		archivedPage.Items[0].VisibilityStatus != "archived" {
		t.Fatalf("archived filter result = %+v", archivedPage)
	}

	allPage, err := ListManagedProducts(ctx, pool, adminSession.SessionToken, ManagedProductListInput{
		Category: category, Status: "all", SortBy: "title", SortOrder: "asc", Limit: 1, Offset: 1,
	})
	if err != nil {
		t.Fatalf("list all products with title pagination: %v", err)
	}
	if allPage.Total != 3 || allPage.Limit != 1 || allPage.Offset != 1 || len(allPage.Items) != 1 ||
		allPage.Items[0].ID != products[1].ID || allPage.Items[0].Title != products[1].Title {
		t.Fatalf("all-products pagination result = %+v; want second title-sorted product", allPage)
	}

	outOfRangePage, err := ListManagedProducts(ctx, pool, adminSession.SessionToken, ManagedProductListInput{
		Category: category, Status: "all", Limit: 1, Offset: 10,
	})
	if err != nil {
		t.Fatalf("list beyond final page: %v", err)
	}
	if outOfRangePage.Total != 3 || len(outOfRangePage.Items) != 0 {
		t.Fatalf("out-of-range page total/items = %d/%d, want 3/0", outOfRangePage.Total, len(outOfRangePage.Items))
	}
}

func TestListManagedProductsRejectsUnboundedOrUnknownFilters(t *testing.T) {
	cases := []ManagedProductListInput{
		{Status: "deleted"},
		{SortBy: "supplier_sql"},
		{SortOrder: "sideways"},
		{Limit: managedProductListMaximumLimit + 1},
		{Offset: managedProductListMaximumOffset + 1},
		{Query: string(make([]byte, 121))},
		{Query: "invalid" + string(rune(0))},
	}
	for index, input := range cases {
		if _, err := normalizeManagedProductListInput(input); !errors.Is(err, ErrInvalidProductListInput) {
			t.Errorf("case %d normalization error = %v, want ErrInvalidProductListInput", index, err)
		}
	}
}
