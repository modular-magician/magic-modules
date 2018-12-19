package google

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/storagetransfer/v1"
	"log"
	"strings"
	"time"
)

func resourceStorageTransferJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceStorageTransferJobCreate,
		Read:   resourceStorageTransferJobRead,
		Update: resourceStorageTransferJobUpdate,
		Delete: resourceStorageTransferJobDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStorageTransferJobStateImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"transfer_spec": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_conditions": objectConditionsSchema(),
						"transfer_options":  transferOptionsSchema(),
						"gcs_data_sink": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     gcsDataSchema(),
						},
						"gcs_data_source": &schema.Schema{
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							Elem:          gcsDataSchema(),
							ConflictsWith: []string{"transfer_spec.aws_s3_data_source", "transfer_spec.http_data_source"},
						},
						"aws_s3_data_source": &schema.Schema{
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							Elem:          awsS3DataSchema(),
							ConflictsWith: []string{"transfer_spec.gcs_data_source", "transfer_spec.http_data_source"},
						},
						"http_data_source": &schema.Schema{
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							Elem:          httpDataSchema(),
							ConflictsWith: []string{"transfer_spec.aws_s3_data_source", "transfer_spec.gcs_data_source"},
						},
					},
				},
			},
			"schedule": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule_start_date": &schema.Schema{
							Type:     schema.TypeList,
							Required: true,
							Optional: false,
							ForceNew: true,
							MaxItems: 1,
							Elem:     dateObjectSchema(),
						},
						"schedule_end_date": &schema.Schema{
							Type:     schema.TypeList,
							Required: false,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem:     dateObjectSchema(),
						},
						"start_time_of_day": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem:     timeObjectSchema(),
						},
					},
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ENABLED",
				ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED", "DELETED"}, false),
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modification_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func objectConditionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"min_time_elapsed_since_last_modification": &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateDuration(),
					Optional:     true,
				},
				"max_time_elapsed_since_last_modification": &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateDuration(),
					Optional:     true,
				},
				"include_prefixes": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						MaxItems: 1000,
						Type:     schema.TypeString,
					},
				},
				"exclude_prefixes": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						MaxItems: 1000,
						Type:     schema.TypeString,
					},
				},
			},
		},
	}
}

func transferOptionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"overwrite_objects_already_existing_in_sink": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"delete_objects_unique_in_sink": &schema.Schema{
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{"transfer_spec.transfer_options.delete_objects_from_source_after_transfer"},
				},
				"delete_objects_from_source_after_transfer": &schema.Schema{
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{"transfer_spec.transfer_options.delete_objects_unique_in_sink"},
				},
			},
		},
	}
}

func timeObjectSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"hours": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 24),
			},
			"minutes": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 59),
			},
			"seconds": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 60),
			},
			"nanos": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 999999999),
			},
		},
	}
}

func dateObjectSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"year": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 9999),
			},

			"month": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 12),
			},

			"day": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 31),
			},
		},
	}
}

func gcsDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket_name": &schema.Schema{
				Required: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func awsS3DataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket_name": &schema.Schema{
				Required: true,
				Type:     schema.TypeString,
			},
			"aws_access_key": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key_id": &schema.Schema{
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"secret_access_key": &schema.Schema{
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
	}
}

func httpDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"list_url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceStorageTransferJobCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	transferJob := &storagetransfer.TransferJob{
		Description:  d.Get("description").(string),
		ProjectId:    project,
		Status:       d.Get("status").(string),
		Schedule:     expandTransferSchedules(d.Get("schedule").([]interface{})),
		TransferSpec: expandTransferSpecs(d.Get("transfer_spec").([]interface{})),
	}

	var res *storagetransfer.TransferJob

	err = retry(func() error {
		res, err = config.clientStorageTransfer.TransferJobs.Create(transferJob).Do()
		return err
	})

	if err != nil {
		fmt.Printf("Error creating transfer job %v: %v", transferJob, err)
		return err
	}

	d.Set("name", res.Name)
	return resourceStorageTransferJobRead(d, meta)
}

func resourceStorageTransferJobRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	res, err := config.clientStorageTransfer.TransferJobs.Get(name).ProjectId(project).Do()
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Transfer Job %q", name))
	}
	log.Printf("[DEBUG] Read transfer job: %v in project: %v \n\n", res.Name, res.ProjectId)

	d.Set("project", res.ProjectId)
	d.Set("description", res.Description)
	d.Set("status", res.Status)
	d.Set("last_modification_time", res.LastModificationTime)
	d.Set("creation_time", res.CreationTime)
	d.Set("deletion_time", res.DeletionTime)

	err = d.Set("schedule", flattenTransferSchedules([]*storagetransfer.Schedule{res.Schedule}))
	if err != nil {
		return err
	}

	d.Set("transfer_spec", flattenTransferSpec(res.TransferSpec))
	if err != nil {
		return err
	}

	return resourceStorageTransferJobRead(d, meta)
}

func resourceStorageTransferJobUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	transferJob := &storagetransfer.TransferJob{}
	fieldMask := []string{}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			fieldMask = append(fieldMask, "description")
			transferJob.Description = v.(string)
		}
	}

	if d.HasChange("status") {
		if v, ok := d.GetOk("status"); ok {
			fieldMask = append(fieldMask, "status")
			transferJob.Status = v.(string)
		}
	}

	if d.HasChange("schedule") {
		if v, ok := d.GetOk("schedule"); ok {
			fieldMask = append(fieldMask, "schedule")
			transferJob.Schedule = expandTransferSchedules(v.([]interface{}))
		}
	}

	if d.HasChange("transfer_spec") {
		if v, ok := d.GetOk("transfer_spec"); ok {
			fieldMask = append(fieldMask, "transfer_spec")
			transferJob.TransferSpec = expandTransferSpecs(v.([]interface{}))
		}
	}

	updateRequest := &storagetransfer.UpdateTransferJobRequest{
		ProjectId:   project,
		TransferJob: transferJob,
	}

	updateRequest.UpdateTransferJobFieldMask = strings.Join(fieldMask, ",")

	res, err := config.clientStorageTransfer.TransferJobs.Patch(d.Get("name").(string), updateRequest).Do()
	if err != nil {
		return err
	}

	d.SetId(res.Name)
	return nil
}

func resourceStorageTransferJobDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	transferJobName := d.Get("name").(string)

	transferJob := &storagetransfer.TransferJob{
		Status: "DELETED",
	}

	fieldMask := "status"

	updateRequest := &storagetransfer.UpdateTransferJobRequest{
		ProjectId:   project,
		TransferJob: transferJob,
	}

	updateRequest.UpdateTransferJobFieldMask = fieldMask

	// Update transfer job with status set to DELETE
	log.Printf("[DEBUG] Setting status to DELETE for: %v\n\n", transferJobName)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := config.clientStorageTransfer.TransferJobs.Patch(transferJobName, updateRequest).Do()
		if err != nil {
			return resource.RetryableError(err)
		}
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
			return resource.RetryableError(gerr)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error deleting transfer job %v: %v\n\n", transferJob, err)
		return err
	}

	log.Printf("[DEBUG] Deleted transfer job %v\n\n", transferJob)

	return nil
}

func resourceStorageTransferJobStateImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		d.Set("project", parts[0])
		d.Set("name", fmt.Sprintf("transferJobs/%s", parts[1]))
	default:
		return nil, fmt.Errorf("Invalid transfer job specifier. Expecting {projectId}/{transferJobName}")
	}
	return []*schema.ResourceData{d}, nil
}

func expandDates(dates []interface{}) *storagetransfer.Date {
	if len(dates) == 0 || dates[0] == nil {
		return nil
	}

	date := dates[0].([]interface{})
	return &storagetransfer.Date{
		Day:   int64(extractFirstMapConfig(date)["day"].(int)),
		Month: int64(extractFirstMapConfig(date)["month"].(int)),
		Year:  int64(extractFirstMapConfig(date)["year"].(int)),
	}
}

func flattenDates(dates []*storagetransfer.Date) []map[string]interface{} {
	datesSchema := make([]map[string]interface{}, 0, len(dates))
	for _, date := range dates {
		datesSchema = append(datesSchema, map[string]interface{}{
			"year":  date.Year,
			"month": date.Month,
			"day":   date.Day,
		})
	}
	return datesSchema
}

func expandTimeOfDays(times []interface{}) *storagetransfer.TimeOfDay {
	if len(times) == 0 || times[0] == nil {
		return nil
	}

	time := times[0].([]interface{})
	return &storagetransfer.TimeOfDay{
		Hours:   int64(extractFirstMapConfig(time)["hours"].(int)),
		Minutes: int64(extractFirstMapConfig(time)["minutes"].(int)),
		Seconds: int64(extractFirstMapConfig(time)["seconds"].(int)),
		Nanos:   int64(extractFirstMapConfig(time)["nanos"].(int)),
	}
}

func flattenTimeOfDays(timeOfDays []*storagetransfer.TimeOfDay) []map[string]interface{} {
	timeOfDaysSchema := make([]map[string]interface{}, 0, len(timeOfDays))
	for _, timeOfDay := range timeOfDays {
		timeOfDaysSchema = append(timeOfDaysSchema, map[string]interface{}{
			"hours":   timeOfDay.Hours,
			"minutes": timeOfDay.Minutes,
			"seconds": timeOfDay.Seconds,
			"nanos":   timeOfDay.Nanos,
		})
	}
	return timeOfDaysSchema
}

func expandTransferSchedules(transferSchedules []interface{}) *storagetransfer.Schedule {
	if len(transferSchedules) == 0 || transferSchedules[0] == nil {
		return nil
	}

	schedule := transferSchedules[0].(map[string]interface{})
	return &storagetransfer.Schedule{
		ScheduleStartDate: expandDates([]interface{}{schedule["schedule_start_date"]}),
		ScheduleEndDate:   expandDates([]interface{}{schedule["schedule_end_date"]}),
		StartTimeOfDay:    expandTimeOfDays([]interface{}{schedule["start_time_of_day"]}),
	}
}

func flattenTransferSchedules(transferSchedules []*storagetransfer.Schedule) []map[string][]map[string]interface{} {
	transferSchedulesSchema := make([]map[string][]map[string]interface{}, 0, len(transferSchedules))
	for _, transferSchedule := range transferSchedules {
		schedule := map[string][]map[string]interface{}{
			"schedule_start_date": flattenDates([]*storagetransfer.Date{transferSchedule.ScheduleStartDate}),
		}

		if transferSchedule.ScheduleEndDate != nil {
			schedule["schedule_end_date"] = flattenDates([]*storagetransfer.Date{transferSchedule.ScheduleEndDate})
		}

		if transferSchedule.StartTimeOfDay != nil {
			schedule["start_time_of_day"] = flattenTimeOfDays([]*storagetransfer.TimeOfDay{transferSchedule.StartTimeOfDay})
		}

		transferSchedulesSchema = append(transferSchedulesSchema, schedule)
	}
	return transferSchedulesSchema
}

func expandGcsData(gcsDatas []interface{}) *storagetransfer.GcsData {
	if len(gcsDatas) == 0 || gcsDatas[0] == nil {
		return nil
	}

	gcsData := gcsDatas[0].(map[string]interface{})
	return &storagetransfer.GcsData{
		BucketName: gcsData["bucket_name"].(string),
	}
}

func flattenGcsData(gcsDatas []*storagetransfer.GcsData) []map[string]interface{} {
	datasSchema := make([]map[string]interface{}, 0, len(gcsDatas))
	for _, data := range gcsDatas {
		datasSchema = append(datasSchema, map[string]interface{}{
			"bucket_name": data.BucketName,
		})
	}
	return datasSchema
}

func expandAwsAccessKeys(awsAccessKeys []interface{}) *storagetransfer.AwsAccessKey {
	if len(awsAccessKeys) == 0 || awsAccessKeys[0] == nil {
		return nil
	}

	awsAccessKey := awsAccessKeys[0].(map[string]interface{})
	return &storagetransfer.AwsAccessKey{
		AccessKeyId:     awsAccessKey["access_key_id"].(string),
		SecretAccessKey: awsAccessKey["secret_access_key"].(string),
	}
}

func flattenAwsAccessKeys(awsAccessKeys []*storagetransfer.AwsAccessKey) []map[string]interface{} {
	datasSchema := make([]map[string]interface{}, 0, len(awsAccessKeys))
	for _, data := range awsAccessKeys {
		datasSchema = append(datasSchema, map[string]interface{}{
			"access_key_id":     data.AccessKeyId,
			"secret_access_key": data.SecretAccessKey,
		})
	}
	return datasSchema
}

func expandAwsS3Data(awsS3Datas []interface{}) *storagetransfer.AwsS3Data {
	if len(awsS3Datas) == 0 || awsS3Datas[0] == nil {
		return nil
	}

	awsS3Data := awsS3Datas[0].(map[string]interface{})
	return &storagetransfer.AwsS3Data{
		BucketName:   awsS3Data["bucket_name"].(string),
		AwsAccessKey: expandAwsAccessKeys(awsS3Data["aws_access_key"].([]interface{})),
	}
}

func flattenAwsS3Data(awsS3Datas []*storagetransfer.AwsS3Data) []map[string]interface{} {
	datasSchema := make([]map[string]interface{}, 0, len(awsS3Datas))
	for _, data := range awsS3Datas {
		datasSchema = append(datasSchema, map[string]interface{}{
			"bucket_name":    data.BucketName,
			"aws_access_key": data.AwsAccessKey,
		})
	}
	return datasSchema
}

func expandHttpData(httpDatas []interface{}) *storagetransfer.HttpData {
	if len(httpDatas) == 0 || httpDatas[0] == nil {
		return nil
	}

	httpData := httpDatas[0].(map[string]interface{})
	return &storagetransfer.HttpData{
		ListUrl: httpData["list_url"].(string),
	}
}

func flattenHttpData(httpDatas []*storagetransfer.HttpData) []map[string]interface{} {
	datasSchema := make([]map[string]interface{}, 0, len(httpDatas))
	for _, data := range httpDatas {
		datasSchema = append(datasSchema, map[string]interface{}{
			"list_url": data.ListUrl,
		})
	}
	return datasSchema
}

func expandObjectConditions(conditions []interface{}) *storagetransfer.ObjectConditions {
	if len(conditions) == 0 || conditions[0] == nil {
		return nil
	}

	condition := conditions[0].(map[string]interface{})
	return &storagetransfer.ObjectConditions{
		ExcludePrefixes:                     convertStringArr(condition["exclude_prefixes"].([]interface{})),
		IncludePrefixes:                     convertStringArr(condition["include_prefixes"].([]interface{})),
		MaxTimeElapsedSinceLastModification: condition["max_time_elapsed_since_last_modification"].(string),
		MinTimeElapsedSinceLastModification: condition["min_time_elapsed_since_last_modification"].(string),
	}
}

func flattenObjectConditions(conditions []*storagetransfer.ObjectConditions) []map[string]interface{} {
	datasSchema := make([]map[string]interface{}, 0, len(conditions))
	for _, data := range conditions {
		datasSchema = append(datasSchema, map[string]interface{}{
			"exclude_prefixes":                         data.ExcludePrefixes,
			"include_prefixes":                         data.IncludePrefixes,
			"max_time_elapsed_since_last_modification": data.MaxTimeElapsedSinceLastModification,
			"min_time_elapsed_since_last_modification": data.MinTimeElapsedSinceLastModification,
		})
	}
	return datasSchema
}

func expandTransferOptions(options []interface{}) *storagetransfer.TransferOptions {
	if len(options) == 0 || options[0] == nil {
		return nil
	}

	option := options[0].(map[string]interface{})
	return &storagetransfer.TransferOptions{
		DeleteObjectsFromSourceAfterTransfer:  option["delete_objects_from_source_after_transfer"].(bool),
		DeleteObjectsUniqueInSink:             option["delete_objects_unique_in_sink"].(bool),
		OverwriteObjectsAlreadyExistingInSink: option["overwrite_objects_already_existing_in_sink"].(bool),
	}
}

func flattenTransferOptions(options []*storagetransfer.TransferOptions) []map[string]interface{} {
	datasSchema := make([]map[string]interface{}, 0, len(options))
	for _, data := range options {
		datasSchema = append(datasSchema, map[string]interface{}{
			"delete_objects_from_source_after_transfer":  data.DeleteObjectsFromSourceAfterTransfer,
			"delete_objects_unique_in_sink":              data.DeleteObjectsUniqueInSink,
			"overwrite_objects_already_existing_in_sink": data.OverwriteObjectsAlreadyExistingInSink,
		})
	}
	return datasSchema
}

func expandTransferSpecs(transferSpecs []interface{}) *storagetransfer.TransferSpec {
	if len(transferSpecs) == 0 || transferSpecs[0] == nil {
		return nil
	}

	transferSpec := transferSpecs[0].(map[string]interface{})
	return &storagetransfer.TransferSpec{
		GcsDataSink:      expandGcsData(transferSpec["gcs_data_sink"].([]interface{})),
		ObjectConditions: expandObjectConditions(transferSpec["object_conditions"].([]interface{})),
		TransferOptions:  expandTransferOptions(transferSpec["transfer_options"].([]interface{})),
		GcsDataSource:    expandGcsData(transferSpec["gcs_data_source"].([]interface{})),
		AwsS3DataSource:  expandAwsS3Data(transferSpec["aws_s3_data_source"].([]interface{})),
		HttpDataSource:   expandHttpData(transferSpec["http_data_source"].([]interface{})),
	}
}

func flattenTransferSpec(transferSpec *storagetransfer.TransferSpec) []map[string][]map[string]interface{} {

	transferSpecSchema := map[string][]map[string]interface{}{
		"gcs_data_sink": flattenGcsData([]*storagetransfer.GcsData{transferSpec.GcsDataSink}),
	}

	if transferSpec.ObjectConditions != nil {
		transferSpecSchema["object_conditions"] = flattenObjectConditions([]*storagetransfer.ObjectConditions{transferSpec.ObjectConditions})
	}
	if transferSpec.TransferOptions != nil {
		transferSpecSchema["transfer_options"] = flattenTransferOptions([]*storagetransfer.TransferOptions{transferSpec.TransferOptions})
	}
	if transferSpec.GcsDataSource != nil {
		transferSpecSchema["gcs_data_source"] = flattenGcsData([]*storagetransfer.GcsData{transferSpec.GcsDataSource})
	} else if transferSpec.AwsS3DataSource != nil {
		transferSpecSchema["aws_s3_data_source"] = flattenAwsS3Data([]*storagetransfer.AwsS3Data{transferSpec.AwsS3DataSource})
	} else if transferSpec.HttpDataSource != nil {
		transferSpecSchema["http_data_source"] = flattenHttpData([]*storagetransfer.HttpData{transferSpec.HttpDataSource})
	}

	return []map[string][]map[string]interface{}{transferSpecSchema}
}
