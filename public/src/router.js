import React from 'react'
import { Router, Route, IndexRoute } from 'react-router'

import App from './containers/App'
import Home from './containers/Home'
import Environment from './containers/Environment'
import EnvironmentHome from './containers/EnvironmentHome'
// releases
import ReleasesList from './containers/releases/ReleasesList'
import ReleaseCreate from './containers/releases/ReleaseCreate'
import Release from './containers/releases/Release'
import ReleaseRollout from './containers/releases/ReleaseRollout'
// deployments
import DeploymentsList from './containers/deployments/DeploymentsList'
import DeploymentBase from './containers/deployments/DeploymentBase'
import Deployment from './containers/deployments/Deployment'
import DeploymentRollouts from './containers/deployments/DeploymentRollouts'
import DeploymentSpec from './containers/deployments/DeploymentSpec'
import DeploymentService from './containers/deployments/DeploymentService'
// jobs
import JobsList from './containers/jobs/JobsList'
import JobBase from './containers/jobs/JobBase'
import Job from './containers/jobs/Job'
import JobSpec from './containers/jobs/JobSpec'
import JobRuns from './containers/jobs/JobRuns'
import JobRun from './containers/jobs/JobRun'
// configMaps
import ConfigMapsList from './containers/configmaps/ConfigMapsList'
import ConfigMap from './containers/configmaps/ConfigMap'
// pods
import PodsList from './containers/pods/PodsList'
import Pod from './containers/pods/Pod'
// nodes
import NodesList from './containers/nodes/NodesList'
import Node from './containers/nodes/Node'

class NotFoundPage extends React.Component {
  render () {
    return (
      <div>NOT FOUND</div>
    )
  }
}

export default function (history) {
  return (
    <Router history={history}>
      <Route path='/' component={App}>
        <IndexRoute component={Home} />
        <Route path='/:env' component={Environment}>
          <IndexRoute component={EnvironmentHome} />
          <Route path='releases' component={ReleasesList} />
          <Route path='releases/create' component={ReleaseCreate} />
          <Route path='releases/:release' component={Release} />
          <Route path='releases/:release/rollouts/:rollout' component={ReleaseRollout} />
          <Route path='deployments' component={DeploymentsList} />
          <Route path='deployments/:deployment' component={DeploymentBase}>
            <IndexRoute component={Deployment} />
            <Route path='rollouts' component={DeploymentRollouts} />
            <Route path='spec' component={DeploymentSpec} />
            <Route path='service' component={DeploymentService} />
          </Route>
          <Route path='jobs' component={JobsList} />
          <Route path='jobs/:job' component={JobBase}>
            <IndexRoute component={Job} />
            <Route path='runs' component={JobRuns} />
            <Route path='runs/:run' component={JobRun} />
            <Route path='spec' component={JobSpec} />
          </Route>
          <Route path='configmaps' component={ConfigMapsList} />
          <Route path='configmaps/:configmap' component={ConfigMap} />
          <Route path='pods' component={PodsList} />
          <Route path='pods/:pod' component={Pod} />
          <Route path='nodes' component={NodesList} />
          <Route path='nodes/:node' component={Node} />
        </Route>
        <Route path='*' component={NotFoundPage} />
      </Route>
    </Router>
  )
}
