import React from 'react';
import _ from 'underscore';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Navbar, Nav, NavDropdown, MenuItem, Modal, Label, Input, Button } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import { viliApi } from '../lib';
import { LinkMenuItem } from '../shared'; // eslint-disable-line no-unused-vars

export class TopNav extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            showCreateEnvModal: false,
        }

        this.showCreateEnvModal = this.showCreateEnvModal.bind(this);
        this.hideCreateEnvModal = this.hideCreateEnvModal.bind(this);
        this.onEnvNameChange = this.onEnvNameChange.bind(this);
        this.onEnvBranchChange = this.onEnvBranchChange.bind(this);
        this.createNewEnvironment = this.createNewEnvironment.bind(this);
    }

    render() {
        var self = this;

        if (window.appconfig) {
            // user
            var user = window.appconfig.user;
            var userText = user.firstName + ' ' + user.lastName + ' (' + user.username + ')';

            // environments
            var path  = window.location.pathname + window.location.search,
                spath = path.split('/');
            var envElements = window.appconfig.envs.map(function(env) {
                spath[1] = env.name;
                var onRemove = null;
                if (self.props.env && env.name !== self.props.env.name && !env.protected) {
                    onRemove = function() {
                        self.deleteEnvironment(env.name);
                    };
                }
                return <LinkMenuItem
                    key={env.name}
                    to={spath.join('/')}
                    active={self.props.env && env.name===self.props.env.name}
                    onRemove={onRemove}>
                    {env}
                </LinkMenuItem>;
            });
            return (
                <Navbar className={this.props.env && this.props.env.prod ? 'prod' : ''}
                        fixedTop={true} fluid={true}>
                    <div className="navbar-header pull-left">
                        <Link className="navbar-brand" to="home">Vili</Link>
                    </div>
                    <Nav key="user" ulClassName="user" pullRight={true}>
                        <NavDropdown id="user-dropdown" title={userText}>
                            <MenuItem title="Logout" href="/logout">Logout</MenuItem>
                        </NavDropdown>
                    </Nav>
                    <Nav key="env" ulClassName="environment" pullRight={true}>
                        <NavDropdown id="env-dropdown"
                                     title={(this.props.env && this.props.env.name) || <span className="text-danger">Select Environment</span>}>
                            {envElements}
                            <MenuItem divider />
                            <MenuItem onSelect={this.showCreateEnvModal}>Create Environment</MenuItem>
                        </NavDropdown>
                    </Nav>
                    <Modal show={this.state.showCreateEnvModal} onHide={this.hideCreateEnvModal}>
                        <Modal.Header closeButton>
                            <Modal.Title>Create New Environment</Modal.Title>
                        </Modal.Header>
                        <Modal.Body>
                            <Label>Environment Name</Label>
                            <Input
                                type="text"
                                value={this.state.envName}
                                placeholder="my-feature-environment"
                                onChange={this.onEnvNameChange}
                            />
                            <Label>Default Branch</Label>
                            <Input
                                type="text"
                                value={this.state.envBranch}
                                placeholder="feature/branch"
                                onChange={this.onEnvBranchChange}
                            />
                        </Modal.Body>
                        <Modal.Footer>
                            <Button onClick={this.hideCreateEnvModal}>Close</Button>
                            <Button bsStyle="primary" onClick={this.createNewEnvironment} disabled={!this.state.envName || !this.state.envBranch}>Create</Button>
                        </Modal.Footer>
                    </Modal>
                </Navbar>
            );
        } else {
            return (
                <Navbar fixedTop={true} fluid={true}>
                    <div className="navbar-header pull-left">
                        <Link className="navbar-brand" to="home">Vili</Link>
                    </div>
                    <Nav key="user" ulClassName="user" pullRight={true}>
                        <MenuItem title="Login" href="/login">Login</MenuItem>
                    </Nav>
                </Navbar>
            );
        }
    }

    showCreateEnvModal() {
        this.setState({showCreateEnvModal: true});
    }

    hideCreateEnvModal() {
        this.setState({
            showCreateEnvModal: false,
            envName: null,
            envBranch: null,
        });
    }

    onEnvNameChange(event) {
        this.setState({envName: event.target.value});
    }

    onEnvBranchChange(event) {
        this.setState({envBranch: event.target.value});
    }

    createNewEnvironment() {
        viliApi.environments.create({
            name: this.state.envName,
            branch: this.state.envBranch,
        });
        window.location.reload();
    }

    deleteEnvironment(env) {
        var envName = prompt('Are you sure you wish to delete this environment? Enter the environment name to confirm');
        if (envName !== env) {
            return
        }
        viliApi.environments.delete(envName);
        window.location.reload();
    }

}
