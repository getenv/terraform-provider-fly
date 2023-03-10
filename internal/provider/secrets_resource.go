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

type secretsResource struct {
	client *graphql.Client
}

func newSecretsResource() resource.Resource {
	return &secretsResource{}
}

type secretsResourceModel struct {
	AppName types.String `tfsdk:"app"`
	Secrets types.Map    `tfsdk:"secrets"`
}

func (r *secretsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secrets"
}

func (r *secretsResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly secrets",

		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "App name",
				Required:            true,
			},
			"secrets": schema.MapAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *secretsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var secrets secretsResourceModel

	diags := req.Plan.Get(ctx, &secrets)
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

	input := fly.SetSecretsInput{AppID: secrets.AppName.ValueString()}

	var secretKvs = make(map[string]string)
	secrets.Secrets.ElementsAs(ctx, &secretKvs, false)

	for k, v := range secretKvs {
		input.Secrets = append(input.Secrets, fly.SetSecretsInputSecret{
			Key:   k,
			Value: v,
		})
	}

	grq := graphql.NewRequest(query)
	grq.Var("input", input)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &secrets)...)
}

func (r *secretsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var secrets secretsResourceModel

	diags := req.State.Get(ctx, &secrets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := `
		query ($appName: String!) {
			app(name: $appName) {
				secrets {
					name
					digest
				}
			}
		}
	`

	grq := graphql.NewRequest(q)
	grq.Var("appName", secrets.AppName.ValueString())

	var fq fly.Query
	if err := r.client.Run(context.Background(), grq, &fq); err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
	}

	// FIX: set secrets kv pairs

	resp.Diagnostics.Append(resp.State.Set(ctx, secrets)...)

}

func (r *secretsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var secrets secretsResourceModel

	diags := req.Plan.Get(ctx, &secrets)
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

	input := fly.SetSecretsInput{AppID: secrets.AppName.ValueString()}

	var secretKvs = make(map[string]string)
	secrets.Secrets.ElementsAs(ctx, &secretKvs, false)

	for k, v := range secretKvs {
		input.Secrets = append(input.Secrets, fly.SetSecretsInputSecret{
			Key:   k,
			Value: v,
		})
	}

	grq := graphql.NewRequest(query)
	grq.Var("input", input)

	var ff fly.Query
	if err := r.client.Run(context.Background(), grq, &ff); err != nil {
		resp.Diagnostics.AddError("Query failed", "client interaction:"+err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &secrets)...)
}

func (r *secretsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var secrets secretsResourceModel

	diags := req.State.Get(ctx, &secrets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := `
	mutation($input: UnsetSecretsInput!) {
		unsetSecrets(input: $input) {
			release {
				id
			}
		}
	}
	`

	var secretKvs = make(map[string]string)
	secrets.Secrets.ElementsAs(ctx, &secretKvs, false)

	keys := make([]string, 0, len(secretKvs))
	for k := range secretKvs {
		keys = append(keys, k)
	}

	input := fly.UnsetSecretsInput{AppID: secrets.AppName.ValueString(), Keys: keys}

	grq := graphql.NewRequest(q)
	grq.Var("appName", secrets.AppName.ValueString())
	grq.Var("input", input)

	var fq fly.Query
	if err := r.client.Run(context.Background(), grq, &fq); err != nil {
		resp.Diagnostics.AddError("Query failed on destroy", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, secrets)...)
}

func (r *secretsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
