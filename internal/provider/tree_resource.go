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
type treeResource struct{
	client *http.DefaultClient
}

type treeResourceModel struct {
	IdempotencyKey types.String `tfsdk:"idempotency_key"`
	Quantity       types.String `tfsdk:"quantity"`
	Frequency      types.String `tfsdk:"frequency"`
	PlantedTrees   types.Int64  `tfsdk:"planted_trees"` // derived value
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
			"planted_trees": schema.Int64Attribute{
				MarkdownDescription: "Total trees planted so far (billed + unbilled).",
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

	apiKey := "123" // You may want to make this configurable

	if data.IdempotencyKey.IsNull() {
		data.IdempotencyKey = types.StringValue(fmt.Sprintf("tf-%d", time.Now().UnixNano()))
	}

	body := fmt.Sprintf(`{"quantity": %s}`, data.Quantity.ValueString())
	request, err := http.NewRequest("POST", "https://api.thetreeapp.org/v1/usage-records", strings.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Treeapp-Api-Key", apiKey)
	request.Header.Add("Idempotency-Key", data.IdempotencyKey.ValueString())

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("API Request Failed", err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected response code: %d", res.StatusCode))
		return
	}

	// Then update state from summary
	readAndSetTreeCount(ctx, &data, apiKey, resp)
	resp.State.Set(ctx, &data)
}

// Read refreshes the Terraform state with the latest data.
func (r *treeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data treeResourceModel
	req.State.Get(ctx, &data)

	apiKey := "123"
	readAndSetTreeCount(ctx, &data, apiKey, resp)
	resp.State.Set(ctx, &data)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *treeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Treat like create (send another POST if quantity changed)
	var data treeResourceModel
	req.Plan.Get(ctx, &data)

	apiKey := "123"
	if data.IdempotencyKey.IsNull() {
		data.IdempotencyKey = types.StringValue(fmt.Sprintf("tf-%d", time.Now().UnixNano()))
	}

	body := fmt.Sprintf(`{"quantity": %s}`, data.Quantity.ValueString())
	request, err := http.NewRequest("POST", "https://api.thetreeapp.org/v1/usage-records", strings.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Treeapp-Api-Key", apiKey)
	request.Header.Add("Idempotency-Key", data.IdempotencyKey.ValueString())

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("API Request Failed", err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected response code: %d", res.StatusCode))
		return
	}

	readAndSetTreeCount(ctx, &data, apiKey, resp)
	resp.State.Set(ctx, &data)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *treeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Do nothing
}


func readAndSetTreeCount(ctx context.Context, data *treeResourceModel, apiKey string, resp resource.Response) {
	request, err := http.NewRequest("GET", "https://api.thetreeapp.org/v1.1/impacts/summary", nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("X-Treeapp-Api-Key", apiKey)

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("API Request Failed", err.Error())
		return
	}
	defer res.Body.Close()

	var summary struct {
		Trees    int64 `json:"trees"`
		Unbilled struct {
			Trees int64 `json:"trees"`
		} `json:"unbilled"`
	}

	err = json.NewDecoder(res.Body).Decode(&summary)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse response", err.Error())
		return
	}

	totalTrees := summary.Trees + summary.Unbilled.Trees
	data.Trees = types.Int64Value(totalTrees)
}