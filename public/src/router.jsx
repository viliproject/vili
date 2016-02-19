import React from 'react'
import Router, { Route, Link, RouteHandler, DefaultRoute } from 'react-router'; // eslint-disable-line no-unused-vars

import { TopNav, SideNav } from './nav' // eslint-disable-line no-unused-vars
import { AppsList, AppBase, App, AppSpec, AppPods, AppService, AppDeployments, AppDeployment } from './apps' // eslint-disable-line no-unused-vars
import { JobsList, JobBase, Job, JobSpec, JobRuns, JobRun } from './jobs' // eslint-disable-line no-unused-vars
import { PodsList, Pod } from './pods' // eslint-disable-line no-unused-vars
import { NodesList, Node } from './nodes' // eslint-disable-line no-unused-vars

class Main extends React.Component {

    render() {
        return (
            <div className="top-nav container-fluid full-height">
                <RouteHandler db={this.props.db} />
            </div>
        );
    }
}

class Home extends React.Component {
    render() {
        var content = '';
        if (window.appconfig) {
            var links = window.appconfig.envs.map(function(env) {
                return <li><Link key={env} to={`/${env}`}>{env}</Link></li>;
            });
            content = (
                <div>
                    <div className="view-header">
                        <ol className="breadcrumb">
                            <li className="active">Select Environment</li>
                        </ol>
                    </div>
                    <ul className="nav nav-pills nav-stacked">{links}</ul>
                </div>
            );
        } else {
            content = (
                <div className="jumbotron">
                    <h1>Welcome to Vili</h1>
                    <p>Please log in to view your applications.</p>
                    <p><a className="btn btn-primary btn-lg" href="/login" role="button">Login</a></p>
                </div>
            );
        }
        return (
            <div>
                <TopNav />
                <div className="page-wrapper">
                    <div className="sidebar">
                        <SideNav ref="sidenav"/>
                    </div>
                    <div className="content-wrapper">{content}</div>
                </div>
            </div>
        );
    }
}

class Environment extends React.Component {
    constructor(props) {
        super(props);
        this.activateSideNavItem = this.activateSideNavItem.bind(this)
    }

    activateSideNavItem(item) {
        this.refs.sidenav.activateItem(item);
    }

    render() {
        return (
            <div>
                <TopNav env={this.props.params.env} />
                <div className="page-wrapper">
                    <div className="sidebar">
                        <SideNav env={this.props.params.env} ref="sidenav"/>
                    </div>
                    <div className="content-wrapper">
                        <RouteHandler db={this.props.db} activateSideNavItem={this.activateSideNavItem} />
                    </div>
                </div>
            </div>
        );
    }
}

class EnvironmentHome extends React.Component {

    render() {
        return (
            <div>
                <div className="view-header">
                    <ol className="breadcrumb">
                        <li className="active">{this.props.params.env}</li>
                    </ol>
                </div>

                <ul className="nav nav-pills nav-stacked">
                    <li><Link to={`/${this.props.params.env}/apps`}>Apps</Link></li>
                    <li><Link to={`/${this.props.params.env}/jobs`}>Jobs</Link></li>
                    <li><Link to={`/${this.props.params.env}/pods`}>Pods</Link></li>
                    <li><Link to={`/${this.props.params.env}/nodes`}>Nodes</Link></li>
                </ul>
            </div>
        );
    }

}

class NotFound extends React.Component {
    render() {
        return (
            <div>NOT FOUND</div>
        );
    }
}

// routes
var routes = (
    <Route path="" handler={Main}>
        <DefaultRoute handler={NotFound} />
        <Route name="home" path="/" handler={Home} />
        <Route name="envhome" path="/:env" handler={Environment}>
            <Route handler={EnvironmentHome} />
            <Route path="apps" handler={AppsList} />
            <Route path="apps/:app" name="apphome" handler={AppBase}>
                <Route handler={App} />
                <Route path="spec" name="appspec" handler={AppSpec} />
                <Route path="pods" name="apppods" handler={AppPods} />
                <Route path="service" name="appservice" handler={AppService} />
                <Route path="deployments" name="appdeployments" handler={AppDeployments} />
                <Route path="deployments/:deployment" name="appdeployment" handler={AppDeployment} />
            </Route>
            <Route path="jobs" handler={JobsList} />
            <Route path="jobs/:job" name="jobhome" handler={JobBase}>
                <Route handler={Job} />
                <Route path="spec" name="jobspec" handler={JobSpec} />
                <Route path="runs" name="jobruns" handler={JobRuns} />
                <Route path="runs/:run" name="jobrun" handler={JobRun} />
            </Route>
            <Route path="pods" handler={PodsList} />
            <Route path="pods/:pod" name="podhome" handler={Pod} />
            <Route path="nodes" handler={NodesList} />
            <Route path="nodes/:node" name="nodehome" handler={Node} />
        </Route>
    </Route>
);

export default Router.create({
    routes: routes,
    location: Router.HistoryLocation
});
