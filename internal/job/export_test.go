package job

import "github.com/osapi-io/gohai/pkg/gohai/source"

// SetDefaultHostnameProvider overrides the default hostname provider for testing.
func SetDefaultHostnameProvider(
	p HostnameProvider,
) {
	defaultHostnameProvider = p
}

// ResetDefaultHostnameProvider restores the default hostname provider.
func ResetDefaultHostnameProvider() {
	defaultHostnameProvider = gohaiHostnameProvider{}
}

// ExportNewGohaiHostnameProvider exposes the gohai hostname provider for
// testing.
func ExportNewGohaiHostnameProvider() HostnameProvider {
	return gohaiHostnameProvider{}
}

// ExportMachineName is the host name of src.
func ExportMachineName(
	src source.Sources,
) (string, error) {
	return machineName(src)
}
