package data_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
	"user/internal/biz"
	"user/internal/data"
	"user/internal/testdata"
)

var _ = Describe("User", func() {
	var ro biz.UserRepo
	var uD *biz.User
	BeforeEach(func() {
		ro = data.NewUserRepo(Db, nil)
		// 这里你可以不引入外部组装好的数据，可以在这里直接写
		uD = testdata.User()
		// 每次使用随机手机号，避免与 Mock 数据或历史测试数据冲突
		uD.Mobile = fmt.Sprintf("138%08d", time.Now().UnixNano()%100000000)
	})
	// 设置 It 块来添加单个规格
	It("CreateUser", func() {
		u, err := ro.CreateUser(ctx, uD)
		Ω(err).ShouldNot(HaveOccurred())
		// 组装的数据 mobile 为 13509876789
		Ω(u.Mobile).Should(Equal(uD.Mobile)) // 手机号应该为创建的时候写入的手机号
	})
	// 设置 It 块来添加单个规格
	It("ListUser", func() {
		_, err := ro.CreateUser(ctx, uD)
		Ω(err).ShouldNot(HaveOccurred())
		user, total, err := ro.ListUser(ctx, 1, 10)
		Ω(err).ShouldNot(HaveOccurred()) // 获取列表不应该出现错误
		Ω(user).ShouldNot(BeEmpty())     // 结果不应该为空
		Ω(total).Should(BeNumerically(">=", 1))
		var found bool
		for _, u := range user {
			if u.Mobile == uD.Mobile {
				found = true
				break
			}
		}
		Ω(found).Should(BeTrue())
	})

	// 设置 It 块来添加单个规格
	It("UpdateUser", func() {
		uD.Mobile = fmt.Sprintf("139%08d", time.Now().UnixNano()%100000000)
		created, err := ro.CreateUser(ctx, uD)
		Ω(err).ShouldNot(HaveOccurred())
		uD.ID = created.ID
		birthDay := time.Unix(int64(693646426), 0)
		uD.NickName = "gyl"
		uD.Birthday = &birthDay
		uD.Gender = "female"
		user, err := ro.UpdateUser(ctx, uD)
		Ω(err).ShouldNot(HaveOccurred()) // 更新不应该出现错误
		Ω(user).Should(BeTrue())         // 结果应该为 true
	})

	It("CheckPassword", func() {
		p1 := "admin"
		encryptedPassword := "$pbkdf2-sha512$5p7doUNIS9I5mvhA$b18171ff58b04c02ed70ea4f39bda036029c107294bce83301a02fb53a1bcae0"
		password, err := ro.CheckPassword(ctx, p1, encryptedPassword)
		Ω(err).ShouldNot(HaveOccurred()) // 密码验证通过
		Ω(password).Should(BeTrue())     // 结果应该为true

		encryptedPassword1 := "$pbkdf2-sha512$5p7doUNIS9I5mvhA$b18171ff58b04c02ed70ea4f39"
		password1, err := ro.CheckPassword(ctx, p1, encryptedPassword1)
		if err != nil {
			return
		}
		Ω(err).ShouldNot(HaveOccurred())
		Ω(password1).Should(BeFalse()) // 密码验证不通过
	})
})
