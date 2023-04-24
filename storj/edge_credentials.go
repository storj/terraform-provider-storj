package storj

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"storj.io/uplink/edge"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"storj.io/uplink"
)

var EdgeCredentialsSchema = map[string]*schema.Schema{
	"access_grant": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		Sensitive:    true,
		Description:  "The access grant to derive edge credentials from.",
		ValidateFunc: validation.NoZeroValues,
	},
	"access_key_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "The access key id derived from the access grant.",
	},
	"allow_public_access": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
		Description: "Whether objects can be read without authentication.",
	},
	"auth_service_address": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "The auth service address to use. Defaults to auth.storjshare.io:7777 if unset.",
	},
	"endpoint": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The url to the gateway.",
	},
	"secret_key": {
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "The secret key derived from the access grant.",
	},
}

func createEdgeCredentials(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {

	authServiceAddress := data.Get("auth_service_address").(string)

	if authServiceAddress == "" {
		authServiceAddress = "auth.storjshare.io:7777"
	}

	accessGrant := data.Get("access_grant").(string)

	access, err := uplink.ParseAccess(accessGrant)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse provided access grant.",
			Detail:   err.Error(),
		})

		return diags
	}

	allowPublicAccess := data.Get("allow_public_access").(bool)

	config := edge.Config{
		AuthServiceAddress: authServiceAddress,
	}

	credentials, err := config.RegisterAccess(ctx, access, &edge.RegisterAccessOptions{
		Public: allowPublicAccess,
	})

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create edge credentials.",
			Detail:   err.Error(),
		})

		return diags
	}

	id := fmt.Sprintf("%s.%s.%s.%s", accessGrant, credentials.Endpoint, credentials.AccessKeyID, credentials.SecretKey)

	hash := sha256.Sum256([]byte(id))

	data.SetId(base64.StdEncoding.EncodeToString(hash[:]))
	data.Set("access_key_id", credentials.AccessKeyID)
	data.Set("endpoint", credentials.Endpoint)
	data.Set("secret_key", credentials.SecretKey)

	return diags
}

func readEdgeCredentials(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	return diags
}

func deleteEdgeCredentials(ctx context.Context, data *schema.ResourceData, c interface{}) (diags diag.Diagnostics) {
	return diags
}
