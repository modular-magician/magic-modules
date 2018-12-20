package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccStorageTransferJob_basic(t *testing.T) {
	t.Parallel()

	testDataSourceId := acctest.RandString(10)
	testDataSinkId := acctest.RandString(10)
	testTransferJobId := acctest.RandString(10)
	testUpdatedDataSourceId := acctest.RandString(10)
	testUpdatedDataSinkId := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStorageTransferJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageTransferJob_basic(getTestProjectFromEnv(), testDataSourceId, testDataSinkId, testTransferJobId),
			},
			{
				ResourceName:      "google_storage_transfer_job.transfer_job",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStorageTransferJob_basic(getTestProjectFromEnv(), testDataSourceId, testUpdatedDataSinkId, testTransferJobId),
				Check:  resource.TestCheckResourceAttr("google_storage_transfer_job.transfer_job.transfer_spec.0.gcs_data_sink.0", "bucket_name", testUpdatedDataSinkId),
			},
			{
				ResourceName:      "google_storage_transfer_job.transfer_job",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStorageTransferJob_basic(getTestProjectFromEnv(), testUpdatedDataSourceId, testUpdatedDataSinkId, testTransferJobId),
				Check:  resource.TestCheckResourceAttr("google_storage_transfer_job.transfer_job.transfer_spec.0.gcs_data_source.0", "bucket_name", testUpdatedDataSinkId),
			},
			{
				ResourceName:      "google_storage_transfer_job.transfer_job",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStorageTransferJob_basic(project string, dataSourceId string, dataSinkId string, transferJobId string) string {
	return fmt.Sprintf(`
data "google_storage_transfer_project_service_account" "default" {
  project       = "%s"
}

resource "google_storage_bucket" "data_source" {
  name          = "test-data-source-bucket-%s"
  project       = "%s"
  force_destroy = true
}

resource "google_storage_bucket_iam_member" "data_source" {
  bucket        = "${google_storage_bucket.data_source.name}"
  role          = "roles/storage.admin"
  member        = "serviceAccount:${data.google_storage_transfer_project_service_account.default.email}"

  depends_on    = [
    "google_storage_bucket.data_source",
    "data.google_storage_transfer_project_service_account.default"
  ]
}

resource "google_storage_bucket" "data_sink" {
  name          = "test-data-sink-bucket-%s"
  project       = "%s"
  force_destroy = true
}

resource "google_storage_bucket_iam_member" "data_sink" {
  bucket        = "${google_storage_bucket.data_sink.name}"
  role          = "roles/storage.admin"
  member        = "serviceAccount:${data.google_storage_transfer_project_service_account.default.email}"

  depends_on    = [
    "google_storage_bucket.data_sink",
    "data.google_storage_transfer_project_service_account.default"
  ]
}

resource "google_storage_transfer_job" "transfer_job" {
	description	= "transfer-job-%s"
	project     = "%s"

	transfer_spec {
		gcs_data_source {
			bucket_name = "${google_storage_bucket.data_source.name}"
		}
		gcs_data_sink {
			bucket_name = "${google_storage_bucket.data_sink.name}"
		}
	}

	schedule {
		schedule_start_date {
			year	= 2018
			month	= 10
			day		= 1
		}
		schedule_end_date {
			year	= 2019
			month	= 10
			day		= 1
		}
		start_time_of_day {
			hours	= 0
			minutes	= 30
			seconds	= 0
			nanos	= 0
		}
	}

	depends_on = [
		"google_storage_bucket_iam_member.data_source",
		"google_storage_bucket_iam_member.data_sink",
	]
}
`, project, dataSourceId, project, dataSinkId, project, transferJobId, project)
}

func testAccStorageTransferJobDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "google_storage_transfer_job" {
			continue
		}

		rs_attr := rs.Primary.Attributes
		name, ok := rs_attr["name"]
		if !ok {
			return fmt.Errorf("No name set")
		}

		project, err := getTestProject(rs.Primary, config)
		if err != nil {
			return err
		}

		res, err := config.clientStorageTransfer.TransferJobs.Get(name).ProjectId(project).Do()
		if res.Status != "DELETED" {
			return fmt.Errorf("Transfer Job not set to DELETED")
		}
		if err != nil {
			return fmt.Errorf("Transfer Job does not exist, should exist and be DELETED")
		}
	}

	return nil
}
