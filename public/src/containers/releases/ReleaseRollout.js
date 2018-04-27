import React from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { Alert } from "react-bootstrap"
import { Link } from "react-router-dom"
import Immutable from "immutable"

import { activateNav } from "../../actions/app"

import ReleaseRolloutWavePanel from "./ReleaseRolloutWavePanel"

function mapStateToProps(state, ownProps) {
  const { envName, releaseName } = ownProps
  const env = state.envs.getIn(["envs", envName])
  const release = state.releases.lookUpObject(envName, releaseName)
  return {
    env,
    release,
  }
}

const dispatchProps = {
  activateNav,
}

export class ReleaseRollout extends React.Component {
  componentDidMount() {
    this.props.activateNav("releases")
  }

  renderMetadata(rollout) {
    const { release } = this.props
    const metadata = []

    switch (rollout.status) {
      case "deployed":
        metadata.push(
          <Alert key="alert" bsStyle="success">
            Release was rolled out <strong>{rollout.rolloutAtHumanize}</strong>{" "}
            by <strong>{rollout.rolloutBy}</strong>
          </Alert>
        )
        break
      case "deploying":
        metadata.push(
          <Alert key="alert" bsStyle="warning">
            Release is rolling out, started by{" "}
            <strong>{rollout.rolloutBy}</strong>
          </Alert>
        )
        break
      case "failed":
        metadata.push(
          <Alert key="alert" bsStyle="danger">
            Release rollout failed at{" "}
            <strong>{rollout.rolloutAtHumanize}</strong>, was started by{" "}
            <strong>{rollout.rolloutBy}</strong>
          </Alert>
        )
        break
    }
    metadata.push(<h5 key="createdAt-title">Created At</h5>)
    metadata.push(<div key="createdAt-value">{release.createdAtHumanize}</div>)
    if (release.link) {
      metadata.push(<h5 key="link-title">Link</h5>)
      metadata.push(
        <div key="link-value">
          <a href={release.link} target="_blank">
            {release.link}
          </a>
        </div>
      )
    }
    return metadata
  }

  renderWavePanels(rollout) {
    const { env, release } = this.props
    const panels = []
    release.waves.forEach((wave, ix) => {
      const rolloutWave = rollout.waves.get(ix, Immutable.Map())
      panels.push(
        <ReleaseRolloutWavePanel
          key={ix}
          ix={ix}
          env={env.name}
          wave={wave.toJS()}
          rolloutWave={rolloutWave.toJS()}
        />
      )
    })
    return (
      <div>
        <h5>Waves</h5>
        {panels}
      </div>
    )
  }

  render() {
    const { env, release, envName, releaseName, rolloutID } = this.props
    const header = [
      <div key="header" className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li>
            <Link to={`/${envName}/releases`}>Releases</Link>
          </li>
          <li>
            <Link to={`/${envName}/releases/${releaseName}`}>
              {releaseName}
            </Link>
          </li>
          <li className="active">{`Rollout ${rolloutID}`}</li>
        </ol>
      </div>,
    ]

    if (!env || !release) {
      return <div>{header}</div>
    }
    const rollouts = release.envRollouts(env.name)
    const rollout = rollouts.find(r => r.id === rolloutID)
    if (!rollout) {
      return <div>{header}</div>
    }

    return (
      <div>
        {header}
        {this.renderMetadata(rollout)}
        {this.renderWavePanels(rollout)}
      </div>
    )
  }
}

ReleaseRollout.propTypes = {
  envName: PropTypes.string,
  releaseName: PropTypes.string,
  rolloutID: PropTypes.number,
  env: PropTypes.object,
  release: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(ReleaseRollout)
