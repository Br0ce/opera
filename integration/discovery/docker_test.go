package discovery_test

import (
	"context"
	"net/url"
	"reflect"
	"testing"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/Br0ce/opera/pkg/db/inmem"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/tool/discovery/docker"
	"github.com/Br0ce/opera/pkg/transport"
)

func TestDiscovery_All(t *testing.T) {
	log := monitor.NewTestLogger(true)
	discoveryTracer := otel.Tracer("DockerDiscovery")

	shark, err := tool.MakeTool(
		tool.WithName("get_shark_warning"),
		tool.WithDescription("Get current shark warning level for the location"),
		tool.WithParameters(map[string]any{
			"location": map[string]any{
				"type": "string",
			},
		}, []string{"location"}),
		tool.WithAddr(url.URL{
			Scheme: "http",
			Host:   "shark:8080",
			Path:   "/",
		}),
	)
	if err != nil {
		t.Fatalf("make shark tool: %s", err.Error())
	}
	weather, err := tool.MakeTool(
		tool.WithName("get_weather"),
		tool.WithDescription("Get weather at the given location"),
		tool.WithParameters(map[string]any{
			"location": map[string]any{
				"type": "string",
			},
		}, []string{"location"}),
		tool.WithAddr(url.URL{
			Scheme: "http",
			Host:   "weather:8080",
			Path:   "/",
		}),
	)
	if err != nil {
		t.Fatalf("make shark tool: %s", err.Error())
	}

	tests := []struct {
		name string
		ctx  context.Context
		want map[string]tool.Tool
	}{
		{
			name: "find both dev tool services",
			ctx:  context.TODO(),
			want: map[string]tool.Tool{
				"get_shark_warning": shark,
				"get_weather":       weather,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trans := transport.NewHTTP(time.Second*5, log)
			di, err := docker.NewDiscovery(inmem.NewToolDB(), trans, discoveryTracer, log)
			if err != nil {
				t.Fatalf("Discovery.All() new test discovery: %s", err.Error())
			}

			err = di.Refresh(test.ctx)
			if err != nil {
				t.Fatalf("Discovery.All() refresh: %s", err.Error())
			}

			got := di.All(test.ctx)
			if len(got) != len(test.want) {
				t.Errorf("Discovery.All() len got = %v, len want %v", len(got), len(test.want))
			}

			for _, g := range got {
				w, ok := test.want[g.Name()]
				if !ok {
					t.Errorf("Discovery.All() find want %v", g.Name())
				}
				if !reflect.DeepEqual(g, w) {
					t.Errorf("Discovery.All() got = %+v, want %+v", g, w)
				}
			}
		})
	}
}

func TestDiscovery_Get(t *testing.T) {
	log := monitor.NewTestLogger(true)
	discoveryTracer := otel.Tracer("DockerDiscovery")

	shark, err := tool.MakeTool(
		tool.WithName("get_shark_warning"),
		tool.WithDescription("Get current shark warning level for the location"),
		tool.WithParameters(map[string]any{
			"location": map[string]any{
				"type": "string",
			},
		}, []string{"location"}),
		tool.WithAddr(url.URL{
			Scheme: "http",
			Host:   "shark:8080",
			Path:   "/",
		}),
	)
	if err != nil {
		t.Fatalf("make shark tool: %s", err.Error())
	}
	weather, err := tool.MakeTool(
		tool.WithName("get_weather"),
		tool.WithDescription("Get weather at the given location"),
		tool.WithParameters(map[string]any{
			"location": map[string]any{
				"type": "string",
			},
		}, []string{"location"}),
		tool.WithAddr(url.URL{
			Scheme: "http",
			Host:   "weather:8080",
			Path:   "/",
		}),
	)
	if err != nil {
		t.Fatalf("make shark tool: %s", err.Error())
	}

	tests := []struct {
		name    string
		ctx     context.Context
		want    tool.Tool
		wantErr bool
	}{
		{
			name:    "find shark",
			ctx:     context.TODO(),
			want:    shark,
			wantErr: false,
		},
		{
			name:    "find weather",
			ctx:     context.TODO(),
			want:    weather,
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trans := transport.NewHTTP(time.Second*5, log)
			di, err := docker.NewDiscovery(inmem.NewToolDB(), trans, discoveryTracer, log)
			if err != nil {
				t.Fatalf("Discovery.All() new test discovery: %s", err.Error())
			}

			err = di.Refresh(test.ctx)
			if err != nil {
				t.Fatalf("Discovery.All() refresh: %s", err.Error())
			}

			got, err := di.Get(test.ctx, test.want.Name())
			if (err != nil) != test.wantErr {
				t.Errorf("Discovery.All() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Discovery.All() got = %v, len want %v", got, test.want)
			}
		})
	}
}
