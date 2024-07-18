package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

const (
	// DebugMode indicates service mode is debug.
	DebugMode = "debug"
	// TestMode indicates service mode is test.
	TestMode = "test"
	// ReleaseMode indicates service mode is release.
	ReleaseMode = "release"
)

type Config struct {
	ServiceName string
	ServiceHost string
	ServicePort string

	Environment string // debug, test, release
	Version     string

	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string

	AuthServiceHost string
	AuthGRPCPort    string

	CompanyServiceHost string
	CompanyServicePort string

	NodeType     string
	K8sNamespace string

	MinioHost        string
	MinioAccessKeyID string
	MinioSecretKey   string

	PostgresMaxConnections int32
}

// Load ...
func Load() Config {
	if err := godotenv.Load("/app/.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			fmt.Println("No .env file found")
		}
		fmt.Println("No .env file found")
	}

	config := Config{}

	config.ServiceName = cast.ToString(getOrReturnDefaultValue("SERVICE_NAME", "ucode"))
	config.ServiceHost = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_SERVICE_HOST", "localhost"))
	config.ServicePort = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_SERVICE_PORT", ":7107"))

	config.Environment = cast.ToString(getOrReturnDefaultValue("ENVIRONMENT", DebugMode))
	config.Version = cast.ToString(getOrReturnDefaultValue("VERSION", "1.0"))

	config.PostgresHost = "65.109.239.69"
	config.PostgresPort = 5432
	config.PostgresUser = "udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs"
	config.PostgresPassword = "599Xx3nma8"
	config.PostgresDatabase = "udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs"

	config.AuthServiceHost = cast.ToString(getOrReturnDefaultValue("AUTH_SERVICE_HOST", "localhost"))
	config.AuthGRPCPort = cast.ToString(getOrReturnDefaultValue("AUTH_GRPC_PORT", ":9103"))

	config.CompanyServiceHost = cast.ToString(getOrReturnDefaultValue("COMPANY_SERVICE_HOST", "localhost"))
	config.CompanyServicePort = cast.ToString(getOrReturnDefaultValue("COMPANY_GRPC_PORT", ":8092"))

	config.NodeType = cast.ToString(getOrReturnDefaultValue("NODE_TYPE", "LOW"))
	config.K8sNamespace = cast.ToString(getOrReturnDefaultValue("K8S_NAMESPACE", "cp-region-type-id"))

	config.MinioAccessKeyID = cast.ToString(getOrReturnDefaultValue("MINIO_ACCESS_KEY", "ongei0upha4DiaThioja6aip8dolai1o"))
	config.MinioSecretKey = cast.ToString(getOrReturnDefaultValue("MINIO_SECRET_KEY", "aew8aeheungohf7vaiphoh7Tusie2vei"))
	config.MinioHost = cast.ToString(getOrReturnDefaultValue("MINIO_ENDPOINT", "cdn.u-code.io"))

	config.PostgresMaxConnections = cast.ToInt32(getOrReturnDefaultValue("POSTGRES_MAX_CONNECTIONS", 500))

	return config
}

func getOrReturnDefaultValue(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
