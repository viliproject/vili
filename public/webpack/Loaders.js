'use strict'

// A webpack loader abstraction
class BaseLoader {
  get test () {
    return null
  }

  get exclude () {
    return null
  }

  get include () {
    return null
  }

  get loader () {
    return null
  }

  get query () {
    return null
  }

  toJSON () {
    let json = {
      test: this.test,
      loader: this.loader
    }

    if (this.query) {
      json.query = this.query
    }

    if (this.exclude) {
      json.exclude = this.exclude
    }

    if (this.include) {
      json.include = this.include
    }

    return json
  }
}

class JavascriptLoader extends BaseLoader {
  get test () {
    return /\.(js|jsx)$/
  }

  get exclude () {
    return [/node_modules/]
  }

  get loader () {
    return 'babel-loader'
  }

  get query () {
    return {
      cacheDirectory: true,
      plugins: [
        'transform-decorators-legacy',
        'syntax-async-functions',
        'transform-regenerator',
        'transform-runtime',
        'transform-object-assign'
      ]
    }
  }
}

class LessLoader extends BaseLoader {
  get test () {
    return /\.less$/
  }

  get loader () {
    // const CSS_SOURCEMAPS = '?sourceMap';
    const CSS_SOURCEMAPS = ''

    return `style!css${CSS_SOURCEMAPS}!postcss!less${CSS_SOURCEMAPS}`
  }
}

class JSONLoader extends BaseLoader {
  get test () {
    return /\.json$/
  }

  get loader () {
    return 'json-loader'
  }
}

class AssetLoader extends BaseLoader {
  get test () {
    return /\.(png|jpg|jpeg|gif|svg|eot|ttf|otf|otf2|woff|woff2?)(\?\S*)?$/
  }

  get loader () {
    return 'url-loader?limit=30000'
  }
}

class HandlebarsLoader extends BaseLoader {
  get test () {
    return /\.hbs$/
  }

  get loader () {
    return 'handlebars-loader'
  }
}

module.exports = {
  BaseLoader: BaseLoader,
  JavascriptLoader: JavascriptLoader,
  LessLoader: LessLoader,
  JSONLoader: JSONLoader,
  AssetLoader: AssetLoader,
  HandlebarsLoader: HandlebarsLoader
}
