import React from 'react';
import { Link } from 'react-router'; // eslint-disable-line no-unused-vars

export class LinkMenuItem extends React.Component {
    constructor(props) {
        super(props);

        this.onRemove = this.onRemove.bind(this);
    }

    render() {
        var className='';
        if (this.props.active) {
            className += ' active';
        }
        if (this.props.subitem) {
            className += ' subitem';
        }
        var removeButton = null;
        if (this.props.onRemove) {
            removeButton = <span className="glyphicon glyphicon-remove remove-item" onClick={this.onRemove} />;
        }
        return (
            <li className={className} role="presentation">
                <Link to={this.props.to} style={{ position: 'relative' }}>
                    {this.props.children}
                    {removeButton}
                </Link>
            </li>
        );
    }

    onRemove(event) {
        this.props.onRemove(event);
        event.preventDefault();
        event.stopPropagation();
    }
}
