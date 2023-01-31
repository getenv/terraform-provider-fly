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
	_ resource.Resource                = &appResource{}
	_ resource.ResourceWithConfigure   = &appResource{}
	_ resource.ResourceWithImportState = &appResource{}
)

type appResource struct {
	client *graphql.Client
}

type appResourceModel struct {
	Name types.String `tfsdk:"name"`
	Org  types.String `tfsdk:"org"`
}

func newAppResource() resource.Resource {
	return &appResource{}
}

func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly app",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "App name",
				Optional:            true,
			},
		},
	}
}

func (r *appResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var app appResourceModel

	diags := req.Plan.Get(ctx, &app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := `
		mutation($input: CreateAppInput!) {
			createApp(input: $input) {
				app {
					id
					name

					regions {
						name
						code
					}
				}
			}
		}
	`

	orgID, err := lookupOrgID(r.client, app.Org.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Org lookup failed", err.Error())
	}

	input := fly.CreateAppInput{
		Name:           app.Name.ValueString(),
		OrganizationID: orgID,
	}

	grq := graphql.NewRequest(query)
	grq.Var("input", input)

	var fq fly.Query
	if err := r.client.Run(context.Background(), grq, &fq); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &app)...)
}

func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var app appResourceModel

	diags := req.State.Get(ctx, &app)
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

	grq := graphql.NewRequest(query)
	grq.Var("appName", app.Name.ValueString())

	var fr fly.Query
	if err := r.client.Run(ctx, grq, &fr); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &app)...)
}

func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *appResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

// lookupOrgID looks up a Fly organization by name and returns the internal ID
func lookupOrgID(client *graphql.Client, name string) (string, error) {
	q := `
	query($slug: String!) {
		organization(slug: $slug) {
			id
		}
	}
`

	grq := graphql.NewRequest(q)
	grq.Var("slug", name)

	var ff fly.Query
	if err := client.Run(context.Background(), grq, &ff); err != nil {
		return "", err
	}

	return ff.Organization.ID, nil
}
