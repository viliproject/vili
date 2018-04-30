import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"

import Loading from "../../components/Loading"
import { activateDeploymentTab } from "../../actions/app"
import { getDeploymentService } from "../../actions/deployments"
import { createService } from "../../actions/services"

function mapStateToProps(state, ownProps) {
  const { envName, deploymentName } = ownProps
  const deployment = state.deployments.lookUpData(envName, deploymentName)
  return {
    deployment,
  }
}

const dispatchProps = {
  activateDeploymentTab,
  getDeploymentService,
  createService,
}

export class DeploymentService extends React.Component {
  componentDidMount() {
    this.props.activateDeploymentTab("service")
    this.subData()
  }

  componentDidUpdate(prevProps) {
    if (
      this.props.envName !== prevProps.envName ||
      this.props.deploymentName !== prevProps.deploymentName
    ) {
      this.subData()
    }
  }

  subData = () => {
    const { envName, deploymentName } = this.props
    this.props.getDeploymentService(envName, deploymentName)
  }

  clickCreateService = event => {
    const { envName, deploymentName } = this.props
    event.currentTarget.setAttribute("disabled", "disabled")
    this.props.createService(envName, deploymentName)
  }

  render() {
    const { deployment } = this.props
    if (!deployment) {
      return <Loading />
    }
    if (!deployment.get("service")) {
      return (
        <div id="service">
          <div className="alert alert-warning" role="alert">
            No Service Defined
          </div>
          <div>
            <button
              className="btn btn-success"
              onClick={this.clickCreateService}
            >
              Create Service
            </button>
          </div>
        </div>
      )
    }
    return (
      <div id="service">
        IP: {deployment.getIn(["service", "spec", "clusterIP"])}
      </div>
    )
  }
}

DeploymentService.propTypes = {
  envName: PropTypes.string,
  deploymentName: PropTypes.string,
  deployment: PropTypes.object,
  activateDeploymentTab: PropTypes.func.isRequired,
  getDeploymentService: PropTypes.func.isRequired,
  createService: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(DeploymentService)
