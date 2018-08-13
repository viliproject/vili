# Vili

[![CircleCI Status](https://circleci.com/gh/airware/vili.svg?style=shield)](https://circleci.com/gh/airware/vili)
[![Slack Status](https://vili-slackin.herokuapp.com/badge.svg)](https://vili-slackin.herokuapp.com/)

Vili is an open source dashboard for managing deployments to a [Kubernetes](http://kubernetes.io/) cluster. It is built to:

- Manage both manual and continuous deployments
- Gate production deployments through our QA process
- Provide transparency into the current state of our infrastructure


### Is Vili right for you?

Vili is opinionated, to be able to set it up you need to:

- use GitHub for version control
- build Docker images to ship code and tag them with the Git SHA and branch they were built from
- push these Docker images to a Docker registry
- use Kubernetes namespaces to manage environments
- use Kubernetes deployments to deploy applications
- use Slack for team messaging

### What does Vili mean anyway?

Vili is a brother of Odin in Norse mythology, and he gives intelligence to the first humans.

<hr>

## Setup
To setup Vili on your Kubernetes cluster, follow these steps:

1. Select a domain name to host Vili under, such as vili.mydomain.com. Create an [Okta](https://www.okta.com/) app with a redirect URL that points to vili.mydomain.com/login/callback. Write down the Okta metadata url.
2. Create [Docker](https://www.docker.com/) repositories for your applications. You may use any standard Docker registry, including [Docker Hub](https://hub.docker.com/), [quay.io](https://quay.io/), or a self-hosted registry. [Amazon ECR](https://aws.amazon.com/ecr/) registries are also supported.
3. Create a new [Firebase](https://www.firebase.com/) app. Set the "Firebase rules" to match [this](docs/installation/firebase_security.json). Write down the Firebase app's URL and secret.
4. Create a GitHub repo with a directory that holds your replication controller templates, pod templates, and environment variables following this [example](docs/examples/github). Also create a GitHub access token following instructions [here](https://help.github.com/articles/creating-an-access-token-for-command-line-use/). Write down your GitHub organization or user name, the path to the directory created above and the authentication token.
5. Create a Slack [bot integration](https://api.slack.com/bot-users). Write down the API token from the integration settings page.
6. Create a [secret](http://kubernetes.io/v1.1/docs/user-guide/secrets.html) in your Kubernetes cluster that stores your GitHub, Docker, Firebase, and Slack credentials following this [example](https://github.com/airware/vili/blob/master/docs/examples/simple/secret.yaml). Populate the values in the secret using the Docker, GitHub, Firebase, and Slack information you wrote down in the previous steps. Don't forget to base64 encode them as required by Kubernetes!
7. Create a deployment in your Kubernetes cluster following this [example](docs/examples/simple/deployment.yaml). Populate the environment variables using the Okta, Docker, GitHub, Firebase, and Slack information you wrote down in the previous steps.
8. Create a [service](http://kubernetes.io/v1.1/docs/user-guide/services.html) for this replication controller, and allow external access to this service under the domain name you chose in step 1.
9. If you want to integrate a Continuous Integration service with Vili, you can do so by adding CI_PROVIDER config variable with value (name of the ci provider in small letters) and other required CI parameters to your config file. 
10. Currently, we are supporting integration only with CircleCI. To integrate with Circle ci, you will need below 3 parameters:

  ```
    CI_PROVIDER="circleci"
    CIRCLECI_TOKEN=XXXX
    CIRCLECI_BASEURL="https://circleci.com/api/v1.1/"
  ```

   You will also have to add a circle job name to your kubernetes namespace's annotations to run after successfull deployment which can be used to run tests or any other post deployment tasks:

  ```
    vili.environment-webhook="circle_jobname"
  ```

You are all set! Vili will use the GitHub and Docker Registry APIs to discover your apps and help you deploy them.

## Local Vili Development

1. Follow the example [sample_devenv.sh](sample_devenv.sh) to create your own environment file with relevant configration.

1. Install `redis`

   ```
   > brew install redis
   ```

1. Start `redis`

   ```
   > brew services start redis
   ```

1. Install Vili frontend node modules

   ```
   > cd /path/to/<vili-root>
   > npm install
   ```

1. Build Vili frontend Webpack

   ```
   > npm run build
   ```

1. Run Vili

   ```
   > go run main.go
   ```

1. Direct your browser to `https://localhost:4001`.  Voila!


## Concepts

[Environment](docs/environments.md): A namespace in Kubernetes that runs an isolated set of apps and jobs.

[App](docs/apps.md): A stateless application controlled by a deployment in Kubernetes, run continuously, and deployed with no downtime.

[Job](docs/jobs.md): A pod in Kubernetes that runs to completion.

[Template](docs/templates.md): YAML configuration files for controllers and pods, using go templates syntax for variable population.

[Approval](docs/approvals.md): An indication by the QA team that a certain build is deployable to prod.
