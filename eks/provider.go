package eks

import (
	"context"
	"encoding/pem"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func New() provider.Provider {
	return &eksProvider{}
}

type eksProvider struct {
	configured bool
	client     *kubernetes.Clientset
}

type eksProviderConfig struct {
	Host                 types.String `tfsdk:"host"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
	Token                types.String `tfsdk:"token"`
}

func (p *eksProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	if !request.Config.Raw.IsFullyKnown() {
		// Some resources that we depend on might not exist during the plan phase.
		return
	}

	var providerConfig eksProviderConfig
	response.Diagnostics.Append(request.Config.Get(ctx, &providerConfig)...)
	if response.Diagnostics.HasError() {
		return
	}

	restConfig := rest.Config{}
	restConfig.Host = providerConfig.Host.Value
	restConfig.BearerToken = providerConfig.Token.Value

	certificateAuthority := []byte(providerConfig.ClusterCACertificate.Value)
	if parsed, _ := pem.Decode(certificateAuthority); parsed == nil || parsed.Type != "CERTIFICATE" {
		response.Diagnostics.AddError(
			"Invalid provider configuration", "'cluster_ca_certificate' is not a valid PEM encoded certificate")
		return
	}
	restConfig.TLSClientConfig.CAData = certificateAuthority

	if client, err := kubernetes.NewForConfig(&restConfig); err != nil {
		response.Diagnostics.AddError("unable to create k8s client", err.Error())
	} else {
		p.configured = true
		p.client = client
	}
}

func (p *eksProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"rok_eks_cluster_addon_disabled": &disabledAddonType{},
		"rok_eks_coredns_running":        &corednsRunningType{},
	}, nil
}

func (p *eksProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
}

func (p *eksProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"cluster_ca_certificate": {
				Type:      types.StringType,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"token": {
				Type:      types.StringType,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
		},
	}, nil
}
