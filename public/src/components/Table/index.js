import React, { PropTypes } from 'react'
import { Table } from 'react-bootstrap'
import _ from 'underscore'

export default class ViliTable extends React.Component {
  static propTypes = {
    rows: PropTypes.array,
    columns: PropTypes.array,
    multiselect: PropTypes.bool,
    isSelectAll: PropTypes.bool,
    hover: PropTypes.bool,
    fill: PropTypes.bool,
    striped: PropTypes.bool,
    onRowSelection: PropTypes.func,
    shouldRenderCheckbox: PropTypes.func
  }

  static defaultProps = {
    hover: true
  }

  render () {
    var headerCells = []
    var subheaderCells = []
    var keys = []
    var hasSubheaders = false
    _.each(this.props.columns, function (col, ix) {
      if (col.subcolumns) {
        headerCells.push(
          <th key={col.key} style={col.style} colSpan={col.subcolumns.length}>{col.title}</th>)
        _.each(col.subcolumns, function (subcol) {
          subheaderCells.push(<th key={subcol.key} data-column={subcol.key}>{subcol.title}</th>)
          keys.push({key: subcol.key, col})
        })
        hasSubheaders = true
      } else {
        headerCells.push(<th key={col.key} style={col.style} data-column={col.key}>{col.title}</th>)
        subheaderCells.push(<th key={col.key} />)
        keys.push({key: col.key, col})
      }
      return
    })

    var header = [<tr key='header-row'>{headerCells}</tr>]
    if (hasSubheaders) {
      header.push(<tr key='subheader-row'>{subheaderCells}</tr>)
    }
    var rows = _.map(this.props.rows, function (row, ix) {
      if (!row) {
        return null
      }
      if (row.component) {
        return row.component
      }
      var cells = _.map(keys, function (key) {
        return <td key={key.key} data-column={key.key} style={key.col.style}>{row[key.key]}</td>
      })
      var className = row._className || ''
      return <tr key={row.key || 'row-' + ix} className={className} bsStyle={row._bsStyle}>{cells}</tr>
    })
    return (
      <Table hover={this.props.hover} fill={this.props.fill} striped={this.props.striped}>
        <thead>{header}</thead>
        <tbody>{rows}</tbody>
      </Table>
    )
  }
}
