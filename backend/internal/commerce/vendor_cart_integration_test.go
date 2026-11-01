package commerce_test

import (
 "context"
 "crypto/sha256"
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

func TestVendorAwareCartLifecycle(t *testing.T) {
 dsn:=os.Getenv("TEST_DATABASE_URL")
 if dsn=="" {t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")}
 ctx,cancel:=context.WithTimeout(context.Background(),25*time.Second);defer cancel()
 pool,err:=pgxpool.New(ctx,dsn);if err!=nil{t.Fatalf("create pool: %v",err)};defer pool.Close()
 if err:=pool.Ping(ctx);err!=nil{t.Fatalf("ping database: %v",err)}
 if err:=database.Migrate(ctx,pool);err!=nil{t.Fatalf("migrate: %v",err)}
 productID:=int64(time.Now().UnixNano()%7000000000000000000)+1000000000000000000
 fixture:=catalog.Product{ID:productID,Title:"Vendor cart fixture",Description:"Integration fixture",Category:"vendor-cart-test",Price:12.34,Rating:4.5,Stock:8,Images:[]string{"https://example.invalid/vendor-cart.webp"},Source:"vendor-cart-integration",SourceID:fmt.Sprintf("vendor-cart-%d",productID),ImportedAt:time.Now().UTC()}
 if err:=database.UpsertProducts(ctx,pool,[]catalog.Product{fixture});err!=nil{t.Fatalf("upsert product fixture: %v",err)}
 svc:=commerce.New(pool)
 vendor,err:=svc.CreateVendor(ctx,fmt.Sprintf("vendor-cart-%d",productID),"Vendor Cart Test")
 if err!=nil{t.Fatalf("create vendor: %v",err)}
 if _,err:=pool.Exec(ctx,"UPDATE vendors SET status='active' WHERE id=$1",vendor.ID);err!=nil{t.Fatalf("activate vendor: %v",err)}
 offer,err:=svc.UpsertVendorProduct(ctx,commerce.VendorProduct{VendorID:vendor.ID,ProductID:productID,Price:"18.50",Currency:"USD",Stock:2,Status:"active"})
 if err!=nil{t.Fatalf("create vendor offer: %v",err)}
 owner:=sha256.Sum256([]byte(fmt.Sprintf("vendor-cart-owner-%d",productID)))
 cart,err:=svc.CreateCart(ctx,owner[:],"USD",time.Hour)
 if err!=nil{t.Fatalf("create cart: %v",err)}
 cart,err=svc.AddVendorItem(ctx,owner[:],offer.ID,1)
 if err!=nil{t.Fatalf("add vendor item: %v",err)}
 if len(cart.Items)!=1||cart.Items[0].VendorProductID==nil||*cart.Items[0].VendorProductID!=offer.ID||cart.Items[0].VendorID==nil||*cart.Items[0].VendorID!=vendor.ID||cart.Items[0].VendorName!="Vendor Cart Test"||cart.Items[0].UnitPrice!="18.50" {
  t.Fatalf("vendor cart line = %+v; offer/vendor/price snapshot mismatch",cart.Items)
 }
 idempotencyHash:=sha256.Sum256([]byte("vendor-cart-checkout-key"))
 if _,err:=svc.Checkout(ctx,owner[:],idempotencyHash[:],15*time.Minute);!errors.Is(err,commerce.ErrVendorCheckoutNotReady) {
  t.Fatalf("legacy checkout error=%v, want ErrVendorCheckoutNotReady",err)
 }
 cart,err=svc.SetVendorItemQuantity(ctx,owner[:],offer.ID,2)
 if err!=nil||cart.Items[0].Quantity!=2{t.Fatalf("set vendor quantity cart=%+v err=%v",cart,err)}
 if _,err:=svc.SetVendorItemQuantity(ctx,owner[:],offer.ID,3);!errors.Is(err,commerce.ErrInsufficientStock){
  t.Fatalf("quantity beyond offer stock error=%v, want ErrInsufficientStock",err)
 }
 cart,err=svc.RemoveVendorItem(ctx,owner[:],offer.ID)
 if err!=nil||len(cart.Items)!=0{t.Fatalf("remove vendor line cart=%+v err=%v",cart,err)}
 if _,err:=svc.AddVendorItem(ctx,owner[:],offer.ID,3);!errors.Is(err,commerce.ErrInsufficientStock){
  t.Fatalf("add beyond offer stock error=%v, want ErrInsufficientStock",err)
 }
}
}
