# Vili

[![Slack Status](https://vili-slackin.herokuapp.com/badge.svg)](https://vili-slackin.herokuapp.com/)

Vili is an open source dashboard for managing deployments to a [Kubernetes] (http://kubernetes.io/) cluster. It is built to:

- Manage both manual and continuous deployments
- Gate production deployments through our QA process
- Provide transparency into the current state of our infrastructure


### Is Vili right for you?

Vili is opinionated, to be able to set it up you need to:

- use GitHub for version control
- build Docker images to ship code and tag them with the Git SHA and branch they were built from
- push these Docker images to Quay.io
- use Kubernetes namespaces to manage environments
- use Kubernetes replication controllers to deploy applications
- use Slack for team messaging

### What does Vili mean anyway?

Vili is a brother of Odin in Norse mythology, and he gives intelligence to the first humans.

<hr>

## Setup
To setup Vili on your Kubernetes cluster, follow these steps:

1. Select a domain name to host Vili under, such as vili.mydomain.com. Create an [Okta](https://www.okta.com/) app with a redirect URL that points to vili.mydomain.com/login/callback. Write down the Okta entrypoint and the certificate.
2. Create [Quay.io](https://quay.io/) repositories for your applications. Also create a Quay.io API application and generate a bearer token following instructions [here](http://docs.quay.io/api/). Write down your Quay organization or user name and your bearer token.
3. Create a new [Firebase](https://www.firebase.com/) app. Set the "Firebase rules" to match [this](docs/installation/firebase_security.json). Write down the Firebase app's URL and secret.
4. Create a GitHub repo with a directory that holds your replication controller templates, pod templates, and environment variables following this [example](docs/examples/github). Also create a GitHub access token following instructions [here](https://help.github.com/articles/creating-an-access-token-for-command-line-use/). Write down your GitHub organization or user name, the path to the directory created above and the authentication token.
5. Create a Slack [bot integration](https://api.slack.com/bot-users). Write down the API token from the integration settings page.
6. Create a [secret](http://kubernetes.io/v1.1/docs/user-guide/secrets.html) in your Kubernetes cluster that stores your GitHub, Quay, Firebase, and Slack tokens, and your Okta certificate following this [example](https://github.com/airware/vili/blob/master/docs/examples/simple/secret.yaml). Populate the values in the secret using the Okta, Quay.io, GitHub, Firebase, and Slack information you wrote down in the previous steps. Don't forget to base64 encode them as required by Kubernetes!
7. Create a replication controller in your Kubernetes cluster following this [example](docs/examples/simple/controller.yaml). Populate the environment variables using the Okta, Quay.io, GitHub, Firebase, and Slack information you wrote down in the previous steps.
8. Create a [service](http://kubernetes.io/v1.1/docs/user-guide/services.html) for this replication controller, and allow external access to this service under the domain name you chose in step 1.

You are all set! Vili will use the GitHub and Quay.io APIs to discover your apps and help you deploy them.

## Concepts

[Environment] (docs/environments.md): A namespace in Kubernetes that runs an isolated set of apps and jobs.

[App] (docs/apps.md): A stateless application controlled by a replication controller in Kubernetes, run continuously, and deployed with no downtime.

[Job] (docs/jobs.md): A pod in Kubernetes that runs to completion.

[Template] (docs/templates.md): YAML configuration files for controllers and pods, using single curly brackets (`{}`) to denote variables.

[Variable] (docs/variables.md): Environment variables used to populate controller and pod templates.

[Approval] (docs/approvals.md): An indication by the QA team that a certain build is deployable to prod.
