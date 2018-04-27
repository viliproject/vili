import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"

import { activateNav } from "../../actions/app"

function mapStateToProps(state, ownProps) {
  const { envName } = ownProps
  const env = state.envs.getIn(["envs", envName])
  return {
    env,
  }
}

const dispatchProps = {
  activateNav,
}

export class EnvironmentHome extends React.Component {
  componentDidMount() {
    this.props.activateNav("")
  }

  render() {
    const { envName, env } = this.props
    if (!env) {
      return null
    }
    const items = []

    items.push(
      <li key="releases">
        <Link to={`/${envName}/releases`}>Releases</Link>
      </li>
    )

    if (env.deployments && !env.deployments.isEmpty()) {
      items.push(
        <li key="deployments">
          <Link to={`/${envName}/deployments`}>Deployments</Link>
        </li>
      )
    }
    if (env.jobs && !env.jobs.isEmpty()) {
      items.push(
        <li key="jobs">
          <Link to={`/${envName}/jobs`}>Jobs</Link>
        </li>
      )
    }
    if (env.functions && !env.functions.isEmpty()) {
      items.push(
        <li key="functions">
          <Link to={`/${envName}/functions`}>Functions</Link>
        </li>
      )
    }
    if (env.configmaps && !env.configmaps.isEmpty()) {
      items.push(
        <li key="configmaps">
          <Link to={`/${envName}/configmaps`}>Config Maps</Link>
        </li>
      )
    }
    items.push(
      <li key="nodes">
        <Link to={`/${envName}/nodes`}>Nodes</Link>
      </li>
    )
    items.push(
      <li key="pods">
        <Link to={`/${envName}/pods`}>Pods</Link>
      </li>
    )
    return (
      <div>
        <div key="header" className="view-header">
          <ol className="breadcrumb">
            <li className="active">{envName}</li>
          </ol>
        </div>
        <ul key="list" className="nav nav-pills nav-stacked">
          {items}
        </ul>
      </div>
    )
  }
}

EnvironmentHome.propTypes = {
  envName: PropTypes.string,
  env: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(EnvironmentHome)
