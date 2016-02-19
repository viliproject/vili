import React from 'react'
import Firebase from 'firebase';
import router from './router';
import style from './less/app.less';
style.use();

document.addEventListener('DOMContentLoaded', function() {
    if (window.appconfig) {
        // initialize firebase
        var db = new Firebase(window.appconfig.firebase.url);
        db.authWithCustomToken(window.appconfig.firebase.token, function(err) {
            if (err) {
                // TODO show user error saying firebase auth failed
                return;
            }
        }, {
            remember: 'sessionOnly'
        });
    }

    // run router
    router.run(function(Handler) { // eslint-disable-line no-unused-vars
        React.render(<Handler db={db} />, document.body);
    });
});
