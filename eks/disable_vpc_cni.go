package eks

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

type disabledAddon struct {
	provider *eksProvider
}

type disabledAddonType struct {
}

type disabledAddonState struct {
	Name types.String `tfsdk:"name"`
}

func (d *disabledAddon) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if !d.provider.configured {
		response.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource.")
		return
	}

	var plan disabledAddonState
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	patch := `{"spec":{"template":{"spec":{"nodeSelector":{"no-such-node": "true"}}}}}`
	_, err := d.provider.client.
		AppsV1().
		DaemonSets("kube-system").
		Patch(ctx, "aws-node", k8stypes.StrategicMergePatchType, []byte(patch), k8smetav1.PatchOptions{})
	if err != nil {
		response.Diagnostics.AddError("unable to patch aws-node ds", err.Error())
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (d *disabledAddon) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// no-op
}

func (d *disabledAddon) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// no-op
}

func (d *disabledAddon) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// no-op
}

func (d disabledAddonType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (d disabledAddonType) NewResource(ctx context.Context, provider provider.Provider) (resource.Resource, diag.Diagnostics) {
	return &disabledAddon{
		provider: provider.(*eksProvider),
	}, nil
}

func (d *disabledAddon) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// no-op
}
