import React from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"

export default class LinkMenuItem extends React.Component {
  static propTypes = {
    active: PropTypes.bool,
    subitem: PropTypes.bool,
    onRemove: PropTypes.func,
    to: PropTypes.string,
    children: PropTypes.node,
  }

  constructor(props) {
    super(props)

    this.onRemove = this.onRemove.bind(this)
  }

  onRemove(event) {
    this.props.onRemove(event)
    event.preventDefault()
    event.stopPropagation()
  }

  render() {
    var className = ""
    if (this.props.active) {
      className += " active"
    }
    if (this.props.subitem) {
      className += " subitem"
    }
    var removeButton = null
    if (this.props.onRemove) {
      removeButton = (
        <span
          className="glyphicon glyphicon-remove remove-item"
          onClick={this.onRemove}
        />
      )
    }
    return (
      <li className={className} role="presentation">
        <Link to={this.props.to} style={{ position: "relative" }}>
          {this.props.children}
          {removeButton}
        </Link>
      </li>
    )
  }
}
