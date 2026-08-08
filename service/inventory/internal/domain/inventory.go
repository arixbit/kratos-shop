package domain

type Inventory struct {
	ID        int64
	SkuID     int64
	Inventory int64
	Locked    int64
	Version   int64
}

type InventoryLock struct {
	ID      int64
	OrderSn string
	SkuID   int64
	Num     int32
	Status  int
}

type InventoryFlow struct {
	ID      int64
	OrderSn string
	SkuID   int64
	Change  int64
	Type    string
}

type SkuItem struct {
	SkuID int64
	Num   int32
}
