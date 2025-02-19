package tests

import (
	"github.com/faelmori/kbx/mods/logz"
	"github.com/goccy/go-json"
	"os"
	"testing"
)

type Opts struct {
	ConfigPath   string `json:"configPath"`
	OutputPath   string `json:"outputPath"`
	OutputFormat string `json:"outputFormat"`
	NeedCheck    bool   `json:"needCheck"`
	CheckMethod  string `json:"checkMethod"`
}

var OptsImpl *Opts

func loadOpts() (*Opts, error) {
	if OptsImpl == nil {
		file, fileErr := os.ReadFile("./config_test.json")
		if fileErr != nil {
			return nil, logz.ErrorLog("erro ao ler arquivo de configuração: "+fileErr.Error(), "Getl")
		}
		OptsImplErr := json.Unmarshal(file, &OptsImpl)
		if OptsImplErr != nil {
			return nil, logz.ErrorLog("erro ao decodificar arquivo de configuração: "+OptsImplErr.Error(), "Getl")
		}
	}
	return OptsImpl, nil
}

func TestSyncProducts(t *testing.T) {
	opts, optsErr := loadOpts()
	if optsErr != nil {
		t.Fatalf("SyncProducts failed: %v", optsErr)
	}
	tests := []struct {
		name string
		opts *Opts
	}{
		{
			name: "TestSyncProducts",
			opts: opts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecuteETL(tt.opts.ConfigPath, tt.opts.OutputPath, tt.opts.OutputFormat, tt.opts.NeedCheck, tt.opts.CheckMethod)
			if err != nil {
				t.Fatalf("SyncProducts failed: %v", err)
			}
		})
	}
}

func BenchmarkSyncProducts(b *testing.B) {
	opts, optsErr := loadOpts()
	if optsErr != nil {
		b.Fatalf("SyncProducts failed: %v", optsErr)
	}
	OptsImpl = opts
	for i := 0; i < b.N; i++ {
		err := ExecuteETL(OptsImpl.ConfigPath, OptsImpl.OutputPath, OptsImpl.OutputFormat, OptsImpl.NeedCheck, OptsImpl.CheckMethod)
		if err != nil {
			b.Fatalf("SyncProducts failed: %v", err)
		}
	}
}
