package data_test

import (
	"inventory/internal/data"
	"inventory/internal/domain"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inventory", func() {
	It("Query/TryLock/ConfirmDeduct/Release", func() {
		repo := data.NewInventoryRepo(Db, nil)

		list, err := repo.Query(ctx, []int64{1})
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(1))
		Expect(list[0].Inventory).To(Equal(int64(10)))

		Expect(repo.TryLock(ctx, "T1", []*domain.SkuItem{{SkuID: 1, Num: 3}})).NotTo(HaveOccurred())
		list, err = repo.Query(ctx, []int64{1})
		Expect(err).NotTo(HaveOccurred())
		Expect(list[0].Locked).To(Equal(int64(3)))

		Expect(repo.ConfirmDeduct(ctx, "T1", []*domain.SkuItem{{SkuID: 1, Num: 3}})).NotTo(HaveOccurred())
		list, err = repo.Query(ctx, []int64{1})
		Expect(err).NotTo(HaveOccurred())
		Expect(list[0].Inventory).To(Equal(int64(7)))
		Expect(list[0].Locked).To(Equal(int64(0)))

		Expect(repo.TryLock(ctx, "T2", []*domain.SkuItem{{SkuID: 1, Num: 2}})).NotTo(HaveOccurred())
		Expect(repo.Release(ctx, "T2", []*domain.SkuItem{{SkuID: 1, Num: 2}})).NotTo(HaveOccurred())
		list, err = repo.Query(ctx, []int64{1})
		Expect(err).NotTo(HaveOccurred())
		Expect(list[0].Locked).To(Equal(int64(0)))
	})

	It("TryLock insufficient", func() {
		repo := data.NewInventoryRepo(Db, nil)
		err := repo.TryLock(ctx, "T3", []*domain.SkuItem{{SkuID: 1, Num: 100}})
		Expect(err).To(HaveOccurred())
	})
})
