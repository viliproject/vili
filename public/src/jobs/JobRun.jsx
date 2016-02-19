import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import { Alert, Button } from 'react-bootstrap'; // eslint-disable-line no-unused-vars
import _ from 'underscore';
import moment from 'moment';
import 'moment-duration-format';
import { viliApi } from '../lib';
import { Table, Loading } from '../shared'; // eslint-disable-line no-unused-vars


class RunButtons extends React.Component { // eslint-disable-line no-unused-vars
    constructor(props) {
        super(props);
        this.start = this.start.bind(this);
        this.terminate = this.terminate.bind(this);
    }

    render() {
        if (!this.state || !this.state.val) {
            return <Loading />;
        }
        var contents = null;
        switch (this.state.val) {
            case 'new':
                contents = <Button bsStyle="success" onClick={this.start}>Start</Button>;
                break;
            case 'running':
                contents = <Button bsStyle="warning" onClick={this.terminate}>Terminate</Button>;
                break;
            case 'terminated':
                contents = <Alert bsStyle="danger">Terminated</Alert>;
                break;
            case 'failed':
                contents = <Alert bsStyle="danger">Failed</Alert>;
                break;
            case 'completed':
                contents = <Alert bsStyle="success">Run complete</Alert>;
                break;
        }
        return <div className="deploy-buttons col-md-10">{contents}</div>;
    }

    componentDidMount() {
        var self = this;
        this.props.runDB.child('state').on('value', function(snapshot) {
            self.setState({val: snapshot.val()});
        });
    }

    componentWillUnmount() {
        this.props.runDB.child('state').off();
    }

    start(event) {
        event.currentTarget.setAttribute('disabled', 'disabled');
        viliApi.runs.start(this.props.env, this.props.job, this.props.run);
    }

    terminate(event) {
        event.currentTarget.setAttribute('disabled', 'disabled');
        viliApi.runs.terminate(this.props.env, this.props.job, this.props.run);
    }
}

class Clock extends React.Component { // eslint-disable-line no-unused-vars
    render() {
        if (!this.state || !this.state.val) {
            return null;
        }
        return (
            <div className="deploy-clock col-md-2">
                {moment.duration(this.state.val).format('m[m]:ss[s]')}
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        this.props.runDB.child('clock').on('value', function(snapshot) {
            self.setState({val: snapshot.val()});
        });
    }

    componentWillUnmount() {
        this.props.runDB.child('clock').off();
    }
}

class Output extends React.Component { // eslint-disable-line no-unused-vars

    render() {
        if (!this.state || !this.state.val) {
            return <div></div>;
        }
        return (
            <div>
                <h3>Job Output</h3>
                <pre>{this.state.val}</pre>
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        this.props.runDB.child('output').on('value', function(snapshot) {
            self.setState({val: snapshot.val()});
        });
    }

    componentWillUnmount() {
        this.props.runDB.child('output').off();
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
        return (
            <div className="logs">
                <h3>Log</h3>
                <Table columns={columns} rows={rows} />
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        this.props.runDB.child('log').on('child_added', function(snapshot) {
            self.state.log.push(snapshot.val());
            self.forceUpdate();
        });
    }

    componentWillUnmount() {
        this.props.runDB.child('log').off();
    }
}

export class JobRun extends React.Component {
    render() {
        var runDB = this.props.db.child(this.props.params.env)
            .child('jobs').child(this.props.params.job)
            .child('runs').child(this.props.params.run);
        return (
            <div className="run">
                <div className="row">
                    <RunButtons env={this.props.params.env}
                                job={this.props.params.job}
                                run={this.props.params.run}
                                runDB={runDB} />
                    <Clock runDB={runDB} />
                </div>
                <Logs runDB={runDB} />
                <Output runDB={runDB} />
            </div>
        );
    }

    componentDidMount() {
        this.props.activateTab('runs');
    }
}
