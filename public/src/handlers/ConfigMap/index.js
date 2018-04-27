import React from "react"
import PropTypes from "prop-types"

import ConfigMap from "../../containers/configmaps/ConfigMap"

export class ConfigMapHandler extends React.Component {
  render() {
    const { env: envName, configmap: configmapName } = this.props.match.params
    return <ConfigMap envName={envName} configmapName={configmapName} />
  }
}

ConfigMapHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default ConfigMapHandler
