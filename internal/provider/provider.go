package provider

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfp "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/superfly/graphql"
)

var _ tfp.Provider = &provider{}

type provider struct {
	configured bool
}

func New() tfp.Provider {
	return &provider{}
}

func (p *provider) Configure(ctx context.Context, req tfp.ConfigureRequest, resp *tfp.ConfigureResponse) {
	token := os.Getenv("FLY_API_TOKEN")

	h := http.Client{
		Timeout:   60 * time.Second,
		Transport: &Transport{UnderlyingTransport: http.DefaultTransport, Token: token, Ctx: ctx},
	}

	client := graphql.NewClient("https://api.fly.io/graphql", graphql.WithHTTPClient(&h))
	resp.DataSourceData = client
	resp.ResourceData = client

	// resp.Diagnostics.AddError("WTF", fmt.Sprintf("%T", resp.ResourceData))

	p.configured = true
}

func (p *provider) Metadata(_ context.Context, _ tfp.MetadataRequest, resp *tfp.MetadataResponse) {
	resp.TypeName = "fly"
}

func (p *provider) Schema(_ context.Context, _ tfp.SchemaRequest, resp *tfp.SchemaResponse) {
}

func (p *provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newAppDataSource,
	}
}

func (p *provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newSecretResource,
	}
}

type Transport struct {
	UnderlyingTransport http.RoundTripper
	Token               string
	Ctx                 context.Context
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)

	return t.UnderlyingTransport.RoundTrip(req)
}
