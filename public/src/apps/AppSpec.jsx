import React from 'react';
import _ from 'underscore';
import hljs from 'highlight.js';
import { Promise } from 'bluebird';
import { viliApi, template } from '../lib';
import { Loading } from '../shared'; // eslint-disable-line no-unused-vars


export class AppSpec extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            display: 'template',
        };

        this.loadData = this.loadData.bind(this);
        this.clickTemplate = this.clickTemplate.bind(this);
        this.clickPopulated = this.clickPopulated.bind(this);
    }

    render() {
        if (!this.state.baseController) {
            return <Loading />;
        }
        var usedVariables = _.map(this.state.baseController.usedVariables, function(variable) {
            return (
                <tr><td>{variable.key}</td><td>{variable.value}</td></tr>
            );
        });
        var missingVariables = _.map(this.state.baseController.missingVariables, function(variable) {
            return (
                <tr><td>{variable}</td><td><span class="text-danger">missing</span></td></tr>
            );
        });
        return (
            <div className="spec">
                <div className="row">
                    <div className="col-md-6">
                        <div id="source-yaml">
                            <pre><code className="nix" ref={
                                  function(node) { if (node) { hljs.highlightBlock(node.getDOMNode()); } } }>
                                {this.state.baseController[this.state.display]}
                            </code></pre>
                        </div>
                    </div>
                    <div className="col-md-6">
                        <div id="template-buttons" className="btn-group" data-toggle="buttons">
                            <label className={'btn btn-primary' + (this.state.display === 'template' ? ' active' : '')}>
                                <input type="radio" autocomplete="off" onClick={this.clickTemplate} />Template
                            </label>
                            <label className={'btn btn-primary' + (this.state.display === 'populated' ? ' active' : '')}>
                                <input type="radio" autocomplete="off" onClick={this.clickPopulated} />Populated
                            </label>
                        </div>
                        <h4>Variables</h4>
                        <table id="variables" className="table">
                            <tr>
                                <th>Key</th><th>Value</th>
                            </tr>
                            {usedVariables}
                            {missingVariables}
                        </table>
                    </div>
                </div>
            </div>
        );
    }

    loadData() {
        var self = this;
        Promise.props({
            app: viliApi.apps.get(this.props.params.env, this.props.params.app)
        }).then(function(state) {
            state.baseController = template(state.app.controllerTemplate, state.app.variables);
            self.setState(state);
        });
    }

    componentDidMount() {
        this.props.activateTab('spec');
        this.loadData();
    }

    componentDidUpdate(prevProps) {
        if (this.props.params != prevProps.params) {
            this.state.app = null;
            this.forceUpdate();
            this.loadData();
        }
    }

    clickTemplate() {
        this.setState({display: 'template'});
    }

    clickPopulated() {
        this.setState({display: 'populated'});
    }
}
