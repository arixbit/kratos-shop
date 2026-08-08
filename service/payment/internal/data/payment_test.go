package data_test

import (
	"payment/internal/data"
	"payment/internal/domain"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Payment", func() {
	It("Create/Get/MarkPaid", func() {
		repo := data.NewPaymentRepo(Db, nil)

		pay, err := repo.Create(ctx, &domain.Payment{
			PaymentNo: "PAY-TEST-001",
			OrderSn:   "ORDER-TEST-001",
			UserID:    1,
			Amount:    100,
			Channel:   "mock",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(pay.ID).NotTo(BeZero())
		Expect(pay.Status).To(Equal(1))

		got, err := repo.GetByPaymentNo(ctx, "PAY-TEST-001")
		Expect(err).NotTo(HaveOccurred())
		Expect(got.OrderSn).To(Equal("ORDER-TEST-001"))

		changed, err := repo.MarkPaid(ctx, "PAY-TEST-001", "TRADE-001")
		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(BeTrue())

		changed, err = repo.MarkPaid(ctx, "PAY-TEST-001", "TRADE-001")
		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(BeFalse())

		got, err = repo.GetByPaymentNo(ctx, "PAY-TEST-001")
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Status).To(Equal(2))
		Expect(got.TradeNo).To(Equal("TRADE-001"))
	})
})
