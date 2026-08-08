package data_test

import (
	"cart/internal/biz"
	"cart/internal/data"
	"cart/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cart", func() {
	var ro biz.CartRepo
	BeforeEach(func() {
		ro = data.NewCartRepo(Db, nil)
		// 清掉历史测试数据，保证用例可重复执行
		db, err := gorm.Open(postgres.Open("postgres://postgres:root@127.0.0.1:5432/shop_cart?sslmode=disable"), &gorm.Config{})
		Ω(err).ShouldNot(HaveOccurred())
		Ω(db.Exec("DELETE FROM shop_carts WHERE user_id = ? AND sku_id = ?", 999, 999).Error).ShouldNot(HaveOccurred())
	})
	// 设置 It 块来添加单个规格
	It("CreateCart", func() {
		cartData := domain.ShopCart{
			UserId:     999,
			GoodsId:    999,
			SkuId:      999,
			GoodsPrice: 1000,
			GoodsNum:   10,
			GoodsSn:    "20232232231",
			GoodsName:  "Mate 40 Pro",
			IsSelect:   true,
		}
		c, err := ro.Create(ctx, &cartData)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(c.UserId).Should(Equal(int64(999)))
		Ω(c.GoodsNum).Should(Equal(int32(10)))

		// 二次验证创建相同商品的数据，只增加商品数量
		cartData2 := domain.ShopCart{
			UserId:     999,
			GoodsId:    999,
			SkuId:      999,
			GoodsPrice: 1000,
			GoodsNum:   10,
			GoodsSn:    "20232232231",
			GoodsName:  "Mate 40 Pro",
			IsSelect:   true,
		}
		c2, err := ro.Create(ctx, &cartData2)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(c2.UserId).Should(Equal(int64(999)))
		Ω(c2.GoodsNum).Should(Equal(int32(20)))
	})

})
