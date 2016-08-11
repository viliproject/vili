import request from 'superagent';
import _ from 'underscore';


function makeRequest(req) {
    return new Promise(function(resolve, reject) {
        req.end(function(err, res) {
            if (err) {
                return reject(err);
            }
            resolve(res.body);
        });
    });
}

function makeGetRequest(endpoint, query) {
    var req = request.get('/api/v1' + endpoint);
    if (query) {
        req.query(query);
    }
    return makeRequest(req);
}

function makePostRequest(endpoint, data) {
    var req = request.post('/api/v1' + endpoint);
    if (data) {
        req.send(data);
    }
    return makeRequest(req);
}

function makePutRequest(endpoint, data) {
    var req = request.put('/api/v1' + endpoint);
    if (data) {
        req.send(data);
    }
    return makeRequest(req);
}

function makeDeleteRequest(endpoint) {
    var req = request.del('/api/v1' + endpoint);
    return makeRequest(req);
}

class ViliApi {
    constructor(opts) {
        this.opts = opts;

        this.apps = {
            get: function(env, name, qs) {
                if (_.isObject(name)) {
                    qs = name;
                    name = null;
                }
                if (name) {
                    return makeGetRequest('/envs/' + env + '/apps/' + name, qs);
                }
                return makeGetRequest('/envs/' + env + '/apps', qs);
            },
            scale: function(env, app, replicas) {
                return makePutRequest('/envs/' + env + '/apps/' + app + '/scale', {
                    replicas: replicas
                });
            },
        };

        this.jobs = {
            get: function(env, name, qs) {
                if (_.isObject(name)) {
                    qs = name;
                    name = null;
                }
                if (name) {
                    return makeGetRequest('/envs/' + env + '/jobs/' + name, qs);
                }
                return makeGetRequest('/envs/' + env + '/jobs', qs);
            },
        };

        this.nodes = {
            get: function(env, name, qs) {
                if (_.isObject(name)) {
                    qs = name;
                    name = null;
                }
                if (name) {
                    return makeGetRequest('/envs/' + env + '/nodes/' + name, qs);
                }
                return makeGetRequest('/envs/' + env + '/nodes', qs);
            },
            setSchedulable: function(env, name, onOff) {
                return makePutRequest('/envs/' + env + '/nodes/' + name + '/' + onOff.toLowerCase());
            },
        };

        this.pods = {
            get: function(env, name, qs) {
                if (_.isObject(name)) {
                    qs = name;
                    name = null;
                }
                if (name) {
                    return makeGetRequest('/envs/' + env + '/pods/' + name, qs);
                }
                return makeGetRequest('/envs/' + env + '/pods', qs);
            },
            delete: function(env, name) {
                return makeDeleteRequest('/envs/' + env + '/pods/' + name);
            },
        };

        this.services = {
            create: function(env, app) {
                return makePostRequest('/envs/' + env + '/apps/' + app + '/service');
            },
        };

        this.deployments = {
            create: function(env, app, spec) {
                var qs = '';
                if (spec.trigger) {
                    qs = '?trigger=true';
                }
                return makePostRequest('/envs/' + env + '/apps/' + app + '/deployments' + qs, spec);
            },
            setRollout: function(env, app, id, rollout) {
                return makePutRequest('/envs/' + env + '/apps/' + app + '/deployments/' + id + '/rollout', rollout);
            },
            resume: function(env, app, id) {
                return makePostRequest('/envs/' + env + '/apps/' + app + '/deployments/' + id + '/resume');
            },
            pause: function(env, app, id) {
                return makePostRequest('/envs/' + env + '/apps/' + app + '/deployments/' + id + '/pause');
            },
            rollback: function(env, app, id) {
                return makePostRequest('/envs/' + env + '/apps/' + app + '/deployments/' + id + '/rollback');
            },
        };

        this.runs = {
            create: function(env, job, spec) {
                var qs = '';
                if (spec.trigger) {
                    qs = '?trigger=true';
                }
                return makePostRequest('/envs/' + env + '/jobs/' + job + '/runs' + qs, spec);
            },
            start: function(env, job, id) {
                return makePostRequest('/envs/' + env + '/jobs/' + job + '/runs/' + id + '/start');
            },
            terminate: function(env, job, id) {
                return makePostRequest('/envs/' + env + '/jobs/' + job + '/runs/' + id + '/terminate');
            },
        };

        this.releases = {
            create: function(name, tag, spec) {
                return makePostRequest('/releases/' + name + '/' + tag, spec);
            },
            delete: function(name, tag) {
                return makeDeleteRequest('/releases/' + name + '/' + tag);
            },
        };

        this.environments = {
            create: function(spec) {
                return makePostRequest('/environments', spec);
            },
            delete: function(name) {
                return makeDeleteRequest('/environments/' + name);
            },
            branches: function() {
                return makeGetRequest('/envBranches');
            },
            spec: function(name, branch) {
                return makeGetRequest('/envSpec?name=' + name + '&branch=' + branch);
            },
        };

    }
}

export default new ViliApi();
