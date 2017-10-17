/* global WebSocket */

export default class WebSocketClient {
  constructor (opts) {
    const loc = window.location
    this.url = ((loc.protocol === 'https:') ? 'wss://' : 'ws://') + loc.host + opts.url
    if (opts.qs) {
      this.url += '?' + this.queryString(opts.qs)
    }
    this.messageHandler = opts.messageHandler

    // create the connection
    this.startWebSocket()
  }

  startWebSocket () {
    const self = this
    this.ws = new WebSocket(this.url)
    this.ws.onmessage = function (event) {
      const data = JSON.parse(event.data)
      if (data.type === 'CLOSED') {
        self.close()
      } else {
        self.messageHandler(data)
      }
    }
    this.ws.onclose = function (event) {
      if (!self.closed) {
        setTimeout(function () {
          self.startWebSocket()
        }, 5000)
      }
    }
  }

  close () {
    this.closed = true
    this.ws.close()
  }

  // utils
  queryString (query) {
    const str = []
    for (var p in query) {
      if (query.hasOwnProperty(p)) {
        str.push(encodeURIComponent(p) + '=' + encodeURIComponent(query[p]))
      }
    }
    return str.join('&')
  }
}
