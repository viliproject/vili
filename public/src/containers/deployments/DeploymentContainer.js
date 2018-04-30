import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import _ from "underscore"

import { activateNav } from "../../actions/app"

const tabs = {
  home: "Home",
  rollouts: "Rollouts",
  spec: "Spec",
  service: "Service",
}

function mapStateToProps(state) {
  return {
    app: state.app,
  }
}

const dispatchProps = {
  activateNav,
}

export class DeploymentContainer extends React.Component {
  componentDidMount() {
    const { deploymentName, activateNav } = this.props
    activateNav("deployments", deploymentName)
  }

  componentDidUpdate(prevProps) {
    const { deploymentName, activateNav } = this.props
    if (deploymentName !== prevProps.deploymentName) {
      activateNav("deployments", deploymentName)
    }
  }

  render() {
    const { envName, deploymentName, app, children } = this.props
    const tabElements = _.map(tabs, (name, key) => {
      let className = ""
      if (app.get("deploymentTab") === key) {
        className = "active"
      }
      let link = `/${envName}/deployments/${deploymentName}`
      if (key !== "home") {
        link += `/${key}`
      }
      return (
        <li key={key} role="presentation" className={className}>
          <Link to={link}>{name}</Link>
        </li>
      )
    })
    return (
      <div>
        <div key="view-header" className="view-header">
          <ol className="breadcrumb">
            <li>
              <Link to={`/${envName}`}>{envName}</Link>
            </li>
            <li>
              <Link to={`/${envName}/deployments`}>Deployments</Link>
            </li>
            <li className="active">{deploymentName}</li>
          </ol>
          <ul className="nav nav-pills pull-right">{tabElements}</ul>
        </div>
        {children}
      </div>
    )
  }
}

DeploymentContainer.propTypes = {
  envName: PropTypes.string,
  deploymentName: PropTypes.string,
  app: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  children: PropTypes.node,
}

export default connect(mapStateToProps, dispatchProps)(DeploymentContainer)
