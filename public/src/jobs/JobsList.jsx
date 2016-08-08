import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Promise } from 'bluebird';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars
import _ from 'underscore';


export class JobsList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};

        this.loadData = this.loadData.bind(this);
    }

    render() {
        var env = _.findWhere(window.appconfig.envs, {name: this.props.params.env});
        var links = env.jobs.map(function(job) {
            return <li><Link key={job} to={`/${env.name}/jobs/${job}`}>{job}</Link></li>;
        });
        return (
            <div>
                <div className="view-header">
                    <ol className="breadcrumb">
                        <li><Link to={`/${this.props.params.env}`}>{this.props.params.env}</Link></li>
                        <li className="active">Jobs</li>
                    </ol>
                </div>
                <ul className="nav nav-pills nav-stacked">{links}</ul>
            </div>
        );
    }

    loadData() {
        var self = this;
        Promise.props({
            // jobs: viliApi.jobs.get(this.props.params.env),
        }).then(function(props) {
            self.setState(props);
        });
    }

    componentDidMount() {
        this.props.activateSideNavItem(['jobs']);
        this.loadData();
    }

    componentDidUpdate(prevProps) {
        if (this.props != prevProps) {
            this.state = {};
            this.forceUpdate();
            this.loadData();
        }
    }
}
