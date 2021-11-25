package passwordrotation_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Password Rotation", func() {
	It("should reencrypt the DB when keys are rotated", func() {
		serviceBroker := helpers.CreateBroker(helpers.BrokerWithPrefix("csb-rotation"))
		defer serviceBroker.Delete()

		By("creating a service")
		serviceInstance := helpers.CreateServiceFromBroker("csb-azure-postgresql", "small", serviceBroker.Name)
		defer serviceInstance.Delete()

		By("getting current passwords")
		encryptionPasswords := helpers.GetBrokerEncryptionEnv(serviceBroker.Name)

		By("rotating the keys")
		Expect(encryptionPasswords.EncryptionEnabled).To(BeTrue())
		Expect(encryptionPasswords.EncryptionPasswords).To(HaveLen(1))

		oldPass := encryptionPasswords.EncryptionPasswords[0]
		oldPass.Primary = false
		newPass := helpers.EncryptionPassword{
			Password: helpers.Password{
				Secret: "someVerySecretPa88wOrd",
			},
			Label:   "second-password",
			Primary: true,
		}
		helpers.SetBrokerEncryptionEnv(serviceBroker.Name, helpers.BrokerEnvVars{
			EncryptionEnabled: encryptionPasswords.EncryptionEnabled,
			EncryptionPasswords: helpers.EncryptionPasswords{
				oldPass,
				newPass,
			},
		})

		By("pushing the unstarted app")
		app := helpers.AppPushUnstarted(apps.PostgreSQL)
		defer helpers.AppDelete(app)

		By("creating a binding")
		serviceInstance.Bind(app)

		By("starting the app")
		helpers.AppStart(app)

		By("restarting the broker with new keys only")
		helpers.SetBrokerEncryptionEnv(serviceBroker.Name, helpers.BrokerEnvVars{
			EncryptionEnabled: encryptionPasswords.EncryptionEnabled,
			EncryptionPasswords: helpers.EncryptionPasswords{
				newPass,
			},
		})
	})
})
