import React from 'react';
import { Promise } from 'bluebird';
import { viliApi } from '../lib';
import { Loading } from '../shared'; // eslint-disable-line no-unused-vars

export class AppService extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};

        this.loadData = this.loadData.bind(this);
        this.clickCreateService = this.clickCreateService.bind(this);
    }

    render() {
        if (!this.state.app) {
            return <Loading />;
        }
        if (!this.state.app.service) {
            return (
                <div id="service">
                    <div className="alert alert-warning" role="alert">No Service Defined</div>
                    <div><button className="btn btn-success" onClick={this.clickCreateService}>Create Service</button></div>
                </div>
            );
        } else {
            return (
                <div id="service">
                    IP: {this.state.app.service.spec.clusterIP}
                </div>
            );
        }
    }

    loadData() {
        var self = this;
        Promise.props({
            app: viliApi.apps.get(this.props.params.env, this.props.params.app, {fields: 'service'})
        }).then(function(state) {
            self.setState(state);
        });
    }

    componentDidMount() {
        this.props.activateTab('service');
        this.loadData();
    }

    componentDidUpdate(prevProps) {
        if (this.props != prevProps) {
            this.state = {};
            this.forceUpdate();
            this.loadData();
        }
    }

    clickCreateService(event) {
        var self = this;
        event.currentTarget.setAttribute('disabled', 'disabled');
        viliApi.services.create(this.props.params.env, this.props.params.app).then(function() {
            self.loadData();
        });
    }
}
