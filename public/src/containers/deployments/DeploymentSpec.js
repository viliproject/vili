import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"

import Loading from "../../components/Loading"
import { activateDeploymentTab } from "../../actions/app"
import { getDeploymentSpec } from "../../actions/deployments"

function mapStateToProps(state, ownProps) {
  const { envName, deploymentName } = ownProps
  const deployment = state.deployments.lookUpData(envName, deploymentName)
  return {
    deployment,
  }
}

const dispatchProps = {
  activateDeploymentTab,
  getDeploymentSpec,
}

export class DeploymentSpec extends React.Component {
  componentDidMount() {
    this.props.activateDeploymentTab("spec")
    this.fetchData()
  }

  componentDidUpdate(prevProps) {
    if (
      this.props.envName !== prevProps.envName ||
      this.props.deploymentName !== prevProps.deploymentName
    ) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { envName, deploymentName, getDeploymentSpec } = this.props
    getDeploymentSpec(envName, deploymentName)
  }

  render() {
    const { deployment } = this.props
    if (!deployment || !deployment.get("spec")) {
      return <Loading />
    }
    return (
      <div className="col-md-8">
        <div id="source-yaml">
          <pre>
            <code>{deployment.get("spec")}</code>
          </pre>
        </div>
      </div>
    )
  }
}

DeploymentSpec.propTypes = {
  envName: PropTypes.string,
  deploymentName: PropTypes.string,
  deployment: PropTypes.object,
  activateDeploymentTab: PropTypes.func,
  getDeploymentSpec: PropTypes.func,
}

export default connect(mapStateToProps, dispatchProps)(DeploymentSpec)
