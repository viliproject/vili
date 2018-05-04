import HttpClient from "./HttpClient"

const httpClient = new HttpClient("/api/v1/")

export { httpClient }
export default {
  deployments: {
    watch(handler, env, qs) {
      qs = qs || {}
      qs.watch = "true"
      return httpClient.ws({
        url: `envs/${env}/deployments`,
        qs: qs,
        messageHandler: handler,
      })
    },
    async getRepository(env, name) {
      return await httpClient.get({
        url: `envs/${env}/deployments/${name}/repository`,
      })
    },
    async getSpec(env, name) {
      return await httpClient.get({
        url: `envs/${env}/deployments/${name}/spec`,
      })
    },
    async getService(env, name) {
      return await httpClient.get({
        url: `envs/${env}/deployments/${name}/service`,
      })
    },
    async scale(env, name, replicas) {
      return await httpClient.put({
        url: `envs/${env}/deployments/${name}/scale`,
        json: { replicas },
      })
    },
    async resume(env, name) {
      return await httpClient.put({
        url: `envs/${env}/deployments/${name}/resume`,
      })
    },
    async pause(env, name) {
      return await httpClient.put({
        url: `envs/${env}/deployments/${name}/pause`,
      })
    },
    async rollback(env, name, toRevision) {
      return await httpClient.put({
        url: `envs/${env}/deployments/${name}/rollback`,
        json: { toRevision },
      })
    },
  },

  replicaSets: {
    watch(handler, env, qs) {
      qs = qs || {}
      qs.watch = "true"
      return httpClient.ws({
        url: `envs/${env}/replicasets`,
        qs: qs,
        messageHandler: handler,
      })
    },
  },

  rollouts: {
    async create(env, deployment, spec) {
      const qs = { async: "true" }
      return await httpClient.post({
        url: `envs/${env}/deployments/${deployment}/rollouts`,
        query: qs,
        json: spec,
      })
    },
    async rollback(env, deployment, id) {
      return await httpClient.post({
        url: `envs/${env}/deployments/${deployment}/rollouts/${id}/rollback`,
      })
    },
  },

  jobs: {
    async getRepository(env, name) {
      return await httpClient.get({
        url: `envs/${env}/jobs/${name}/repository`,
      })
    },
    async getSpec(env, name) {
      return await httpClient.get({ url: `envs/${env}/jobs/${name}/spec` })
    },
  },
  jobRuns: {
    watch(handler, env, qs) {
      qs = qs || {}
      qs.watch = "true"
      return httpClient.ws({
        url: `envs/${env}/jobs`,
        qs: qs,
        messageHandler: handler,
      })
    },
    async create(env, job, spec) {
      const qs = { async: "true" }
      return await httpClient.post({
        url: `envs/${env}/jobs/${job}/runs`,
        query: qs,
        json: spec,
      })
    },
    async del(env, run) {
      return await httpClient.delete({ url: `envs/${env}/jobs/${run}` })
    },
  },
  configmaps: {
    watch(handler, env, qs) {
      qs = qs || {}
      qs.watch = "true"
      return httpClient.ws({
        url: `envs/${env}/configmaps`,
        qs: qs,
        messageHandler: handler,
      })
    },
    async getSpec(env, name) {
      return await httpClient.get({
        url: `envs/${env}/configmaps/${name}/spec`,
      })
    },
    async create(env, name) {
      return await httpClient.post({ url: `envs/${env}/configmaps/${name}` })
    },
    async del(env, name) {
      return await httpClient.delete({ url: `envs/${env}/configmaps/${name}` })
    },
    async setKeys(env, name, values) {
      return await httpClient.put({
        url: `envs/${env}/configmaps/${name}/keys`,
        json: values,
      })
    },
    async delKey(env, name, key) {
      return await httpClient.delete({
        url: `envs/${env}/configmaps/${name}/${key}`,
      })
    },
  },

  pods: {
    watch(handler, env, qs) {
      qs = qs || {}
      qs.watch = "true"
      return httpClient.ws({
        url: `envs/${env}/pods`,
        qs: qs,
        messageHandler: handler,
      })
    },
    watchLog(handler, env, name, qs) {
      qs = qs || {}
      qs.follow = "true"
      return httpClient.ws({
        url: `envs/${env}/pods/${name}/log`,
        qs: qs,
        messageHandler: handler,
      })
    },
    async del(env, name) {
      return await httpClient.delete({ url: `envs/${env}/pods/${name}` })
    },
  },

  nodes: {
    watch(handler, env, qs) {
      qs = qs || {}
      qs.watch = "true"
      return httpClient.ws({
        url: `envs/${env}/nodes`,
        qs: qs,
        messageHandler: handler,
      })
    },
    async setSchedulable(env, name, onOff) {
      return await httpClient.put({
        url: `envs/${env}/nodes/${name}/${onOff.toLowerCase()}`,
      })
    },
  },

  services: {
    async create(env, deployment) {
      return await httpClient.post({
        url: `envs/${env}/deployments/${deployment}/service`,
      })
    },
  },

  functions: {
    watch(handler, env, qs) {
      qs = qs || {}
      qs.watch = "true"
      return httpClient.ws({
        url: `envs/${env}/functions`,
        qs: qs,
        messageHandler: handler,
      })
    },
    async getRepository(env, name) {
      return await httpClient.get({
        url: `envs/${env}/functions/${name}/repository`,
      })
    },
    async getSpec(env, name) {
      return await httpClient.get({
        url: `envs/${env}/functions/${name}/spec`,
      })
    },
    async deploy(env, name, spec) {
      return await httpClient.put({
        url: `envs/${env}/functions/${name}/deploy`,
        json: spec,
      })
    },
    async rollback(env, name, toVersion) {
      return await httpClient.put({
        url: `envs/${env}/functions/${name}/rollback`,
        json: { toVersion },
      })
    },
  },

  releases: {
    watch(handler, env) {
      return httpClient.ws({
        url: `envs/${env}/releases`,
        messageHandler: handler,
      })
    },
    async getSpec(env) {
      return await httpClient.get({ url: `envs/${env}/releases/spec` })
    },
    async create(env, spec) {
      return await httpClient.post({ url: `envs/${env}/releases`, json: spec })
    },
    async createFromLatest(env, spec) {
      return await httpClient.post({ url: `envs/${env}/releases?latest=true` })
    },
    async deploy(env, name) {
      return await httpClient.put({
        url: `envs/${env}/releases/${name}/deploy`,
      })
    },
    async del(env, name) {
      return await httpClient.delete({ url: `envs/${env}/releases/${name}` })
    },
  },

  branches: {
    async get() {
      return await httpClient.get({ url: `branches` })
    },
  },

  environments: {
    async create(spec) {
      return await httpClient.post({ url: `environments`, json: spec })
    },
    async del(name) {
      return await httpClient.delete({ url: `environments/${name}` })
    },
    async getSpec(name, branch) {
      const qs = { name: name, branch: branch }
      return await httpClient.get({ url: `environments/spec`, query: qs })
    },
  },
}
