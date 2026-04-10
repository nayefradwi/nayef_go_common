package new

type VpsDockerComposeView struct {
	Name            string
	ShouldAddDb     bool
	ShouldAddRedis  bool
	ShouldAddSecret bool
	ShouldAddCaddy  bool
	Domain          string
	Environment     string
}

func newVpsDockerComposeView(req CreateNewProjectRequest, deployment DeploymentType, environment string) VpsDockerComposeView {
	return VpsDockerComposeView{
		Name:            req.Name,
		ShouldAddDb:     req.HasPostgres(),
		ShouldAddRedis:  req.HasRedis(),
		ShouldAddSecret: req.HasAuth(),
		ShouldAddCaddy:  deployment == DeploymentTypeVps,
		Domain:          req.Name,
		Environment:     environment,
	}
}

type BootstrapView struct {
	GoModule string
}

func newBootstrapView(req CreateNewProjectRequest) BootstrapView {
	return BootstrapView{GoModule: req.GoModule}
}

type EnvView struct {
	ShouldAddDb     bool
	ShouldAddRedis  bool
	ShouldAddSecret bool
}

func newEnvView(req CreateNewProjectRequest) EnvView {
	return EnvView{
		ShouldAddDb:     req.HasPostgres(),
		ShouldAddRedis:  req.HasRedis(),
		ShouldAddSecret: req.HasAuth(),
	}
}

type ConfigView struct {
	ShouldAddDb     bool
	ShouldAddRedis  bool
	ShouldAddSecret bool
}

func newConfigView(req CreateNewProjectRequest) ConfigView {
	return ConfigView{
		ShouldAddDb:     req.HasPostgres(),
		ShouldAddRedis:  req.HasRedis(),
		ShouldAddSecret: req.HasAuth(),
	}
}

type DiView struct {
	ShouldAddDb     bool
	ShouldAddRedis  bool
	ShouldAddSecret bool
	Imports         []string
}

func newDiView(req CreateNewProjectRequest) DiView {
	imports := []string{"context", req.GoModule + "/" + INTERNAL + "/" + CONFIG}
	if req.HasPostgres() {
		imports = append(imports, PGUTIL, PGX+"/"+"pgxpool")
	}

	if req.HasRedis() {
		imports = append(imports, REDISUTIL, REDIS)
	}
	return DiView{
		Imports:         imports,
		ShouldAddDb:     req.HasPostgres(),
		ShouldAddRedis:  req.HasRedis(),
		ShouldAddSecret: req.HasAuth(),
	}
}

type LocalDockerComposeView struct {
	Name            string
	ShouldAddDb     bool
	ShouldAddRedis  bool
	ShouldAddSecret bool
}

func newLocalDockerComposeView(req CreateNewProjectRequest) LocalDockerComposeView {
	return LocalDockerComposeView{
		Name:            req.Name,
		ShouldAddDb:     req.HasPostgres(),
		ShouldAddRedis:  req.HasRedis(),
		ShouldAddSecret: req.HasAuth(),
	}
}

type HealthView struct {
	IsRest bool
}

func newHealthView(req CreateNewProjectRequest) HealthView {
	return HealthView{IsRest: req.ServiceType == ServiceTypeRest}
}

type LocalEnvView struct {
	ShouldAddSecret bool
}

func newLocalEnvView(req CreateNewProjectRequest) LocalEnvView {
	return LocalEnvView{ShouldAddSecret: req.HasAuth()}
}

type RouterView struct {
	IsRest        bool
	HasPagination bool
	GoModule      string
}

func newRouterView(req CreateNewProjectRequest) RouterView {
	return RouterView{
		IsRest:        req.ServiceType == ServiceTypeRest,
		HasPagination: req.HasFeature(FeaturePagination),
		GoModule:      req.GoModule,
	}
}
