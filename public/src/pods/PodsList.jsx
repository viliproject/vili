import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Promise } from 'bluebird';
import _ from 'underscore';
import { viliApi, displayTime } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars

export class PodsList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};
    }

    render() {
        var self = this;
        var header = (
            <div className="view-header">
                <ol className="breadcrumb">
                    <li><Link to={`/${this.props.params.env}`}>{this.props.params.env}</Link></li>
                    <li className="active">Pods</li>
                </ol>
            </div>
        );
        if (!this.state.pods) {
            return (
                <div>
                    {header}
                    <Loading />
                </div>
            );
        }
        var columns = [
            {title: 'Name', key: 'name'},
            {title: 'App', key: 'app'},
            {title: 'Node', key: 'node'},
            {title: 'Phase', key: 'phase'},
            {title: 'Ready', key: 'ready'},
            {title: 'Created', key: 'created'},
        ];

        var rows = _.map(this.state.pods.items, function(pod) {
            var ready = pod.status.phase === 'Running' &&
            _.every(pod.status.containerStatuses, function(cs) {
                return cs.ready;
            });
            var app = null;
            if (pod.metadata.labels && pod.metadata.labels) {
                app = <Link to={`/${self.props.params.env}/apps/${pod.metadata.labels.app}`}>{pod.metadata.labels.app}</Link>;
            }
            return {
                name: <Link to={`/${self.props.params.env}/pods/${pod.metadata.name}`}>{pod.metadata.name}</Link>,
                app: app,
                node: <Link to={`/${self.props.params.env}/nodes/${pod.spec.nodeName}`}>{pod.spec.nodeName}</Link>,
                phase: pod.status.phase,
                ready: ready ? String.fromCharCode('10003') : '',
                created: displayTime(new Date(pod.metadata.creationTimestamp)),
            };
        });

        return (
            <div>
                {header}
                <Table columns={columns} rows={rows} />
            </div>
        );
    }

    loadData() {
        var self = this;
        Promise.props({
            pods: viliApi.pods.get(this.props.params.env),
        }).then(function(props) {
            self.setState(props);
        });
    }

    componentDidMount() {
        this.props.activateSideNavItem(['pods']);
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
