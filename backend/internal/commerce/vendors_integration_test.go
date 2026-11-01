package commerce_test

import (
 "context"
 "errors"
 "fmt"
 "os"
 "testing"
 "time"

 "github.com/jackc/pgx/v5/pgxpool"
 "github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
 "github.com/pooya16pg-spec/globalcart/backend/internal/commerce"
 "github.com/pooya16pg-spec/globalcart/backend/internal/database"
)

func TestVendorServiceTransactions(t *testing.T) {
 dsn:=os.Getenv("TEST_DATABASE_URL")
 if dsn=="" { t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped") }
 ctx,cancel:=context.WithTimeout(context.Background(),20*time.Second); defer cancel()
 pool,err:=pgxpool.New(ctx,dsn); if err!=nil {t.Fatalf("create pool: %v",err)}; defer pool.Close()
 if err:=pool.Ping(ctx);err!=nil {t.Fatalf("ping database: %v",err)}
 if err:=database.Migrate(ctx,pool);err!=nil {t.Fatalf("migrate: %v",err)}
 productID:=int64(time.Now().UnixNano()%8000000000000000000)+1000000000000000000
 fixture:=catalog.Product{ID:productID,Title:"Vendor service fixture",Description:"Integration fixture",Category:"vendor-test",Price:12.34,Rating:4.5,Stock:5,Images:[]string{"https://example.invalid/vendor.webp"},Source:"vendor-integration",SourceID:fmt.Sprintf("vendor-%d",productID),ImportedAt:time.Now().UTC()}
 if err:=database.UpsertProducts(ctx,pool,[]catalog.Product{fixture});err!=nil {t.Fatalf("upsert fixture: %v",err)}
 var variantID int64
 if err:=pool.QueryRow(ctx,`INSERT INTO product_variants(product_id,sku,title,price,stock,attributes,source,source_id)
 VALUES($1,$2,'Blue / Large',12.34,5,'{"color":"blue","size":"L"}'::jsonb,'vendor-integration',$3) RETURNING id`,
 productID,fmt.Sprintf("VND-%d",productID),fmt.Sprintf("variant-%d",productID)).Scan(&variantID);err!=nil {t.Fatalf("insert variant: %v",err)}
 svc:=commerce.New(pool)
 slug:=fmt.Sprintf("vendor-%d",productID)
 vendor,err:=svc.CreateVendor(ctx,slug,"Integration Vendor")
 if err!=nil {t.Fatalf("create vendor: %v",err)}
 if vendor.Status!="pending" {t.Fatalf("new vendor status=%q, want pending",vendor.Status)}
 if _,err:=svc.UpsertVendorProduct(ctx,commerce.VendorProduct{VendorID:vendor.ID,ProductID:productID,VariantID:&variantID,Price:"15.25",Currency:"USD",Stock:4,Status:"active"});!errors.Is(err,commerce.ErrVendorInactive) {
  t.Fatalf("offer for pending vendor error=%v, want ErrVendorInactive",err)
 }
 if _,err:=pool.Exec(ctx,"UPDATE vendors SET status='active' WHERE id=$1",vendor.ID);err!=nil {t.Fatalf("activate fixture vendor: %v",err)}
 offer,err:=svc.UpsertVendorProduct(ctx,commerce.VendorProduct{VendorID:vendor.ID,ProductID:productID,VariantID:&variantID,SellerSKU:"VENDOR-BLUE-L",Price:"15.25",Currency:"USD",Stock:4,Status:"active"})
 if err!=nil {t.Fatalf("upsert offer: %v",err)}
 if offer.ID<=0 {t.Fatalf("offer id=%d, want positive",offer.ID)}
 if err:=svc.UpdateVendorPrice(ctx,vendor.ID,offer.ID,"16.75");err!=nil {t.Fatalf("update price: %v",err)}
 stock,err:=svc.AdjustVendorInventory(ctx,vendor.ID,offer.ID,3,"restock","PO-1001")
 if err!=nil||stock!=7 {t.Fatalf("restock result=%d err=%v, want 7",stock,err)}
 if _,err:=svc.AdjustVendorInventory(ctx,vendor.ID,offer.ID,-99,"stock_adjustment","bad-adjustment");!errors.Is(err,commerce.ErrInsufficientStock) {
  t.Fatalf("stock underflow error=%v, want ErrInsufficientStock",err)
 }
 got,err:=svc.GetVendorProduct(ctx,vendor.ID,offer.ID)
 if err!=nil||got.Price!="16.75"||got.Stock!=7 {t.Fatalf("read offer=%+v err=%v, want price 16.75 stock 7",got,err)}
 var prices,events int
 if err:=pool.QueryRow(ctx,"SELECT count(*)::int FROM vendor_price_history WHERE vendor_product_id=$1",offer.ID).Scan(&prices);err!=nil {t.Fatalf("count prices: %v",err)}
 if err:=pool.QueryRow(ctx,"SELECT count(*)::int FROM vendor_inventory_events WHERE vendor_product_id=$1",offer.ID).Scan(&events);err!=nil {t.Fatalf("count inventory events: %v",err)}
 if prices!=2||events!=2 {t.Fatalf("audit counts prices=%d events=%d, want 2 each",prices,events)}
 if err:=svc.UpdateVendorPrice(ctx,vendor.ID+999999,offer.ID,"17.00");!errors.Is(err,commerce.ErrVendorProductNotFound) {
  t.Fatalf("cross-vendor price update error=%v, want not found",err)
 }
}
