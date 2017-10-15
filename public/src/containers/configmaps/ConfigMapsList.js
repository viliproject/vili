import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { makeLookUpObjects } from '../../selectors'
import ConfigMapRow from './ConfigMapRow'

function makeMapStateToProps () {
  const lookUpObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const { env: envName } = ownProps.params
    const env = state.envs.getIn(['envs', envName])
    const configmaps = lookUpObjects(state.configmaps, env.name)
    return {
      env,
      configmaps
    }
  }
}

const dispatchProps = {
  activateNav
}

export class ConfigMapsList extends React.Component {
  componentDidMount () {
    this.props.activateNav('configmaps')
  }

  render () {
    const { params, env, configmaps } = this.props
    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li className='active'>Config Maps</li>
        </ol>
      </div>
    )

    const columns = [
      {title: 'Name', key: 'name'},
      {title: 'Key Count', key: 'key-count'},
      {title: 'Created', key: 'created'}
    ]

    const rows = []
    env.configmaps.forEach((configmapName) => {
      const configmap = configmaps.find((d) => d.getIn(['metadata', 'name']) === configmapName)
      rows.push({
        component: (
          <ConfigMapRow
            key={configmapName}
            env={params.env}
            name={configmapName}
            configmap={configmap}
          />
        ),
        key: configmapName
      })
    })

    return (
      <div>
        {header}
        <Table columns={columns} rows={rows} />
      </div>
    )
  }

}

ConfigMapsList.propTypes = {
  activateNav: PropTypes.func,
  params: PropTypes.object,
  location: PropTypes.object,
  env: PropTypes.object,
  configmaps: PropTypes.object
}

export default connect(makeMapStateToProps, dispatchProps)(ConfigMapsList)
