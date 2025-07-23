package constants

// Kubernetes configuration paths
const (
	// KubeConfigDir is the standard directory name for Kubernetes configuration
	KubeConfigDir = ".kube"

	// KubeConfigFile is the standard filename for Kubernetes configuration
	KubeConfigFile = "config"
)

// LazyOC application paths
const (
	// LazyOCConfigDir is the directory name for LazyOC configuration
	LazyOCConfigDir = ".lazyoc"

	// ConfigFileName is the filename for LazyOC configuration
	ConfigFileName = "config.json"

	// LogFileName is the default log file name
	LogFileName = "lazyoc.log"

	// LogFilePermissions defines the permissions for log files
	LogFilePermissions = 0666
)