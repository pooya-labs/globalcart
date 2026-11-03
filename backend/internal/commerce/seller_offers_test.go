package commerce

import (
	"errors"
	"testing"
)

func TestNormalizeSellerOfferPrice(t *testing.T) {
	cases:=map[string]string{"1":"1.00","1.2":"1.20","19.95":"19.95","9999999999.99":"9999999999.99"}
	for input,want:=range cases {
		got,err:=normalizeSellerOfferPrice(input)
		if err!=nil || got!=want { t.Fatalf("%q => %q,%v want %q",input,got,err,want) }
	}
	for _,input:=range []string{"","0","0.00","-1.00","+1.00","1e3","1.001","1.","01.25","10000000000.00","۱۲.۵۰"} {
		if _,err:=normalizeSellerOfferPrice(input); !errors.Is(err,ErrInvalidSellerOfferInput) {
			t.Fatalf("%q error=%v want invalid input",input,err)
		}
	}
}

func TestNormalizeSellerOfferUpdate(t *testing.T) {
	price,sku,stock,status:="12.5"," SELLER-42 ",int32(7)," PAUSED "
	got,err:=normalizeSellerOfferUpdate(SellerOfferUpdateInput{
		ExpectedRevision:3,SellerSKU:&sku,Price:&price,Stock:&stock,Status:&status,
	})
	if err!=nil { t.Fatal(err) }
	if *got.Price!="12.50" || *got.SellerSKU!="SELLER-42" || *got.Stock!=7 || *got.Status!="paused" {
		t.Fatalf("normalized patch=%+v",got)
	}
	active:="active"
	if _,err:=normalizeSellerOfferUpdate(SellerOfferUpdateInput{Status:&active}); !errors.Is(err,ErrInvalidSellerOfferInput) {
		t.Fatalf("self publish err=%v",err)
	}
	if _,err:=normalizeSellerOfferUpdate(SellerOfferUpdateInput{}); !errors.Is(err,ErrSellerOfferNoChanges) {
		t.Fatalf("empty patch err=%v",err)
	}
}
