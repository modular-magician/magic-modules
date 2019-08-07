---
layout: "google"
page_title: "Google: google_app_engine_application"
sidebar_current: "docs-google-app-engine-application"
description: |-
 Allows management of an App Engine application.
---

# google_app_engine_application

Allows creation and management of an App Engine application.

~> App Engine applications cannot be deleted once they're created; you have to delete the
   entire project to delete the application. Terraform will report the application has been
   successfully deleted; this is a limitation of Terraform, and will go away in the future.
   Terraform is not able to delete App Engine applications.

## Example Usage

```hcl
resource "google_project" "my_project" {
  name       = "My Project"
  project_id = "your-project-id"
  org_id     = "1234567"
}

resource "google_app_engine_standard_app_version" "app" {
  version_id = "v2"
  service = "default"
  runtime = "nodejs10"
  entrypoint {
    shell = "node app.js"
  }
  handlers {
    security_level = "SECURE_OPTIONAL"
    static_files {
      path = "stylesheets/"
      upload_path_regex = "stylesheets/.*"
    }
    url_regex = "/stylesheets/(.*)"
  }
  deployment {
    zip {
      source_url = "https://storage.googleapis.com/sample/hello-world.zip"
    }
    
  } 
  env_variables = {
    port = "8080"
      } 
}
```

## Argument Reference

The following arguments are supported:

* `version_id` - (Required) The name of the version e.g. **v1**

* `service` - (Required) The name of the service in google app engine e.g. **default**

* `runtime` - (Required) Runtime for this version e.g. **nodejs10**.

* `entry_point` - (Optional) The serving status of the app.

## Import

Applications can be imported using the ID of the project the application belongs to, e.g.

```
$ terraform import google_app_engine_standard_app_version.app your-project-id
```
