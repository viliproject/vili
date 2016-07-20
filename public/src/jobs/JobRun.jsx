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
        if (!this.props.status) {
            return <Loading />;
        }

        var contents = null;
        switch (this.props.status) {
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

class Variables extends React.Component {
    constructor(props) {
        super(props);
        this.setVariables = this.setVariables.bind(this);
        this.onVariableChange = this.onVariableChange.bind(this);
    }

    componentDidMount() {
        var self = this;
        this.props.runDB.child('variables').on('value', function(snapshot) {
            self.setState({val: snapshot.val()});
        });
    }

    componentWillUnmount() {
        this.props.runDB.child('variables').off();
    }

    render() {
        if (!this.state || !this.state.val) {
            return <div></div>;
        }
        var self = this;
        var variables = _.map(_.pairs(this.state.val), function(pair) {
            return <VariableRow name={pair[0]} value={pair[1]} onChange={self.onVariableChange(pair[0])} disabled={self.props.status != 'new'} />;
        });

        return <table id="variables" className="table">
            <tr>
                <th>Key</th><th>Value</th>
            </tr>
            {variables}
        </table>
    }

    setVariables(vars) {
        viliApi.runs.setVariables(this.props.env, this.props.job, this.props.run, vars);
    }

    onVariableChange(variable) {
        var self = this;
        return function(event) {
            var newVal = {};
            newVal[variable] = event.target.value;
            newVal = _.extend(self.state.val, newVal);
            self.setVariables(newVal);
            self.setState({val: newVal});
        }
    }
}

class VariableRow extends React.Component {
    render() {
        return <tr>
            <td>{this.props.name}</td>
            <td><input type="text" defaultValue={this.props.value} onChange={this.props.onChange} disabled={this.props.disabled}/></td>
        </tr>
    }
}

class Output extends React.Component { // eslint-disable-line no-unused-vars

    render() {
        if (!this.state || !this.state.val) {
            return <div />;
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
        if (!this.state || !this.state.runDB) {
            return <div />;
        }
        return (
            <div className="run">
                <div className="row">
                    <RunButtons env={this.props.params.env}
                                job={this.props.params.job}
                                run={this.props.params.run}
                                status={this.state.status}
                                runDB={this.state.runDB} />
                    <Clock runDB={this.state.runDB} />
                </div>
                <Variables env={this.props.params.env}
                           job={this.props.params.job}
                           run={this.props.params.run}
                           status={this.state.status}
                           runDB={this.state.runDB} />
                <Logs runDB={this.state.runDB} />
                <Output runDB={this.state.runDB} />
            </div>
        );
    }

    componentDidMount() {
        var self = this;
        var runDB = this.props.db.child(this.props.params.env)
            .child('jobs').child(this.props.params.job)
            .child('runs').child(this.props.params.run);
        this.props.activateTab('runs');
        runDB.child('state').on('value', function(snapshot) {
            self.setState({status: snapshot.val()});
        });
        this.setState({runDB: runDB});
    }

    componentWillUnmount() {
        if (this.state && this.state.runDB) {
            this.state.runDB.child('state').off();
        }
    }
}
