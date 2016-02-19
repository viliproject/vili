export VILI_URI=http://localhost:4001
export VILI_ENVS=tools,staging,preprod,prodtools,prod

export VILI_SESSION_SECRET=looselipssinkships

export VILI_ROBOTS_TEAMCITY=test

export VILI_REDIS_DB=8

export VILI_OKTA_ENTRYPOINT=https://airware.okta.com/app/entrypoint
export VILI_OKTA_ISSUER=http://www.okta.com/issuer
export VILI_OKTA_CERT="cert"
export VILI_OKTA_DOMAIN=airware.com

export VILI_GITHUB_TOKEN=token
export VILI_GITHUB_CONTENTSURL="https://api.github.com/repos/airware/loki/contents/k8s/<%= path %>"
export VILI_GITHUB_ENVS_TOOLS_CONTENTSURL="https://api.github.com/repos/airware/loki/contents/k8s-tools/<%= path %>"
export VILI_GITHUB_ENVS_PRODTOOLS_CONTENTSURL="https://api.github.com/repos/airware/loki/contents/k8s-tools/<%= path %>"

export VILI_QUAY_TOKEN=token
export VILI_QUAY_NAMESPACE=airware

export VILI_KUBE_TOOLS_URL=https://kubemasters-staging.airware.io
export VILI_KUBE_STAGING_URL=https://kubemasters-staging.airware.io
export VILI_KUBE_PREPROD_URL=https://kubemasters-staging.airware.io
export VILI_KUBE_PRODTOOLS_URL=https://kubemasters-prod.airware.io
export VILI_KUBE_PRODTOOLS_NAMESPACE=tools
export VILI_KUBE_PROD_URL=https://kubemasters-prod.airware.io

export VILI_FIREBASE_URL=https://test.firebaseio.com/
export VILI_FIREBASE_SECRET=secret

export VILI_SLACK_URI=https://hooks.slack.com/services/myservice
export VILI_SLACK_CHANNEL="#slacktest"
