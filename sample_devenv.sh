export LISTEN_ADDR=":4001"
export VILI_URI=http://localhost:4001
export BUILD_DIR=$HOME/go/src/github.com/viliproject/vili/public/build

export APP_CERT="TODO"
export APP_PRIVATE_KEY="TODO"

export STATIC_LIVE_RELOAD=1

export AUTH_SERVICE=basic
export BASIC_AUTH_USERS='username1:bcrypthash username2:bcrypthash'

# saml configuration instead of basic
# export AUTH_SERVICE=saml
# export SAML_METADATA_URL=https://acmeinc.okta.com/app/metadata-url

export ENVIRONMENTS="tools staging preprod prodtools prod"
export APPROVAL_PROD_ENVS="preprod prod tools prodtools"
export IGNORED_ENVS="myenv"
export DEFAULT_ENV="staging"

export STAGING_KUBECONFIG_PATH=$HOME/.kube/config

# kubernetes url configs instead of a kubeconfig file
# export KUBE_TOOLS_URL=https://kubemasters-staging.acme.com
# export KUBE_STAGING_URL=https://kubemasters-staging.acme.com
# export KUBE_PREPROD_URL=https://kubemasters-staging.acme.com
# export KUBE_PRODTOOLS_URL=https://kubemasters-prod.acme.com
# export KUBE_PRODTOOLS_NAMESPACE=tools
# export KUBE_PROD_URL=https://kubemasters-prod.acme.com

export LOG_DEBUG=1

export REDIS_PORT=redis://localhost:6379
export REDIS_DB=8

export HARDCODED_TOKEN_USERS="mytoken myusername"

export GITHUB_TOKEN=token
export GITHUB_OWNER=viliproject
export GITHUB_REPO=vili
export GITHUB_CONTENTS_PATH="vili/conf/%s"
export GITHUB_ENVS_TOOLS_CONTENTS_PATH="vili/toolsconf/%s"
export GITHUB_ENVS_PRODTOOLS_CONTENTS_PATH="vili/prodtoolsconf/%s"

# Set to either "registry" or "ecr"
export DOCKER_MODE=registry
export REGISTRY_BRANCH_DELIMITER="-"
export REGISTRY_NAMESPACE=viliproject # optional, omit for top-level repositories

# registry mode
export REGISTRY_URL=https://quay.io
export REGISTRY_USERNAME=username
export REGISTRY_PASSWORD=password

# ecr mode
# export AWS_REGION=us-east-1
# export AWS_ACCESS_KEY_ID="accesskeyid"
# export AWS_SECRET_ACCESS_KEY="secretaccesskey"
# export ECR_ACCOUNT_ID=123456789012 # only needed if accessing another account's ecr registry

export FIREBASE_URL=https://test.firebaseio.com/
export FIREBASE_SECRET=secret

export SLACK_TOKEN=token
export SLACK_EMOJI=":party_parrot:"
export SLACK_CHANNEL="#slacktest"
export SLACK_USERNAME=vilibot
export SLACK_DEPLOY_USERNAMES="user1 user2"

export CI_PROVIDER="CI Name"
export CI_PROVIDER_TOKEN=token
