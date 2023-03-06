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
	VolID   types.String `tfsdk:"volid"`
	Name    types.String `tfsdk:"name"`
	Region  types.String `tfsdk:"region"`
	SizeGB  types.Int64  `tfsdk:"sizegb"`
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
			"name": schema.StringAttribute{
				MarkdownDescription: "Volume name",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Deployment region",
				Required:            true,
			},
			"sizegb": schema.Int64Attribute{
				MarkdownDescription: "Volume size in GB",
				Required:            true,
			},
			"volid": schema.StringAttribute{
				MarkdownDescription: "Volume ID",
				Optional:            true,
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
		}
	}
	`

	createVolMutationInput := fly.CreateVolumeInput{
		AppID:  volume.AppName.ValueString(),
		Name:   volume.Name.ValueString(),
		Region: volume.Region.ValueString(),
		SizeGb: int(volume.SizeGB.ValueInt64()),
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
		resp.Diagnostics.AddWarning("Query failed fetching Read", err.Error())
	}

	respDiagnostics := resp.State.Set(ctx, volume)
	if respDiagnostics.HasError() {
		resp.Diagnostics.AddError("Read response error on setting state", "abc")
	}
	resp.Diagnostics.Append(respDiagnostics...)
}

func (r *volumesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	mutation($input: ExtendVolumeInput!) {
		extendVolume(input: $input) {
			app {
				name
			}
		}
	}
`

	extendVolumeInput := fly.ExtendVolumeInput{
		VolumeID: volume.VolID.ValueString(),
		SizeGb:   int(volume.SizeGB.ValueInt64()),
	}

	grq := graphql.NewRequest(query)
	grq.Var("input", extendVolumeInput)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Extend Volume failed", "client interaction:"+err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &extendVolumeInput)...)
}

func (r *volumesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var volume volumesResourceModel
	diags := req.State.Get(ctx, &volume)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := `
	mutation($input: DeleteVolumeInput!) {
		deleteVolume(input: $input) {
			app {
				name
			}
		}
	}
`

	grq := graphql.NewRequest(q)
	deleteInput := fly.DeleteVolumeInput{
		VolumeID: "vol_2en7r1poodwrk6yx",
	}

	grq.Var("input", deleteInput)

	var fq fly.Query
	if err := r.client.Run(context.Background(), grq, &fq); err != nil {
		resp.Diagnostics.AddError("Query failed deleting the volume", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, volume)...)
}

func (r *volumesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
