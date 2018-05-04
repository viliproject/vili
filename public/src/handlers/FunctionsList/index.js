import React from "react"
import PropTypes from "prop-types"

import FunctionsList from "../../containers/functions/FunctionsList"

export class FunctionsListHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <FunctionsList envName={envName} />
  }
}

FunctionsListHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default FunctionsListHandler
