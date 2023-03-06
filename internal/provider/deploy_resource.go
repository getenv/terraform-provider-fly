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
	_ resource.Resource                = &appDeployResource{}
	_ resource.ResourceWithConfigure   = &appDeployResource{}
	_ resource.ResourceWithImportState = &appDeployResource{}
)

type appDeployResource struct {
	client *graphql.Client
}

func newAppDeployResource() resource.Resource {
	return &appDeployResource{}
}

type appDeployResourceModel struct {
	AppName types.String `tfsdk:"app"`
	Image   types.String `tfsdk:"image"`
}

func (r *appDeployResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deploy"
}

func (r *appDeployResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly App Deployment",

		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Docker Imagename",
				Required:            true,
			},
		},
	}
}

func (r *appDeployResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *appDeployResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var appDeploy appDeployResourceModel

	diags := req.Plan.Get(ctx, &appDeploy)

	data, err := json.Marshal(diags)
	if err != nil {
		resp.Diagnostics.AddError("unmarshall error", err.Error())
	}

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Query failed appending deploy details to diagnostics", string(data))
		return
	}

	query := `
	mutation($input: DeployImageInput!) {
		deployImage(input: $input) {
			release {
				id
			}
			releaseCommand {
				id
			}
		}
	}
`

	deployAppMutationInput := fly.DeployImageInput{
		AppID: appDeploy.AppName.ValueString(),
		Image: appDeploy.Image.ValueString(),
	}

	grq := graphql.NewRequest(query)
	grq.Var("input", deployAppMutationInput)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Query failed deploying the app with image", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &appDeploy)...)
}

func (r *appDeployResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var appDeploy appDeployResourceModel
	diags := req.State.Get(ctx, &appDeploy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := `
	query ($appName: String!, $deploymentId: ID!, $evaluationId: String!) {
		app(name: $appName) {
			deploymentStatus(id: $deploymentId, evaluationId: $evaluationId) {
				id
				inProgress
				status
				successful
				description
				version
				desiredCount
				placedCount
				healthyCount
				unhealthyCount
				allocations {
					id
					idShort
					status
					region
					desiredStatus
					version
					healthy
					failed
					canary
					restarts
					checks {
						status
						serviceName
					}
				}
			}
		}
	}
	`

	grq := graphql.NewRequest(q)
	grq.Var("appName", appDeploy.AppName.ValueString())
	grq.Var("deploymentId", appDeploy.AppName.ValueString())
	grq.Var("evaluationId", appDeploy.AppName.ValueString())

	var fq fly.Query
	if err := r.client.Run(context.Background(), grq, &fq); err != nil {
		resp.Diagnostics.AddError("Query failed fetching Read", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, appDeploy)...)
}

func (r *appDeployResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *appDeployResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *appDeployResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
