package storj

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"storj.io/uplink"
)

var AccessGrantBucketSchema = map[string]*schema.Schema{
	"name": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "The name of the bucket to grant access to.",
		ValidateFunc: validation.NoZeroValues,
	},
	"paths": {
		Type:        schema.TypeList,
		Optional:    true,
		Description: "The name of the bucket to grant access to.",
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
}

var AccessGrantSchema = map[string]*schema.Schema{
	"access_grant": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		Sensitive:    true,
		Description:  "The root access grant that we should derive the new access grant from.",
		ValidateFunc: validation.NoZeroValues,
	},
	"allow_download": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Description: "Allow downloads to occur when using the derived grant.",
	},
	"allow_upload": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Description: "Allow uploads to occur when using the derived grant.",
	},
	"allow_list": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Description: "Allow list operations to occur when using the derived grant.",
	},
	"allow_delete": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Description: "Allow deletes to occur when using the derived grant.",
	},
	"bucket": {
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Description: "The list of buckets and associated prefixes the derived access grant should allow access to.",
		Elem:        &schema.Resource{Schema: AccessGrantBucketSchema},
	},
	"derived_access_grant": {
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "The new access grant that has been derived from the provided grant.",
	},
}

func createAccessGrant(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	shares := make([]uplink.SharePrefix, 0)

	buckets := data.Get("bucket").([]interface{})
	for _, bucket := range buckets {
		b := bucket.(map[string]interface{})
		name := b["name"].(string)
		paths := b["paths"].([]interface{})

		if len(paths) == 0 {
			shares = append(shares, uplink.SharePrefix{
				Bucket: name,
			})
		} else {
			for _, path := range paths {
				shares = append(shares, uplink.SharePrefix{
					Bucket: name,
					Prefix: path.(string),
				})
			}
		}
	}

	accessGrant := data.Get("access_grant").(string)
	permission := uplink.Permission{
		AllowDownload: data.Get("allow_download").(bool),
		AllowUpload:   data.Get("allow_upload").(bool),
		AllowList:     data.Get("allow_list").(bool),
		AllowDelete:   data.Get("allow_delete").(bool),
	}

	access, err := uplink.ParseAccess(accessGrant)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse provided access grant.",
			Detail:   err.Error(),
		})

		return diags
	}

	access, err = access.Share(permission, shares...)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to restrict provided access grant to new scope.",
			Detail:   err.Error(),
		})

		return diags
	}

	derived, err := access.Serialize()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to serialize derived access grant.",
			Detail:   err.Error(),
		})

		return diags
	}

	hash := sha256.Sum256([]byte(derived))

	data.SetId(base64.StdEncoding.EncodeToString(hash[:]))
	data.Set("derived_access_grant", derived)

	return diags
}

func readAccessGrant(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	return diags
}

func deleteAccessGrant(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	return diags
}
