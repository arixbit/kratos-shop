package domain

type Goods struct {
	ID              int64
	CategoryID      int32
	BrandsID        int32
	TypeID          int64
	Name            string
	NameAlias       string
	GoodsSn         string
	GoodsTags       string
	MarketPrice     int64
	GoodsBrief      string
	GoodsFrontImage string
	GoodsImages     []string
	OnSale          bool
	ShipFree        bool
	ShipID          int32
	IsNew           bool
	IsHot           bool
	ClickNum        int64
	SoldNum         int64
	FavNum          int64
	Sku             []*GoodsSku
}

type GoodsInfoResponse struct {
	GoodsID int64
}

type GoodsListResponse struct {
	Total int64
	List  []*Goods
}

type GoodsImage struct {
	ID       int64
	GoodsID  int64
	Link     string
	Position int32
	IsMaster bool
}
