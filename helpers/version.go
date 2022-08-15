package helpers

import (
	"fmt"
	"runtime"

	jibberJabber "github.com/cloudfoundry/jibber_jabber"
	. "github.com/klauspost/cpuid/v2"
)

// GitCommit returns the git commit that was compiled. This will be filled in by the compiler.
var GitCommit = "" //nolint:gochecknoglobals // We need this as a variable for setting the value at build time.

// Version returns the main version number that is being run at the moment.
var Version = "unstable" //nolint:gochecknoglobals // We need this as a variable for setting the value at build time.

// BuildDate returns the date the binary was built.
var BuildDate = "" //nolint:gochecknoglobals // We need this as a variable for setting the value at build time.

// GetGoVersion returns the version of the go runtime used to compile the binary.
func GetGoVersion() string {
	return runtime.Version()
}

// GetOsArch returns the os and arch used to build the binary.
func GetOsArch() string {
	return fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
}

func GetSystemLocale() string {
	userLocale, err := jibberJabber.DetectIETF()
	if err != nil {
		return "unknown"
	}

	return userLocale
}

func GetCPUName() string {
	return CPU.BrandName
}

func GetCPUCoreCount() int {
	return CPU.PhysicalCores
}
