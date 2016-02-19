import React from 'react';
import { Button, ButtonToolbar } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import _ from 'underscore';
import { Promise } from 'bluebird';
import { viliApi, displayTime } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars

export class AppPods extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};
        this.loadData = this.loadData.bind(this);
        this.scale = this.scale.bind(this);
    }

    render() {
        if (!this.state.appPods) {
            return <Loading />;
        }
        var self = this;
        var columns = _.union([
            {title: 'Name', key: 'name'},
            {title: 'Host', key: 'host'},
            {title: 'Deployment', key: 'deployment'},
            {title: 'Phase', key: 'phase'},
            {title: 'Ready', key: 'ready'},
            {title: 'Pod IP', key: 'pod_ip'},
            {title: 'Created', key: 'created'},
            {title: 'Actions', key: 'actions'},
        ]);

        var rows = _.map(this.state.appPods.items, function(pod) {
            var deployment = pod.metadata.labels.deployment;
            var ready = pod.status.phase === 'Running' &&
                _.every(pod.status.containerStatuses, function(cs) {
                    return cs.ready;
                });

            var nameLink = <Link to={`/${self.props.params.env}/pods/${pod.metadata.name}`}>{pod.metadata.name}</Link>;
            var hostLink = <Link to={`/${self.props.params.env}/nodes/${pod.spec.nodeName}`}>{pod.spec.nodeName}</Link>;
            var deploymentLink = <Link to={`/${self.props.params.env}/apps/${self.props.params.app}/deployments/${deployment}`}>{deployment}</Link>;
            var actions = [
                <Button onClick={self.deletePod.bind(self, pod.metadata.name)} bsStyle="danger" bsSize="xs">Delete</Button>
            ];
            return {
                name: nameLink,
                host: hostLink,
                deployment: deploymentLink,
                phase: pod.status.phase,
                ready: ready ? String.fromCharCode('10003') : '',
                pod_ip: pod.status.podIP,
                created: displayTime(new Date(pod.metadata.creationTimestamp)),
                actions: actions,
            };
        });

        return (
            <div>
                <ButtonToolbar className="pull-right">
                    <Button onClick={this.scale} bsStyle="success" bsSize="small">Scale</Button>
                </ButtonToolbar>
                <Table columns={columns} rows={rows} />
            </div>
        );
    }

    loadData() {
        var self = this;
        if (this.dataInterval) {
            clearInterval(this.dataInterval);
        }
        var loader = function() {
            Promise.props({
                appPods: viliApi.pods.get(
                    self.props.params.env,
                    {labelSelector: 'app=' + self.props.params.app}),
            }).then(function(state) {
                self.setState(state);
            });
        };
        loader();
        this.dataInterval = setInterval(loader, 3000);
    }

    componentDidMount() {
        this.props.activateTab('pods');
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
        if (this.dataInterval) {
            clearInterval(this.dataInterval);
        }
    }

    scale() {
        var replicas = prompt('Enter the number of replicas to scale to');
        if (!replicas) {
            return;
        }
        replicas = parseInt(replicas);
        if (_.isNaN(replicas) || replicas < 0) {
            return;
        }
        viliApi.apps.scale(this.props.params.env, this.props.params.app, replicas);
    }

    deletePod(pod) {
        viliApi.pods.delete(this.props.params.env, pod);
    }

}
