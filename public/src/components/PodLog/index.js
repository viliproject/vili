import React, { PropTypes } from 'react'

export default class Log extends React.Component {
  static propTypes = {
    log: PropTypes.string
  }

  render () {
    return (
      <div>
        <h4>Log</h4>
        <pre>
          {this.props.log}
        </pre>
      </div>
    )
  }
}
