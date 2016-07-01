import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import _ from 'underscore';
import { Promise } from 'bluebird';
import { viliApi, displayTime, template } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars
import router from '../router';


class Row extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.state = {};

        this.runTag = this.runTag.bind(this);
        this.approveTag = this.approveTag.bind(this);
        this.unapproveTag = this.unapproveTag.bind(this);
    }

    render() {
        var data = this.props.data;
        var tag = data.tag;
        var date = new Date(data.lastModified);

        var actions = [];
        if (this.props.canRun && (!this.props.envNeedsApprovalForRun || this.state.approval)) {
            actions.push(<button type="button" className="btn btn-xs btn-primary" onClick={this.runTag}>Run</button>);
        }
        if (this.props.envCanApproveTag) {
            if (this.state.approval) {
                actions.push(<button type="button" className="btn btn-xs btn-danger" onClick={this.unapproveTag}>Unapprove</button>);

            } else {
                actions.push(<button type="button" className="btn btn-xs btn-success" onClick={this.approveTag}>Approve</button>);
            }
        }

        var cells = _.union([
            <td data-column="tag">{tag}</td>,
            <td data-column="branch">{data.branch}</td>,
            <td data-column="revision">{data.revision || 'unknown'}</td>,
            <td data-column="buildtime">{displayTime(date)}</td>,
        ], this.props.hasApprovalColumn ? [
            <td data-column="approved">{this.state.approvalContents}</td>
        ] : [], [
            <td data-column="actions">{actions}</td>
        ]);

        return <tr>{cells}</tr>;
    }

    componentDidMount() {
        var self = this;
        if (this.props.approvalDB) {
            this.props.approvalDB.child(this.props.data.tag).on('value', function(snapshot) {
                var approval = snapshot.val();
                var approvalContents = [];
                if (approval) {
                    approvalContents.push(<span>{displayTime(new Date(approval.time)) + ' by ' + approval.username}</span>);
                    if (approval.url) {
                        approvalContents.push(<br/>);
                        approvalContents.push(<a href={approval.url} target="_blank">release notes</a>);
                    }
                }
                self.setState({
                    approval: approval,
                    approvalContents: approvalContents,
                });
            });
        }
    }

    componentWillUnmount() {
        if (this.props.approvalDB) {
            this.props.approvalDB.child(this.props.data.tag).off();
        }
    }

    runTag() {
        var self = this;
        viliApi.runs.create(this.props.env, this.props.job, {
            tag: this.props.data.tag,
            branch: this.props.data.branch,
            trigger: false
        }).then(function(run) {
            router.transitionTo(`/${self.props.env}/jobs/${self.props.job}/runs/${run.id}`);
        });
    }

    approveTag() {
        var url = prompt('Please enter the release url');
        if (!url) {
            return;
        }
        viliApi.releases.create(this.props.job, this.props.data.tag, {
            url: url,
        });
    }

    unapproveTag() {
        viliApi.releases.delete(this.props.job, this.props.data.tag);
    }
}

export class Job extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};
        this.loadData = this.loadData.bind(this);
    }

    render() {
        if (!this.state.job) {
            return <Loading />;
        }
        var self = this;
        var columns = _.union([
            {title: 'Tag', key: 'tag'},
            {title: 'Branch', key: 'branch'},
            {title: 'Revision', key: 'revision'},
            {title: 'Build Time', key: 'buildtime'},
        ], this.state.hasApprovalColumn ? [{title: 'Approved', key: 'approved'}] : [], [
            {title: 'Actions', key: 'actions'},
        ]);

        var rows = [];

        _.each(this.state.job.repository, function(data) {
            var date = new Date(data.lastModified);
            var row = <Row data={data} currentTag={self.state.currentTag}
                           canRun={self.state.canRun}
                           envNeedsApprovalForRun={self.state.envNeedsApprovalForRun}
                           hasApprovalColumn={self.state.hasApprovalColumn}
                           envCanApproveTag={self.state.envCanApproveTag}
                           approvalDB={self.state.approvalDB}
                           env={self.props.params.env}
                           job={self.props.params.job}
                      />;
            rows.push({
                _row: row,
                time: date.getTime()
            });
        });

        rows = _.sortBy(rows, function(row) {
            return -row.time;
        });

        return <Table columns={columns} rows={rows} />;
    }

    loadData() {
        var self = this;
        Promise.props({
            job: viliApi.jobs.get(this.props.params.env, this.props.params.job)
        }).then(function(state) {
            state.envCanApproveTag = _.contains(window.appconfig.approvalEnvs, self.props.params.env);
            state.envNeedsApprovalForDeploy = _.contains(window.appconfig.prodEnvs, self.props.params.env);
            state.hasApprovalColumn = state.envCanApproveTag || state.envNeedsApprovalForRun;
            if (state.hasApprovalColumn) {
                state.approvalDB = self.props.db.child('releases').child(self.props.params.job);
            }

            state.basePod = template(state.job.podTemplate, state.job.variables);
            state.canRun = state.basePod.valid;
            if (!state.canRun) {
                // TODO show message saying pod not valid
            }
            self.setState(state);
        });
    }

    componentDidMount() {
        this.props.activateTab('home');
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
