import React from 'react';
import Router, { Link } from 'react-router'; // eslint-disable-line no-unused-vars
import _ from 'underscore';
import { displayTime } from '../lib';
import { Table } from '../shared'; // eslint-disable-line no-unused-vars

export class JobRuns extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            runs: {}
        };
        this.subscribeData = this.subscribeData.bind(this);
    }

    render() {
        var self = this;
        var columns = [
            {title: 'Time', key: 'time'},
            {title: 'Run', key: 'run'},
            {title: 'Tag', key: 'tag'},
            {title: 'Status', key: 'status'},
        ];

        var runs = _.map(this.state.runs, (x) => x);
        runs.reverse();
        var rows = _.map(runs, function(run) {
            var runLink = <Link to={`/${self.props.params.env}/jobs/${self.props.params.job}/runs/${run.id}`}>{run.id}</Link>;
            return {
                time: displayTime(new Date(run.time)),
                run: runLink,
                tag: run.tag,
                status: run.state || 'new',
            };
        });
        return <Table columns={columns} rows={rows} />;
    }

    subscribeData() {
        var self = this;
        if (this.runsDB) {
            this.runsDB.off();
        }
        this.runsDB = this.props.db.child(this.props.params.env)
            .child('jobs').child(this.props.params.job).child('runs');
        this.runsDB.orderByChild('time').on('child_added', function(snapshot) {
            var run = snapshot.val();
            self.state.runs[run.id] = run;
            self.setState({runs: self.state.runs});
        });
        this.runsDB.on('child_changed', function(snapshot) {
            var run = snapshot.val();
            self.state.runs[run.id] = run;
            self.setState({runs: self.state.runs});
        });
    }

    componentDidMount() {
        this.props.activateTab('runs');
        this.subscribeData();
    }

    componentDidUpdate(prevProps) {
        if (this.props.params != prevProps.params) {
            this.state.runs = [];
            this.forceUpdate();
            this.subscribeData();
        }
    }

    componentWillUnmount() {
        if (this.runsDB) {
            this.runsDB.off();
        }
    }
}
