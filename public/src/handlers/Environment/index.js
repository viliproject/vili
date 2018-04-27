import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Route, Switch } from "react-router"

import EnvironmentHome from "../../handlers/EnvironmentHome"
import Releases from "../../handlers/Releases"
import Deployments from "../../handlers/Deployments"
import Functions from "../../handlers/Functions"
import Jobs from "../../handlers/Jobs"
import ConfigMaps from "../../handlers/ConfigMaps"
import Pods from "../../handlers/Pods"
import Nodes from "../../handlers/Nodes"
import NotFoundPage from "../../components/NotFoundPage"

import { setEnv } from "../../actions/app"
import { subReleases } from "../../actions/releases"
import { subDeployments } from "../../actions/deployments"
import { subReplicaSets } from "../../actions/replicaSets"
import { subJobRuns } from "../../actions/jobRuns"
import { subFunctions } from "../../actions/functions"
import { subConfigMaps } from "../../actions/configmaps"
import { subPods } from "../../actions/pods"
import { subNodes } from "../../actions/nodes"

const dispatchProps = {
  setEnv,
  subReleases,
  subDeployments,
  subReplicaSets,
  subJobRuns,
  subFunctions,
  subConfigMaps,
  subPods,
  subNodes,
}

export class Environment extends React.Component {
  componentDidMount() {
    this.subData()
  }

  componentDidUpdate(prevProps) {
    if (this.props.match.params.env !== prevProps.match.params.env) {
      this.subData()
    }
  }

  componentWillUnmount() {
    const { setEnv } = this.props
    setEnv(null)
  }

  subData = () => {
    const {
      match: { params: { env } },
      setEnv,
      subReleases,
      subDeployments,
      subReplicaSets,
      subJobRuns,
      subFunctions,
      subConfigMaps,
      subPods,
      subNodes,
    } = this.props
    setEnv(env)
    subReleases(env)
    subDeployments(env)
    subReplicaSets(env)
    subJobRuns(env)
    subFunctions(env)
    subConfigMaps(env)
    subPods(env)
    subNodes(env)
  }

  render() {
    const prefix = this.props.match.path

    return (
      <Switch>
        <Route exact path={`${prefix}`} component={EnvironmentHome} />
        <Route path={`${prefix}/releases`} component={Releases} />
        <Route path={`${prefix}/deployments`} component={Deployments} />
        <Route path={`${prefix}/jobs`} component={Jobs} />
        <Route path={`${prefix}/functions`} component={Functions} />
        <Route path={`${prefix}/configmaps`} component={ConfigMaps} />
        <Route path={`${prefix}/pods`} component={Pods} />
        <Route path={`${prefix}/nodes`} component={Nodes} />
        <Route component={NotFoundPage} />
      </Switch>
    )
  }
}

Environment.propTypes = {
  setEnv: PropTypes.func.isRequired,
  subReleases: PropTypes.func.isRequired,
  subDeployments: PropTypes.func.isRequired,
  subReplicaSets: PropTypes.func.isRequired,
  subJobRuns: PropTypes.func.isRequired,
  subFunctions: PropTypes.func.isRequired,
  subConfigMaps: PropTypes.func.isRequired,
  subPods: PropTypes.func.isRequired,
  subNodes: PropTypes.func.isRequired,
  match: PropTypes.object.isRequired,
}

export default connect(null, dispatchProps)(Environment)
