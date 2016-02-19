import React from 'react';
import _ from 'underscore';
import humanSize from 'human-size';
import { Promise } from 'bluebird';
import { viliApi, displayTime, template } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars
import router from '../router';


class Row extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.state = {};

        this.loadData = this.loadData.bind(this);
        this.deployTag = this.deployTag.bind(this);
        this.approveTag = this.approveTag.bind(this);
        this.unapproveTag = this.unapproveTag.bind(this);
    }

    render() {
        var data = this.props.data;
        var className = '';
        if (this.props.deployedAt) {
            className = 'success';
        }
        var tag = data.tag;
        var date = new Date(data.lastModified);

        var actions = [];
        if (this.props.canDeploy && (!this.props.envNeedsApprovalForDeploy || this.state.approval)) {
            actions.push(<button type="button" className="btn btn-xs btn-primary" onClick={this.deployTag}>Deploy</button>);
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
            <td data-column="buildtime">{displayTime(date)}</td>,
            <td data-column="size">{humanSize(data.size, 1)}</td>,
            <td data-column="deployed_at">{this.props.deployedAt}</td>,
        ], this.props.hasApprovalColumn ? [
            <td data-column="approved">{this.state.approvalContents}</td>,
        ] : [], [
            <td data-column="actions">{actions}</td>,
        ]);

        return <tr className={className}>{cells}</tr>;
    }

    loadData() {
        var self = this;
        if (this.props.approvalDB) {
            var db = this.props.approvalDB.child(this.props.data.tag);
            db.off();
            db.on('value', function(snapshot) {
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

    componentDidMount() {
        this.loadData();
    }

    componentDidUpdate(prevProps) {
        if (this.props != prevProps) {
            this.state = {};
            this.forceUpdate();
            this.loadData();
        }
    }

    componentWillUnmount() {
        if (this.props.approvalDB) {
            this.props.approvalDB.child(this.props.data.tag).off();
        }
    }

    deployTag(event) {
        var self = this;
        event.target.setAttribute('disabled', 'disabled');
        viliApi.deployments.create(this.props.env, this.props.app, {
            tag: this.props.data.tag,
            trigger: false
        }).then(function(deployment) {
            router.transitionTo(`/${self.props.env}/apps/${self.props.app}/deployments/${deployment.id}`);
        });
    }

    approveTag() {
        var url = prompt('Please enter the release url');
        if (!url) {
            return;
        }
        viliApi.releases.create(this.props.app, this.props.data.tag, {
            url: url,
        });
    }

    unapproveTag() {
        viliApi.releases.delete(this.props.app, this.props.data.tag);
    }
}

export class App extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};
        this.loadData = this.loadData.bind(this);
    }

    render() {
        if (!this.state.app) {
            return <Loading />;
        }
        var self = this;
        var columns = _.union([
            {title: 'Tag', key: 'tag'},
            {title: 'Branch', key: 'branch'},
            {title: 'Build Time', key: 'buildtime'},
            {title: 'Size', key: 'size'},
            {title: 'Deployed', key: 'deployed_at'},
        ], this.state.hasApprovalColumn ? [{title: 'Approved', key: 'approved'}] : [], [
            {title: 'Actions', key: 'actions'},
        ]);

        var rows = [];

        _.each(this.state.app.repository, function(data) {
            var date = new Date(data.lastModified);
            var deployed = data.tag === self.state.currentTag;
            var row = <Row data={data} currentTag={self.state.currentTag}
                           deployedAt={deployed ? self.state.deployedAt : ''}
                           canDeploy={self.state.canDeploy}
                           envNeedsApprovalForDeploy={self.state.envNeedsApprovalForDeploy}
                           hasApprovalColumn={self.state.hasApprovalColumn}
                           envCanApproveTag={self.state.envCanApproveTag}
                           approvalDB={self.state.approvalDB}
                           env={self.props.params.env}
                           app={self.props.params.app}
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
            app: viliApi.apps.get(this.props.params.env, this.props.params.app)
        }).then(function(state) {
            state.envCanApproveTag = _.contains(window.appconfig.approvalEnvs, self.props.params.env);
            state.envNeedsApprovalForDeploy = _.contains(window.appconfig.prodEnvs, self.props.params.env);
            state.hasApprovalColumn = state.envCanApproveTag || state.envNeedsApprovalForDeploy;
            if (state.hasApprovalColumn) {
                state.approvalDB = self.props.db.child('releases').child(self.props.params.app);
            }

            state.baseController = template(state.app.controllerTemplate, state.app.variables);
            state.canDeploy = state.baseController.valid;
            if (!state.canDeploy) {
                // TODO show message saying controller not valid
            }
            if (!state.app.controller || state.app.controller.status === 'Failure') {
                state.currentTag = null;
            } else {
                state.currentTag = state.app.controller.spec.template.spec.containers[0].image.split(':')[1];
            }
            state.deployedAt = state.app.controller ? displayTime(new Date(state.app.controller.metadata.creationTimestamp)) : '';
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
