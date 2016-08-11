import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Alert, ButtonGroup, Button, FormControl } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import _ from 'underscore';
import moment from 'moment';
import 'moment-duration-format';
import { viliApi, displayTime } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars


const rolloutStrategies = [
    'RollingUpdate',
    'Recreate',
];

class DeploymentHeader extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.state = this.stateFromRollout(props.rollout)

        this.deploy = this.deploy.bind(this);
        this.rollback = this.rollback.bind(this);
        this.pause = this.pause.bind(this);

        this.onStrategyChange = this.onStrategyChange.bind(this);
        this.saveRollout = this.saveRollout.bind(this);
    }

    onStrategyChange(event) {
        var self = this;
        this.setState({strategy: event.target.value}, function() {
            self.saveRollout();
        });
    }

    saveRollout() {
        var rollout = {
            strategy: this.state.strategy,
        };
        viliApi.deployments.setRollout(this.props.env, this.props.app, this.props.deployment, rollout);
    }

    render() {
        var banner = null;
        var buttons = null;
        var readOnlyForm = true;
        switch (this.props.state) {
            case 'new':
                buttons = <Button bsStyle="success" onClick={this.deploy} disabled={this.state.disabled}>Deploy</Button>;
                readOnlyForm = false;
                break;
            case 'running':
                buttons = <Button bsStyle="warning" onClick={this.pause} disabled={this.state.disabled}>Pause</Button>;
                break;
            case 'pausing':
                buttons = [
                    <Button bsStyle="warning" disabled={true}>Pausing...</Button>,
                    <Button bsStyle="warning" onClick={this.pause} disabled={this.state.disabled}>Force Pause</Button>,
                ];
                break;
            case 'paused':
                buttons = [
                    <Button bsStyle="success" onClick={this.deploy} disabled={this.state.disabled}>Resume</Button>,
                    <Button bsStyle="danger" onClick={this.rollback} disabled={this.state.disabled}>Rollback</Button>,
                ];
                readOnlyForm = false;
                break;
            case 'rollingback':
                buttons = <Button bsStyle="danger" disabled={true}>Rolling Back...</Button>;
                break;
            case 'rolledback':
                banner = <Alert bsStyle="danger">Rolled back</Alert>;
                break;
            case 'completed':
                banner = <Alert bsStyle="success">Deployment complete</Alert>;
                break;
        }
        var strategies = _.map(rolloutStrategies, function(strategy) {
            return <option value={strategy}>{strategy}</option>;
        });
        var strategySelect = readOnlyForm ? (
            <p className="form-control-static">
                {this.state.strategy}
            </p>
          ) : (
            <FormControl
                componentClass="select"
                value={this.state.strategy}
                onChange={this.onStrategyChange}>
                {strategies}
            </FormControl>
        );
        var form = (
            <form className="form-horizontal">
                <label labelClassName="col-xs-4" wrapperClassName="col-xs-8">Rollout Strategy</label>
                {strategySelect}
            </form>
        );
        return (
            <div className="deployment-header">
                {banner}
                <div className="row">
                    <div className="col-md-6">{form}</div>
                    <div className="col-md-4">
                        <ButtonGroup className="pull-right">
                            {buttons}
                        </ButtonGroup>
                    </div>
                    <div className="col-md-2">
                        <Clock deploymentDB={this.props.deploymentDB} />
                    </div>
                </div>
            </div>
        );
    }

    deploy() {
        this.setState({disabled: true});
        viliApi.deployments.resume(this.props.env, this.props.app, this.props.deployment);
    }

    pause() {
        this.setState({disabled: true});
        viliApi.deployments.pause(this.props.env, this.props.app, this.props.deployment);
    }

    rollback() {
        this.setState({disabled: true});
        viliApi.deployments.rollback(this.props.env, this.props.app, this.props.deployment);
    }

    stateFromRollout(rollout) {
        if (!rollout) {
            return {
                strategy: 'RollingUpdate',
            };
        }
        return {
            strategy: rollout.strategy ? rollout.strategy : 'RollingUpdate',
        };
    }

    componentDidUpdate(prevProps) {
        if (this.props.state != prevProps.state) {
            this.setState({disabled: false});
        }
        if (this.props.rollout != prevProps.rollout) {
            this.setState(this.stateFromRollout(this.props.rollout));
        }
    }
}

class Clock extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.state = {};
    }

    render() {
        return (
            <div className="deploy-clock">
                {moment.duration(this.state.val || 0).format('m[m]:ss[s]')}
            </div>
        );
    }

    loadData() {
        var self = this;
        this.props.deploymentDB.child('clock').off();
        this.props.deploymentDB.child('clock').on('value', function(snapshot) {
            self.setState({val: snapshot.val()});
        });
    }

    componentDidMount() {
        this.loadData();
    }

    componentWillUnmount() {
        this.props.deploymentDB.child('clock').off();
    }
}

class FromPods extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.state = {};
    }

    render() {
        var self = this;
        var columns = [
            {title: 'Name', key: 'name'},
            {title: 'Created', key: 'created'},
            {title: 'Phase', key: 'phase'},
            {title: 'Ready', key: 'ready'},
            {title: 'Host', key: 'host'},
        ];

        var podsMap = {};
        var originalKeys = [];
        _.each(this.state.originalPods, function(pod) {
            podsMap[pod.name] = pod;
            originalKeys.push(pod.name);
        });

        var fromKeys = [];
        _.each(this.state.fromPods, function(pod) {
            podsMap[pod.name] = pod;
            fromKeys.push(pod.name);
        });

        var allKeys = _.union(originalKeys, fromKeys);

        var readyCount = 0;
        var desiredReplicas = this.props.desiredReplicas || 0;
        var rows = _.map(allKeys, function(key) {
            var pod = podsMap[key]
            var deleted = !_.contains(fromKeys, key);

            var nameLink = deleted ? pod.name : (<Link to={`/${self.props.env}/pods/${pod.name}`}>{pod.name}</Link>);
            var hostLink = pod.host ? (<Link to={`/${self.props.env}/nodes/${pod.host}`}>{pod.host}</Link>) : '';

            if (!deleted && pod.ready) {
                readyCount += 1;
            }
            return {
                _className: deleted ? 'text-muted' : '',
                name: nameLink,
                created: displayTime(new Date(pod.created)),
                phase: deleted ? 'Deleted' : pod.phase,
                ready: !deleted && pod.ready ? String.fromCharCode('10003') : '',
                host: hostLink,
            };
        });
        return (
            <div className="col-md-6">
                <h3>{`From Pods (${readyCount}/${desiredReplicas})`}</h3>
                <Table columns={columns} rows={rows} />
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        _.each(['fromPods', 'originalPods'], function(key) {
            self.props.deploymentDB.child(key).on('value', function(snapshot) {
                var d = {};
                d[key] = snapshot.val();
                self.setState(d);
            });
        });
    }

    componentWillUnmount() {
        var self = this;
        _.each(['fromPods', 'originalPods'], function(key) {
            self.props.deploymentDB.child(key).off();
        });
    }
}

class ToPods extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.state = {};
    }

    render() {
        var self = this;
        var columns = [
            {title: 'Name', key: 'name'},
            {title: 'Created', key: 'created'},
            {title: 'Phase', key: 'phase'},
            {title: 'Ready', key: 'ready'},
            {title: 'Host', key: 'host'},
        ];

        var readyCount = 0;
        var desiredReplicas = this.props.desiredReplicas || 0;
        var rows = _.map(this.state.toPods, function(pod) {
            var nameLink = <Link to={`/${self.props.env}/pods/${pod.name}`}>{pod.name}</Link>;
            var hostLink = pod.host ? <Link to={`/${self.props.env}/nodes/${pod.host}`}>{pod.host}</Link> : '';
            if (pod.ready) {
                readyCount += 1;
            }
            return {
                name: nameLink,
                created: displayTime(new Date(pod.created)),
                phase: pod.phase,
                ready: pod.ready ? String.fromCharCode('10003') : '',
                host: hostLink,
            };
        });
        return (
            <div className="col-md-6">
                <h3>{`To Pods (${readyCount}/${desiredReplicas})`}</h3>
                <Table columns={columns} rows={rows} />
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        _.each(['toPods'], function(key) {
            self.props.deploymentDB.child(key).on('value', function(snapshot) {
                var d = {};
                d[key] = snapshot.val();
                self.setState(d);
            });
        });
    }

    componentWillUnmount() {
        var self = this;
        _.each(['toPods'], function(key) {
            self.props.deploymentDB.child(key).off();
        });
    }
}

class Logs extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.state = {
            log: []
        };
    }

    render() {
        var columns = [
            {title: 'Time', key: 'time'},
            {title: 'Message', key: 'message'},
        ];

        var rows = _.map(this.state.log, function(item) {
            return {
                time: moment(new Date(item.time)).format('YYYY-MM-DD HH:mm:ss'),
                message: item.msg,
            };
        });
        rows.reverse();
        return (
            <div className="logs">
                <h3>Log</h3>
                <Table columns={columns} rows={rows} />
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        this.props.deploymentDB.child('log').on('child_added', function(snapshot) {
            self.state.log.push(snapshot.val());
            self.forceUpdate();
        });
    }

    componentWillUnmount() {
        this.props.deploymentDB.child('log').off();
    }
}

export class AppDeployment extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};

        this.loadData = this.loadData.bind(this);
        this.getDeploymentDB = this.getDeploymentDB.bind(this);
    }

    render() {
        if (!this.state.state) {
            return <Loading />;
        }
        var deploymentDB = this.getDeploymentDB();
        return (
            <div className="deployment">
                <DeploymentHeader env={this.props.params.env}
                                  app={this.props.params.app}
                                  deployment={this.props.params.deployment}
                                  deploymentDB={deploymentDB}
                                  state={this.state.state}
                                  desiredReplicas={this.state.desiredReplicas}
                                  rollout={this.state.rollout}
                />
                <div className="row">
                    <FromPods env={this.props.params.env} desiredReplicas={this.state.desiredReplicas}
                              deploymentDB={deploymentDB} />
                    <ToPods env={this.props.params.env} desiredReplicas={this.state.desiredReplicas}
                            deploymentDB={deploymentDB} />
                </div>
                <Logs deploymentDB={deploymentDB} />
            </div>
        );
    }

    componentDidMount() {
        this.props.activateTab('deployments');
        this.loadData();
    }

    loadData() {
        var self = this;
        var deploymentDB = this.getDeploymentDB();
        _.each(['state', 'desiredReplicas', 'rollout'], function(key) {
            deploymentDB.child(key).off();
            deploymentDB.child(key).on('value', function(snapshot) {
                var d = {};
                d[key] = snapshot.val();
                self.setState(d);
            });
        });
    }

    componentWillUnmount() {
        var deploymentDB = this.getDeploymentDB();
        _.each(['state', 'desiredReplicas', 'rollout'], function(key) {
            deploymentDB.child(key).off();
        });
    }

    getDeploymentDB() {
        return this.props.db.child(this.props.params.env)
                   .child('apps').child(this.props.params.app)
                   .child('deployments').child(this.props.params.deployment);
    }

}
