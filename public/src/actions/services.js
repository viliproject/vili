
export function createService (env, name) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.services.create(env, name)
    return { results, error }
  }
}
