import React from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { Button, ButtonToolbar } from "react-bootstrap"
import { Link } from "react-router-dom"

import Table from "../../components/Table"
import { activateNav } from "../../actions/app"
import { deployRelease } from "../../actions/releases"

import ReleaseWavePanel from "./ReleaseWavePanel"

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
  deployRelease,
}

export class Release extends React.Component {
  componentDidMount() {
    this.props.activateNav("releases")
  }

  deployRelease = event => {
    event.target.setAttribute("disabled", "disabled")
    const { deployRelease, env, release } = this.props
    deployRelease(env.name, release.name)
  }

  renderActions() {
    const buttons = []
    buttons.push(
      <Button
        key="deploy"
        onClick={this.deployRelease}
        bsStyle="primary"
        bsSize="small"
      >
        Deploy
      </Button>
    )
    return (
      <ButtonToolbar key="toolbar" className="pull-right">
        {buttons}
      </ButtonToolbar>
    )
  }

  renderMetadata() {
    const { env, release } = this.props
    if (!env || !release) {
      return null
    }

    const metadata = []
    if (env.approvedFromEnv) {
      metadata.push(<h5 key="approved-env-title">Approved From</h5>)
      metadata.push(<div key="approved-env-value">{env.approvedFromEnv}</div>)
    }

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

    metadata.push(<h5 key="approvedBy-title">Approved By</h5>)
    metadata.push(<div key="approvedBy-value">{release.createdBy}</div>)
    metadata.push(<h5 key="createdAt-title">Created At</h5>)
    metadata.push(<div key="createdAt-value">{release.createdAtHumanize}</div>)

    const rollouts = release.envRollouts(env.name)
    if (rollouts.size > 0) {
      metadata.push(<h5 key="rollouts-title">Rollouts</h5>)

      const columns = [
        { title: "ID", key: "id", style: { width: "50px" } },
        { title: "Rollout At", key: "rolloutAtHumanize" },
        {
          title: "Rollout By",
          key: "rolloutBy",
          style: { width: "200px", textAlign: "right" },
        },
        {
          title: "Status",
          key: "status",
          style: { width: "200px", textAlign: "right" },
        },
      ]
      const rows = []
      rollouts.forEach(rollout => {
        rows.push({
          id: (
            <Link
              to={`/${env.name}/releases/${release.name}/rollouts/${
                rollout.id
              }`}
            >
              {rollout.id}
            </Link>
          ),
          rolloutAtHumanize: rollout.rolloutAtHumanize,
          rolloutBy: rollout.rolloutBy,
          status: rollout.status,
        })
      })
      metadata.push(
        <Table key="rollouts-value" columns={columns} rows={rows} />
      )
    }
    return metadata
  }

  renderWavePanels() {
    const { env, release } = this.props
    if (release) {
      const panels = []
      release.waves.forEach((wave, ix) => {
        panels.push(
          <ReleaseWavePanel
            key={ix}
            ix={ix}
            env={env.name}
            wave={wave.toJS()}
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
    return null
  }

  render() {
    const { envName, releaseName } = this.props
    const header = (
      <div key="header" className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li>
            <Link to={`/${envName}/releases`}>Releases</Link>
          </li>
          <li className="active">{releaseName}</li>
        </ol>
        {this.renderActions()}
      </div>
    )

    return (
      <div>
        {header}
        {this.renderMetadata()}
        {this.renderWavePanels()}
      </div>
    )
  }
}

Release.propTypes = {
  envName: PropTypes.string,
  releaseName: PropTypes.string,
  env: PropTypes.object,
  release: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  deployRelease: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(Release)
