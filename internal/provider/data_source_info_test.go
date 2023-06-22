package provider

import (
	"runtime"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	gopsutil_cpu "github.com/shirou/gopsutil/cpu"
)

func TestAccCpuInfoDataSource(t *testing.T) {
	cpu_stats, _ := gopsutil_cpu.Info()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "cpuinfo_info" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.cpuinfo_info.test", "isa", runtime.GOARCH),
					resource.TestCheckResourceAttr("data.cpuinfo_info.test", "name", cpu_stats[0].ModelName),
				),
			},
		},
	})
}
