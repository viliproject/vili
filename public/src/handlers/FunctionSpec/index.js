import React from "react"
import PropTypes from "prop-types"

import FunctionSpec from "../../containers/functions/FunctionSpec"

export class FunctionSpecHandler extends React.Component {
  render() {
    const { env: envName, function: functionName } = this.props.match.params
    return <FunctionSpec envName={envName} functionName={functionName} />
  }
}

FunctionSpecHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default FunctionSpecHandler
