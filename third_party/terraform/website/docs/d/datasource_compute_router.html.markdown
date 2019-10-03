---
layout: "google"
page_title: "Google: google_compute_router"
sidebar_current: "docs-google-datasource-compute-router"
description: |-
  Get a Cloud Router within GCE.
---

# google\_compute\_router

Get a router within GCE from its name.

## Example Usage

```hcl
data "google_compute_router" "my-router" {
  name   = "default-us-east1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the router. One of `name` or `self_link`
must be specified.

* `project` - (Optional) The ID of the project in which the resource belongs. If it
    is not provided, the provider project is used.

* `region` - (Optional) The region this router has been created in. If
    unspecified, this defaults to the region configured in the provider.

    
## Attributes Reference

See [google_compute_router](https://www.terraform.io/docs/providers/google/r/compute_router.html) resource for details of the available attributes.
