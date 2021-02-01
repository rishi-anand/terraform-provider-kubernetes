package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceManifest() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceManifestApply,
		ReadContext:   resourceManifestApply,
		UpdateContext: resourceManifestApply,
		DeleteContext: resourceManifestDelete,
		Schema: map[string]*schema.Schema{
			Content: {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"all_namespace_override": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"skip_resources": {
				Type:     schema.TypeSet,
				Optional: true,
			},
		},
	}
}

func resourceManifestApply(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	content := d.Get(Content).(string)
	namespace := d.Get("namespace").(string)
	var diags diag.Diagnostics
	//all_namespace_override := d.Get("all_namespace_override").(bool)
	if err := doManifestAction(ApplyAction, content, namespace, nil); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceManifestDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	content := d.Get(Content).(string)
	namespace := d.Get("namespace").(string)
	var diags diag.Diagnostics
	if err := doManifestAction(DeleteAction, content, namespace, nil); err != nil {
		return diag.FromErr(err)
	}
	return diags
}
