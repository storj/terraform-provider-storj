package storj

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"storj.io/uplink"
)

var ObjectSchema = map[string]*schema.Schema{
	"bucket": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		Description:  "The name of the bucket to create.",
		ValidateFunc: validation.NoZeroValues,
	},
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		Description:  "The key of the object in the bucket.",
		ValidateFunc: validation.NoZeroValues,
	},
	"content": {
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"source", "content_base64"},
		Description:   "Literal string value to use as the object content, which will be uploaded as UTF-8-encoded text.",
	},
	"content_base64": {
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"source", "content"},
		Description:   "Base64-encoded data that will be decoded and uploaded as raw bytes for the object content.",
	},
	"metadata": {
		Type:        schema.TypeMap,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "Map of keys/values to provision metadata.",
	},
	"source": {
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"content", "content_base64"},
		Description:   "Path to a file that will be read and uploaded as raw bytes for the object content.",
	},
}

var ObjectListKind = map[string]*schema.Schema{
	"bucket": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		Description:  "The name of the bucket to create.",
		ValidateFunc: validation.NoZeroValues,
	},
	"prefix": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The prefix to scan.",
	},
	"max_keys": {
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "The maximum number of keys to return.",
	},
	"start_offset": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The offset to start at.",
	},
	"keys": {
		Type:     schema.TypeList,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	},
}

func createObject(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	return updateObject(ctx, data, c)
}

func readObject(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	access := c.(*uplink.Access)

	project, err := uplink.OpenProject(ctx, access)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to open project",
			Detail:   err.Error(),
		})
	}

	defer project.Close()

	bucket := data.Get("bucket").(string)
	key := data.Id()
	if key == "" {
		key = data.Get("key").(string)
	}

	download, err := project.DownloadObject(ctx, bucket, key, &uplink.DownloadOptions{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to download object",
			Detail:   err.Error(),
		})

		return diags
	}

	d, err := io.ReadAll(download)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to download object",
			Detail:   err.Error(),
		})

		return diags
	}

	data.Set("content", string(d))
	data.Set("content_base64", base64.StdEncoding.EncodeToString(d))

	return diags
}

func readObjects(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	access := c.(*uplink.Access)

	project, err := uplink.OpenProject(ctx, access)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to open project",
			Detail:   err.Error(),
		})
	}

	defer project.Close()

	bucket := data.Get("bucket").(string)

	// optional
	prefix := data.Get("prefix").(string)
	maxKeys := data.Get("max_keys").(int)
	startAfter := data.Get("start_after").(string)

	iterator := project.ListObjects(ctx, bucket, &uplink.ListObjectsOptions{
		Prefix:    prefix,
		Cursor:    startAfter,
		Recursive: true,
	})

	keys := make([]string, 0)

	for iterator.Next() {
		if err := iterator.Err(); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to iterate object",
				Detail:   err.Error(),
			})

			return diags
		}

		keys = append(keys, iterator.Item().Key)
		if len(keys) == maxKeys {
			break
		}
	}

	data.SetId(bucket)
	data.Set("keys", keys)

	return diags
}

func updateObject(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	access := c.(*uplink.Access)

	project, err := uplink.OpenProject(ctx, access)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to open project",
			Detail:   err.Error(),
		})
	}

	defer project.Close()

	// required
	bucket := data.Get("bucket").(string)
	key := data.Get("key").(string)

	// one of
	source := data.Get("source").(string)
	content := data.Get("content").(string)
	contentB64 := data.Get("content_base64").(string)

	var reader io.Reader
	switch {
	case source != "":
		// file
		f, err := os.Open(source)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to open source file",
				Detail:   err.Error(),
			})

			return diags
		}

		defer f.Close()
		reader = f
	case content != "":
		// utf8
		reader = strings.NewReader(content)
	case contentB64 != "":
		// binary
		reader = base64.NewDecoder(base64.StdEncoding, strings.NewReader(contentB64))
	}

	// optional
	m := data.Get("metadata").(map[string]interface{})
	metadata := make(map[string]string, len(m))

	for k, v := range m {
		metadata[k] = v.(string)
	}

	upload, err := project.UploadObject(ctx, bucket, key, &uplink.UploadOptions{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to start object upload.",
			Detail:   err.Error(),
		})

		return diags
	}

	defer upload.Abort()

	err = upload.SetCustomMetadata(ctx, metadata)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set metadata for object.",
			Detail:   err.Error(),
		})

		return diags
	}

	_, err = io.Copy(upload, reader)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to upload object content.",
			Detail:   err.Error(),
		})

		return diags
	}

	err = upload.Commit()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to commit object upload.",
			Detail:   err.Error(),
		})

		return diags
	}

	data.SetId(key)

	return diags
}

func deleteObject(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	access := c.(*uplink.Access)

	project, err := uplink.OpenProject(ctx, access)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to open project",
			Detail:   err.Error(),
		})
	}

	defer project.Close()

	bucket := data.Get("bucket").(string)
	key := data.Get("key").(string)

	_, err = project.DeleteObject(ctx, bucket, key)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete object",
			Detail:   err.Error(),
		})
	}

	return diags
}
