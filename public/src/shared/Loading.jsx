import React from 'react';


export class Loading extends React.Component {
    render() {
        return (
            <div className="loading">
                <span className="glyphicon glyphicon-refresh glyphicon-refresh-animate"></span>
                <span>Loading</span>
            </div>
        );
    }
}
