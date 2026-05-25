package configapp

import (
	"testing"

	"github.com/nobuenhombre/suikat/pkg/fico"
)

func TestConfigLoadSave(t *testing.T) {
	fileName := "config-app_test_save.yaml"

	cfg := &Config{}
	err := cfg.Load("config-app_test_load.yaml")
	if err != nil {
		t.Fatalf("cfg.Load error: %v", err)
	}

	err = cfg.Save(fileName)
	if err != nil {
		t.Fatalf("cfg.Save error: %v", err)
	}

	txtConfigFile := fico.TxtFile(fileName)
	_, err = txtConfigFile.Read()
	if err != nil {
		t.Fatalf("Read saved file error: %v", err)
	}
}
