import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Button, Label } from 'react-bootstrap'

import { deployTag } from '../../actions/deployments'

const dispatchProps = {
  deployTag
}

export class DeploymentRow extends React.Component {
  static propTypes = {
    deployTag: PropTypes.func.isRequired,
    env: PropTypes.string,
    deployment: PropTypes.string,
    isActive: PropTypes.bool,
    tag: PropTypes.string,
    branch: PropTypes.string,
    revision: PropTypes.string,
    buildTime: PropTypes.string,
    replicaSets: PropTypes.object
  }

  deployTag = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { deployTag, env, deployment, tag, branch } = this.props
    deployTag(env, deployment, tag, branch)
  }

  render () {
    const { isActive, tag, branch, revision, buildTime, replicaSets } = this.props
    var className = ''
    const deployedAt = []
    replicaSets.forEach((replicaSet) => {
      var bsStyle = 'default'
      if (isActive) {
        className = 'success'
        bsStyle = 'success'
      } else if (replicaSet.getIn(['status', 'replicas'], 0) > 0) {
        className = 'warning'
        bsStyle = 'warning'
      }
      deployedAt.push(
        <div key={replicaSet.getIn(['metadata', 'name'])}>
          <Label bsStyle={bsStyle}>{replicaSet.revision} - {replicaSet.deployedAt}</Label>
        </div>
      )
    })
    return (
      <tr className={className}>
        <td>{tag}</td>
        <td>{branch}</td>
        <td>{revision || 'unknown'}</td>
        <td>{buildTime}</td>
        <td style={{textAlign: 'right'}}>{deployedAt}</td>
        <td style={{textAlign: 'right'}}>
          <Button onClick={this.deployTag} bsStyle='primary' bsSize='xs'>Deploy</Button>
        </td>
      </tr>
    )
  }

}

export default connect(null, dispatchProps)(DeploymentRow)
