import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { Button } from 'react-bootstrap'

import { deletePod } from '../../actions/pods'

const dispatchProps = {
  deletePod
}

class PodRow extends React.Component {
  static propTypes = {
    deletePod: PropTypes.func.isRequired,
    env: PropTypes.string,
    pod: PropTypes.object
  }

  get nameLink () {
    const { env, pod } = this.props
    return (
      <Link to={`/${env}/pods/${pod.getIn(['metadata', 'name'])}`}>{pod.getIn(['metadata', 'name'])}</Link>
    )
  }

  get deploymentJobLink () {
    const { env, pod } = this.props
    if (pod.getLabel('app')) {
      return (
        <Link to={`/${env}/deployments/${pod.getLabel('app')}`}>{pod.getLabel('app')}</Link>
      )
    } else if (pod.getLabel('job')) {
      return (
        <Link to={`/${env}/jobs/${pod.getLabel('job')}`}>{pod.getLabel('job')}</Link>
      )
    }
  }

  get nodeLink () {
    const { env, pod } = this.props
    return (
      <Link to={`/${env}/nodes/${pod.getIn(['spec', 'nodeName'])}`}>{pod.getIn(['spec', 'nodeName'])}</Link>
    )
  }

  deletePod = () => {
    const { env, pod, deletePod } = this.props
    deletePod(env, pod.getIn(['metadata', 'name']))
  }

  render () {
    const { pod } = this.props
    return (
      <tr>
        <td>{this.nameLink}</td>
        <td>{this.deploymentJobLink}</td>
        <td>{this.nodeLink}</td>
        <td>{pod.getIn(['status', 'phase'])}</td>
        <td>{pod.isReady ? String.fromCharCode('10003') : ''}</td>
        <td>{pod.createdAt}</td>
        <td style={{textAlign: 'right'}}><Button onClick={this.deletePod} bsStyle='danger' bsSize='xs'>Delete</Button></td>
      </tr>
    )
  }
}

export default connect(null, dispatchProps)(PodRow)
