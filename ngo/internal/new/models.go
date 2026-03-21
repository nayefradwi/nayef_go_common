package new

const (
	CMD        = "cmd"
	API        = "api"
	BOOTSTRAP  = "bootstrap"
	ROUTER     = "router"
	MAIN       = "main"
	INTERNAL   = "internal"
	CONFIG     = "config"
	DI         = "di"
	HEALTH     = "health"
	INFRA      = "infra"
	MIGRATIONS = "migrations"
	SQLC       = "sqlc"
	QUERIES    = "queries"
	WEB        = "web"
	BUILD      = "build"
	SERVICE    = "service"
	REQUESTS   = "requests"
	RESPONSES  = "responses"
	HANDLER    = "handler"
	ERRORS     = "errors"
	TEST       = "test"
	GO         = "go"
	YAML       = "yaml"
	ENV        = ".env"
	GITIGNORE  = ".gitignore"
	README     = "README.md"
	DOCKERFILE = "Dockerfile"
)

const (
	CHI           = "github.com/go-chi/chi/v5"
	PGX           = "github.com/jackc/pgx/v5"
	REDIS         = "github.com/redis/go-redis/v9"
	GODOTENV      = "github.com/joho/godotenv"
	GRPC          = "google.golang.org/grpc"
	PGUTIL        = "github.com/nayefradwi/nayef_go_common/pgutil"
	AUTH          = "github.com/nayefradwi/nayef_go_common/auth"
	REDISUTIL     = "github.com/nayefradwi/nayef_go_common/redisutil"
	HTTPUTIL      = "github.com/nayefradwi/nayef_go_common/httputil"
	GRPCUTIL      = "github.com/nayefradwi/nayef_go_common/grpcutil"
	VALIDATION    = "github.com/nayefradwi/nayef_go_common/validation"
	PAGINATION    = "github.com/nayefradwi/nayef_go_common/pagination"
	OTP           = "github.com/nayefradwi/nayef_go_common/otp"
	LOCKING       = "github.com/nayefradwi/nayef_go_common/locking"
	COMMON_ERRORS = "github.com/nayefradwi/nayef_go_common/errors"
	ERRORSPB      = "github.com/nayefradwi/nayef_go_common/errorspb"
)

type file struct {
	name      string
	extension string
}

type dir struct {
	name        string
	directories []dir
	files       []file
}

func (d dir) clone() dir {
	c := dir{name: d.name, files: make([]file, len(d.files))}
	copy(c.files, d.files)
	c.directories = make([]dir, len(d.directories))
	for i, sub := range d.directories {
		c.directories[i] = sub.clone()
	}
	return c
}

func (d *dir) addSubDir(path []string, node dir) {
	if len(path) == 0 {
		d.directories = append(d.directories, node)
		return
	}

	for i := range d.directories {
		if d.directories[i].name == path[0] {
			d.directories[i].addSubDir(path[1:], node)
			return
		}
	}
}

var (
	baseDirectories = dir{
		directories: []dir{
			{
				name: CMD,
				directories: []dir{
					{
						name: API,
						files: []file{
							{name: BOOTSTRAP, extension: GO},
							{name: ROUTER, extension: GO},
							{name: MAIN, extension: GO},
						},
					},
				},
			},
			{
				name: INTERNAL,
				directories: []dir{
					{name: CONFIG, files: []file{{name: CONFIG, extension: GO}}},
					{name: DI, files: []file{{name: DI, extension: GO}}},
					{name: HEALTH, files: []file{{name: HANDLER, extension: GO}}},
					{name: INFRA},
				},
			},
		},
	}

	basePackages = []string{
		GODOTENV,
		COMMON_ERRORS,
	}
)

var featureToPackage = map[Feature]string{
	FeatureLocking:    LOCKING,
	FeatureOtp:        OTP,
	FeaturePagination: PAGINATION,
}
