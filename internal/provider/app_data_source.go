package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fly "github.com/superfly/flyctl/api"
	"github.com/superfly/graphql"
)

var _ datasource.DataSource = &appDataSource{}

func newAppDataSource() datasource.DataSource {
	return &appDataSource{}
}

type appDataSource struct {
	client *graphql.Client
}

type appDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

func (d *appDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *appDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly app",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
		},
	}
}

func (d *appDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *appDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var app appDataSourceModel

	diags := req.Config.Get(ctx, &app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := `
	query ($appName: String!) {
		app(name: $appName) {
			name
		}
	}
`

	r := graphql.NewRequest(query)
	r.Var("appName", app.Name.ValueString())

	var fr fly.Query
	if err := d.client.Run(ctx, r, &fr); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	app.Name = types.StringValue(fr.App.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &app)...)
}
