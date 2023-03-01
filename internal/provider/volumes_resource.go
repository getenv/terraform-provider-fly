package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fly "github.com/superfly/flyctl/api"
	"github.com/superfly/graphql"
)

var (
	_ resource.Resource                = &volumesResource{}
	_ resource.ResourceWithConfigure   = &volumesResource{}
	_ resource.ResourceWithImportState = &volumesResource{}
)

type volumesResource struct {
	client *graphql.Client
}

func newVolumesResource() resource.Resource {
	return &volumesResource{}
}

type volumesResourceModel struct {
	AppName types.String `tfsdk:"app"`
	VolName types.String `tfsdk:"volumename"`
}

func (r *volumesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

func (r *volumesResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly Volumes",

		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
			"volumename": schema.StringAttribute{
				MarkdownDescription: "Volume name",
				Required:            true,
			},
		},
	}
}

func (r *volumesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volumesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var volume volumesResourceModel

	diags := req.Plan.Get(ctx, &volume)

	data, err := json.Marshal(diags)
	if err != nil {
		resp.Diagnostics.AddError("unmarshall error", err.Error())
	}

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Query failed appending volume details to diagnostics", string(data))
		return
	}

	query := `
	mutation($input: CreateVolumeInput!) {
		createVolume(input: $input) {
			app {
				name
			}
			volume {
				name
			}
		}
	}
	`

	createVolMutationInput := fly.CreateVolumeInput{
		AppID: volume.AppName.String(),
		Name:  volume.VolName.String(),
	}

	grq := graphql.NewRequest(query)
	grq.Var("input", createVolMutationInput)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Query failed creating a volume to the app", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &volume)...)
}

func (r *volumesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var volume volumesResourceModel
	diags := req.State.Get(ctx, &volume)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := `
	query($appName: String!) {
		app(name: $appName) {
			volumes {
				nodes {
					id
					name
					state
				}
			}
		}
	}
	`

	grq := graphql.NewRequest(q)
	grq.Var("appName", volume.AppName.ValueString())

	var fq fly.Query
	if err := r.client.Run(context.Background(), grq, &fq); err != nil {
		resp.Diagnostics.AddError("Query failed fetching Read", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, volume)...)
}

func (r *volumesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *volumesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *volumesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
