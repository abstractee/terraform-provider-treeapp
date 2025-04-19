// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TreeappProvider satisfies various provider interfaces.
var _ provider.Provider = &TreeappProvider{}

// TreeappProvider defines the provider implementation.
type TreeappProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
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
	var data TreeappProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

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

    if Api_key == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("api_key"),
            "Missing Treeapp API Key",
            "The provider cannot create the Treeapp API client as there is a missing or empty value for the Treeapp API Key. If either is already set, ensure the value is not empty.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

	client := http.DefaultClient
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
