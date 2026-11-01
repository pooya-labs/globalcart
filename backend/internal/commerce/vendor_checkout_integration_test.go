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

func TestVendorSplitCheckoutAndCouponAllocation(t *testing.T) {
 dsn:=os.Getenv("TEST_DATABASE_URL");if dsn==""{t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")}
 ctx,cancel:=context.WithTimeout(context.Background(),40*time.Second);defer cancel()
 pool,err:=pgxpool.New(ctx,dsn);if err!=nil{t.Fatalf("create pool: %v",err)};defer pool.Close()
 if err=pool.Ping(ctx);err!=nil{t.Fatalf("ping database: %v",err)}
 if err=database.Migrate(ctx,pool);err!=nil{t.Fatalf("migrate: %v",err)}
 productID:=int64(3000000000000+time.Now().UnixNano()%1000000000000)
 fixture:=catalog.Product{ID:productID,Title:"Split checkout fixture",Description:"Integration fixture",Category:"split-checkout-test",Price:9.99,Rating:4.5,Stock:17,Images:[]string{"https://example.invalid/split.webp"},Source:"split-checkout-integration",SourceID:fmt.Sprintf("split-%d",productID),ImportedAt:time.Now().UTC()}
 if err=database.UpsertProducts(ctx,pool,[]catalog.Product{fixture});err!=nil{t.Fatalf("upsert product: %v",err)}
 svc:=commerce.New(pool)
 createVendor:=func(suffix,name string) commerce.Vendor {
  v,e:=svc.CreateVendor(ctx,fmt.Sprintf("split-%s-%d",suffix,productID),name);if e!=nil{t.Fatalf("create vendor: %v",e)}
  if _,e=pool.Exec(ctx,"UPDATE vendors SET status='active' WHERE id=$1",v.ID);e!=nil{t.Fatalf("activate vendor: %v",e)}
  return v
 }
 a:=createVendor("a","Split Vendor A");b:=createVendor("b","Split Vendor B")
 offerA,err:=svc.UpsertVendorProduct(ctx,commerce.VendorProduct{VendorID:a.ID,ProductID:productID,Price:"10.00",Currency:"USD",Stock:4,Status:"active"});if err!=nil{t.Fatalf("create offer A: %v",err)}
 offerB,err:=svc.UpsertVendorProduct(ctx,commerce.VendorProduct{VendorID:b.ID,ProductID:productID,Price:"20.00",Currency:"USD",Stock:5,Status:"active"});if err!=nil{t.Fatalf("create offer B: %v",err)}
 owner:=sha256.Sum256([]byte(fmt.Sprintf("split-owner-%d",productID)))
 cart,err:=svc.CreateCart(ctx,owner[:],"USD",time.Hour);if err!=nil{t.Fatalf("create cart: %v",err)}
 if _,err=svc.AddVendorItem(ctx,owner[:],offerA.ID,2);err!=nil{t.Fatalf("add A: %v",err)}
 if _,err=svc.AddVendorItem(ctx,owner[:],offerB.ID,1);err!=nil{t.Fatalf("add B: %v",err)}
 starts:=time.Now().Add(-time.Minute);ends:=time.Now().Add(time.Hour);limit:=int64(10)
 if err=svc.CreateCoupon(ctx,fmt.Sprintf("SPLIT-%d",productID),commerce.CouponSpec{DiscountType:"percentage",DiscountValue:"10",MinimumSubtotal:"1",StartsAt:starts,EndsAt:ends,UsageLimit:&limit,PerCustomerLimit:1});err!=nil{t.Fatalf("create coupon: %v",err)}
 if err=svc.ApplyCoupon(ctx,owner[:],fmt.Sprintf("split-%d",productID));err!=nil{t.Fatalf("apply coupon: %v",err)}
 key:=sha256.Sum256([]byte(fmt.Sprintf("split-key-%d",productID)))
 order,err:=svc.CheckoutVendorCart(ctx,owner[:],key[:],20*time.Minute);if err!=nil{t.Fatalf("vendor split checkout: %v",err)}
 if order.Subtotal!="40.00"||order.DiscountAmount!="4.00"||order.Total!="36.00"{t.Fatalf("parent totals=%+v",order)}
 var childCount int;var childSubtotal,childDiscount,childTotal string
 if err=pool.QueryRow(ctx,`SELECT count(*),sum(subtotal)::numeric(12,2)::text,sum(discount_amount)::numeric(12,2)::text,sum(total)::numeric(12,2)::text FROM demo_vendor_orders WHERE order_id=$1`,order.ID).Scan(&childCount,&childSubtotal,&childDiscount,&childTotal);err!=nil{t.Fatalf("read child totals: %v",err)}
 if childCount!=2||childSubtotal!="40.00"||childDiscount!="4.00"||childTotal!="36.00"{t.Fatalf("child count=%d subtotal=%s discount=%s total=%s",childCount,childSubtotal,childDiscount,childTotal)}
 var reservationCount,linkedCount int
 if err=pool.QueryRow(ctx,"SELECT count(*),count(vendor_order_id) FROM inventory_reservations WHERE order_id=$1 AND vendor_product_id IS NOT NULL AND status='active'",order.ID).Scan(&reservationCount,&linkedCount);err!=nil{t.Fatalf("read vendor reservations: %v",err)}
 if reservationCount!=2||linkedCount!=2{t.Fatalf("vendor reservations=%d linked=%d",reservationCount,linkedCount)}
 replay,err:=svc.CheckoutVendorCart(ctx,owner[:],key[:],20*time.Minute);if err!=nil||!replay.Replayed||replay.ID!=order.ID{t.Fatalf("idempotent replay=%+v err=%v",replay,err)}
 var legacyStock int32
 if err=pool.QueryRow(ctx,"SELECT stock FROM products WHERE id=$1",productID).Scan(&legacyStock);err!=nil{t.Fatalf("read legacy stock: %v",err)}
 if legacyStock!=17{t.Fatalf("legacy product stock mutated by vendor checkout: %d",legacyStock)}
 if _,err=svc.CompleteDemoPayment(ctx,owner[:],order.ID);err!=nil{t.Fatalf("complete simulated payment: %v",err)}
 var paidChildren int
 if err=pool.QueryRow(ctx,"SELECT count(*) FROM demo_vendor_orders WHERE order_id=$1 AND status='paid_simulated'",order.ID).Scan(&paidChildren);err!=nil{t.Fatalf("read paid children: %v",err)}
 if paidChildren!=2{t.Fatalf("paid child orders=%d, want 2",paidChildren)}
 _=cart
}

func TestVendorSplitCheckoutStockRace(t *testing.T) {
 dsn:=os.Getenv("TEST_DATABASE_URL");if dsn==""{t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")}
 ctx,cancel:=context.WithTimeout(context.Background(),40*time.Second);defer cancel()
 pool,err:=pgxpool.New(ctx,dsn);if err!=nil{t.Fatalf("create pool: %v",err)};defer pool.Close()
 if err=database.Migrate(ctx,pool);err!=nil{t.Fatalf("migrate: %v",err)}
 productID:=int64(5000000000000+time.Now().UnixNano()%1000000000000)
 fixture:=catalog.Product{ID:productID,Title:"Vendor race fixture",Description:"Integration fixture",Category:"vendor-race-test",Price:3,Rating:4,Stock:9,Images:[]string{"https://example.invalid/race.webp"},Source:"vendor-race-integration",SourceID:fmt.Sprintf("race-%d",productID),ImportedAt:time.Now().UTC()}
 if err=database.UpsertProducts(ctx,pool,[]catalog.Product{fixture});err!=nil{t.Fatalf("upsert product: %v",err)}
 svc:=commerce.New(pool);v,err:=svc.CreateVendor(ctx,fmt.Sprintf("race-%d",productID),"Race Vendor");if err!=nil{t.Fatalf("create vendor: %v",err)}
 if _,err=pool.Exec(ctx,"UPDATE vendors SET status='active' WHERE id=$1",v.ID);err!=nil{t.Fatalf("activate vendor: %v",err)}
 offer,err:=svc.UpsertVendorProduct(ctx,commerce.VendorProduct{VendorID:v.ID,ProductID:productID,Price:"3.00",Currency:"USD",Stock:1,Status:"active"});if err!=nil{t.Fatalf("create offer: %v",err)}
 type outcome struct{err error};results:=make(chan outcome,2)
 for i:=0;i<2;i++{i:=i;go func(){owner:=sha256.Sum256([]byte(fmt.Sprintf("race-owner-%d-%d",productID,i)));key:=sha256.Sum256([]byte(fmt.Sprintf("race-key-%d-%d",productID,i)));if _,e:=svc.CreateCart(ctx,owner[:],"USD",time.Hour);e!=nil{results<-outcome{e};return};if _,e:=svc.AddVendorItem(ctx,owner[:],offer.ID,1);e!=nil{results<-outcome{e};return};_,e:=svc.CheckoutVendorCart(ctx,owner[:],key[:],20*time.Minute);results<-outcome{e}}()}
 success,stockFailure:=0,0
 for i:=0;i<2;i++{r:=<-results;if r.err==nil{success++}else if errors.Is(r.err,commerce.ErrInsufficientStock){stockFailure++}else{t.Fatalf("unexpected concurrent checkout error: %v",r.err)}}
 if success!=1||stockFailure!=1{t.Fatalf("success=%d insufficient-stock=%d; want one each",success,stockFailure)}
}
