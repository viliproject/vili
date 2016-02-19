import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars

export class LinkMenuItem extends React.Component {
    render() {
        var className='';
        if (this.props.active) {
            className += ' active';
        }
        if (this.props.subitem) {
            className += ' subitem';
        }
        return (
            <li className={className} role="presentation">
                <Link to={this.props.to}>{this.props.children}</Link>
            </li>
        );
    }
}
