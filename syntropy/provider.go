package syntropy

import (
	"context"
	"fmt"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"
	"os"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.Provider = &provider{}

type provider struct {
	client     *syntropy.APIClient
	configured bool
	version    string
	token      string
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"access_token": {
				MarkdownDescription: "Syntropy platform access token",
				Type:                types.StringType,
				Required:            true,
			},
			"api_url": {
				MarkdownDescription: "Syntropy platform API URL",
				Type:                types.StringType,
				Optional:            true,
			},
		},
	}, nil
}

type providerData struct {
	AccessToken types.String `tfsdk:"access_token"`
	ApiUrl      types.String `tfsdk:"api_url"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var accessToken string
	if config.AccessToken.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as access_token",
		)
		return
	}

	if config.AccessToken.Null {
		accessToken = os.Getenv("SYNTROPY_ACCESS_TOKEN")
	} else {
		accessToken = config.AccessToken.Value
	}

	var apiUrl string
	if config.ApiUrl.Unknown {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as api_url",
		)
		return
	}

	if config.ApiUrl.Null {
		apiUrl = os.Getenv("SYNTROPY_API_URL")
	} else {
		apiUrl = config.ApiUrl.Value
	}

	p.client = NewClient(ctx, accessToken, apiUrl)
	p.token = accessToken
	p.configured = true
}

func NewClient(ctx context.Context, accessKey, apiURL string) *syntropy.APIClient {
	cfg := syntropy.NewConfiguration()
	cfg.HTTPClient = http.DefaultClient

	if apiURL != "" {
		cfg.Servers = syntropy.ServerConfigurations{
			{
				URL: apiURL,
			},
		}
	}
	return syntropy.NewAPIClient(cfg)
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"syntropystack_network_connection_mesh": networkConnectionMeshResourceType{},
		"syntropystack_network_connection":      networkConnectionResourceType{},
		"syntropystack_agent":                   agentResourceType{},
	}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"syntropystack_agent":                      agentDataSourceType{},
		"syntropystack_agent_search":               agentSearchDataSourceType{},
		"syntropystack_network_connection_service": networkConnectionServiceDataSourceType{},
	}, nil
}

func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}
	return *p, diags
}

func (p *provider) createAuthContext(parent context.Context) context.Context {
	return context.WithValue(parent, syntropy.ContextAPIKeys, map[string]syntropy.APIKey{
		"accessToken": {
			Key: p.token,
		},
	})
}
