package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fly "github.com/superfly/flyctl/api"
	"github.com/superfly/graphql"
)

var (
	_ resource.Resource                = &ipResource{}
	_ resource.ResourceWithConfigure   = &ipResource{}
	_ resource.ResourceWithImportState = &ipResource{}
)

type ipResource struct {
	client *graphql.Client
}

func newIpResource() resource.Resource {
	return &ipResource{}
}

type ipResourceModel struct {
	Address types.String `tfsdk:"address"`
	AppName types.String `tfsdk:"app"`
}

func (r *ipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip"
}

func (r *ipResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly IP address",

		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				MarkdownDescription: "IP address",
				Computed:            true,
			},
			"app": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
		},
	}
}

func (r *ipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*graphql.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *graphql.Client, got: %T.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var ip ipResourceModel

	diags := req.Plan.Get(ctx, &ip)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := `
		mutation($input: AllocateIPAddressInput!) {
			allocateIpAddress(input: $input) {
				ipAddress {
					address
				}
			}
		}
	`

	input := fly.AllocateIPAddressInput{AppID: ip.AppName.ValueString(), Type: "v6"}

	grq := graphql.NewRequest(query)
	grq.Var("input", input)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	ip.Address = types.StringValue(ff.AllocateIPAddress.IPAddress.Address)
	resp.Diagnostics.Append(resp.State.Set(ctx, &ip)...)
}

func (r *ipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var ip ipResourceModel

	diags := req.State.Get(ctx, &ip)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &ip)...)
}

func (r *ipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *ipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *ipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
