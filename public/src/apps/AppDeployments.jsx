import React from 'react';
import Router, { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import _ from 'underscore';
import { displayTime } from '../lib';
import { Table } from '../shared'; // eslint-disable-line no-unused-vars

export class AppDeployments extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            deployments: {}
        };
        this.subscribeData = this.subscribeData.bind(this);
    }

    render() {
        var self = this;
        var columns = [
            {title: 'Time', key: 'time'},
            {title: 'Deployment', key: 'deployment'},
            {title: 'From Tag', key: 'fromtag'},
            {title: 'To Tag', key: 'totag'},
            {title: 'Status', key: 'status'},
        ];

        var deployments = _.map(this.state.deployments, (x) => x);
        deployments.reverse();
        var rows = _.map(deployments, function(deployment) {
            var deploymentLink = <Link to={`/${self.props.params.env}/apps/${self.props.params.app}/deployments/${deployment.id}`}>{deployment.id}</Link>;
            return {
                time: displayTime(new Date(deployment.time)),
                deployment: deploymentLink,
                fromtag: deployment.fromTag,
                totag: deployment.tag,
                status: deployment.state || 'new',
            };
        });
        return <Table columns={columns} rows={rows} />;
    }

    subscribeData() {
        var self = this;
        if (this.deploymentsDB) {
            this.deploymentsDB.off();
        }
        this.deploymentsDB = this.props.db.child(this.props.params.env)
            .child('apps').child(this.props.params.app).child('deployments');
        this.deploymentsDB.orderByChild('time').on('child_added', function(snapshot) {
            var deployment = snapshot.val();
            self.state.deployments[deployment.id] = deployment;
            self.setState({deployments: self.state.deployments});
        });
        this.deploymentsDB.on('child_changed', function(snapshot) {
            var deployment = snapshot.val();
            self.state.deployments[deployment.id] = deployment;
            self.setState({deployments: self.state.deployments});
        });
    }

    componentDidMount() {
        this.props.activateTab('deployments');
        this.subscribeData();
    }

    componentDidUpdate(prevProps) {
        if (this.props != prevProps) {
            this.state = {
                deployments: {}
            };
            this.forceUpdate();
            this.subscribeData();
        }
    }

    componentWillUnmount() {
        if (this.deploymentsDB) {
            this.deploymentsDB.off();
        }
    }
}
