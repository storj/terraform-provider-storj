package storj

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"storj.io/uplink"
)

var BucketSchema = map[string]*schema.Schema{
	"bucket": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		Description:  "The name of the bucket to create.",
		ValidateFunc: validation.NoZeroValues,
	},
}

func createBucket(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	project := c.(*uplink.Project)

	bucket := data.Get("bucket").(string)

	_, err := project.CreateBucket(ctx, bucket)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create bucket",
			Detail:   err.Error(),
		})

		return diags
	}

	data.SetId(bucket)

	return diags
}

func readBucket(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	project := c.(*uplink.Project)

	bucket := data.Id()
	if bucket == "" {
		bucket = data.Get("bucket").(string) // as data source
	}

	_, err := project.StatBucket(ctx, bucket)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to read bucket",
			Detail:   err.Error(),
		})

		return diags
	}

	data.SetId(bucket)

	return diags
}

func deleteBucket(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	project := c.(*uplink.Project)

	bucket := data.Id()

	_, err := project.DeleteBucket(ctx, bucket)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete bucket",
			Detail:   err.Error(),
		})
	}

	return diags
}
