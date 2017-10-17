import PropTypes from 'prop-types'
import React from 'react'
import { Link } from 'react-router'

export class ConfigMapRow extends React.Component {
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

ConfigMapRow.propTypes = {
  env: PropTypes.string,
  name: PropTypes.string,
  configmap: PropTypes.object
}

export default ConfigMapRow
