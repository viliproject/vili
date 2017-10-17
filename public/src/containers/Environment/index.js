import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'

import { subReleases } from '../../actions/releases'
import { subDeployments } from '../../actions/deployments'
import { subReplicaSets } from '../../actions/replicaSets'
import { subJobRuns } from '../../actions/jobRuns'
import { subConfigMaps } from '../../actions/configmaps'
import { subPods } from '../../actions/pods'
import { subNodes } from '../../actions/nodes'

const dispatchProps = {
  subReleases,
  subDeployments,
  subReplicaSets,
  subJobRuns,
  subConfigMaps,
  subPods,
  subNodes
}

@connect(null, dispatchProps)
export default class Environment extends React.Component {
  static propTypes = {
    subReleases: PropTypes.func.isRequired,
    subDeployments: PropTypes.func.isRequired,
    subReplicaSets: PropTypes.func.isRequired,
    subJobRuns: PropTypes.func.isRequired,
    subConfigMaps: PropTypes.func.isRequired,
    subPods: PropTypes.func.isRequired,
    subNodes: PropTypes.func.isRequired,
    params: PropTypes.object,
    children: PropTypes.node
  }

  componentDidMount () {
    this.subData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params.env !== prevProps.params.env) {
      this.subData()
    }
  }

  subData = () => {
    const { params } = this.props
    this.props.subReleases(params.env)
    this.props.subDeployments(params.env)
    this.props.subReplicaSets(params.env)
    this.props.subJobRuns(params.env)
    this.props.subConfigMaps(params.env)
    this.props.subPods(params.env)
    this.props.subNodes(params.env)
  }

  render () {
    return (
      <div>{this.props.children}</div>
    )
  }
}
