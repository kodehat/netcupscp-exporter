package build

// Variables that can be injected during build.
var (

	// BuildTime time when the application was build.
	BuildTime string = "unknown"

	// CommitHash Git commit hash of the built application.
	CommitHash string = "unknown"

	// Version version of the built application.
	Version string = "dev"

	// GoVersion Go version the application was build with.
	GoVersion string = "unknown"
)
