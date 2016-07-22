export LISTEN_ADDR=":4001"
export VILI_URI=http://localhost:4001
export BUILD_DIR=$HOME/go/src/github.com/airware/vili/public/build

export STATIC_LIVE_RELOAD=1

export ENVIRONMENTS="tools staging preprod prodtools prod"
export PROD_ENVS="prod prodtools"
export APPROVAL_ENVS="preprod tools"
export DEFAULT_ENV="staging"

export LOG_DEBUG=1

export REDIS_PORT=redis://localhost:6379
export REDIS_DB=8

export OKTA_ENTRYPOINT=https://airware.okta.com/app/entrypoint
export OKTA_ISSUER=http://www.okta.com/issuer
export OKTA_CERT="cert"
export OKTA_DOMAIN=airware.com

export GITHUB_TOKEN=token
export GITHUB_OWNER=airware
export GITHUB_REPO=loki
export GITHUB_CONTENTS_PATH="vili/conf/%s"
export GITHUB_ENVS_TOOLS_CONTENTSURL="https://api.github.com/repos/airware/loki/contents/k8s-tools/<%= path %>"
export GITHUB_ENVS_PRODTOOLS_CONTENTSURL="https://api.github.com/repos/airware/loki/contents/k8s-tools/<%= path %>"

# Set to either "registry" or "ecr"
export DOCKER_MODE=registry
export REGISTRY_BRANCH_DELIMITER="-"
export REGISTRY_NAMESPACE=airware # optional, omit for top-level repositories

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

export KUBE_TOOLS_URL=https://kubemasters-staging.airware.io
export KUBE_STAGING_URL=https://kubemasters-staging.airware.io
export KUBE_PREPROD_URL=https://kubemasters-staging.airware.io
export KUBE_PRODTOOLS_URL=https://kubemasters-prod.airware.io
export KUBE_PRODTOOLS_NAMESPACE=tools
export KUBE_PROD_URL=https://kubemasters-prod.airware.io
