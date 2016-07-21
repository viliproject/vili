import React from 'react';
import _ from 'underscore';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Navbar, Nav, NavDropdown, MenuItem } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import { viliApi } from '../lib';
import { LinkMenuItem } from '../shared'; // eslint-disable-line no-unused-vars

export class TopNav extends React.Component {
    render() {
        var self = this;

        if (window.appconfig) {
            var isProd = _.contains(window.appconfig.prodEnvs, this.props.env);

            // user
            var user = window.appconfig.user;
            var userText = user.firstName + ' ' + user.lastName + ' (' + user.username + ')';

            // environments
            var path  = window.location.pathname + window.location.search,
                spath = path.split('/');
            var envElements = window.appconfig.envs.map(function(env) {
                spath[1] = env;
                return <LinkMenuItem key={env} to={spath.join('/')} active={env===self.props.env} onRemove={env===self.props.env ? null : self.deleteEnvironment(env)}>{env}</LinkMenuItem>;
            });
            return (
                <Navbar className={isProd ? 'prod' : ''}
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
                                     title={this.props.env || <span className="text-danger">Select Environment</span>}>
                            {envElements}
                            <MenuItem divider />
                            <MenuItem onSelect={this.createNewEnvironment}>Create Environment</MenuItem>
                        </NavDropdown>
                    </Nav>
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

    createNewEnvironment() {
        var envName = prompt('Enter the name of the new environment to create');
        if (!envName) {
            return;
        }
        viliApi.environments.create(envName);
    }

    deleteEnvironment(env) {
        return function() {
            var envName = prompt('Are you sure you wish to delete this environment? Enter the environment name to confirm');
            if (envName != env) {
                return
            }
            viliApi.environments.delete(envName);
        }
    }

}
