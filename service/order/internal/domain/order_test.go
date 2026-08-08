package domain

import "testing"

func TestCartItemListFindById(t *testing.T) {
	list := CartItemList{
		{CartId: 1, SkuId: 11},
		{CartId: 2, SkuId: 22},
	}
	if got := list.FindById(2); got == nil || got.SkuId != 22 {
		t.Fatalf("FindById(2) = %+v, want sku 22", got)
	}
	if got := list.FindById(99); got != nil {
		t.Fatalf("FindById(99) = %+v, want nil", got)
	}
}

func TestCartItemListGetSkuId(t *testing.T) {
	list := CartItemList{
		{CartId: 1, SkuId: 11},
		{CartId: 2, SkuId: 22},
	}
	ids := list.GetSkuId()
	if len(ids) != 2 || ids[0] != 11 || ids[1] != 22 {
		t.Fatalf("GetSkuId() = %v, want [11 22]", ids)
	}
}
