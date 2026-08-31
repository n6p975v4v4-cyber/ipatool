package cmd

import (
	"errors"

	"github.com/majd/ipatool/v2/pkg/appstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Purchase command", func() {
	It("accepts an app ID selector", func() {
		cmd := purchaseCmd()

		Expect(cmd.Flags().Set("app-id", "123")).To(Succeed())

		appID, err := cmd.Flags().GetInt64("app-id")
		Expect(err).ToNot(HaveOccurred())
		Expect(appID).To(Equal(int64(123)))
	})

	It("requires an app ID or bundle identifier", func() {
		cmd := purchaseCmd()

		err := cmd.RunE(cmd, nil)
		Expect(err).To(MatchError("either the app ID or the bundle identifier must be specified"))
	})

	It("purchases directly by app ID without looking up a bundle identifier", func() {
		lookupCalled := false

		app, err := resolvePurchaseApp(123, "", func() (appstore.LookupOutput, error) {
			lookupCalled = true

			return appstore.LookupOutput{}, nil
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(lookupCalled).To(BeFalse())
		Expect(app).To(Equal(appstore.App{ID: 123}))
	})

	It("uses the bundle identifier lookup when both selectors are supplied", func() {
		lookupErr := errors.New("lookup failed")
		app, err := resolvePurchaseApp(123, "com.example.app", func() (appstore.LookupOutput, error) {
			return appstore.LookupOutput{}, lookupErr
		})

		Expect(err).To(MatchError(lookupErr))
		Expect(app).To(Equal(appstore.App{}))
	})

	It("uses the bundle identifier lookup result when supplied", func() {
		expectedApp := appstore.App{ID: 456, BundleID: "com.example.app"}

		app, err := resolvePurchaseApp(123, "com.example.app", func() (appstore.LookupOutput, error) {
			return appstore.LookupOutput{App: expectedApp}, nil
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(app).To(Equal(expectedApp))
	})
})
