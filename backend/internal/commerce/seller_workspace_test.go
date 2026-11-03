package commerce

import (
	"errors"
	"testing"
)

func TestNormalizeSellerDisplayName(t *testing.T) {
	if got,err:=normalizeSellerDisplayName("  GlobalCart Studio  "); err!=nil || got!="GlobalCart Studio" {
		t.Fatalf("normalize display name = %q, %v",got,err)
	}
	for _,value:=range []string{"","   ","bad\x00name",string(make([]byte,1)),string(make([]rune,201))} {
		if _,err:=normalizeSellerDisplayName(value); !errors.Is(err,ErrInvalidSellerProfileInput) {
			t.Fatalf("invalid display name %q error=%v",value,err)
		}
	}
}
