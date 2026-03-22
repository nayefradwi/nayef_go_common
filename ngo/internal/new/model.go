package new

type ServiceType string

const (
	ServiceTypeRest ServiceType = "REST"
	ServiceTypeGrpc ServiceType = "gRPC"
	ServiceTypeBoth ServiceType = "Both"
)

type AuthType string

const (
	AuthTypeJWT     AuthType = "JWT"
	AuthTypeRefresh AuthType = "JWT + Revokable Refresh"
	AuthTypeOpaque  AuthType = "Opaque Tokens"
	AuthTypeNone    AuthType = "Something Else"
)

type InfraType string

const (
	InfraTypePostgres InfraType = "postgresql"
	InfraTypeRedis    InfraType = "redis"
)

type Feature string

const (
	FeatureLocking    Feature = "Locking"
	FeatureOtp        Feature = "OTP"
	FeaturePagination Feature = "Pagination"
)

var AllFeatures = []Feature{
	FeatureLocking,
	FeatureOtp,
}

var AllInfraTypes = []InfraType{
	InfraTypePostgres,
	InfraTypeRedis,
}

type DBLibrary string

const (
	DBLibrarySqlc DBLibrary = "sqlc"
	DBLibraryNone DBLibrary = "Something Else"
)

type CreateNewProjectRequest struct {
	Name           string
	ServiceType    ServiceType
	AuthType       AuthType
	WithValidation bool
	Features       []Feature
	InfraTypes     []InfraType
	DBLibrary      DBLibrary
	HeadDir        Dir
	Packages       []string
	RootDirPath    string
	GoModule       string
}

type EnvTemplateInput struct {
	ShouldAddDb     bool
	ShouldAddRedis  bool
	ShouldAddSecret bool
}

type DiTemplateInput struct {
	Imports        []string
	ShouldAddDb    bool
	ShouldAddRedis bool
}

type File struct {
	Name      string
	Extension string
}

type Dir struct {
	Name        string
	Directories []Dir
	Files       []File
}

func (d Dir) Clone() Dir {
	c := Dir{Name: d.Name, Files: make([]File, len(d.Files))}
	copy(c.Files, d.Files)
	c.Directories = make([]Dir, len(d.Directories))
	for i, sub := range d.Directories {
		c.Directories[i] = sub.Clone()
	}
	return c
}

func (d *Dir) AddSubDir(path []string, node Dir) {
	if len(path) == 0 {
		d.Directories = append(d.Directories, node)
		return
	}

	for i := range d.Directories {
		if d.Directories[i].Name == path[0] {
			d.Directories[i].AddSubDir(path[1:], node)
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
