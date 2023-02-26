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
	_ resource.Resource                = &secretsResource{}
	_ resource.ResourceWithConfigure   = &secretsResource{}
	_ resource.ResourceWithImportState = &secretsResource{}
)

type certificatessResource struct {
	client *graphql.Client
}

func newCertificatessResource() resource.Resource {
	return &certificatessResource{}
}

type certificatessResourceModel struct {
	AppName  types.String `tfsdk:"app"`
	HostName types.String `tfsdk:"host"`
}

func (r *certificatessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (r *certificatessResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly Certificates",

		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
			"host": schema.MapAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *certificatessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *certificatessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var certificate certificatessResourceModel

	diags := req.Plan.Get(ctx, &certificate)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := `
	mutation($appId: ID!, $hostname: String!) {
		addCertificate(appId: $appId, hostname: $hostname) {
			certificate {
				hostname
				id
			}
		}
	}
	`

	appCert := fly.AppCertificate{
		ID:       certificate.AppName.ValueString(),
		Hostname: certificate.HostName.ValueString(),
	}

	// hostNameCheck := fly.HostnameCheck{}

	// input := fly.AppCertificate{}

	grq := graphql.NewRequest(query)
	grq.Var("input", appCert)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Query failed setting a cert to the app", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &certificate)...)
}

func (r *certificatessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *certificatessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *certificatessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *certificatessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
