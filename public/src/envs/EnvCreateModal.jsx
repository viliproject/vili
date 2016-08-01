import React from 'react';
import * as _ from 'underscore';
import { Promise } from 'bluebird';
import { viliApi, template } from '../lib';
import { Modal, Label, Input, Button, ListGroup, ListGroupItem, Panel } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import router from '../router';


export class EnvCreateModal extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            show: true,
        };

        this.hide = this.hide.bind(this);
        this.onNameChange = this.onNameChange.bind(this);
        this.onBranchChange = this.onBranchChange.bind(this);
        this.loadTemplate = _.debounce(this.loadTemplate.bind(this), 200);
        this.onSpecChange = this.onSpecChange.bind(this);
        this.createNewEnvironment = this.createNewEnvironment.bind(this);
        this.loadApps = this.loadApps.bind(this);
        this.deployApps = this.deployApps.bind(this);
        this.loadJobs = this.loadJobs.bind(this);
        this.runJobs = this.runJobs.bind(this);
    }

    render() {
        var self = this;
        var actionButton = null;
        if (!this.state.createdResources) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.createNewEnvironment}
                    disabled={!this.state.spec || this.state.error}>
                    Create
                </Button>
            );
        } else if (!this.state.apps) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.loadApps}>
                    Deploy Apps
                </Button>
            );
        } else if (!this.state.appsDeployed) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.deployApps}>
                    Confirm
                </Button>
            );
        } else if (!this.state.jobs) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.loadJobs}>
                    Run Jobs
                </Button>
            );
        } else if (!this.state.jobsRan) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.runJobs}>
                    Confirm
                </Button>
            );
        }
        var specForm = null;
        if (this.state.name && this.state.branch && !this.state.apps) {
            specForm =
                [
                    <Label>Environment Spec</Label>,
                    <Input
                        type="textarea"
                        value={this.state.spec}
                        onChange={this.onSpecChange}
                        style={{'height': '400px'}}
                        disabled={this.state.createdResources}
                    />,
                ];
        }
        var output = null;
        if (this.state.jobs) {
            output = <Jobs jobs={this.state.jobs} />;
        } else if (this.state.apps) {
            output = <Apps apps={this.state.apps} />;
        } else if (this.state.createdResources) {
            output = <CreatedResources envName={this.state.name} resources={this.state.createdResources} />;
        } else if (this.state.error) {
            var errorMessage = _.map(this.state.error.response.body.split("\n"), function(text) {
                return <div>{text}</div>;
            })
            output = <Panel header='Error' bsStyle='danger'>{errorMessage}</Panel>;
        }
        return (
            <Modal show="true" onHide={this.hide}>
                <Modal.Header closeButton>
                    <Modal.Title>Create New Environment</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Label>Environment Name</Label>
                    <Input
                        type="text"
                        value={this.state.name}
                        placeholder="my-feature-environment"
                        onChange={this.onNameChange}
                        disabled={this.state.createdResources}
                    />
                    <Label>Default Branch</Label>
                    <Input
                        type="text"
                        value={this.state.branch}
                        placeholder="feature/branch"
                        onChange={this.onBranchChange}
                        disabled={this.state.createdResources}
                    />
                    {specForm}
                    {output}
                </Modal.Body>
                <Modal.Footer>
                    <Button onClick={this.hide}>Close</Button>
                    {actionButton}
                </Modal.Footer>
            </Modal>
        );
    }

    hide() {
        this.setState({
            name: null,
            branch: null,
            template: null,
        });
        if (this.props.onHide) {
            this.props.onHide();
        }
        if (this.state.createdResources) {
            // Need to reload the page to properly populate environment info
            window.location.pathname=`/${this.state.name}/apps`;
        }
    }

    onNameChange(event) {
        var name = event.target.value;
        this.setState({
            name: name,
            createdResources: null,
            error: null,
        });
        this.loadTemplate(name, this.state.branch);
    }

    onBranchChange(event) {
        var branch = event.target.value;
        this.setState({
            branch: branch,
            createdResources: null,
            error: null,
        });
        this.loadTemplate(this.state.name, branch);
    }

    loadTemplate(name, branch) {
        if (!name || !branch) {
            return;
        }
        var self = this;
        viliApi.environments.template(branch).then(function(resp) {
            var templ = template(resp.template, {
                NAMESPACE: name,
                BRANCH: branch,
            });
            self.setState({spec: templ.populated});
        });
    }

    onSpecChange(event) {
        this.setState({
            spec: event.target.value,
            createdResources: null,
            error: null,
        });
    }

    createNewEnvironment() {
        var self = this;
        viliApi.environments.create({
            name: this.state.name,
            branch: this.state.branch,
            spec: this.state.spec,
        }).then(function(resp) {
            self.setState({createdResources: resp})
        }, function(error) {
            self.setState({error: error})
        });
    }

    loadApps() {
        var self = this;
        var envApps = window.appconfig.envApps[window.appconfig.defaultEnv];
        var apps = {};
        _.each(envApps, function(appName) {
            apps[appName] = {
                name: appName,
                loading: true,
            };
        });
        this.setState({apps: apps});
        _.each(envApps, function(appName) {
            viliApi.apps.get(self.state.name, appName).then(function(app) {
                var image = _.findWhere(app.repository, {branch: self.state.branch});
                if (!image && app.repository) {
                    image = app.repository[0];
                }
                var apps = _.clone(self.state.apps);
                apps[appName].image = image;
                apps[appName].loading = false;
                self.setState({apps: apps});
            }, function(error) {
                var apps = _.clone(self.state.apps);
                apps[appName].error = error;
                apps[appName].loading = false;
                self.setState({apps: apps});
            });
        });
    }

    deployApps() {
        var self = this;
        _.mapObject(this.state.apps, function(app, appName) {
            if (app.image) {
                viliApi.deployments.create(self.state.name, appName, {
                    tag: app.image.tag,
                    branch: app.image.branch,
                    trigger: true,
                    desiredReplicas: 1,
                }).then(function() {
                    var apps = _.clone(self.state.apps);
                    apps[appName].deployed = true;
                    self.setState({apps: apps});
                }, function(error) {
                    var apps = _.clone(self.state.apps);
                    apps[appName].error = error;
                    self.setState({apps: apps});
                });
            }
        });
        this.setState({appsDeployed: true});
    }

    loadJobs() {
        var self = this;
        var envJobs = window.appconfig.envJobs[window.appconfig.defaultEnv];
        var jobs = {};
        _.each(envJobs, function(jobName) {
            jobs[jobName] = {
                name: jobName,
                loading: true,
            };
        });
        this.setState({jobs: jobs});
        _.each(envJobs, function(jobName) {
            viliApi.jobs.get(self.state.name, jobName).then(function(job) {
                var image = _.findWhere(job.repository, {branch: self.state.branch});
                if (!image && job.repository) {
                    image = job.repository[0];
                }
                var jobs = _.clone(self.state.jobs);
                jobs[jobName].image = image;
                jobs[jobName].loading = false;
                self.setState({jobs: jobs});
            }, function(error) {
                var jobs = _.clone(self.state.jobs);
                jobs[jobName].error = error;
                jobs[jobName].loading = false;
                self.setState({jobs: jobs});
            });
        });
    }

    runJobs() {
        var self = this;
        _.mapObject(this.state.jobs, function(job, jobName) {
            if (job.image) {
                viliApi.runs.create(self.state.name, jobName, {
                    tag: job.image.tag,
                    branch: job.image.branch,
                    trigger: true,
                }).then(function() {
                    var jobs = _.clone(self.state.jobs);
                    jobs[jobName].started = true;
                    self.setState({jobs: jobs});
                }, function(error) {
                    var jobs = _.clone(self.state.jobs);
                    jobs[jobName].error = error;
                    self.setState({jobs: jobs});
                });
            }
        });
        this.setState({jobsRan: true});
    }
}

class CreatedResources extends React.Component {
    render() {
        var self = this;
        var createdResources = _.clone(this.props.resources);
        var namespaces = _.map(createdResources.namespace, function(name) {
            var style = 'success';
            if (name != self.props.envName) {
                style = 'danger';
            }
            return <ListGroupItem bsStyle={style}>Created namespace {name}</ListGroupItem>;
        })
        delete(createdResources.namespace)
        var resources = _.mapObject(createdResources, function(names, type) {
            return _.map(names, function(name) {
                return <ListGroupItem bsStyle="success">Created {type} {name}</ListGroupItem>;
            });
        });
        return (
            <ListGroup>
                {namespaces}
                {resources}
            </ListGroup>
        );
    }
}

class Apps extends React.Component {
    render() {
        var self = this;
        var apps = _.mapObject(this.props.apps, function(app, name) {
            app = _.clone(app);
            app.name = name;
            return app;
        });
        apps = _.sortBy(apps, 'name');
        var appItems = _.map(apps, function(app) {
            if (app.loading) {
                return <ListGroupItem header={app.name}>Loading...</ListGroupItem>;
            }
            if (!app.image) {
                return <ListGroupItem header={app.name} bsStyle='warning'>No deployable image found</ListGroupItem>;
            }
            var style = null;
            if (app.deployed) {
                style = 'success';
            } else if (app.error) {
                style = 'danger';
            }
            return <ListGroupItem header={app.name} bsStyle={style}>{app.image.revision} from {app.image.branch}</ListGroupItem>;
        });
        return (
            <ListGroup>
                {appItems}
            </ListGroup>
        );
    }
}

class Jobs extends React.Component {
    render() {
        var self = this;
        var jobs = _.mapObject(this.props.jobs, function(job, name) {
            job = _.clone(job);
            job.name = name;
            return job;
        });
        jobs = _.sortBy(jobs, 'name');
        var jobItems = _.map(jobs, function(job) {
            if (job.loading) {
                return <ListGroupItem header={job.name}>Loading...</ListGroupItem>;
            }
            if (!job.image) {
                return <ListGroupItem header={job.name} bsStyle='warning'>No runnable image found</ListGroupItem>;
            }
            var style = null;
            if (job.started) {
                style = 'success';
            } else if (job.error) {
                style = 'danger';
            }
            return <ListGroupItem header={job.name} bsStyle={style}>{job.image.revision} from {job.image.branch}</ListGroupItem>;
        });
        return (
            <ListGroup>
                {jobItems}
            </ListGroup>
        );
    }
}
