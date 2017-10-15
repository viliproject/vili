import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { makeLookUpObjects } from '../../selectors'
import PodRow from './PodRow'

function makeMapStateToProps () {
  const lookUpObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const pods = lookUpObjects(state.pods, ownProps.params.env)
    return {
      pods
    }
  }
}

const dispatchProps = {
  activateNav
}

export class PodsList extends React.Component {
  static propTypes = {
    activateNav: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    pods: PropTypes.object
  }

  componentDidMount () {
    this.props.activateNav('pods')
  }

  render () {
    const { params, pods } = this.props
    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li className='active'>Pods</li>
        </ol>
      </div>
    )

    const columns = [
      {title: 'Name', key: 'name'},
      {title: 'Deployment/Job', key: 'deployment-job'},
      {title: 'Node', key: 'node'},
      {title: 'Phase', key: 'phase'},
      {title: 'Ready', key: 'ready'},
      {title: 'Created', key: 'created'},
      {title: 'Actions', key: 'actions'}
    ]

    const rows = []
    pods.map((pod, key) => {
      rows.push({
        component: (
          <PodRow
            key={key}
            env={params.env}
            pod={pod}
          />
        ),
        key: key
      })
    })
    const sortedRows = _.sortBy(rows, 'key')

    return (
      <div>
        {header}
        <Table columns={columns} rows={sortedRows} />
      </div>
    )
  }

}

export default connect(makeMapStateToProps, dispatchProps)(PodsList)
