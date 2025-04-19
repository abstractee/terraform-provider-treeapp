package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/validator"
	"github.com/hashicorp/terraform-plugin-framework/validator/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/stringdefault"
)

// Ensure the implementation satisfies the expected interfaces.
var (
    _ resource.Resource = &treeResource{}
)

// NewtreeResource is a helper function to simplify the provider implementation.
func NewTreeResource() resource.Resource {
    return &treeResource{}
}

// treeResource is the resource implementation.
type treeResource struct{}

// Metadata returns the resource type name.
func (r *treeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_tree"
}

// Schema defines the schema for the resource.
func (r *treeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"idempotency_key": schema.StringAttribute{
				MarkdownDescription: "Idempotency key",
				Optional:            true,
			},
			"quantity": schema.StringAttribute{
				MarkdownDescription: "Quantity of tree to plant",
				Required:            true,
			},
			"frequency": schema.StringAttribute{
				MarkdownDescription: "How often to plant the trees. One of: `per_month`, `per_deployment`, `one_time`.",
				Optional:            true,
				Computed:            true, 
				Default:             stringdefault.StaticString("one_time"),
				Validators: []validator.String{
					stringvalidator.OneOf("per_month", "per_deployment", "one_time"),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *treeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

// Read refreshes the Terraform state with the latest data.
func (r *treeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *treeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *treeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Do nothing
}
