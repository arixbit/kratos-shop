package data_test

import (
	"cart/internal/conf"
	"cart/internal/data"
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// 测试 data 方法
func TestData(t *testing.T) {
	//  Ginkgo 测试通过调用 Fail(description string) 功能来表示失败
	// 使用 RegisterFailHandler 将此函数传递给 Gomega 。这是 Ginkgo 和 Gomega 之间的唯一连接点
	RegisterFailHandler(Fail)
	// 通知 Ginkgo 启动测试套件。如果您的任何 specs 失败，Ginkgo 将自动使 testing.T 失败。
	RunSpecs(t, "biz data test user")
}

var Db *data.Data       // 用于测试的 data
var ctx context.Context // 上下文

// initialize  AutoMigrate gorm自动建表
func initialize(db *gorm.DB) error {
	err := db.AutoMigrate(
		&data.ShopCart{},
	)
	return errors.WithStack(err)
}

// ginkgo 使用 BeforeEach 为您的 Specs 设置状态
var _ = BeforeSuite(func() {
	config := &conf.Data{
		Database: &conf.Data_Database{
			Driver: "postgres",
			Source: "postgres://postgres:root@127.0.0.1:5432/shop_cart?sslmode=disable&TimeZone=Asia/Shanghai",
		},
	}
	db := data.NewDB(config)
	postgresDb, _, err := data.NewData(config, nil, db, nil)
	Expect(err).NotTo(HaveOccurred())
	Db = postgresDb
	err = initialize(db)
	Expect(err).NotTo(HaveOccurred())
})
