import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import FunctionsList from "../../handlers/FunctionsList"
import FunctionContainer from "../../handlers/FunctionContainer"

export class Functions extends React.Component {
  render() {
    const prefix = this.props.match.path
    return (
      <Switch>
        <Route exact path={`${prefix}`} component={FunctionsList} />
        <Route path={`${prefix}/:function`} component={FunctionContainer} />
      </Switch>
    )
  }
}

Functions.propTypes = {
  match: PropTypes.object.isRequired,
}

export default Functions
