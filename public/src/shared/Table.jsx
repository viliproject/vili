import React from 'react';
import _ from 'underscore';

export class Table extends React.Component {
    render() {
        var headerCells = [];
        var subheaderCells = [];
        var keys = [];
        var hasSubheaders = false;
        _.each(this.props.columns, function(col) {
            if (col.subcolumns) {
                headerCells.push(
                    <th colSpan={col.subcolumns.length}>{col.title}</th>
                );
                _.each(col.subcolumns, function(subcol) {
                    subheaderCells.push(<th data-column={subcol.key}>{subcol.title}</th>);
                    keys.push(subcol.key);
                });
                hasSubheaders = true;
            } else {
                headerCells.push(<th data-column={col.key}>{col.title}</th>);
                subheaderCells.push(<th></th>);
                keys.push(col.key);
            }
            return;
        });

        var header = [<tr key="header-row">{headerCells}</tr>];
        if (hasSubheaders) {
            header.push(<tr key="subheader-row">{subheaderCells}</tr>);
        }
        var rows = _.map(this.props.rows, function(row, ix) {
            if (row._row) {
                return row._row;
            }
            var cells = _.map(keys, function(key) {
                return <td data-column={key}>{row[key]}</td>;
            });
            var className = row._className || '';
            return <tr key={'row-'+ix} className={className}>{cells}</tr>
        });
        return (
            <table className="table table-hover">
                <thead key="thead">{header}</thead>
                <tbody key="tbody">{rows}</tbody>
            </table>
        );
    }
}
