import React from 'react';
import * as _ from 'underscore';
import { Promise } from 'bluebird';
import { viliApi, template } from '../lib';
import { Modal, Label, Input, Button, ListGroup, ListGroupItem, Panel } from 'react-bootstrap'; // eslint-disable-line no-unused-vars


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
    }

    render() {
        var self = this;
        var createButton = (
            <Button
                bsStyle="primary"
                onClick={this.createNewEnvironment}
                disabled={!this.state.spec || this.state.error}>
                Create
            </Button>
        );
        var specForm = null;
        if (this.state.name && this.state.branch) {
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
        if (this.state.createdResources) {
            var createdResources = _.clone(this.state.createdResources);
            var namespaces = _.map(createdResources.namespace, function(name) {
                var style = 'success';
                if (name != self.state.name) {
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
            output = (
                <ListGroup>
                    {namespaces}
                    {resources}
                </ListGroup>
            );
            createButton = null;
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
                    {createButton}
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
            window.location.reload();
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
}
