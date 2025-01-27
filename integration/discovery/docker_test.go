package discovery_test

import (
	"context"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/tool/db/inmem"
	"github.com/Br0ce/opera/pkg/tool/discovery/docker"
	"github.com/Br0ce/opera/pkg/transport"
)

func TestDiscovery_All(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    map[string]tool.Tool
		wantErr bool
	}{
		{
			name: "find both dev tool services",
			ctx:  context.TODO(),
			want: map[string]tool.Tool{
				"get_shark_warning": {
					Name:        "get_shark_warning",
					Description: "Get current shark warning level for the location",
					Parameters: tool.Parameters{
						Properties: map[string]any{
							"location": map[string]any{
								"type": "string",
							},
						},
						Required: []string{"location"},
					},
					Addr: url.URL{
						Scheme: "http",
						Host:   "shark:8080",
						Path:   "/",
					},
				},
				"get_weather": {
					Name:        "get_weather",
					Description: "Get weather at the given location",
					Parameters: tool.Parameters{
						Properties: map[string]any{
							"location": map[string]any{
								"type": "string",
							},
						},
						Required: []string{"location"},
					},
					Addr: url.URL{
						Scheme: "http",
						Host:   "weather:8080",
						Path:   "/",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			log := monitor.NewTestLogger(true)
			trans := transport.NewHTTP(time.Second*5, log)
			db := inmem.NewDB(log)
			di := docker.NewDiscovery(db, trans, log)

			got, err := di.All(test.ctx)
			if (err != nil) != test.wantErr {
				t.Errorf("Discovery.All() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if len(got) != len(test.want) {
				t.Errorf("Discovery.All() len got = %v, len want %v", len(got), len(test.want))
			}

			for _, g := range got {
				w, ok := test.want[g.Name]
				if !ok {
					t.Errorf("Discovery.All() find want %v", g.Name)
				}
				if !reflect.DeepEqual(g, w) {
					t.Errorf("Discovery.All() got = %+v, want %+v", g, w)
				}
			}
			// for i, g := range got {
			// 	fmt.Println(g)
			// 	fmt.Println(test.want[i])
			// }
		})
	}
}
