package docker

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"slices"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
	mockDB "github.com/Br0ce/opera/pkg/tool/mock"
	mockTransport "github.com/Br0ce/opera/pkg/transport/mock"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type mockClient struct {
	client.APIClient
	containerListFn func(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	contListInvoked bool
	closeInvoked    bool
}

func (mc *mockClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	mc.contListInvoked = true
	return mc.containerListFn(ctx, options)
}

func (mc *mockClient) Close() error {
	mc.closeInvoked = true
	return nil
}

func testTool() tool.Tool {
	props := map[string]any{
		"location": map[string]any{
			"type": "string",
		},
	}
	req := []string{"location"}
	to, _ := tool.MakeTool(
		tool.WithName("myTool"),
		tool.WithAddr(url.URL{Host: "myHost"}),
		tool.WithDescription("my description"),
		tool.WithParameters(props, req),
	)
	return to
}

func TestDiscovery_Get(t *testing.T) {
	t.Parallel()

	ttool := testTool()
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name       string
		getFn      func(ctx context.Context, name string) (tool.Tool, error)
		getInvoked bool
		args       args
		want       tool.Tool
		wantErr    bool
	}{
		{
			name: "pass",
			getFn: func(ctx context.Context, name string) (tool.Tool, error) {
				if name != ttool.Name() {
					t.Fatalf("Discovery.Get() name = %v, want %v", name, ttool.Name())
				}
				return ttool, nil
			},
			getInvoked: true,
			args: args{
				ctx:  context.TODO(),
				name: ttool.Name(),
			},
			want:    ttool,
			wantErr: false,
		},
		{
			name: "db error",
			getFn: func(ctx context.Context, name string) (tool.Tool, error) {
				return tool.Tool{}, errors.New("some error")
			},
			getInvoked: true,
			args: args{
				ctx:  context.TODO(),
				name: "myName",
			},
			wantErr: true,
		},
	}

	log := monitor.NewTestLogger(false)
	ctx := context.TODO()
	shutdown, err := monitor.StartTestTracing(ctx, false, "")
	if err != nil {
		t.Fatalf("init test tracing")
	}
	defer func(ctx context.Context) {
		err := shutdown(ctx)
		if err != nil {
			t.Logf("shutdown tracing: %s", err.Error())
		}
	}(ctx)
	tr := otel.Tracer("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &mockDB.ToolDB{
				GetFn: test.getFn,
			}
			di := &Discovery{
				db:  db,
				tr:  tr,
				log: log,
			}
			got, err := di.Get(test.args.ctx, test.args.name)
			if (err != nil) != test.wantErr {
				t.Errorf("Discovery.Get() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Discovery.Get() = %v, want %v", got, test.want)
			}
			if test.getInvoked != db.GetInvoked {
				t.Errorf("Discovery.Get() getInvoked = %v, want %v", db.GetInvoked, test.getInvoked)
			}
		})
	}
}

func TestDiscovery_All(t *testing.T) {
	t.Parallel()

	props := map[string]any{
		"location": map[string]any{
			"type": "string",
		},
	}
	req := []string{"location"}
	t1, err := tool.MakeTool(
		tool.WithName("myTool"),
		tool.WithAddr(url.URL{Host: "myHost"}),
		tool.WithDescription("my description"),
		tool.WithParameters(props, req),
	)
	if err != nil {
		t.Fatalf("make test tool: %s", err.Error())
	}
	t2, err := tool.MakeTool(
		tool.WithName("myOtherTool"),
		tool.WithAddr(url.URL{Host: "myHost"}),
		tool.WithDescription("my description"),
		tool.WithParameters(props, req),
	)
	if err != nil {
		t.Fatalf("make other test tool: %s", err.Error())
	}

	tests := []struct {
		name       string
		ctx        context.Context
		allFn      func(ctx context.Context) ([]tool.Tool, error)
		allInvoked bool
		want       []tool.Tool
		wantErr    bool
	}{
		{
			name: "no tools present",
			allFn: func(ctx context.Context) ([]tool.Tool, error) {
				return []tool.Tool{}, nil
			},
			allInvoked: true,
			ctx:        context.TODO(),
			want:       []tool.Tool{},
			wantErr:    false,
		},
		{
			name: "db error",
			allFn: func(ctx context.Context) ([]tool.Tool, error) {
				return nil, errors.New("some error")
			},
			allInvoked: true,
			ctx:        context.TODO(),
			wantErr:    true,
		},
		{
			name: "pass",
			allFn: func(ctx context.Context) ([]tool.Tool, error) {
				return []tool.Tool{t1, t2}, nil
			},
			allInvoked: true,
			ctx:        context.TODO(),
			want:       []tool.Tool{t1, t2},
			wantErr:    false,
		},
	}

	log := monitor.NewTestLogger(false)
	ctx := context.TODO()
	shutdown, err := monitor.StartTestTracing(ctx, false, "")
	if err != nil {
		t.Fatalf("init test tracing")
	}
	defer func(ctx context.Context) {
		err := shutdown(ctx)
		if err != nil {
			t.Logf("shutdown tracing: %s", err.Error())
		}
	}(ctx)
	tr := otel.Tracer("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &mockDB.ToolDB{
				AllFn: test.allFn,
			}
			di := &Discovery{
				db:  db,
				tr:  tr,
				log: log,
			}
			got, err := di.All(test.ctx)
			if (err != nil) != test.wantErr {
				t.Errorf("Discovery.All() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Discovery.All() = %v, want %v", got, test.want)
			}
			if test.allInvoked != db.AllInvoked {
				t.Errorf("Discovery.All() allInvoked = %v, want %v", db.AllInvoked, test.allInvoked)
			}
		})
	}
}

func TestDiscovery_Refresh(t *testing.T) {

	tests := []struct {
		name             string
		ctx              context.Context
		clearFn          func(ctx context.Context) error
		clearInvoked     bool
		listContainersFn func(ctx context.Context, options container.ListOptions) ([]types.Container, error)
		contListInvoked  bool
		closeInvoked     bool
		getFn            func(ctx context.Context, addr string, header map[string][]string) ([]byte, error)
		getInvoked       bool
		addFn            func(ctx context.Context, tool tool.Tool) error
		addInvoked       bool
		wantErr          bool
	}{
		{
			name: "db clear error",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return errors.New("some error")
			},
			clearInvoked:    true,
			contListInvoked: false,
			closeInvoked:    false,
			getInvoked:      false,
			addInvoked:      false,
			wantErr:         true,
		},
		{
			name: "client error",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return nil, errors.New("some error")
			},
			contListInvoked: true,
			closeInvoked:    true,
			getInvoked:      false,
			addInvoked:      false,
			wantErr:         true,
		},
		{
			name: "client error",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return nil, errors.New("some error")
			},
			contListInvoked: true,
			closeInvoked:    true,
			getInvoked:      false,
			addInvoked:      false,
			wantErr:         true,
		},
		{
			name: "no tool containers",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{
					{
						Image: "cont1",
					},
					{
						Image: "cont2",
					},
				}, nil
			},
			contListInvoked: true,
			closeInvoked:    true,
			getInvoked:      false,
			addInvoked:      false,
			wantErr:         false,
		},
		{
			name: "transport error",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{
					{
						Image: "tool",
						Labels: map[string]string{
							name:         "myTool",
							host:         "myHost",
							port:         "8888",
							path:         "myPath",
							"otherLabel": "value",
						},
					},
					{
						Image: "cont2",
					},
				}, nil
			},
			contListInvoked: true,
			closeInvoked:    true,
			getFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				return nil, errors.New("some error")
			},
			getInvoked: true,
			addInvoked: false,
			wantErr:    true,
		},
		{
			name: "invalid config error",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{
					{
						Image: "tool",
						Labels: map[string]string{
							name:         "myTool",
							host:         "myHost",
							port:         "8888",
							path:         "myPath",
							"otherLabel": "value",
						},
					},
					{
						Image: "cont2",
					},
				}, nil
			},
			contListInvoked: true,
			closeInvoked:    true,
			getFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				waddr := "http://myHost:8888/myPath/config"
				if addr != waddr {
					t.Errorf("Discovery.Refresh() addr = %v, wantErr %v", addr, waddr)
				}
				c := config{
					Name:        "myTool",
					Description: "",
					Properties: map[string]any{
						"myparam": map[string]any{
							"type": "string",
						},
					},
					Required: []string{"myparam"},
				}
				bb, err := json.Marshal(c)
				if err != nil {
					t.Errorf("Discovery.Refresh() marshal config: %s", err.Error())
				}
				return bb, nil
			},
			getInvoked: true,
			addInvoked: false,
			wantErr:    true,
		},
		{
			name: "add tool error",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{
					{
						Image: "tool",
						Labels: map[string]string{
							name:         "myTool",
							host:         "myHost",
							port:         "8888",
							path:         "myPath",
							"otherLabel": "value",
						},
					},
					{
						Image: "cont2",
					},
				}, nil
			},
			contListInvoked: true,
			closeInvoked:    true,
			getFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				waddr := "http://myHost:8888/myPath/config"
				if addr != waddr {
					t.Errorf("Discovery.Refresh() addr = %v, wantErr %v", addr, waddr)
				}
				c := config{
					Name:        "myTool",
					Description: "my description",
					Properties: map[string]any{
						"myparam": map[string]any{
							"type": "string",
						},
					},
					Required: []string{"myparam"},
				}
				bb, err := json.Marshal(c)
				if err != nil {
					t.Errorf("Discovery.Refresh() marshal config: %s", err.Error())
				}
				return bb, nil
			},
			getInvoked: true,
			addFn: func(ctx context.Context, tool tool.Tool) error {
				return errors.New("some error")
			},
			addInvoked: true,
			wantErr:    true,
		},
		{
			name: "one tool",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{
					{
						Image: "tool",
						Labels: map[string]string{
							name:         "myTool",
							host:         "myHost",
							port:         "8888",
							path:         "myPath",
							"otherLabel": "value",
						},
					},
					{
						Image: "cont2",
					},
				}, nil
			},
			contListInvoked: true,
			closeInvoked:    true,
			getFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				waddr := "http://myHost:8888/myPath/config"
				if addr != waddr {
					t.Errorf("Discovery.Refresh() addr = %v, wantErr %v", addr, waddr)
				}
				c := config{
					Name:        "myTool",
					Description: "my description",
					Properties: map[string]any{
						"myparam": map[string]any{
							"type": "string",
						},
					},
					Required: []string{"myparam"},
				}
				bb, err := json.Marshal(c)
				if err != nil {
					t.Errorf("Discovery.Refresh() marshal config: %s", err.Error())
				}
				return bb, nil
			},
			getInvoked: true,
			addFn: func(ctx context.Context, to tool.Tool) error {
				wantTool, err := tool.MakeTool(
					tool.WithName("myTool"),
					tool.WithAddr(url.URL{Host: "myHost:8888", Scheme: "http", Path: "myPath"}),
					tool.WithDescription("my description"),
					tool.WithParameters(map[string]any{
						"myparam": map[string]any{
							"type": "string",
						},
					},
						[]string{"myparam"}))
				if err != nil {
					t.Errorf("Discovery.Refresh() make want tool: %s", err.Error())
				}
				if !reflect.DeepEqual(to, wantTool) {
					t.Errorf("Discovery.Refresh() add: got = %v want %v", to, wantTool)
				}
				return nil
			},
			addInvoked: true,
			wantErr:    false,
		},
		{
			name: "two tools",
			ctx:  context.TODO(),
			clearFn: func(ctx context.Context) error {
				return nil
			},
			clearInvoked: true,
			listContainersFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{
					{
						Image: "tool",
						Labels: map[string]string{
							name:         "myTool",
							host:         "myHost",
							port:         "8888",
							path:         "myPath",
							"otherLabel": "value",
						},
					},
					{
						Image: "otherTool",
						Labels: map[string]string{
							name:         "myOtherTool",
							host:         "myOtherHost",
							port:         "8888",
							path:         "myPath",
							"otherLabel": "value",
						},
					},
				}, nil
			},
			contListInvoked: true,
			closeInvoked:    true,
			getFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				waddr := []string{"http://myHost:8888/myPath/config", "http://myOtherHost:8888/myPath/config"}
				if !slices.Contains(waddr, addr) {
					t.Errorf("Discovery.Refresh() addr = %v, wantErr %v", addr, waddr)
				}
				c := config{
					Name:        "myTool",
					Description: "my description",
					Properties: map[string]any{
						"myparam": map[string]any{
							"type": "string",
						},
					},
					Required: []string{"myparam"},
				}
				bb, err := json.Marshal(c)
				if err != nil {
					t.Errorf("Discovery.Refresh() marshal config: %s", err.Error())
				}
				return bb, nil
			},
			getInvoked: true,
			addFn: func(ctx context.Context, to tool.Tool) error {
				return nil
			},
			addInvoked: true,
			wantErr:    false,
		},
	}

	log := monitor.NewTestLogger(false)
	ctx := context.TODO()
	shutdown, err := monitor.StartTestTracing(ctx, false, "")
	if err != nil {
		t.Fatalf("init test tracing")
	}
	defer func(ctx context.Context) {
		err := shutdown(ctx)
		if err != nil {
			t.Logf("shutdown tracing: %s", err.Error())
		}
	}(ctx)
	tr := otel.Tracer("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &mockDB.ToolDB{
				ClearFn: test.clearFn,
				AddFn:   test.addFn,
			}
			cli := &mockClient{
				containerListFn: test.listContainersFn,
			}
			trans := &mockTransport.Transporter{
				GetFn: test.getFn,
			}
			di := &Discovery{
				db:        db,
				client:    cli,
				transport: trans,
				tr:        tr,
				log:       log,
			}

			if err := di.Refresh(test.ctx); (err != nil) != test.wantErr {
				t.Errorf("Discovery.Refresh() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.clearInvoked != db.ClearInvoked {
				t.Errorf("Discovery.containers() clear invoked = %v, want %v", db.ClearInvoked, test.clearInvoked)
			}
			if test.contListInvoked != cli.contListInvoked {
				t.Errorf("Discovery.containers() containerList invoked = %v, want %v", cli.contListInvoked, test.contListInvoked)
			}
			if test.closeInvoked != cli.closeInvoked {
				t.Errorf("Discovery.containers() close invoked = %v, want %v", cli.closeInvoked, test.closeInvoked)
			}
			if test.getInvoked != trans.GetInvoked {
				t.Errorf("Discovery.containers() get invoked = %v, want %v", trans.GetInvoked, test.getInvoked)
			}
			if test.addInvoked != db.AddInvoked {
				t.Errorf("Discovery.containers() add invoked = %v, want %v", db.AddInvoked, test.addInvoked)
			}
		})
	}
}

func TestDiscovery_config(t *testing.T) {
	t.Parallel()

	testAddr := url.URL{
		Scheme: "https",
		Host:   "myHost:8888",
		Path:   "mypath",
	}
	wantConfig := config{
		Name:        "myTool",
		Description: "my description",
		Properties: map[string]any{
			"myparam": map[string]any{
				"type": "string",
			},
		},
		Required: []string{"myparam"},
	}

	tests := []struct {
		name    string
		getFn   func(ctx context.Context, addr string, header map[string][]string) ([]byte, error)
		ctx     context.Context
		addr    url.URL
		want    config
		wantErr bool
	}{
		{
			name: "pass",
			getFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				if spanContext := trace.SpanContextFromContext(ctx); !spanContext.IsValid() {
					t.Fatal("Discovery.config() spanContext invalid")
				}

				if addr != testAddr.JoinPath("config").String() {
					t.Fatalf("Discovery.config() addr got = %s want %s", addr, testAddr.String())
				}

				bb, err := json.Marshal(wantConfig)
				if err != nil {
					t.Fatalf("Discovery.config() marshal test config: %s", err.Error())
				}

				return bb, nil
			},
			ctx:     context.TODO(),
			addr:    testAddr,
			want:    wantConfig,
			wantErr: false,
		},
		{
			name: "transport fail",
			getFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				return nil, errors.New("some error")
			},
			ctx:     context.TODO(),
			wantErr: true,
		},
	}

	log := monitor.NewTestLogger(false)
	ctx := context.TODO()
	shutdown, err := monitor.StartTestTracing(ctx, false, "")
	if err != nil {
		t.Fatalf("init test tracing")
	}
	defer func(ctx context.Context) {
		err := shutdown(ctx)
		if err != nil {
			t.Logf("shutdown tracing: %s", err.Error())
		}
	}(ctx)
	tr := otel.Tracer("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := &mockTransport.Transporter{
				GetFn: test.getFn,
			}
			di := &Discovery{
				transport: transport,
				tr:        tr,
				log:       log,
			}

			got, err := di.config(test.ctx, test.addr)
			if (err != nil) != test.wantErr {
				t.Errorf("Discovery.config() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Discovery.config() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestDiscovery_toTool(t *testing.T) {
	t.Parallel()

	toolContainer := types.Container{
		Image: "myImage",
		Labels: map[string]string{
			name:         "myTool",
			host:         "myHost",
			port:         "8888",
			path:         "myPath",
			"otherLabel": "value",
		},
	}
	wantTool, err := tool.MakeTool(
		tool.WithName("myTool"),
		tool.WithAddr(url.URL{Host: "myHost:8888", Scheme: "http", Path: "myPath"}),
		tool.WithDescription("my description"),
		tool.WithParameters(map[string]any{
			"myparam": map[string]any{
				"type": "string",
			},
		},
			[]string{"myparam"}))
	if err != nil {
		t.Fatalf("test tool")
	}

	otherContainer := types.Container{
		Image: "otherImage",
		Labels: map[string]string{
			name: "myTool",
			host: "myHost",
			port: "8888",
		},
	}
	tests := []struct {
		name           string
		ctx            context.Context
		container      types.Container
		transportGetFn func(ctx context.Context, addr string, header map[string][]string) ([]byte, error)
		wantInvoked    bool
		wantErr        bool
		want           tool.Tool
	}{
		{
			name: "pass",
			transportGetFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				if spanContext := trace.SpanContextFromContext(ctx); !spanContext.IsValid() {
					t.Fatal("Discovery.toTool() spanContext invalid")
				}

				if addr != "http://myHost:8888/myPath/config" {
					t.Fatalf("Discovery.toTool() config addr = %s", addr)
				}

				cfg := config{
					Name:        "myTool",
					Description: "my description",
					Properties: map[string]any{
						"myparam": map[string]any{
							"type": "string",
						},
					},
					Required: []string{"myparam"},
				}
				bb, err := json.Marshal(cfg)
				if err != nil {
					t.Fatalf("Discovery.toTool() marshal test config: %s", err.Error())
				}

				return bb, nil
			},
			wantInvoked: true,
			ctx:         context.TODO(),
			container:   toolContainer,
			want:        wantTool,
			wantErr:     false,
		},
		{
			name:        "other container",
			wantInvoked: false,
			ctx:         context.TODO(),
			container:   otherContainer,
			wantErr:     true,
		},
		{
			name: "config error",
			transportGetFn: func(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
				return nil, errors.New("some config error")
			},
			wantInvoked: false,
			ctx:         context.TODO(),
			container:   toolContainer,
			wantErr:     true,
		},
	}

	log := monitor.NewTestLogger(false)
	ctx := context.TODO()
	shutdown, err := monitor.StartTestTracing(ctx, false, "")
	if err != nil {
		t.Fatalf("init test tracing")
	}
	defer func(ctx context.Context) {
		err := shutdown(ctx)
		if err != nil {
			t.Logf("shutdown tracing: %s", err.Error())
		}
	}(ctx)
	tr := otel.Tracer("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := &mockTransport.Transporter{
				GetFn: test.transportGetFn,
			}
			di := &Discovery{
				transport: transport,
				tr:        tr,
				log:       log,
			}
			got, err := di.toTool(test.ctx, test.container)
			if (err != nil) != test.wantErr {
				t.Errorf("Discovery.toTool() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Discovery.toTool() got = %v, want %v", got, test.want)
			}
			if transport.GetInvoked != test.wantInvoked {
				t.Errorf("Discovery.toTool() getInvoked = %v, wantInvoked %v", transport.GetInvoked, test.wantInvoked)
			}
		})
	}
}

func TestDiscovery_containers(t *testing.T) {
	cont1 := types.Container{
		Image: "myImage",
	}
	cont2 := types.Container{
		Image: "myOtherImage",
	}

	tests := []struct {
		name                string
		containerListFn     func(ctx context.Context, options container.ListOptions) ([]types.Container, error)
		wantContlistInvoked bool
		wantCloseInvoked    bool
		ctx                 context.Context
		want                []types.Container
		wantErr             bool
	}{
		{
			name: "empty list",
			containerListFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{}, nil
			},
			ctx:                 context.TODO(),
			want:                []types.Container{},
			wantContlistInvoked: true,
			wantCloseInvoked:    true,
			wantErr:             false,
		},
		{
			name: "pass",
			containerListFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return []types.Container{cont1, cont2}, nil
			},
			ctx:                 context.TODO(),
			want:                []types.Container{cont1, cont2},
			wantContlistInvoked: true,
			wantCloseInvoked:    true,
			wantErr:             false,
		},
		{
			name: "client error",
			containerListFn: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return nil, errors.New("some error")
			},
			ctx:                 context.TODO(),
			wantContlistInvoked: true,
			wantCloseInvoked:    true,
			wantErr:             true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli := &mockClient{
				containerListFn: test.containerListFn,
			}
			di := &Discovery{
				client: cli,
			}

			got, err := di.containers(test.ctx)
			if (err != nil) != test.wantErr {
				t.Errorf("Discovery.containers() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Discovery.containers() = %v, want %v", got, test.want)
			}
			if test.wantContlistInvoked != cli.contListInvoked {
				t.Errorf("Discovery.containers() containerList invoked = %v, want %v", cli.contListInvoked, test.wantContlistInvoked)
			}
			if test.wantCloseInvoked != cli.closeInvoked {
				t.Errorf("Discovery.containers() close invoked = %v, want %v", cli.closeInvoked, test.wantCloseInvoked)
			}

		})
	}
}
