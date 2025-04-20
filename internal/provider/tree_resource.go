package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	// "github.com/hashicorp/terraform-plugin-framework/validator"
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
type treeResource struct {
	client *TreeappClient
}

type treeResourceModel struct {
	IdempotencyKey    types.String `tfsdk:"idempotency_key"`
	Quantity          types.Int64  `tfsdk:"quantity"` 
	Frequency         types.String `tfsdk:"frequency"`
	PlantedTrees      types.Map    `tfsdk:"planted_trees"` // map[string]int64
}

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
			"quantity": schema.Int64Attribute{
				MarkdownDescription: "Desired quantity of trees to plant",
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
			"planted_trees": schema.MapAttribute{
				MarkdownDescription: "Trees planted so far (billed, unbilled).",
				ElementType:         types.Int64Type,
				Computed:            true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *treeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data treeResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get state from summary
	stats, err := r.client.GetPlantedTreeStats()
	if err != nil {
		resp.Diagnostics.AddError("GetPlantedTreeStats: Request Error", err.Error())
		return
	}
	billed := stats["billed"]
	unbilled := stats["unbilled"]

	// Reconcile quantity if needed
	quantity := 0
	switch data.Frequency.ValueString() {
		case "one_time":
			total := billed + unbilled
			if total < data.Quantity.ValueInt64() {
				quantity = int(data.Quantity.ValueInt64() - total)
			}
		
		case "per_month":
			if unbilled < data.Quantity.ValueInt64() {
				quantity = int(data.Quantity.ValueInt64() - unbilled)
			}
		
		case "per_deployment":
			quantity  = int(data.Quantity.ValueInt64() )
	}

	_, err = r.client.CreateUsageRecord(int(quantity), "")

	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}

	// Then update state from summary
	stats, err = r.client.GetPlantedTreeStats()
	if err != nil {
		resp.Diagnostics.AddError("GetPlantedTreeStats: Request Error", err.Error())
		return
	}
	billed = stats["billed"]
	unbilled = stats["unbilled"]
	// Construct the planted_trees map
	plantedTrees := map[string]attr.Value{
		"billed":   types.Int64Value(billed),
		"unbilled": types.Int64Value(unbilled),
	}
	data.PlantedTrees, diags = types.MapValue(types.Int64Type, plantedTrees)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, &data)
}

// Read refreshes the Terraform state with the latest data.
func (r *treeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data treeResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve tree stats (billed/unbilled) from external system
	stats, err := r.client.GetPlantedTreeStats() 
	if err != nil {
		resp.Diagnostics.AddError("GetPlantedTreeStats: Request Error", err.Error())
		return
	} 
	billed := stats["billed"]
	unbilled := stats["unbilled"]

	// Construct the planted_trees map
	plantedTrees := map[string]attr.Value{
		"billed":   types.Int64Value(billed),
		"unbilled": types.Int64Value(unbilled),
	}
	data.PlantedTrees, diags = types.MapValue(types.Int64Type, plantedTrees)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reconcile quantity if needed
	switch data.Frequency.ValueString() {
		case "one_time":
			total := billed + unbilled
			if total < data.Quantity.ValueInt64() {
				data.Quantity = types.Int64Value(data.Quantity.ValueInt64() - total)
			}
		
		case "per_month":
			if unbilled < data.Quantity.ValueInt64() {
				data.Quantity = types.Int64Value(data.Quantity.ValueInt64() - unbilled)
			}
		
		case "per_deployment":
			// Do not update quantity
	}

	// Set final state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}


// Update updates the resource and sets the updated Terraform state on success.
func (r *treeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Treat like create (send another POST if quantity changed)
	var data treeResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get state from summary
	stats, err := r.client.GetPlantedTreeStats()
	if err != nil {
		resp.Diagnostics.AddError("GetPlantedTreeStats: Request Error", err.Error())
		return
	}
	billed := stats["billed"]
	unbilled := stats["unbilled"]

	// Reconcile quantity if needed
	quantity := 0
	switch data.Frequency.ValueString() {
		case "one_time":
			total := billed + unbilled
			if total < data.Quantity.ValueInt64() {
				quantity = int(data.Quantity.ValueInt64() - total)
			}
		
		case "per_month":
			if unbilled < data.Quantity.ValueInt64() {
				quantity = int(data.Quantity.ValueInt64() - unbilled)
			}
		
		case "per_deployment":
			quantity  = int(data.Quantity.ValueInt64() )
	}

	_, err = r.client.CreateUsageRecord(int(quantity), "")

	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}

	// Then update state from summary
	stats, err = r.client.GetPlantedTreeStats()
	if err != nil {
		resp.Diagnostics.AddError("GetPlantedTreeStats: Request Error", err.Error())
		return
	}
	billed = stats["billed"]
	unbilled = stats["unbilled"]

	// Construct the planted_trees map
	plantedTrees := map[string]attr.Value{
		"billed":   types.Int64Value(billed),
		"unbilled": types.Int64Value(unbilled),
	}
	data.PlantedTrees, diags = types.MapValue(types.Int64Type, plantedTrees)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, &data)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *treeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Do nothing
}

// Configure adds the provider configured client to the resource.
func (r *treeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*TreeappClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *TreeappClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}
