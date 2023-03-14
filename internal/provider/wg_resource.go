package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fly "github.com/superfly/flyctl/api"
	"github.com/superfly/graphql"
	"golang.org/x/crypto/curve25519"
)

var (
	_ resource.Resource                = &wgResource{}
	_ resource.ResourceWithConfigure   = &wgResource{}
	_ resource.ResourceWithImportState = &wgResource{}
)

type wgResource struct {
	client *graphql.Client
}

func newWGResource() resource.Resource {
	return &wgResource{}
}

type wgResourceModel struct {
	Name   types.String `tfsdk:"name"`
	Org    types.String `tfsdk:"org"`
	Region types.String `tfsdk:"region"`
}

func (r *wgResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wg"
}

func (r *wgResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly Wireguard Resources",

		Attributes: map[string]schema.Attribute{
			// required
			"org": schema.StringAttribute{
				MarkdownDescription: "Org name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Wireguard peer name",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Deployment region",
				Required:            true,
			},

			// computed
			"endpoint_ip": schema.StringAttribute{
				MarkdownDescription: "The Wireguard endpoint IP",
				Computed:            true,
			},
			"peer_ip": schema.StringAttribute{
				MarkdownDescription: "The Wireguard peer IP",
				Computed:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "The Wireguard private key",
				Computed:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "The Wireguard public key",
				Computed:            true,
			},
		},
	}
}

func (r *wgResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var wg wgResourceModel

	diags := req.Plan.Get(ctx, &wg)

	data, err := json.Marshal(diags)
	if err != nil {
		resp.Diagnostics.AddError("unmarshall error", err.Error())
	}

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Query failed appending wg details to diagnostics", string(data))
		return
	}

	pubkey, _ := C25519pair()

	query := `
	mutation($input: AddWireGuardPeerInput!) {
		addWireGuardPeer(input: $input) {
			peerip
			endpointip
			pubkey
		}
	}
	`

	inputs := map[string]interface{}{
		"name":           wg.Name,
		"organizationId": wg.Org,
		"pubkey":         pubkey,
		"region":         wg.Region,
	}

	grq := graphql.NewRequest(query)
	grq.Var("input", inputs)

	var wgp fly.CreatedWireGuardPeer
	if err := r.client.Run(context.Background(), grq, &wgp); err != nil {
		resp.Diagnostics.AddError("Query failed creating a Wireguard peer", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &wg)...)
}

func (r *wgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var wg wgResourceModel
	diags := req.State.Get(ctx, &wg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := `
	`

	grq := graphql.NewRequest(q)

	var fq fly.Query
	if err := r.client.Run(context.Background(), grq, &fq); err != nil {
		resp.Diagnostics.AddError("Query failed fetching Read", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, wg)...)
}

func (r *wgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *wgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *wgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

func C25519pair() (string, string) {
	var private [32]byte
	_, err := rand.Read(private[:])
	if err != nil {
		panic(fmt.Sprintf("reading from random: %s", err))
	}

	public, err := curve25519.X25519(private[:], curve25519.Basepoint)
	if err != nil {
		panic(fmt.Sprintf("can't mult: %s", err))
	}

	return base64.StdEncoding.EncodeToString(public),
		base64.StdEncoding.EncodeToString(private[:])
}
