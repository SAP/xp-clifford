package cli

import "testing"

func TestDefaultCLIConfigurationHasTimeoutParam(t *testing.T) {
	config := defaultCLIConfiguration()
	if !config.HasTimeoutParam {
		t.Fatal("expected timeout parameter to be enabled by default")
	}
}

func TestConfigureCLIAttachesTimeoutParam(t *testing.T) {
	originalConfig := Configuration.CLIConfiguration
	originalRootCommand := rootCommand
	defer func() {
		Configuration.CLIConfiguration = originalConfig
		rootCommand = originalRootCommand
	}()

	Configuration.CLIConfiguration = defaultCLIConfiguration()

	if err := configureCLI(); err != nil {
		t.Fatal(err)
	}

	flag := rootCommand.PersistentFlags().Lookup("timeout")
	if flag == nil {
		t.Fatal("expected timeout flag to be registered")
	}
	if flag.DefValue != "30s" {
		t.Fatalf("expected default timeout 30s, got %q", flag.DefValue)
	}
}

func TestConfigureCLICanDisableTimeoutParam(t *testing.T) {
	originalConfig := Configuration.CLIConfiguration
	originalRootCommand := rootCommand
	defer func() {
		Configuration.CLIConfiguration = originalConfig
		rootCommand = originalRootCommand
	}()

	config := defaultCLIConfiguration()
	config.HasTimeoutParam = false
	Configuration.CLIConfiguration = config

	if err := configureCLI(); err != nil {
		t.Fatal(err)
	}

	if flag := rootCommand.PersistentFlags().Lookup("timeout"); flag != nil {
		t.Fatal("expected timeout flag to be disabled")
	}
}
