package cmd

import (
	"errors"
	"time"

	"github.com/avast/retry-go"
	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/spf13/cobra"
)

// nolint:wrapcheck
func purchaseCmd() *cobra.Command {
	var (
		appID    int64
		bundleID string
	)

	cmd := &cobra.Command{
		Use:   "purchase",
		Short: "Obtain a license for the app from the App Store",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appID == 0 && bundleID == "" {
				return errors.New("either the app ID or the bundle identifier must be specified")
			}

			var lastErr error
			var acc appstore.Account

			return retry.Do(func() error {
				infoResult, err := dependencies.AppStore.AccountInfo()
				if err != nil {
					return err
				}

				acc = infoResult.Account

				if errors.Is(lastErr, appstore.ErrPasswordTokenExpired) {
					loginResult, err := dependencies.AppStore.Login(appstore.LoginInput{
						Email:    acc.Email,
						Password: acc.Password,
					})
					if err != nil {
						return err
					}

					acc = loginResult.Account
				}

				app, err := resolvePurchaseApp(appID, bundleID, func() (appstore.LookupOutput, error) {
					return dependencies.AppStore.Lookup(appstore.LookupInput{Account: acc, BundleID: bundleID})
				})
				if err != nil {
					return err
				}

				err = dependencies.AppStore.Purchase(appstore.PurchaseInput{Account: acc, App: app})
				if err != nil && !errors.Is(err, appstore.ErrLicenseAlreadyExists) {
					return err
				}

				dependencies.Logger.Log().
					Bool("alreadyOwned", errors.Is(err, appstore.ErrLicenseAlreadyExists)).
					Bool("success", true).
					Send()

				return nil
			},
				retry.LastErrorOnly(true),
				retry.DelayType(retry.FixedDelay),
				retry.Delay(time.Millisecond),
				retry.Attempts(2),
				retry.RetryIf(func(err error) bool {
					lastErr = err

					return errors.Is(err, appstore.ErrPasswordTokenExpired)
				}),
			)
		},
	}

	cmd.Flags().Int64VarP(&appID, "app-id", "i", 0, "ID of the target app (required)")
	cmd.Flags().StringVarP(&bundleID, "bundle-identifier", "b", "", "The bundle identifier of the target app (overrides the app ID)")

	return cmd
}

func resolvePurchaseApp(appID int64, bundleID string, lookup func() (appstore.LookupOutput, error)) (appstore.App, error) {
	app := appstore.App{ID: appID}
	if bundleID == "" {
		return app, nil
	}

	lookupResult, err := lookup()
	if err != nil {
		return appstore.App{}, err
	}

	return lookupResult.App, nil
}
