# Environments

An environment is a namespace in kubernetes that runs an isolated set of apps and jobs.

Deployment and pod definitions span all environments with a shared GitHub contents path, and are loaded from the environment's branch, or the default branch if none is specified by the namespace's `vili.environment-branch` annotation.
