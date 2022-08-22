package eks

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type corednsRunning struct {
	provider *eksProvider
}

type corednsRunningType struct {
}

type corednsRunningState struct {
	Name types.String `tfsdk:"name"`
}

func (c *corednsRunning) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if !c.provider.configured {
		response.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource.")
		return
	}

	var plan corednsRunningState
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if watch, err := c.provider.client.AppsV1().Deployments("kube-system").Watch(ctx, k8smetav1.ListOptions{}); err != nil {
		response.Diagnostics.AddError("unable to create watcher", err.Error())
		return
	} else {
		defer watch.Stop()
	await:
		for {
			select {
			case event := <-watch.ResultChan():
				deployment := event.Object.(*k8sappsv1.Deployment)
				if deployment.Name == "coredns" && deployment.Status.Replicas == deployment.Status.ReadyReplicas {
					break await
				}
			case <-ctx.Done():
				response.Diagnostics.AddError("context has been cancelled", "cancelled by caller")
				return
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (c *corednsRunning) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// no-op
}

func (c *corednsRunning) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// no-op
}

func (c *corednsRunning) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// no-op
}

func (c *corednsRunningType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (c *corednsRunningType) NewResource(ctx context.Context, provider provider.Provider) (resource.Resource, diag.Diagnostics) {
	return &corednsRunning{
		provider: provider.(*eksProvider),
	}, nil
}
