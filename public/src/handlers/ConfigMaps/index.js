import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import ConfigMapsList from "../../handlers/ConfigMapsList"
import ConfigMap from "../../handlers/ConfigMap"

export class ConfigMaps extends React.Component {
  render() {
    const prefix = this.props.match.path
    return (
      <Switch>
        <Route exact path={`${prefix}`} component={ConfigMapsList} />
        <Route exact path={`${prefix}/:configmap`} component={ConfigMap} />
      </Switch>
    )
  }
}

ConfigMaps.propTypes = {
  match: PropTypes.object.isRequired,
}

export default ConfigMaps
