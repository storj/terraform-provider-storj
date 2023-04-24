package storj

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"storj.io/uplink"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_grant": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The access grant used to authenticate the user with the server.",
				DefaultFunc: schema.EnvDefaultFunc("STORJ_ACCESS_GRANT", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"storj_bucket": {
				Schema:        BucketSchema,
				CreateContext: createBucket,
				ReadContext:   readBucket,
				DeleteContext: deleteBucket,
			},
			"storj_object": {
				Schema:        ObjectSchema,
				CreateContext: createObject,
				ReadContext:   readObject,
				UpdateContext: updateObject,
				DeleteContext: deleteObject,
			},
			"storj_access_grant": {
				Schema:        AccessGrantSchema,
				CreateContext: createAccessGrant,
				ReadContext:   readAccessGrant,
				DeleteContext: deleteAccessGrant,
			},
			"storj_edge_credentials": {
				Schema:        EdgeCredentialsSchema,
				CreateContext: createEdgeCredentials,
				ReadContext:   readEdgeCredentials,
				DeleteContext: deleteEdgeCredentials,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"storj_bucket": {
				Schema:      BucketSchema,
				ReadContext: readBucket,
			},
			"storj_object": {
				Schema:      ObjectSchema,
				ReadContext: readObject,
			},
			"storj_objects": {
				Schema:      ObjectListKind,
				ReadContext: readObjects,
			},
		},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (_ interface{}, diags diag.Diagnostics) {
			accessGrant := data.Get("access_grant").(string)

			access, err := uplink.ParseAccess(accessGrant)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to parse access grant",
					Detail:   err.Error(),
				})

				return nil, diags
			}

			return access, diags
		},
	}
}
