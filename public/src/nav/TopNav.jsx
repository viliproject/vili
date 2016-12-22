import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Navbar, Nav, NavDropdown, MenuItem } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import { viliApi } from '../lib';
import { LinkMenuItem } from '../shared'; // eslint-disable-line no-unused-vars
import { EnvCreateModal } from '../envs';

export class TopNav extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            showCreateEnvModal: false,
        }

        this.showCreateEnvModal = this.showCreateEnvModal.bind(this);
        this.hideCreateEnvModal = this.hideCreateEnvModal.bind(this);
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
                    {env.name}
                </LinkMenuItem>;
            });
            var envCreateModal = null;
            if (this.state.showCreateEnvModal) {
                envCreateModal = <EnvCreateModal onHide={this.hideCreateEnvModal} />;
            }
            return (
                <Navbar className={this.props.env && this.props.env.prod ? 'prod' : ''}
                        fixedTop={true} fluid={true}>
                    <div className="navbar-header pull-left">
                        <Link className="navbar-brand" to="home">Vili</Link>
                    </div>
                    <Nav key="user" className="user" pullRight={true}>
                        <NavDropdown id="user-dropdown" title={userText}>
                            <MenuItem title="Logout" href="/logout">Logout</MenuItem>
                        </NavDropdown>
                    </Nav>
                    <Nav key="env" className="environment" pullRight={true}>
                        <NavDropdown id="env-dropdown"
                                     title={(this.props.env && this.props.env.name) || <span className="text-danger">Select Environment</span>}>
                            {envElements}
                            <MenuItem divider />
                            <MenuItem onSelect={this.showCreateEnvModal}>Create Environment</MenuItem>
                        </NavDropdown>
                    </Nav>
                    {envCreateModal}
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
        this.setState({
            showCreateEnvModal: true,
        });
    }

    hideCreateEnvModal() {
        this.setState({
            showCreateEnvModal: false,
        });
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
