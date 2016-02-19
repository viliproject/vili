import React from 'react';
import { RouteHandler, Link } from 'react-router'; // eslint-disable-line no-unused-vars
import * as _ from 'underscore';

const tabs = {
    'home': 'Home',
    'spec': 'Spec',
    'runs': 'Runs',
};

export class JobBase extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            activeTab: 'home'
        };
        this.activateTab = this.activateTab.bind(this)
    }

    activateTab(tab) {
        this.setState({
            activeTab: tab
        });
    }

    componentDidMount() {
        this.props.activateSideNavItem(['jobs', this.props.params.job]);
    }

    componentDidUpdate() {
        this.props.activateSideNavItem(['jobs', this.props.params.job]);
    }

    render() {
        var self = this;
        var tabElements = _.map(tabs, function(name, key) {
            var className = '';
            if (self.state.activeTab === key) {
                className = 'active';
            }
            var link = `/${self.props.params.env}/jobs/${self.props.params.job}`;
            if (key !== 'home') {
                link += `/${key}`;
            }
            return (
                <li role="presentation" className={className}>
                    <Link to={link}>{name}</Link>
                </li>
            );
        });
        return (
            <div>
                <div className="view-header">
                    <ol className="breadcrumb">
                        <li><Link to={`/${this.props.params.env}`}>{this.props.params.env}</Link></li>
                        <li><Link to={`/${this.props.params.env}/jobs`}>Jobs</Link></li>
                        <li className="active">{this.props.params.job}</li>
                    </ol>
                    <ul className="nav nav-pills pull-right">
                        {tabElements}
                    </ul>
                </div>
                <RouteHandler db={this.props.db} activateTab={this.activateTab} />
            </div>
        );
    }
}
