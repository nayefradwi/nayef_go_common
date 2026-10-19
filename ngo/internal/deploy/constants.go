package deploy

const (
	DEPLOYMENTS = "deployments"
	TERRAFORM   = "terraform"
	GITHUB      = ".github"
	WORKFLOWS   = "workflows"
)

const (
	PROVIDER_AWS   = "aws"
	PROVIDER_GCP   = "gcp"
	PROVIDER_AZURE = "azure"
	PROVIDER_VPS   = "vps"
)

const (
	TF_SMALL  = "small.tf.tmpl"
	TF_MEDIUM = "medium.tf.tmpl"
	TF_LARGE  = "large.tf.tmpl"
	TF_VARS   = "terraform.tfvars.tmpl"
)

const (
	TMPL_DOCKER_COMPOSE = "docker-compose.yaml.tmpl"
	TMPL_DEPLOY         = "deploy.yml.tmpl"
	TMPL_TODO           = "todo.md.tmpl"
	TMPL_DEPLOY_SCRIPT  = "deploy.sh.tmpl"
	TMPL_RUNNER_INIT    = "runner-init.sh.tmpl"
)

// Azure container app and job names cap at 32 chars, and the job adds "-migrate"
const MAX_RESOURCE_NAME = 24
