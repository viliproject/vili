import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import _ from 'underscore';
import { Promise } from 'bluebird';
import { viliApi, displayTime } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars


export class Node extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};
        this.loadData = this.loadData.bind(this);
    }

    render() {
        var header = (
            <div className="view-header">
                <ol className="breadcrumb">
                    <li><Link to={`/${this.props.params.env}`}>{this.props.params.env}</Link></li>
                    <li><Link to={`/${this.props.params.env}/nodes`}>Nodes</Link></li>
                    <li className="active">{this.props.params.node}</li>
                </ol>
            </div>
        );
        if (!this.state.node) {
            return (
                <div>
                    {header}
                    <Loading />
                </div>
            );
        }
        var self = this;
        var columns = _.union([
            {title: 'Name', key: 'name'},
            {title: 'App', key: 'app'},
            {title: 'Pod IP', key: 'pod_ip'},
            {title: 'Created', key: 'created'},
            {title: 'Phase', key: 'phase'},
        ]);

        var rows = _.map(this.state.nodePods.items, function(pod) {
            var app = null;
            if (pod.metadata.labels && pod.metadata.labels) {
                app = <Link to={`/${self.props.params.env}/apps/${pod.metadata.labels.app}`}>{pod.metadata.labels.app}</Link>;
            }
            return {
                name: <Link to={`/${self.props.params.env}/pods/${pod.metadata.name}`}>{pod.metadata.name}</Link>,
                app: app,
                phase: pod.status.phase,
                pod_ip: pod.status.podIP,
                created: displayTime(new Date(pod.metadata.creationTimestamp)),
            };
        });

        return (
            <div>
                {header}
                <div>
                    <h3>Pods</h3>
                    <Table columns={columns} rows={rows} />
                </div>
            </div>
        );
    }

    loadData() {
        var self = this;
        Promise.props({
            node: viliApi.nodes.get(this.props.params.env, this.props.params.node),
            nodePods: viliApi.pods.get(
                this.props.params.env,
                {fieldSelector: 'spec.nodeName=' + this.props.params.node}),
        }).then(function(state) {
            self.setState(state);
        });
    }

    componentDidMount() {
        this.props.activateSideNavItem(['nodes']);
        this.loadData();
    }

    componentDidUpdate(prevProps) {
        if (this.props.params != prevProps.params) {
            this.state.node = null;
            this.forceUpdate();
            this.loadData();
        }
    }
}
