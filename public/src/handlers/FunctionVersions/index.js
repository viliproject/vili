import React from "react"
import PropTypes from "prop-types"

import FunctionVersions from "../../containers/functions/FunctionVersions"

export class FunctionVersionsHandler extends React.Component {
  render() {
    const { env: envName, function: functionName } = this.props.match.params
    return <FunctionVersions envName={envName} functionName={functionName} />
  }
}

FunctionVersionsHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default FunctionVersionsHandler
