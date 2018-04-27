import React from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import {
  Button,
  ButtonToolbar,
  FormGroup,
  FormControl,
  ControlLabel,
  HelpBlock,
} from "react-bootstrap"
import { Link } from "react-router-dom"
import _ from "underscore"

import { activateNav } from "../../actions/app"
import { getReleaseSpec, createRelease } from "../../actions/releases"
import { makeLookUpObjects } from "../../selectors"

import ReleaseCreateWavePanel from "./ReleaseCreateWavePanel"

function makeMapStateToProps() {
  const lookUpDeployments = makeLookUpObjects()
  const lookUpReplicaSets = makeLookUpObjects()
  const lookUpJobRuns = makeLookUpObjects()
  return (state, ownProps) => {
    const { envName } = ownProps
    const env = state.envs.getIn(["envs", envName])
    const releaseEnv = state.releases.lookUp(envName)
    const deployments = lookUpDeployments(state.deployments, envName)
    const replicaSets = lookUpReplicaSets(state.replicaSets, envName)
    const jobRuns = lookUpJobRuns(state.jobRuns, envName)
    return {
      env,
      releaseEnv,
      deployments,
      replicaSets,
      jobRuns,
    }
  }
}

const dispatchProps = {
  activateNav,
  getReleaseSpec,
  createRelease,
}

function updateTargetVersion(target, env, deployments, replicaSets, jobRuns) {
  switch (target.type) {
    case "action":
      target.branch = env.branch
      return
    case "job":
      const run = jobRuns
        .filter(x => x.hasLabel("job", target.name))
        .sortBy(x => -x.creationTimestamp)
        .first()
      if (run) {
        target.tag = run.imageTag
        target.branch = run.imageBranch || env.branch
        target.runAt = run.runAt
      }
      return
    case "app":
      const deployment = deployments.find(
        d => d.getIn(["metadata", "name"]) === target.name
      )
      if (deployment) {
        target.tag = deployment.imageTag
        target.branch = deployment.imageBranch || env.branch
        const replicaSet = replicaSets
          .filter(
            x =>
              x.hasLabel("app", target.name) &&
              x.revision === deployment.revision
          )
          .sortBy(x => -x.creationTimestamp)
          .first()
        if (replicaSet) {
          target.deployedAt = replicaSet.deployedAt
        }
      }
  }
}

export class ReleaseCreate extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      releaseName: "",
      releaseNameValidation: "warning",
      releaseNameHelp: "Release name cannot be empty",
      releaseLink: "",
    }
  }

  componentDidMount() {
    this.props.activateNav("releases")
    this.props.getReleaseSpec(this.props.envName)
  }

  handleNameChange = e => {
    const releaseName = e.target.value
    let releaseNameValidation = null
    let releaseNameHelp = null
    if (releaseName !== releaseName.replace(/([^a-z0-9]+)/gi, "")) {
      releaseNameValidation = "error"
      releaseNameHelp = "Release name must be alphanumeric"
    } else if (!releaseName) {
      releaseNameValidation = "warning"
      releaseNameHelp = "Release name cannot be empty"
    }
    this.setState({ releaseName, releaseNameValidation, releaseNameHelp })
  }

  handleLinkChange = e => {
    this.setState({ releaseLink: e.target.value })
  }

  getSpec() {
    const { releaseEnv, env, deployments, replicaSets, jobRuns } = this.props
    if (!releaseEnv.spec) {
      return
    }
    const spec = JSON.parse(JSON.stringify(releaseEnv.spec))
    _.each(spec.waves, (wave, ix) => {
      _.each(wave.targets, target => {
        updateTargetVersion(target, env, deployments, replicaSets, jobRuns)
      })
    })
    return spec
  }

  createRelease = event => {
    event.target.setAttribute("disabled", "disabled")
    const { releaseName, releaseNameValidation, releaseLink } = this.state
    if (releaseNameValidation) {
      return
    }
    const { envName } = this.props
    const spec = this.getSpec()
    const release = {
      name: releaseName,
      link: releaseLink,
      waves: spec.waves,
    }
    this.props.createRelease(envName, release)
  }

  renderForm() {
    const { env } = this.props
    return (
      <form>
        <FormGroup>
          <ControlLabel>Deployed To</ControlLabel>
          <FormControl.Static>
            {(env && env.deployedToEnv) || ""}
          </FormControl.Static>
        </FormGroup>
        <FormGroup validationState={this.state.releaseNameValidation}>
          <ControlLabel>Name</ControlLabel>
          <FormControl
            type="text"
            value={this.state.releaseName}
            placeholder="Release name"
            onChange={this.handleNameChange}
          />
          <FormControl.Feedback />
          <HelpBlock>{this.state.releaseNameHelp}</HelpBlock>
        </FormGroup>
        <FormGroup>
          <ControlLabel>Link</ControlLabel>
          <FormControl
            type="text"
            value={this.state.releaseLink}
            placeholder="Release link"
            onChange={this.handleLinkChange}
          />
        </FormGroup>
      </form>
    )
  }

  renderWavePanels() {
    const { env } = this.props
    const spec = this.getSpec()
    if (spec) {
      return _.map(spec.waves, (wave, ix) => {
        return (
          <ReleaseCreateWavePanel
            key={ix}
            ix={ix}
            env={env.name}
            targets={wave.targets}
          />
        )
      })
    }
    return null
  }

  render() {
    const { envName } = this.props
    const header = [
      <div key="header" className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li>
            <Link to={`/${envName}/releases`}>Releases</Link>
          </li>
          <li className="active">New</li>
        </ol>
        <ButtonToolbar key="toolbar" className="pull-right">
          <Button
            onClick={this.createRelease}
            bsStyle="primary"
            bsSize="small"
            disabled={!!this.state.releaseNameValidation}
          >
            Create
          </Button>
        </ButtonToolbar>
      </div>,
    ]

    return (
      <div>
        {header}
        {this.renderForm()}
        <h5>Waves</h5>
        {this.renderWavePanels()}
      </div>
    )
  }
}

ReleaseCreate.propTypes = {
  envName: PropTypes.string,
  env: PropTypes.object,
  releaseEnv: PropTypes.object,
  deployments: PropTypes.object,
  replicaSets: PropTypes.object,
  jobRuns: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  getReleaseSpec: PropTypes.func.isRequired,
  createRelease: PropTypes.func.isRequired,
}

export default connect(makeMapStateToProps, dispatchProps)(ReleaseCreate)
