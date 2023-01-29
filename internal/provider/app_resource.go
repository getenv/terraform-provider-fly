package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &app{}
	_ resource.ResourceWithConfigure   = &app{}
	_ resource.ResourceWithImportState = &app{}
)

type app struct {
	Name types.String `tfsdk:"name"`
}

func (a *app) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (a *app) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {}

func (a *app) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

func (a *app) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (a *app) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (a *app) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (a *app) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (a *app) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
