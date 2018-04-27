import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import FunctionContainer from "../../containers/functions/FunctionContainer"
import Function from "../../handlers/Function"
import FunctionVersions from "../../handlers/FunctionVersions"
import FunctionSpec from "../../handlers/FunctionSpec"
import NotFoundPage from "../../components/NotFoundPage"

export class FunctionContainerHandler extends React.Component {
  render() {
    const prefix = this.props.match.path
    const { env: envName, function: functionName } = this.props.match.params
    return (
      <FunctionContainer envName={envName} functionName={functionName}>
        <Switch>
          <Route exact path={`${prefix}`} component={Function} />
          <Route
            exact
            path={`${prefix}/versions`}
            component={FunctionVersions}
          />
          <Route exact path={`${prefix}/spec`} component={FunctionSpec} />
          <Route component={NotFoundPage} />
        </Switch>
      </FunctionContainer>
    )
  }
}

FunctionContainerHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default FunctionContainerHandler
