package cloud

import (
	"os"
	"reflect"
	"testing"

	cli "gopkg.in/urfave/cli.v1"
)

func TestWorkerOptions_DeepCopy(t *testing.T) {
	opts := NewWorkerOptions()
	dep := opts.DeepCopy()
	deep := &dep

	if !reflect.DeepEqual(opts, deep) {
		t.Errorf("WorkerOptions.DeepCopy() = %v, want %v", deep, opts)
	}
	if deep == opts {
		t.Errorf("WorkerOptions.DeepCopy() = %v, want %v", &deep, &opts)
	}

	// change it
	deep.Quota[ResourceLimitsCPU] = ZeroQuantity
	deep.Quota[ResourceLimitsMemory] = ZeroQuantity

	if reflect.DeepEqual(opts, deep) {
		t.Errorf("WorkerOptions.DeepCopy() = %v, want %v", deep, opts)
	}

}

func TestWorkerOptions_AddFlags(t *testing.T) {

	opts := NewWorkerOptions()
	app := &cli.App{
		Action: func(c *cli.Context) {
			tests := []struct {
				name string
				got  interface{}
				want interface{}
			}{
				{CycloneServer, opts.WorkerEnvs.CycloneServer, "http://127.0.0.1:7099"},
				{ConsoleWebEndpoint, opts.WorkerEnvs.ConsoleWebEndpoint, "http://127.0.0.1:3000"},
				{RegistryLocation, opts.WorkerEnvs.RegistryLocation, "cargo.caicloud.io"},
				{ClairDisable, opts.WorkerEnvs.ClairDisable, false},
				{GitlabURL, opts.WorkerEnvs.GitlabURL, "https://gitlab.com"},
				{WorkerImage, opts.WorkerEnvs.WorkerImage, "cargo.caicloud.io/caicloud/cyclone-worker"},
				{ResourceLimitsMemory, opts.Quota[ResourceLimitsMemory], DefaultLimitMemory},
				{ResourceLimitsCPU, opts.Quota[ResourceLimitsCPU], DefaultLimitCPU},
			}

			for _, tt := range tests {
				if !reflect.DeepEqual(tt.got, tt.want) {
					t.Errorf("WorkerOptions.AddFlags(): %s = %v, want %v", tt.name, tt.got, tt.want)
				}
			}
		},
	}
	opts.AddFlags(app)
	app.Run(os.Args)
}
