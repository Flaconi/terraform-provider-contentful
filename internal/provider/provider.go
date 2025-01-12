package provider

import (
	"context"
	"github.com/flaconi/contentful-go"
	client2 "github.com/flaconi/contentful-go/pkgs/client"
	"github.com/flaconi/contentful-go/pkgs/util"
	"github.com/flaconi/terraform-provider-contentful/internal/resources/api_key"
	"github.com/flaconi/terraform-provider-contentful/internal/resources/app_definition"
	"github.com/flaconi/terraform-provider-contentful/internal/resources/app_installation"
	"github.com/flaconi/terraform-provider-contentful/internal/resources/contenttype"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

var (
	_ provider.Provider = &contentfulProvider{}
)

func New(version string, debug bool) provider.Provider {
	return &contentfulProvider{
		version: version,
		debug:   debug,
	}
}

type contentfulProvider struct {
	version string
	debug   bool
}

// Provider schema struct
type contentfulProviderModel struct {
	CmaToken       types.String `tfsdk:"cma_token"`
	OrganizationId types.String `tfsdk:"organization_id"`
}

func (c contentfulProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "contentful"
}

func (c contentfulProvider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cma_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Contentful Management API token",
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The organization ID",
			},
		},
	}
}

func (c contentfulProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config contentfulProviderModel

	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var cmaToken string

	if config.CmaToken.IsUnknown() || config.CmaToken.IsNull() {
		cmaToken = os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")
	} else {
		cmaToken = config.CmaToken.ValueString()
	}

	var organizationId string

	if config.OrganizationId.IsUnknown() || config.OrganizationId.IsNull() {
		organizationId = os.Getenv("CONTENTFUL_ORGANIZATION_ID")
	} else {
		organizationId = config.OrganizationId.ValueString()
	}

	cma := contentful.NewCMA(cmaToken)
	cma.SetOrganization(organizationId)

	debug := c.debug

	if os.Getenv("TF_LOG") != "" {
		debug = true
	}

	cma.Debug = debug

	client, err := contentful.NewCMAV2(client2.ClientConfig{
		Debug:     debug,
		UserAgent: util.ToPointer("terraform-provider-contentful"),
		Token:     cmaToken,
	})

	if err != nil {
		response.Diagnostics.AddError(
			"error during creation of cma client",
			err.Error(),
		)
		return
	}

	data := utils.ProviderData{
		Client:         cma,
		CMAClient:      client,
		OrganizationId: organizationId,
	}

	response.ResourceData = data
	response.DataSourceData = data
}

func (c contentfulProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (c contentfulProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		contenttype.NewContentTypeResource,
		app_definition.NewAppDefinitionResource,
		app_installation.NewAppInstallationResource,
		api_key.NewApiKeyResource,
	}
}
