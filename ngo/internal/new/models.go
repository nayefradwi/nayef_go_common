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
	DEFAULT_MOD_PATH = "github.com/nayefradwi"
	TESTIFY          = "github.com/stretchr/testify"
	CHI              = "github.com/go-chi/chi/v5"
	PGX              = "github.com/jackc/pgx/v5"
	REDIS            = "github.com/redis/go-redis/v9"
	GODOTENV         = "github.com/joho/godotenv"
	GRPC             = "google.golang.org/grpc"
	PGUTIL           = "github.com/nayefradwi/nayef_go_common/pgutil"
	AUTH             = "github.com/nayefradwi/nayef_go_common/auth"
	REDISUTIL        = "github.com/nayefradwi/nayef_go_common/redisutil"
	HTTPUTIL         = "github.com/nayefradwi/nayef_go_common/httputil"
	GRPCUTIL         = "github.com/nayefradwi/nayef_go_common/grpcutil"
	VALIDATION       = "github.com/nayefradwi/nayef_go_common/validation"
	PAGINATION       = "github.com/nayefradwi/nayef_go_common/pagination"
	OTP              = "github.com/nayefradwi/nayef_go_common/otp"
	LOCKING          = "github.com/nayefradwi/nayef_go_common/locking"
	COMMON_ERRORS    = "github.com/nayefradwi/nayef_go_common/errors"
	ERRORSPB         = "github.com/nayefradwi/nayef_go_common/errorspb"
)

type File struct {
	Name      string
	Extension string
}

type Dir struct {
	Name        string
	Directories []Dir
	Files       []File
}

func (d Dir) clone() Dir {
	c := Dir{Name: d.Name, Files: make([]File, len(d.Files))}
	copy(c.Files, d.Files)
	c.Directories = make([]Dir, len(d.Directories))
	for i, sub := range d.Directories {
		c.Directories[i] = sub.clone()
	}
	return c
}

func (d *Dir) addSubDir(path []string, node Dir) {
	if len(path) == 0 {
		d.Directories = append(d.Directories, node)
		return
	}

	for i := range d.Directories {
		if d.Directories[i].Name == path[0] {
			d.Directories[i].addSubDir(path[1:], node)
			return
		}
	}
}

var (
	baseDirectories = Dir{
		Directories: []Dir{
			{
				Name: CMD,
				Directories: []Dir{
					{
						Name: API,
						Files: []File{
							{Name: BOOTSTRAP, Extension: GO},
							{Name: ROUTER, Extension: GO},
							{Name: MAIN, Extension: GO},
						},
					},
				},
			},
			{
				Name: INTERNAL,
				Directories: []Dir{
					{Name: CONFIG, Files: []File{{Name: CONFIG, Extension: GO}}},
					{Name: DI, Files: []File{{Name: DI, Extension: GO}}},
					{Name: HEALTH, Files: []File{{Name: HANDLER, Extension: GO}}},
					{Name: INFRA},
				},
			},
		},
	}

	basePackages = []string{
		GODOTENV,
		COMMON_ERRORS,
		TESTIFY,
	}
)

var featureToPackage = map[Feature]string{
	FeatureLocking:    LOCKING,
	FeatureOtp:        OTP,
	FeaturePagination: PAGINATION,
}
