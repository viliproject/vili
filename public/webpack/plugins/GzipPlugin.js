"use strict"
const zlib = require("zlib")
const RawSource = require("webpack-sources").RawSource

class GzipPlugin {
  constructor() {
    this.regExp = /\.js$|\.css$/
    this.compress = this.compress.bind(this)
  }

  apply(compiler) {
    compiler.plugin("this-compilation", compilation => {
      compilation.plugin("optimize-assets", this.compress)
    })
  }

  compress(assets, cb) {
    var tasks = []

    Object.keys(assets).forEach(file => {
      const task = new Promise((resolve, reject) => {
        if (!this.regExp.test(file)) {
          return resolve()
        }

        const asset = assets[file]
        const content = asset.source()

        if (!content.length) {
          return resolve()
        }

        zlib.gzip(content, (err, result) => {
          if (err) {
            return reject(err)
          }

          const fileParts = file.split(".")
          const ext = fileParts.pop()
          assets[fileParts.join(".") + ".gz." + ext] = new RawSource(result)
          return resolve()
        })
      })

      tasks.push(task)
    })

    Promise.all(tasks).then(() => cb())
  }
}

module.exports = GzipPlugin
