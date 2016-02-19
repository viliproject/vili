import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Button } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import { Promise } from 'bluebird';
import _ from 'underscore';
import humanSize from 'human-size';
import { viliApi, displayTime } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars

export class NodesList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};

        this.setNodeSchedulable = this.setNodeSchedulable.bind(this);
    }

    render() {
        var header = (
            <div className="view-header">
                <ol className="breadcrumb">
                    <li><Link to={`/${this.props.params.env}`}>{this.props.params.env}</Link></li>
                    <li className="active">Nodes</li>
                </ol>
            </div>
        );
        if (!this.state.nodes) {
            return (
                <div>
                    {header}
                    <Loading />
                </div>
            );
        }
        var self = this;
        var columns = [
            {title: 'Host', key: 'host'},
            {title: 'Instance Type', key: 'instance_type'},
            {title: 'Role', key: 'role'},
            {title: 'Capacity', subcolumns: [
                {title: 'CPU', key: 'cpu_capacity'},
                {title: 'Memory', key: 'memory_capacity'},
                {title: 'Pods', key: 'pods_capacity'},
            ]},
            {title: 'Versions', subcolumns: [
                {title: 'CoreOS', key: 'os_version'},
                {title: 'Kubelet', key: 'kubelet_version'},
                {title: 'Proxy', key: 'proxy_version'},
            ]},
            {title: 'Created', key: 'created'},
            {title: 'Status', key: 'status'},
            {title: 'Actions', key: 'actions'},
        ];

        var rows = _.map(
            this.state.nodes.items,
            function(node) {
                var name = node.metadata.name;
                var memory = /(\d+)Ki/g.exec(node.status.capacity.memory);
                if (memory) {
                    memory = humanSize(parseInt(memory[1]) * 1024, 1);
                } else {
                    memory = node.status.capacity.memory;
                }
                var node_statuses = [];
                if (node.status.conditions[0].status === 'Unknown') {
                    node_statuses.push('NotReady');
                } else {
                    node_statuses.push(node.status.conditions[0].type);
                }
                var actions;
                if (node.spec.unschedulable === true) {
                    actions = (
                        <Button bsStyle="success" bsSize="xs"
                                onClick={self.setNodeSchedulable.bind(self, name, 'enable')}>
                            Enable
                        </Button>
                    );
                    node_statuses.push('Disabled')
                } else {
                    actions = (
                        <Button bsStyle="danger" bsSize="xs"
                                onClick={self.setNodeSchedulable.bind(self, name, 'disable')}>
                            Disable
                        </Button>
                    );
                }
                return {
                    host: <Link to={`/${self.props.params.env}/nodes/${name}`}>{name}</Link>,
                    instance_type: node.metadata.labels['airware.io/instance-type'],
                    role: node.metadata.labels['airware.io/role'],
                    cpu_capacity: node.status.capacity.cpu,
                    memory_capacity: memory,
                    pods_capacity: node.status.capacity.pods,
                    os_version: node.status.nodeInfo.osImage,
                    kubelet_version: node.status.nodeInfo.kubeletVersion,
                    proxy_version: node.status.nodeInfo.kubeProxyVersion,
                    created: displayTime(new Date(node.metadata.creationTimestamp)),
                    status: node_statuses.join(','),
                    actions: actions
                };
            });

        return (
            <div>
                {header}
                <Table columns={columns} rows={rows} />
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        this.props.activateSideNavItem(['nodes']);
        Promise.props({
            nodes: viliApi.nodes.get(this.props.params.env),
        }).then(function(props) {
            self.setState(props);
        });
    }

    setNodeSchedulable(node, action) {
        var self = this;
        viliApi.nodes.setSchedulable(this.props.params.env, node, action)
               .then(function() {
                   return viliApi.nodes.get(self.props.params.env);
               }).then(function(nodes) {
                   self.setState({nodes: nodes});
               });
    }
}
