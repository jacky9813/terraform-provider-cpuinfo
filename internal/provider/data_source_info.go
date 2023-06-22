package provider

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	gopsutil_cpu "github.com/shirou/gopsutil/cpu"
	"golang.org/x/sys/cpu"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type cpuInfoDataSource struct{}

type cpuInfoDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Isa      types.String   `tfsdk:"isa"`
	Name     types.String   `tfsdk:"name"`
	Features []types.String `tfsdk:"features"`
}

var (
	_ datasource.DataSource = &cpuInfoDataSource{}
)

func NewCpuInfoDataSource() datasource.DataSource {
	return &cpuInfoDataSource{}
}

func (d *cpuInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_info"
}

func (d *cpuInfoDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get CPU information about the host that running Terraform.",
		Attributes: map[string]schema.Attribute{
			"isa": schema.StringAttribute{
				Computed:    true,
				Description: "The instruction set that CPU can use. amd64 and arm64 are common examples.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The model name of the CPU.",
			},
			"features": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The hardware features that CPU supports (SSE4, AVX, AES, etc.)",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *cpuInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state cpuInfoDataSourceModel

	cpu_stats, _ := gopsutil_cpu.Info()

	state.Isa = types.StringValue(runtime.GOARCH)
	state.Name = types.StringValue(cpu_stats[0].ModelName)
	state.Features = make([]basetypes.StringValue, 0)
	state.ID = types.StringValue("placeholder")

	cpu_features := reflect.ValueOf(make(chan struct{}))

	switch runtime.GOARCH {
	case "amd64", "386":
		cpu_features = reflect.ValueOf(cpu.X86)
	case "arm64":
		cpu_features = reflect.ValueOf(cpu.ARM64)
	case "arm":
		cpu_features = reflect.ValueOf(cpu.ARM)
	case "s390x":
		cpu_features = reflect.ValueOf(cpu.S390X)
	case "ppc64", "ppc64le":
		cpu_features = reflect.ValueOf(cpu.PPC64)
	case "mips64", "mips64le":
		cpu_features = reflect.ValueOf(cpu.MIPS64X)
	}

	cpu_features_fields := cpu_features.Type()
	for i := 0; i < cpu_features.NumField(); i++ {
		feature_name, is_feature_flag := strings.CutPrefix(cpu_features_fields.Field(i).Name, "Has")
		if is_feature_flag {
			has_feature := cpu_features.Field(i).Interface()
			if has_feature == true {
				state.Features = append(state.Features, types.StringValue(feature_name))
			}
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
