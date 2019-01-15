variable "ssl_policy" {
  type = "map"
}

variable "topic" {
  type = "map"
}

variable "subscription" {
  type = "map"
}

variable "instance_group_manager" {
  type = "map"
}

resource "google_compute_ssl_policy" "custom-ssl-policy" {
  name            = "${var.ssl_policy["name"]}"
  min_tls_version = "${var.ssl_policy["min_tls_version"]}"
  profile         = "${var.ssl_policy["profile"]}"
  custom_features = ["${var.ssl_policy["custom_feature"]}", "${var.ssl_policy["custom_feature2"]}"]
  project         = "${var.gcp_project_id}"
}

resource "google_pubsub_topic" "topic" {
  project = "${var.gcp_project_id}"
  name    = "${var.topic["name"]}"
}

resource "google_pubsub_subscription" "default" {
  project              = "${var.gcp_project_id}"
  name                 = "${var.subscription["name"]}"
  topic                = "${google_pubsub_topic.topic.name}"
  ack_deadline_seconds = "${var.subscription["ack_deadline_seconds"]}"
}

resource "google_compute_instance_group_manager" "gcp-inspec-igm" {
  project           = "${var.gcp_project_id}"
  zone              = "${var.gcp_zone}"
  name              = "${var.instance_group_manager["name"]}"
  instance_template = "${google_compute_instance_template.default.self_link}"
  base_instance_name        = "${var.instance_group_manager["base_instance_name"]}"
  target_pools = []
  target_size  = 0
  named_port {
    name = "${var.instance_group_manager["named_port_name"]}"
    port = "${var.instance_group_manager["named_port_port"]}"
  }
}