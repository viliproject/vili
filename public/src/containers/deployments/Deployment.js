import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import _ from "underscore"

import displayTime from "../../lib/displayTime"
import Table from "../../components/Table"
import Loading from "../../components/Loading"
import DeploymentRow from "../../components/deployments/DeploymentRow"
import { activateDeploymentTab } from "../../actions/app"
import { getDeploymentRepository } from "../../actions/deployments"
import { makeLookUpObjectsByLabel } from "../../selectors"

function makeMapStateToProps() {
  const lookUpObjectsByLabel = makeLookUpObjectsByLabel()
  return (state, ownProps) => {
    const { envName, deploymentName } = ownProps
    const deployment = state.deployments.lookUpData(envName, deploymentName)
    const replicaSets = lookUpObjectsByLabel(
      state.replicaSets,
      envName,
      "app",
      deploymentName
    )
    return {
      deployment,
      replicaSets,
    }
  }
}

const dispatchProps = {
  activateDeploymentTab,
  getDeploymentRepository,
}

export class Deployment extends React.Component {
  componentDidMount() {
    this.props.activateDeploymentTab("home")
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
    const { envName, deploymentName, getDeploymentRepository } = this.props
    getDeploymentRepository(envName, deploymentName)
  }

  render() {
    const { envName, deploymentName, deployment, replicaSets } = this.props
    if (!deployment || !deployment.get("repository")) {
      return <Loading />
    }

    const columns = [
      { title: "Tag", key: "tag", style: { width: "180px" } },
      { title: "Branch", key: "branch", style: { width: "120px" } },
      { title: "Revision", key: "revision", style: { width: "90px" } },
      { title: "Build Time", key: "buildTime", style: { width: "180px" } },
      { title: "Deployed", key: "deployedAt", style: { textAlign: "right" } },
      { title: "Actions", key: "actions", style: { textAlign: "right" } },
    ]

    let rows = []
    deployment.get("repository").forEach(image => {
      const imageReplicaSets = replicaSets.filter(
        rs => rs.imageTag === image.get("tag")
      )
      const buildTime = new Date(image.get("lastModified"))
      rows.push({
        component: (
          <DeploymentRow
            key={image.get("tag")}
            env={envName}
            deployment={deploymentName}
            isActive={
              deployment.get("object", {}).imageTag === image.get("tag")
            }
            tag={image.get("tag")}
            branch={image.get("branch")}
            revision={image.get("revision")}
            buildTime={displayTime(buildTime)}
            replicaSets={imageReplicaSets}
          />
        ),
        time: buildTime.getTime(),
      })
    })

    rows = _.sortBy(rows, function(row) {
      return -row.time
    })

    return <Table columns={columns} rows={rows} />
  }
}

Deployment.propTypes = {
  envName: PropTypes.string,
  deploymentName: PropTypes.string,
  activateDeploymentTab: PropTypes.func,
  getDeploymentRepository: PropTypes.func,
  deployment: PropTypes.object,
  replicaSets: PropTypes.object,
}

export default connect(makeMapStateToProps, dispatchProps)(Deployment)
