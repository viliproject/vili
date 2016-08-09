import React from 'react';
import * as _ from 'underscore';
import { Promise } from 'bluebird';
import { viliApi } from '../lib';
import { Modal, Button, FormGroup, ControlLabel, FormControl, ListGroup, ListGroupItem, Panel } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import Typeahead from 'react-bootstrap-typeahead';
import router from '../router';


export class EnvCreateModal extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            show: true,
        };

        this.hide = this.hide.bind(this);
        this.loadBranches = this.loadBranches.bind(this);
        this.onNameChange = this.onNameChange.bind(this);
        this.onBranchChange = this.onBranchChange.bind(this);
        this.loadSpec = _.debounce(this.loadSpec.bind(this), 200);
        this.onSpecChange = this.onSpecChange.bind(this);
        this.createNewEnvironment = this.createNewEnvironment.bind(this);
        this.loadJobs = this.loadJobs.bind(this);
        this.runJobs = this.runJobs.bind(this);
        this.loadApps = this.loadApps.bind(this);
        this.deployApps = this.deployApps.bind(this);
    }

    render() {
        var self = this;
        var actionButton = null;
        if (!this.state.createdResources) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.createNewEnvironment}
                    disabled={!this.state.spec || Boolean(this.state.error)}>
                    Create
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
        } else if (!this.state.jobsTriggered) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.runJobs}
                    disabled={_.some(this.state.jobs, function(job) { return job.loading;} )}>
                    Confirm
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
        } else if (!this.state.appsTriggered) {
            actionButton = (
                <Button
                    bsStyle="primary"
                    onClick={this.deployApps}
                    disabled={_.some(this.state.apps, function(app) { return app.loading;} )}>
                    Confirm
                </Button>
            );
        }
        var specForm = null;
        if (this.state.name && this.state.branch && !this.state.jobs) {
            specForm = (
                <FormGroup controlId="environmentSpec">
                    <ControlLabel>Environment Spec</ControlLabel>
                    <FormControl
                        componentClass="textarea"
                        value={this.state.spec}
                        onChange={this.onSpecChange}
                        style={{'height': '400px'}}
                        disabled={this.state.createdResources} />
                </FormGroup>
            );
        }
        var output = null;
        if (this.state.apps) {
            output = <Apps apps={this.state.apps} />;
        } else if (this.state.jobs) {
            output = <Jobs jobs={this.state.jobs} />;
        } else if (this.state.createdResources) {
            output = <CreatedResources envName={this.state.name} resources={this.state.createdResources} />;
        } else if (this.state.error) {
            var errorMessage = _.map(this.state.error.response.body.split("\n"), function(text) {
                return <div>{text}</div>;
            })
            output = <Panel header='Error' bsStyle='danger'>{errorMessage}</Panel>;
        }
        return (
            <Modal show={true} onHide={this.hide}>
                <Modal.Header closeButton>
                    <Modal.Title>Create New Environment</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <FormGroup controlId="environmentName">
                        <ControlLabel>Environment Name</ControlLabel>
                        <FormControl
                            componentClass="input"
                            type="text"
                            value={this.state.name}
                            placeholder="my-feature-environment"
                            onChange={this.onNameChange}
                            disabled={this.state.createdResources}
                        />
                    </FormGroup>
                    <ControlLabel>Default Branch</ControlLabel>
                    <Typeahead
                        options={this.state.branches || []}
                        labelKey='branch'
                        onInputChange={this.onBranchChange}
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

    loadBranches() {
        var self = this;
        viliApi.environments.branches().then(function(resp) {
            self.setState({branches: _.map(resp.branches, function(branch) {
                return {branch: branch};
            })});
        });
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
        this.loadSpec(name, this.state.branch);
    }

    onBranchChange(branch) {
        this.setState({
            branch: branch,
            createdResources: null,
            error: null,
        });
        this.loadSpec(this.state.name, branch);
    }

    loadSpec(name, branch) {
        if (!name || !branch) {
            return;
        }
        var self = this;
        viliApi.environments.spec(name, branch).then(function(resp) {
            self.setState({spec: resp.spec});
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

    loadJobs() {
        var self = this;
        var defaultEnv = _.findWhere(window.appconfig.envs, {name: window.appconfig.defaultEnv});
        var envJobs = defaultEnv.jobs;
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
        this.setState({jobsTriggered: true});
        var self = this;
        _.mapObject(this.state.jobs, function(job) {
            if (job.image) {
                viliApi.runs.create(self.state.name, job.name, {
                    tag: job.image.tag,
                    branch: job.image.branch,
                    trigger: true,
                }).then(function() {
                    var jobs = _.clone(self.state.jobs);
                    jobs[job.name].started = true;
                    self.setState({jobs: jobs});
                }, function(error) {
                    var jobs = _.clone(self.state.jobs);
                    jobs[job.name].error = error;
                    self.setState({jobs: jobs});
                });
            }
        });
    }

    loadApps() {
        var self = this;
        var defaultEnv = _.findWhere(window.appconfig.envs, {name: window.appconfig.defaultEnv});
        var envApps = defaultEnv.apps;
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
        this.setState({appsTriggered: true});
        var self = this;
        _.mapObject(this.state.apps, function(app) {
            if (app.image) {
                var deployApp = function() {
                    viliApi.deployments.create(self.state.name, app.name, {
                        tag: app.image.tag,
                        branch: app.image.branch,
                        trigger: true,
                        desiredReplicas: 1,
                    }).then(function() {
                        var apps = _.clone(self.state.apps);
                        apps[app.name].deployed = true;
                        self.setState({apps: apps});
                    }, function(error) {
                        var apps = _.clone(self.state.apps);
                        apps[app.name].error = error;
                        self.setState({apps: apps});
                    });
                };
                viliApi.services.create(self.state.name, app.name).then(deployApp, deployApp);
            }
        });
    }

    componentWillMount() {
        this.loadBranches();
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
        var resources = _.flatten(_.map(createdResources, function(names, type) {
            return _.map(names, function(name) {
                return <ListGroupItem bsStyle="success">Created {type} {name}</ListGroupItem>;
            });
        }));
        return (
            <ListGroup>
                {namespaces}
                {resources}
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
