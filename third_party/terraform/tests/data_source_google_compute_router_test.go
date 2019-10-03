package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceComputeRouter(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeRouterConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDataSourceStateMatchesResourceStateWithIgnores(
						"data.google_compute_router.myrouter",
						"google_compute_router.foobar",
						map[string]struct{}{
							"name": {},
							"bgp": {},
						},
					),
				),
			},
		},
	})
}

func testAccDataSourceComputeRouterConfig() string {
	return fmt.Sprintf(`
		resource "google_compute_network" "foobar" {
			name = "router-test-%s"
			auto_create_subnetworks = false
		}
		resource "google_compute_subnetwork" "foobar" {
			name = "router-test-subnetwork-%s"
			network = "${google_compute_network.foobar.self_link}"
			ip_cidr_range = "10.0.0.0/16"
		}
		resource "google_compute_router" "foobar" {
			name = "router-test-%s"
			region = "${google_compute_subnetwork.foobar.region}"
			network = "${google_compute_network.foobar.name}"
			bgp {
				asn = 64514
			}
		}

data "google_compute_router" "myrouter" {
	name     = "${google_compute_router.foobar.name}"
}
`, acctest.RandString(10), acctest.RandString(10), acctest.RandString(10))
}
