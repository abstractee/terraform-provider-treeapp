// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TreeappProvider satisfies various provider interfaces.
var _ provider.Provider = &TreeappProvider{}

// TreeappProvider defines the provider implementation.
type TreeappProvider struct {
	version string
	client  *TreeappClient
}

// TreeappProviderModel describes the provider data model.
type TreeappProviderModel struct {
	Api_key types.String `tfsdk:"api_key"`
}

func (p *TreeappProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "treeapp"
	resp.Version = p.version
}

func (p *TreeappProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API Key for TreeApp",
				Required:            true,
			},
		},
	}
}

func (p *TreeappProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TreeappProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Api_key.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Treeapp API Key",
			"The provider cannot create the Treeapp API client as there is an unknown configuration value for the Treeapp API key. ",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

    Api_key := os.Getenv("TREEAPP_API_KEY")

    if !config.Api_key.IsNull() {
        Api_key = config.Api_key.ValueString()
    }

	if config.Api_key.String() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Treeapp API Key",
			"The provider cannot create the Treeapp API client as there is a missing or empty value for the Treeapp API Key. If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := NewTreeappClient(Api_key)
	p.client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TreeappProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTreeResource,
	}
}

func (p *TreeappProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
	// return []func() datasource.DataSource{
	// 	NewExampleDataSource,
	// }
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TreeappProvider{
			version: version,
		}
	}
}
