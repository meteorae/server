package helpers

import (
	"fmt"
	"runtime"

	"github.com/cloudfoundry/jibber_jabber"
	. "github.com/klauspost/cpuid/v2"
)

// GitCommit returns the git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version returns the main version number that is being run at the moment.
var Version = "unstable"

// BuildDate returns the date the binary was built.
var BuildDate = ""

// GoVersion returns the version of the go runtime used to compile the binary.
var GoVersion = runtime.Version()

// OsArch returns the os and arch used to build the binary.
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

var (
	CPUName      = CPU.BrandName
	CPUCoreCount = CPU.PhysicalCores
	SystemLocale = "en-US"
)

func init() {
	userLocale, err := jibber_jabber.DetectIETF()
	if err != nil {
		userLocale = "unknown"
	}

	SystemLocale = userLocale
}
