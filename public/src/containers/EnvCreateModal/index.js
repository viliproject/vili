import React from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import {
  Modal,
  Button,
  FormGroup,
  ControlLabel,
  FormControl,
  Panel,
} from "react-bootstrap"
import { Typeahead } from "react-bootstrap-typeahead"
import _ from "underscore"
import Immutable from "immutable"

import {
  hideCreateEnvModal,
  getBranches,
  createEnvironment,
  getEnvironmentSpec,
} from "../../actions/envs"
import history from "../../lib/history"

function mapStateToProps(state) {
  const envs = state.envs
  return {
    envs,
    showCreateModal: envs.get("showCreateModal"),
  }
}

const dispatchProps = {
  hideCreateEnvModal,
  getBranches,
  createEnvironment,
  getEnvironmentSpec,
}

export class EnvCreateModal extends React.Component {
  constructor(props) {
    super(props)

    this.state = {}

    this.loadSpec = _.debounce(this.loadSpec.bind(this), 200)
    this.createNewEnvironment = this.createNewEnvironment.bind(this)
  }

  componentWillMount() {
    const { showCreateModal } = this.props
    if (showCreateModal) {
      this.loadData()
    }
  }

  componentWillReceiveProps(nextProps) {
    const { showCreateModal } = this.props
    const { showCreateModal: nextShowCreateModal } = nextProps
    if (nextShowCreateModal && !showCreateModal) {
      this.loadData()
    }
  }

  loadData = () => {
    this.props.getBranches()
  }

  async createNewEnvironment() {
    const { results, error } = await this.props.createEnvironment({
      name: this.state.name,
      branch: this.state.branch,
      spec: this.state.spec,
    })
    if (error) {
      this.setState({ error })
      return
    }
    history.push(`/${this.state.name}/releases/${results.release.name}`)
    this.hide()
  }

  hide = () => {
    this.setState({
      name: null,
      branch: null,
      template: null,
      error: null,
    })
    this.props.hideCreateEnvModal()
  }

  onNameChange = event => {
    var name = event.target.value
    this.setState({
      name: name,
      createdResources: null,
      error: null,
    })
    this.loadSpec(name, this.state.branch)
  }

  onBranchChange = results => {
    const branch = results[0].label
    this.setState({
      branch: branch,
      createdResources: null,
      error: null,
    })
    this.loadSpec(this.state.name, branch)
  }

  async loadSpec(name, branch) {
    if (!name || !branch) {
      return
    }
    const { results, error } = await this.props.getEnvironmentSpec(name, branch)
    if (!error) {
      this.setState({ spec: results.spec })
    }
  }

  onSpecChange = event => {
    this.setState({
      spec: event.target.value,
      createdResources: null,
      error: null,
    })
  }

  render() {
    const { envs } = this.props
    if (!envs.get("showCreateModal")) {
      return null
    }

    let actionButton = null
    if (!this.state.createdResources) {
      actionButton = (
        <Button
          bsStyle="primary"
          onClick={this.createNewEnvironment}
          disabled={!this.state.spec || Boolean(this.state.error)}
        >
          Create
        </Button>
      )
    }

    let specForm = null
    if (this.state.name && this.state.branch) {
      specForm = (
        <FormGroup controlId="environmentSpec">
          <ControlLabel>Environment Spec</ControlLabel>
          <FormControl
            componentClass="textarea"
            value={this.state.spec}
            onChange={this.onSpecChange}
            style={{ height: "400px" }}
            disabled={this.state.createdResources}
          />
        </FormGroup>
      )
    }

    let output = null
    if (this.state.error) {
      var errorMessage = _.map(this.state.error.split("\n"), function(
        text,
        ix
      ) {
        return <div key={ix}>{text}</div>
      })
      output = (
        <Panel bsStyle="danger">
          <Panel.Heading>Error</Panel.Heading>
          <Panel.Body>{errorMessage}</Panel.Body>
        </Panel>
      )
    }

    const branches = envs
      .get("branches", Immutable.List())
      .map(branch => {
        return {
          label: branch,
        }
      })
      .toJS()
    return (
      <Modal show onHide={this.hide}>
        <Modal.Header closeButton>
          <Modal.Title>Create New Environment</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <FormGroup controlId="environmentName">
            <ControlLabel>Environment Name</ControlLabel>
            <FormControl
              componentClass="input"
              type="text"
              value={this.state.name}
              placeholder="my-feature-environment"
              onChange={this.onNameChange}
              disabled={this.state.createdResources}
            />
          </FormGroup>
          <FormGroup>
            <ControlLabel>Default Branch</ControlLabel>
            <Typeahead
              options={branches}
              labelKey="label"
              onChange={this.onBranchChange}
              disabled={this.state.createdResources}
            />
          </FormGroup>
          {specForm}
          {output}
        </Modal.Body>
        <Modal.Footer>
          <Button onClick={this.hide}>Close</Button>
          {actionButton}
        </Modal.Footer>
      </Modal>
    )
  }
}

EnvCreateModal.propTypes = {
  envs: PropTypes.object,
  showCreateModal: PropTypes.bool,
  hideCreateEnvModal: PropTypes.func.isRequired,
  getBranches: PropTypes.func.isRequired,
  createEnvironment: PropTypes.func.isRequired,
  getEnvironmentSpec: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(EnvCreateModal)
