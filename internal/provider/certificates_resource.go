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
	_ resource.Resource                = &certificatesResource{}
	_ resource.ResourceWithConfigure   = &certificatesResource{}
	_ resource.ResourceWithImportState = &certificatesResource{}
)

type certificatesResource struct {
	client *graphql.Client
}

func newCertificatesResource() resource.Resource {
	return &certificatesResource{}
}

type certificatesResourceModel struct {
	AppName  types.String `tfsdk:"app"`
	HostName types.String `tfsdk:"host"`
}

func (r *certificatesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (r *certificatesResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly Certificates",

		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Host name",
				Required:            true,
			},
		},
	}
}

func (r *certificatesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *certificatesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var certificate certificatesResourceModel

	diags := req.Plan.Get(ctx, &certificate)

	data, err := json.Marshal(certificate)
	if err != nil {
		resp.Diagnostics.AddError("unmarshall error", err.Error())
	}

	resp.Diagnostics.AddError("Query failed setting a cert to the app", string(data))
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

func (r *certificatesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *certificatesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *certificatesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *certificatesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
