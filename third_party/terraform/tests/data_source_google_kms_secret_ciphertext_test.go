package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccKmsSecretCiphertext_basic(t *testing.T) {
	t.Parallel()

	projectOrg := getTestOrgFromEnv(t)
	projectBillingAccount := getTestBillingAccountFromEnv(t)

	projectId := "terraform-" + acctest.RandString(10)
	keyRingName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	cryptoKeyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	plaintext := fmt.Sprintf("secret-%s", acctest.RandString(10))

	// The first test creates resources needed to encrypt plaintext and produces the ciphertext to assert the data source with
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testGoogleKmsCryptoKey_basic(projectId, projectOrg, projectBillingAccount, keyRingName, cryptoKeyName),
				Check: func(s *terraform.State) error {
					ciphertext, cryptoKeyId, err := testAccEncryptSecretDataWithCryptoKey(s, "google_kms_crypto_key.crypto_key", plaintext)

					if err != nil {
						return err
					}

					// The second test asserts that the data source has the correct ciphertext, given the plaintext
					resource.Test(t, resource.TestCase{
						PreCheck:  func() { testAccPreCheck(t) },
						Providers: testAccProviders,
						Steps: []resource.TestStep{
							{
								Config: testGoogleKmsSecretCiphertext_datasource(cryptoKeyId.terraformId(), plaintext),
								Check:  resource.TestCheckResourceAttr("data.google_kms_secret_ciphertext.acceptance", "ciphertext", ciphertext),
							},
						},
					})

					return nil
				},
			},
		},
	})
}

func testGoogleKmsSecretCiphertext_datasource(cryptoKeyTerraformId, plaintext string) string {
	return fmt.Sprintf(`
data "google_kms_secret_ciphertext" "acceptance" {
	crypto_key = "%s"
	plaintext = "%s"
}
	`, cryptoKeyTerraformId, plaintext)
}
