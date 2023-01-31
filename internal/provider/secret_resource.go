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
	_ resource.Resource                = &secretResource{}
	_ resource.ResourceWithConfigure   = &secretResource{}
	_ resource.ResourceWithImportState = &secretResource{}
)

type secretResource struct {
	client *graphql.Client
}

func newSecretResource() resource.Resource {
	return &secretResource{}
}

type secretResourceModel struct {
	AppName types.String `tfsdk:"app"`
	Name    types.String `tfsdk:"name"`
	Value   types.String `tfsdk:"value"`
}

func (r *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly secret",

		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Secret name",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Secret value",
				Required:            true,
			},
		},
	}
}

func (r *secretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var secret secretResourceModel

	diags := req.Plan.Get(ctx, &secret)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := `
	mutation($input: SetSecretsInput!) {
		setSecrets(input: $input) {
			release {
				id
			}
		}
	}
	`

	input := fly.SetSecretsInput{AppID: secret.AppName.ValueString()}
	input.Secrets = append(input.Secrets, fly.SetSecretsInputSecret{
		Key:   secret.Name.ValueString(),
		Value: secret.Value.ValueString(),
	})

	grq := graphql.NewRequest(query)
	grq.Var("input", input)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &secret)...)
}

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var secret secretResourceModel

	diags := req.State.Get(ctx, &secret)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := `
		query ($appName: String!) {
			app(name: $appName) {
				secrets(name: $secretName) {
					name
				}
			}
		}
	`

	grq := graphql.NewRequest(query)
	grq.Var("appName", secret.AppName.ValueString())
	grq.Var("secretName", secret.Name.ValueString())

	var fr fly.Query
	if err := r.client.Run(ctx, grq, &fr); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	// FIX: validate that secret was found
	s := fr.App.Secrets[0]

	secret.Name = types.StringValue(s.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &secret)...)
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
