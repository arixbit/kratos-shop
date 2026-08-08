package data_test

import (
	"context"
	"testing"
	"time"

	"inventory/internal/conf"
	"inventory/internal/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "inventory data test")
}

var (
	ctx   = context.Background()
	Db    *data.Data
	rawDB *gorm.DB
)

var _ = BeforeSuite(func() {
	config := &conf.Data{
		Database: &conf.Data_Database{
			Driver: "postgres",
			Source: "postgres://postgres:root@127.0.0.1:5432/shop_inventory?sslmode=disable&TimeZone=Asia/Shanghai",
		},
		Redis: &conf.Data_Redis{
			Addr:         "127.0.0.1:6379",
			Password:     "root",
			Db:           0,
			DialTimeout:  durationpb.New(time.Second),
			ReadTimeout:  durationpb.New(time.Second),
			WriteTimeout: durationpb.New(time.Second),
		},
	}
	db := data.NewDB(config)
	rdb := data.NewRedis(config)
	Expect(rdb.FlushDB(ctx).Err()).NotTo(HaveOccurred())

	d, _, err := data.NewData(config, nil, db, rdb)
	Expect(err).NotTo(HaveOccurred())
	Db = d
	rawDB = db

	Expect(rawDB.Exec("TRUNCATE inventories, inventory_locks, inventory_flows, consumed_event RESTART IDENTITY CASCADE").Error).NotTo(HaveOccurred())
	Expect(rawDB.Create(&data.Inventory{SkuID: 1, Inventory: 10, Locked: 0}).Error).NotTo(HaveOccurred())
	Expect(rawDB.Create(&data.Inventory{SkuID: 2, Inventory: 30, Locked: 0}).Error).NotTo(HaveOccurred())
	Expect(rawDB.Create(&data.Inventory{SkuID: 3, Inventory: 60, Locked: 0}).Error).NotTo(HaveOccurred())
	Expect(rawDB.Create(&data.Inventory{SkuID: 4, Inventory: 40, Locked: 0}).Error).NotTo(HaveOccurred())
})
