import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import _ from 'underscore'

import displayTime from '../../lib/displayTime'
import Table from '../../components/Table'
import Loading from '../../components/Loading'
import DeploymentRow from '../../components/deployments/DeploymentRow'
import { activateDeploymentTab } from '../../actions/app'
import { getDeploymentRepository } from '../../actions/deployments'

function mapStateToProps (state, ownProps) {
  const { env, deployment: deploymentName } = ownProps.params
  const deployment = state.deployments.lookUpData(env, deploymentName)
  const replicaSets = state.replicaSets.lookUpObjectsByFunc(env, (obj) => {
    return obj.hasLabel('app', deploymentName)
  })
  return {
    deployment,
    replicaSets
  }
}

@connect(mapStateToProps)
export default class Deployment extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    deployment: PropTypes.object,
    replicaSets: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateDeploymentTab('home'))
    this.fetchData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { params } = this.props
    this.props.dispatch(getDeploymentRepository(params.env, params.deployment))
  }

  render () {
    const { params, deployment, replicaSets } = this.props
    if (!deployment) {
      return (<Loading />)
    }

    const columns = [
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Branch', key: 'branch', style: {width: '120px'}},
      {title: 'Revision', key: 'revision', style: {width: '90px'}},
      {title: 'Build Time', key: 'buildTime', style: {width: '180px'}},
      {title: 'Deployed', key: 'deployedAt', style: {textAlign: 'right'}},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]
    var rows = _.map(deployment.repository, (image) => {
      const imageReplicaSets = Object.keys(replicaSets).filter(
        (key) => replicaSets[key].imageTag === image.tag
      ).reduce((obj, key) => {
        obj[key] = replicaSets[key]
        return obj
      }, {})

      const buildTime = new Date(image.lastModified)
      return {
        component: (
          <DeploymentRow key={image.tag}
            env={params.env}
            deployment={params.deployment}
            currentRevision={deployment.object && deployment.object.revision}
            tag={image.tag}
            branch={image.branch}
            revision={image.revision}
            buildTime={displayTime(buildTime)}
            replicaSets={imageReplicaSets}
          />),
        time: buildTime.getTime()
      }
    })

    rows = _.sortBy(rows, function (row) {
      return -row.time
    })

    return (<Table columns={columns} rows={rows} />)
  }

}
