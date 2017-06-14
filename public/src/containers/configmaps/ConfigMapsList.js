import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'

function mapStateToProps (state, ownProps) {
  const env = _.findWhere(state.envs.toJS().envs, {name: ownProps.params.env})
  const configmaps = state.configmaps.lookUpObjects(ownProps.params.env)
  _.each(env.configmaps, (key) => {
    if (_.isUndefined(configmaps[key])) {
      configmaps[key] = null
    }
  })
  return {
    configmaps
  }
}

@connect(mapStateToProps)
export default class ConfigMapsList extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    configmaps: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateNav('configmaps'))
  }

  render () {
    const { params, configmaps } = this.props
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

    const rows = _.map(configmaps, function (configmap, name) {
      return {
        component: (
          <Row key={name}
            env={params.env}
            name={name}
            configmap={configmap}
          />
        ),
        key: name
      }
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

@connect()
class Row extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    env: PropTypes.string,
    name: PropTypes.string,
    configmap: PropTypes.object
  }

  get nameLink () {
    const { env, name } = this.props
    return (
      <Link to={`/${env}/configmaps/${name}`}>{name}</Link>
    )
  }

  render () {
    const { configmap } = this.props
    return (
      <tr>
        <td data-column='name'>{this.nameLink}</td>
        <td data-column='key-count'>{configmap && configmap.keyCount || '-'}</td>
        <td data-column='created_at'>{configmap && configmap.createdAt || '-'}</td>
      </tr>
    )
  }

}
