import querystring from 'querystring'
import request from 'browser-request'
import { Promise } from 'bluebird'
import WebSocketClient from './WebSocketClient'

class Response {
  constructor (results, error, res) {
    this.results = results || null
    this.error = error || null
    this.res = res
  }
}

export default class HttpClient {
  static all = function (httpRequests) {
    return Promise.all(httpRequests)
  };

  static series = function (httpRequestPairs) {
    return Promise.reduce(httpRequestPairs, (i, pair) => {
      const [httpRequest, params] = pair
      return httpRequest.apply(null, params)
    }, 0)
  };

  constructor (baseURL) {
    this.baseURL = baseURL || ''
    this.hooks = []
  }

  addHook (cb) {
    this.hooks.push(cb)
  }

  get (opts) {
    return this.send('GET', opts)
  }

  post (opts) {
    return this.send('POST', opts)
  }

  postForm (opts) {
    let o = Object.assign({}, opts)
    delete o.form
    let form = opts.form
    o.body = querystring.stringify(form)

    o.headers = Object.assign({}, o.headers, {
      'content-type': 'application/x-www-form-urlencoded'
    })

    return this.send('POST', o)
  }

  put (opts) {
    return this.send('PUT', opts)
  }

  delete (opts) {
    return this.send('DELETE', opts)
  }

  ws (opts) {
    opts.url = this.baseURL + opts.url
    return new WebSocketClient(opts)
  }

  isErrorCode (statusCode) {
    return statusCode >= 400
  }

  send (method, opts) {
    if (!opts.body && opts.json) {
      opts.body = JSON.stringify(opts.json)
    }
    const o = {
      url: this.baseURL + opts.url,
      method: method,
      qs: opts.query || {},
      body: opts.body,
      headers: opts.headers
    }

    if (o.method.toLowerCase() === 'get') {
      if (opts.pagination && typeof opts.pagination.next_offset === 'number') {
        o.qs.offset = opts.pagination.next_offset
      } else if (opts.pagination && typeof opts.pagination.previous_offset === 'number') {
        o.qs.offset = opts.pagination.previous_offset
      }

      if (opts.limit) {
        o.qs.limit = opts.limit
      }

      if (opts.sort) {
        o.qs.sort = opts.sort
      }
    }

    return new Promise((resolve, reject) => {
      request(o, (err, res, body) => {
        for (let i = 0; i < this.hooks.length; i++) {
          this.hooks[i](res.statusCode)
        }

        if (err) {
          return resolve(new Response(null, err, res))
        }

        if (this.isErrorCode(res.statusCode)) {
          try {
            const err = JSON.parse(body)
            return resolve(new Response(null, err, res))
          } catch (e) {
            return resolve(new Response(null, {}, res))
          }
        }

        if (res.statusCode === 204) {
          return resolve(new Response(null, null, res))
        }

        try {
          const data = JSON.parse(body)
          return resolve(new Response(data, null, res))
        } catch (e) {
          return resolve(new Response(null, e, res))
        }
      })
    })
  }
}
