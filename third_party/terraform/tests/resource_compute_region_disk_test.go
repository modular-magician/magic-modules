package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	computeBeta "google.golang.org/api/compute/v0.beta"
)

func TestAccComputeRegionDisk_basic(t *testing.T) {
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRegionDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeRegionDisk_basic(diskName, "self_link"),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccComputeRegionDisk_basic(diskName, "name"),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComputeRegionDisk_basicUpdate(t *testing.T) {
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRegionDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeRegionDisk_basic(diskName, "self_link"),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccComputeRegionDisk_basicUpdated(diskName, "self_link"),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComputeRegionDisk_encryption(t *testing.T) {
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk computeBeta.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRegionDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeRegionDisk_encryption(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeRegionDiskExists(
						"google_compute_region_disk.regiondisk", &disk),
					testAccCheckRegionDiskEncryptionKey(
						"google_compute_region_disk.regiondisk", &disk),
				),
			},
		},
	})
}

func TestAccComputeRegionDisk_deleteDetach(t *testing.T) {
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	regionDiskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	regionDiskName2 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	instanceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRegionDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeRegionDisk_deleteDetach(instanceName, diskName, regionDiskName),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// this needs to be an additional step so we refresh and see the instance
			// listed as attached to the disk; the instance is created after the
			// disk. and the disk's properties aren't refreshed unless there's
			// another step
			{
				Config: testAccComputeRegionDisk_deleteDetach(instanceName, diskName, regionDiskName),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Change the disk name to destroy it, which detaches it from the instance
			{
				Config: testAccComputeRegionDisk_deleteDetach(instanceName, diskName, regionDiskName2),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Add the extra step like before
			{
				Config: testAccComputeRegionDisk_deleteDetach(instanceName, diskName, regionDiskName2),
			},
			{
				ResourceName:      "google_compute_region_disk.regiondisk",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckComputeRegionDiskExists(n string, disk *computeBeta.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := getTestProjectFromEnv()
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.clientComputeBeta.RegionDisks.Get(
			p, rs.Primary.Attributes["region"], rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("RegionDisk not found")
		}

		*disk = *found

		return nil
	}
}
func testAccCheckRegionDiskEncryptionKey(n string, disk *computeBeta.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		attr := rs.Primary.Attributes["disk_encryption_key.0.sha256"]
		if disk.DiskEncryptionKey == nil {
			return fmt.Errorf("RegionDisk %s has mismatched encryption key.\nTF State: %+v\nGCP State: <empty>", n, attr)
		} else if attr != disk.DiskEncryptionKey.Sha256 {
			return fmt.Errorf("RegionDisk %s has mismatched encryption key.\nTF State: %+v.\nGCP State: %+v",
				n, attr, disk.DiskEncryptionKey.Sha256)
		}
		return nil
	}
}

func testAccComputeRegionDisk_basic(diskName, refSelector string) string {
	return fmt.Sprintf(`
resource "google_compute_disk" "disk" {
	name = "%s"
	image = "debian-cloud/debian-9"
	size = 50
	type = "pd-ssd"
	zone = "us-central1-a"
}

resource "google_compute_snapshot" "snapdisk" {
	name = "%s"
	source_disk = "${google_compute_disk.disk.name}"
	zone = "us-central1-a"
}

resource "google_compute_region_disk" "regiondisk" {
	name = "%s"
	snapshot = "${google_compute_snapshot.snapdisk.%s}"
	type = "pd-ssd"
	region = "us-central1"

	replica_zones = ["us-central1-a", "us-central1-f"]
}`, diskName, diskName, diskName, refSelector)
}

func testAccComputeRegionDisk_basicUpdated(diskName, refSelector string) string {
	return fmt.Sprintf(`
resource "google_compute_disk" "disk" {
	name = "%s"
	image = "debian-cloud/debian-9"
	size = 50
	type = "pd-ssd"
	zone = "us-central1-a"
}

resource "google_compute_snapshot" "snapdisk" {
	name = "%s"
	source_disk = "${google_compute_disk.disk.name}"
	zone = "us-central1-a"
}

resource "google_compute_region_disk" "regiondisk" {
	name     = "%s"
	snapshot = "${google_compute_snapshot.snapdisk.%s}"
	type     = "pd-ssd"
	region   = "us-central1"

	replica_zones = ["us-central1-a", "us-central1-f"]

	size = 100
	labels = {
		my-label = "my-updated-label-value"
		a-new-label = "a-new-label-value"
	}

}`, diskName, diskName, diskName, refSelector)
}

func testAccComputeRegionDisk_encryption(diskName string) string {
	return fmt.Sprintf(`
resource "google_compute_disk" "disk" {
	name  = "%s"
	image = "debian-cloud/debian-9"
	size  = 50
	type  = "pd-ssd"
	zone  = "us-central1-a"
}

resource "google_compute_snapshot" "snapdisk" {
	name = "%s"
	zone = "us-central1-a"

	source_disk = "${google_compute_disk.disk.name}"
}

resource "google_compute_region_disk" "regiondisk" {
	name     = "%s"
	snapshot = "${google_compute_snapshot.snapdisk.self_link}"
	type     = "pd-ssd"
	region   = "us-central1"

	replica_zones = ["us-central1-a", "us-central1-f"]

	disk_encryption_key {
		raw_key = "SGVsbG8gZnJvbSBHb29nbGUgQ2xvdWQgUGxhdGZvcm0="
	}
}`, diskName, diskName, diskName)
}

func testAccComputeRegionDisk_deleteDetach(instanceName, diskName, regionDiskName string) string {
	return fmt.Sprintf(`
resource "google_compute_disk" "disk" {
	name = "%s"
	image = "debian-cloud/debian-9"
	size = 50
	type = "pd-ssd"
	zone = "us-central1-a"
}

resource "google_compute_snapshot" "snapdisk" {
	name = "%s"
	source_disk = "${google_compute_disk.disk.name}"
	zone = "us-central1-a"
}

resource "google_compute_region_disk" "regiondisk" {
	name = "%s"
	snapshot = "${google_compute_snapshot.snapdisk.self_link}"
	type = "pd-ssd"
	region = "us-central1"

	replica_zones = ["us-central1-a", "us-central1-f"]
}

resource "google_compute_instance" "inst" {
	name = "%s"
	machine_type = "n1-standard-1"
	zone = "us-central1-a"

	boot_disk {
		initialize_params {
			image = "debian-cloud/debian-9"
		}
	}

	attached_disk {
		source = "${google_compute_region_disk.regiondisk.self_link}"
	}

	network_interface {
		network = "default"
	}
}`, diskName, diskName, regionDiskName, instanceName)
}
