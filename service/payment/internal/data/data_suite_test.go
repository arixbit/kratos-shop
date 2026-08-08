package data_test

import (
	"context"
	"testing"

	"payment/internal/conf"
	"payment/internal/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

func TestData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "payment data test")
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
			Source: "postgres://postgres:root@127.0.0.1:5432/shop_payment?sslmode=disable&TimeZone=Asia/Shanghai",
		},
	}
	db := data.NewDB(config)
	d, _, err := data.NewData(config, nil, db, nil)
	Expect(err).NotTo(HaveOccurred())
	Db = d
	rawDB = db
	Expect(rawDB.Exec("TRUNCATE payments RESTART IDENTITY CASCADE").Error).NotTo(HaveOccurred())
})
