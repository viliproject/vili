import React from "react"
import PropTypes from "prop-types"

import ConfigMapsList from "../../containers/configmaps/ConfigMapsList"

export class ConfigMapsListHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <ConfigMapsList envName={envName} />
  }
}

ConfigMapsListHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default ConfigMapsListHandler
