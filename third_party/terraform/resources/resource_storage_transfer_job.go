package google

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/storagetransfer/v1"
	"log"
	"regexp"
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
				Type:     schema.TypeString,
				Required: true,
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
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: gcsDataSchema(),
							},
							ConflictsWith: []string{"transfer_spec.aws_s3_data_source", "transfer_spec.http_data_source"},
						},
						"aws_s3_data_source": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: awsS3DataSchema(),
							},
							ConflictsWith: []string{"transfer_spec.gcs_data_source", "transfer_spec.http_data_source"},
						},
						"http_data_source": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: httpDataSchema(),
							},
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
						"schedule_start_date": dateObjectSchema(true, false),
						"schedule_end_date":   dateObjectSchema(false, true),
						"start_time_of_day":   timeObjectSchema(),
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

func timeObjectSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		MaxItems: 1,
		Elem: &schema.Resource{
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
		},
	}

}

func dateObjectSchema(required bool, optional bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: required,
		Optional: optional,
		ForceNew: true,
		MaxItems: 1,
		Elem: &schema.Resource{
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
		},
	}
}

func gcsDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bucket_name": &schema.Schema{
			Required: true,
			Type:     schema.TypeString,
		},
	}
}

func awsS3DataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func httpDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"list_url": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
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
		Schedule:     expandTransferSchedules(d.Get("schedule").([]interface{}))[0],
		TransferSpec: expandTransferSpecs(d.Get("transfer_spec").([]interface{}))[0],
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
	jobId, err := extractTransferJobId(res.Name)
	if err != nil {
		fmt.Printf("Error extracting transfer job id %v: %v", transferJob, err)
		return err
	}

	log.Printf("[DEBUG] Created transfer job %v \n\n", jobId)
	d.SetId(jobId)
	return nil
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

	d.Set("transfer_spec", flattenTransferSpecs([]*storagetransfer.TransferSpec{res.TransferSpec}))
	if err != nil {
		return err
	}

	jobId, err := extractTransferJobId(res.Name)
	if err != nil {
		fmt.Printf("Error extracting transfer job id %v: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] Patched transfer job: %v\n\n", jobId)
	d.SetId(jobId)
	return nil
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
			transferJob.Schedule = expandTransferSchedules(v.([]interface{}))[0]
		}
	}

	if d.HasChange("transfer_spec") {
		if v, ok := d.GetOk("transfer_spec"); ok {
			fieldMask = append(fieldMask, "transfer_spec")
			transferJob.TransferSpec = expandTransferSpecs(v.([]interface{}))[0]
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

	jobId, err := extractTransferJobId(res.Name)
	if err != nil {
		fmt.Printf("Error extracting transfer job id %v: %v", transferJob, err)
		return err
	}

	log.Printf("[DEBUG] Patched transfer job: %v\n\n", jobId)
	d.SetId(jobId)
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

func extractTransferJobId(id string) (string, error) {
	if !regexp.MustCompile("^transferJobs/.+$").Match([]byte(id)) {
		return "", fmt.Errorf("Invalid transferJob id format, expecting transferJob/{id}")
	}
	parts := strings.Split(id, "/")
	return parts[1], nil
}

func expandDates(dates []interface{}) []*storagetransfer.Date {
	expandedDates := make([]*storagetransfer.Date, 0, len(dates))
	for _, raw := range dates {
		date := raw.([]interface{})
		expandedDates = append(expandedDates, &storagetransfer.Date{
			Day:   int64(extractFirstMapConfig(date)["day"].(int)),
			Month: int64(extractFirstMapConfig(date)["month"].(int)),
			Year:  int64(extractFirstMapConfig(date)["year"].(int)),
		})
	}
	return expandedDates
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

func expandTimeOfDays(times []interface{}) []*storagetransfer.TimeOfDay {
	expandedTimes := make([]*storagetransfer.TimeOfDay, 0, len(times))
	for _, raw := range times {
		time := raw.([]interface{})
		expandedTimes = append(expandedTimes, &storagetransfer.TimeOfDay{
			Hours:   int64(extractFirstMapConfig(time)["hours"].(int)),
			Minutes: int64(extractFirstMapConfig(time)["minutes"].(int)),
			Seconds: int64(extractFirstMapConfig(time)["seconds"].(int)),
			Nanos:   int64(extractFirstMapConfig(time)["nanos"].(int)),
		})
	}
	return expandedTimes
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

func expandTransferSchedules(transferSchedules []interface{}) []*storagetransfer.Schedule {
	schedules := make([]*storagetransfer.Schedule, 0, len(transferSchedules))
	for _, raw := range transferSchedules {
		schedule := raw.(map[string]interface{})
		sched := &storagetransfer.Schedule{
			ScheduleStartDate: expandDates([]interface{}{schedule["schedule_start_date"]})[0],
		}

		if v, ok := schedule["schedule_end_date"]; ok && len(v.([]interface{})) > 0 {
			sched.ScheduleEndDate = expandDates([]interface{}{v})[0]
		}
		if v, ok := schedule["start_time_of_day"]; ok && len(v.([]interface{})) > 0 {
			sched.StartTimeOfDay = expandTimeOfDays([]interface{}{v})[0]
		}

		schedules = append(schedules, sched)
	}
	return schedules
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

func expandGcsData(gcsDatas []interface{}) []*storagetransfer.GcsData {
	datas := make([]*storagetransfer.GcsData, 0, len(gcsDatas))
	for _, raw := range gcsDatas {
		data := raw.(map[string]interface{})
		datas = append(datas, &storagetransfer.GcsData{
			BucketName: data["bucket_name"].(string),
		})
	}
	return datas
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

func expandAwsAccessKeys(awsAccessKeys []interface{}) []*storagetransfer.AwsAccessKey {
	datas := make([]*storagetransfer.AwsAccessKey, 0, len(awsAccessKeys))
	for _, raw := range awsAccessKeys {
		data := raw.(map[string]interface{})
		datas = append(datas, &storagetransfer.AwsAccessKey{
			AccessKeyId:     data["access_key_id"].(string),
			SecretAccessKey: data["secret_access_key"].(string),
		})
	}
	return datas
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

func expandAwsS3Data(awsS3Datas []interface{}) []*storagetransfer.AwsS3Data {
	datas := make([]*storagetransfer.AwsS3Data, 0, len(awsS3Datas))
	for _, raw := range awsS3Datas {
		data := raw.(map[string]interface{})
		datas = append(datas, &storagetransfer.AwsS3Data{
			BucketName:   data["bucket_name"].(string),
			AwsAccessKey: expandAwsAccessKeys(data["aws_access_key"].([]interface{}))[0],
		})
	}
	return datas
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

func expandHttpData(httpDatas []interface{}) []*storagetransfer.HttpData {
	datas := make([]*storagetransfer.HttpData, 0, len(httpDatas))
	for _, raw := range httpDatas {
		data := raw.(map[string]interface{})
		datas = append(datas, &storagetransfer.HttpData{
			ListUrl: data["list_url"].(string),
		})
	}
	return datas
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

func expandObjectConditions(conditions []interface{}) []*storagetransfer.ObjectConditions {
	datas := make([]*storagetransfer.ObjectConditions, 0, len(conditions))
	for _, raw := range conditions {
		data := raw.(map[string]interface{})
		datas = append(datas, &storagetransfer.ObjectConditions{
			ExcludePrefixes:                     convertStringArr(data["exclude_prefixes"].([]interface{})),
			IncludePrefixes:                     convertStringArr(data["include_prefixes"].([]interface{})),
			MaxTimeElapsedSinceLastModification: data["max_time_elapsed_since_last_modification"].(string),
			MinTimeElapsedSinceLastModification: data["min_time_elapsed_since_last_modification"].(string),
		})
	}
	return datas
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

func expandTransferOptions(options []interface{}) []*storagetransfer.TransferOptions {
	datas := make([]*storagetransfer.TransferOptions, 0, len(options))
	for _, raw := range options {
		data := raw.(map[string]interface{})
		datas = append(datas, &storagetransfer.TransferOptions{
			DeleteObjectsFromSourceAfterTransfer:  data["delete_objects_from_source_after_transfer"].(bool),
			DeleteObjectsUniqueInSink:             data["delete_objects_unique_in_sink"].(bool),
			OverwriteObjectsAlreadyExistingInSink: data["overwrite_objects_already_existing_in_sink"].(bool),
		})
	}
	return datas
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

func expandTransferSpecs(transferSpecs []interface{}) []*storagetransfer.TransferSpec {
	specs := make([]*storagetransfer.TransferSpec, 0, len(transferSpecs))
	for _, raw := range transferSpecs {
		spec := raw.(map[string]interface{})

		transferSpec := &storagetransfer.TransferSpec{
			GcsDataSink: expandGcsData(spec["gcs_data_sink"].([]interface{}))[0],
		}

		if v, ok := spec["object_conditions"]; ok && len(v.([]interface{})) > 0 {
			transferSpec.ObjectConditions = expandObjectConditions(v.([]interface{}))[0]
		}
		if v, ok := spec["transfer_options"]; ok && len(v.([]interface{})) > 0 {
			transferSpec.TransferOptions = expandTransferOptions(v.([]interface{}))[0]
		}

		if v, ok := spec["gcs_data_source"]; ok && len(v.([]interface{})) > 0 {
			transferSpec.GcsDataSource = expandGcsData(v.([]interface{}))[0]
		} else if v, ok := spec["aws_s3_data_source"]; ok && len(v.([]interface{})) > 0 {
			transferSpec.AwsS3DataSource = expandAwsS3Data(v.([]interface{}))[0]
		} else if v, ok := spec["http_data_source"]; ok && len(v.([]interface{})) > 0 {
			transferSpec.HttpDataSource = expandHttpData(v.([]interface{}))[0]
		}

		specs = append(specs, transferSpec)
	}
	return specs
}

func flattenTransferSpecs(transferSpecs []*storagetransfer.TransferSpec) []map[string][]map[string]interface{} {
	transferSpecsSchema := make([]map[string][]map[string]interface{}, 0, len(transferSpecs))
	for _, transferSpec := range transferSpecs {
		schema := map[string][]map[string]interface{}{
			"gcs_data_sink": flattenGcsData([]*storagetransfer.GcsData{transferSpec.GcsDataSink}),
		}

		if transferSpec.ObjectConditions != nil {
			schema["object_conditions"] = flattenObjectConditions([]*storagetransfer.ObjectConditions{transferSpec.ObjectConditions})
		}
		if transferSpec.TransferOptions != nil {
			schema["transfer_options"] = flattenTransferOptions([]*storagetransfer.TransferOptions{transferSpec.TransferOptions})
		}
		if transferSpec.GcsDataSource != nil {
			schema["gcs_data_source"] = flattenGcsData([]*storagetransfer.GcsData{transferSpec.GcsDataSource})
		} else if transferSpec.AwsS3DataSource != nil {
			schema["aws_s3_data_source"] = flattenAwsS3Data([]*storagetransfer.AwsS3Data{transferSpec.AwsS3DataSource})
		} else if transferSpec.HttpDataSource != nil {
			schema["http_data_source"] = flattenHttpData([]*storagetransfer.HttpData{transferSpec.HttpDataSource})
		}

		transferSpecsSchema = append(transferSpecsSchema, schema)
	}
	return transferSpecsSchema
}
