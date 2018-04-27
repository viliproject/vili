import React from "react"

export default class Loading extends React.Component {
  render() {
    return (
      <div className="loading">
        <span className="glyphicon glyphicon-refresh glyphicon-refresh-animate" />
        <span>Loading</span>
      </div>
    )
  }
}
