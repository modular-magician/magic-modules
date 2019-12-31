package google

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGoogleSQLCaCerts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGoogleSQLCaCertsRead,

		Schema: map[string]*schema.Schema{
			"active_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certs": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cert": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"common_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expiration_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sha1_fingerprint": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"instance": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"project": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"instance_self_link": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
		},
	}
}

func dataSourceGoogleSQLCaCertsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	var project, instance string
	if v, ok := d.GetOk("instance"); ok {
		p, err := getProject(d, config)
		if err != nil {
			return err
		}
		project = p
		instance = v.(string)
	} else if selfLink, ok := d.GetOk("instance_self_link"); ok {
		fv, err := parseProjectFieldValue("instances", selfLink.(string), "project", d, config, false)
		if err != nil {
			return err
		}
		project = fv.Project
		instance = fv.Name
		d.Set("instance_self_link", selfLink)
	} else {
		return fmt.Errorf("one of instance or instance_self_link must be set")
	}

	log.Printf("[DEBUG] Fetching CA certs from instance %s", instance)

	response, err := config.clientSqlAdmin.Instances.ListServerCas(project, instance).Do()
	if err != nil {
		return fmt.Errorf("error retrieving CA certs: %s", err)
	}

	log.Printf("[DEBUG] Fetched CA certs from instance %s", instance)

	d.Set("project", project)
	d.Set("instance", instance)
	d.Set("certs", flattenServerCaCerts(response.Certs))
	d.Set("active_version", response.ActiveVersion)
	d.SetId(fmt.Sprintf("projects/%s/instance/%s", project, instance))

	return nil
}
