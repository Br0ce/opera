package inmem

import (
	"reflect"
	"sync"
	"testing"

	"github.com/Br0ce/opera/pkg/tool"
)

func TestTool_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tools   []tool.Tool
		wantErr bool
	}{
		{
			name:    "pass",
			tools:   tool.TestTools(),
			wantErr: false,
		},
		{
			name:    "already exists err",
			tools:   append(tool.TestTools(), tool.TestToolA()),
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			to := &Tool{}
			var wg sync.WaitGroup
			wg.Add(len(test.tools))

			var gotErr bool
			for _, item := range test.tools {
				go func(item tool.Tool) {
					defer wg.Done()
					err := to.Add(item)
					if err != nil {
						gotErr = true
					}
				}(item)
			}
			wg.Wait()

			if gotErr != test.wantErr {
				t.Errorf("Tool.Add() error = %v, wantErr %v", gotErr, test.wantErr)
			}
			if test.wantErr {
				return
			}
			founds := make([]bool, len(test.tools))
			to.tools.Range(func(key, value any) bool {
				got, ok := value.(tool.Tool)
				if !ok {
					t.Error("Tool.Add() tool invalid")
				}
				if got.Name() != key {
					t.Errorf("Tool.Add() key = %v, wantErr %v", key, got.Name())
				}
				for i, want := range test.tools {
					if want.Name() == key {
						if !reflect.DeepEqual(got, want) {
							t.Errorf("Tool.Add() = %v, want %v", got, want)
						}
						founds[i] = true
					}
				}
				return true
			})
			for _, found := range founds {
				if !found {
					t.Error("Tool.Add() not found")
				}
			}
		})
	}
}

func TestTool_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fields  []tool.Tool
		want    []tool.Tool
		wantErr bool
	}{
		{
			name:    "found",
			fields:  tool.TestTools(),
			want:    tool.TestTools(),
			wantErr: false,
		},
		{
			name:    "not found",
			fields:  []tool.Tool{tool.TestToolA()},
			want:    []tool.Tool{tool.TestToolB()},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			to := &Tool{}
			for _, f := range test.fields {
				to.tools.Store(f.Name(), f)
			}

			var wg sync.WaitGroup
			wg.Add(len(test.fields))

			var gotErr bool
			for _, w := range test.want {
				go func(want tool.Tool) {
					defer wg.Done()
					got, err := to.Get(want.Name())
					if err != nil {
						gotErr = true
						return
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("Tool.Get() = %v, want %v", got, want)
					}
				}(w)
			}
			wg.Wait()

			if gotErr != test.wantErr {
				t.Errorf("Tool.Add() error = %v, wantErr %v", gotErr, test.wantErr)
			}
		})
	}
}

func TestTool_All(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields []tool.Tool
		wants  []tool.Tool
	}{
		{
			name:   "found",
			fields: tool.TestTools(),
			wants:  tool.TestTools(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			to := &Tool{}
			for _, f := range test.fields {
				to.tools.Store(f.Name(), f)
			}

			var cnt int
			for got := range to.All() {
				for _, want := range test.wants {
					if want.Name() != got.Name() {
						continue
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("Tool.All() = %v, want %v", got, want)
					}
					cnt++
				}
			}
			if cnt != len(test.wants) {
				t.Errorf("Tool.All() fount cnt = %v, want %v", cnt, len(test.wants))
			}
		})
	}
}

func TestTool_Clear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields []tool.Tool
	}{
		{
			name:   "multiple",
			fields: tool.TestTools(),
		},
		{
			name:   "one tool",
			fields: []tool.Tool{tool.TestToolA()},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			to := &Tool{}
			for _, field := range test.fields {
				to.tools.Store(field.Name(), field)
			}
			to.Clear()
			var cnt int
			to.tools.Range(func(key, value any) bool {
				cnt++
				return true
			})
			if cnt != 0 {
				t.Errorf("Tool.Clear() fount cnt = %v, want %v", cnt, 0)
			}
		})
	}
}
