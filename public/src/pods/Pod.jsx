import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Promise } from 'bluebird';
import { viliApi } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars


export class Pod extends React.Component {
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
                    <li><Link to={`/${this.props.params.env}/pods`}>Pods</Link></li>
                    <li className="active">{this.props.params.pod}</li>
                </ol>
            </div>
        );
        var pod = this.state.pod;
        if (!pod) {
            return (
                <div>
                    {header}
                    <Loading />
                </div>
            );
        }

        var app = null;
        if (pod.metadata.labels && pod.metadata.labels.app) {
            app = <Link to={`/${this.props.params.env}/apps/${pod.metadata.labels.app}`}>{pod.metadata.labels.app}</Link>;
        }

        return (
            <div>
                {header}
                <div>
                    <dl className="dl-horizontal">
                        <dt>Pod IP</dt>
                        <dd>{pod.status.podIP}</dd>
                        <dt>Phase</dt>
                        <dd>{pod.status.phase}</dd>
                        <dt>App</dt>
                        <dd>{app}</dd>
                        <dt>Node</dt>
                        <dd>
                            <Link to={`/${this.props.params.env}/nodes/${pod.spec.nodeName}`}>{pod.spec.nodeName}</Link>
                        </dd>
                    </dl>

                </div>
            </div>
        );
    }

    loadData() {
        var self = this;
        Promise.props({
            pod: viliApi.pods.get(this.props.params.env, this.props.params.pod)
        }).then(function(state) {
            self.setState(state);
        });
    }

    componentDidMount() {
        this.props.activateSideNavItem(['pods']);
        this.loadData();
    }

    componentDidUpdate(prevProps) {
        if (this.props.params != prevProps.params) {
            this.state.pod = null;
            this.forceUpdate();
            this.loadData();
        }
    }
}
