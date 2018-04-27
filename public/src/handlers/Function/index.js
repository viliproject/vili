import React from "react"
import PropTypes from "prop-types"

import Function from "../../containers/functions/Function"

export class FunctionHandler extends React.Component {
  render() {
    const { env: envName, function: functionName } = this.props.match.params
    return <Function envName={envName} functionName={functionName} />
  }
}

FunctionHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default FunctionHandler
