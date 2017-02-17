import React from 'react';
import { viliApi } from '../lib';
import { Loading } from '../shared'; // eslint-disable-line no-unused-vars


export class AppSpec extends React.Component {
    constructor(props) {
        super(props);

        this.loadData = this.loadData.bind(this);
    }

    render() {
        if (!this.state || !this.state.deploymentSpec) {
            return <Loading />;
        }
        return (
            <div className="col-md-8">
                <div id="source-yaml">
                    <pre><code className="nix">
                        {this.state.deploymentSpec}
                    </code></pre>
                </div>
            </div>
        );
    }

    loadData() {
        var self = this;
        viliApi.apps.get(this.props.params.env, this.props.params.app).then(function(state) {
            self.setState(state)
        });
    }

    componentDidMount() {
        this.props.activateTab('spec');
        this.loadData();
    }
}
