import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { ButtonToolbar, Button } from "react-bootstrap"
import _ from "underscore"

import Table from "../../components/Table"
import Loading from "../../components/Loading"
import { activateDeploymentTab } from "../../actions/app"
import {
  resumeDeployment,
  pauseDeployment,
  scaleDeployment,
} from "../../actions/deployments"
import { makeLookUpObjectsByLabel } from "../../selectors"

import DeploymentRolloutPanel from "./DeploymentRolloutPanel"
import DeploymentRolloutHistoryRow from "./DeploymentRolloutHistoryRow"

function getReplicaSetPods(replicaSet, pods) {
  const generateName = replicaSet.getIn(["metadata", "name"]) + "-"
  return pods.filter(
    pod => pod.getIn(["metadata", "generateName"]) === generateName
  )
}

function makeMapStateToProps() {
  const lookUpReplicaSetsByLabel = makeLookUpObjectsByLabel()
  const lookUpPodsByLabel = makeLookUpObjectsByLabel()
  return (state, ownProps) => {
    const { envName, deploymentName } = ownProps
    const deployment = state.deployments.lookUpData(envName, deploymentName)
    const replicaSets = lookUpReplicaSetsByLabel(
      state.replicaSets,
      envName,
      "app",
      deploymentName
    )
    const pods = lookUpPodsByLabel(state.pods, envName, "app", deploymentName)
    return {
      deployment,
      replicaSets,
      pods,
    }
  }
}

const dispatchProps = {
  activateDeploymentTab,
  resumeDeployment,
  pauseDeployment,
  scaleDeployment,
}

export class DeploymentRollouts extends React.Component {
  componentDidMount() {
    this.props.activateDeploymentTab("rollouts")
  }

  resumeDeployment = event => {
    event.target.setAttribute("disabled", "disabled")
    const { envName, deploymentName, resumeDeployment } = this.props
    resumeDeployment(envName, deploymentName)
  }

  pauseDeployment = event => {
    event.target.setAttribute("disabled", "disabled")
    const { envName, deploymentName, pauseDeployment } = this.props
    pauseDeployment(envName, deploymentName)
  }

  scaleDeployment = () => {
    var replicas = prompt("Enter the number of replicas to scale to")
    if (!replicas) {
      return
    }
    replicas = parseInt(replicas)
    if (_.isNaN(replicas)) {
      return
    }
    const { envName, deploymentName, scaleDeployment } = this.props
    scaleDeployment(envName, deploymentName, replicas)
  }

  renderHeader() {
    const buttons = []
    const { deployment } = this.props
    if (deployment.getIn(["object", "spec", "paused"])) {
      buttons.push(
        <Button
          key="resume"
          bsStyle="success"
          bsSize="small"
          onClick={this.resumeDeployment}
        >
          Resume
        </Button>
      )
    } else {
      buttons.push(
        <Button
          key="pause"
          bsStyle="warning"
          bsSize="small"
          onClick={this.pauseDeployment}
        >
          Pause
        </Button>
      )
    }
    buttons.push(
      <Button
        key="scale"
        bsStyle="info"
        bsSize="small"
        onClick={this.scaleDeployment}
      >
        Scale
      </Button>
    )
    return (
      <div className="clearfix" style={{ marginBottom: "10px" }}>
        <ButtonToolbar className="pull-right">{buttons}</ButtonToolbar>
      </div>
    )
  }

  renderPodPanels(replicaSets, activeReplicaSet) {
    const { envName, deployment, pods } = this.props
    const podPanels = []
    replicaSets.forEach(replicaSet => {
      const replicaSetPods = getReplicaSetPods(replicaSet, pods)
      if (replicaSetPods.isEmpty() && activeReplicaSet !== replicaSet) {
        return null
      }
      podPanels.push(
        <DeploymentRolloutPanel
          key={replicaSet.getIn(["metadata", "name"])}
          env={envName}
          deployment={deployment.get("object")}
          replicaSet={replicaSet}
          pods={replicaSetPods}
        />
      )
    })
    return podPanels
  }

  renderHistoryTable(replicaSets, activeReplicaSet) {
    const { envName, deploymentName, pods } = this.props
    const columns = [
      { title: "Revision", key: "revision", style: { width: "90px" } },
      { title: "Tag", key: "tag", style: { width: "180px" } },
      { title: "Branch", key: "branch" },
      { title: "Rollout Time", key: "time", style: { textAlign: "right" } },
      {
        title: "Deployed By",
        key: "deployedBy",
        style: { textAlign: "right" },
      },
      { title: "Actions", key: "actions", style: { textAlign: "right" } },
    ]

    const rows = []
    replicaSets.forEach(replicaSet => {
      const replicaSetPods = getReplicaSetPods(replicaSet, pods)
      if (!replicaSetPods.isEmpty() || activeReplicaSet === replicaSet) {
        return null
      }
      rows.push({
        component: (
          <DeploymentRolloutHistoryRow
            key={replicaSet.getIn(["metadata", "name"])}
            env={envName}
            deployment={deploymentName}
            replicaSet={replicaSet}
          />
        ),
      })
    })
    return (
      <div>
        <h4>Rollout History</h4>
        <Table columns={columns} rows={rows} />
      </div>
    )
  }

  render() {
    const { deployment, replicaSets } = this.props
    if (!deployment || !deployment.get("object")) {
      return <Loading />
    }
    const sortedReplicaSets = replicaSets.sort((a, b) => {
      if (
        a.revision > b.revision ||
        a.creationTimestamp < b.creationTimestamp
      ) {
        return 1
      }
      if (
        a.revision === b.revision &&
        a.creationTimestamp === b.creationTimestamp
      ) {
        return 0
      }
      return -1
    })
    const activeReplicaSet = sortedReplicaSets
      .filter(x => x.revision === deployment.get("object").revision)
      .first()

    return (
      <div>
        {this.renderHeader()}
        {this.renderPodPanels(sortedReplicaSets, activeReplicaSet)}
        {this.renderHistoryTable(sortedReplicaSets, activeReplicaSet)}
      </div>
    )
  }
}

DeploymentRollouts.propTypes = {
  envName: PropTypes.string.isRequired,
  deploymentName: PropTypes.string.isRequired,
  deployment: PropTypes.object,
  replicaSets: PropTypes.object,
  pods: PropTypes.object,
  activateDeploymentTab: PropTypes.func,
  resumeDeployment: PropTypes.func,
  pauseDeployment: PropTypes.func,
  scaleDeployment: PropTypes.func,
}

export default connect(makeMapStateToProps, dispatchProps)(DeploymentRollouts)
