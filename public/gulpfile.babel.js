import gulp from 'gulp';
import eslint from 'gulp-eslint';
import webpack, { DefinePlugin } from 'webpack';
import del from 'del';
import fs from 'fs';
import path from 'path';
import _ from 'underscore';
import runSequence from 'run-sequence';
import gutil from 'gutil';

process.env.PACKAGE = path.resolve('./package.json');

try {
    var notifier = require('node-notifier');
} catch (e) {
    notifier = null;
}

gulp.task('lint', function() {
  return gulp.src('src/**/*.js')
      .pipe(eslint())
      .pipe(eslint.format());
});

// webpack
const STYLE_LOADER = 'style-loader/useable';
function makeConfig(debug) {
    var cssLoader = debug ? 'css-loader' : 'css-loader?minimize';
    return {
        entry: './src/app.jsx',
        output: {
            path: './build/',
            publicPath: './',
            sourcePrefix: '  ',
            filename: 'app-[hash].js',
        },

        cache: debug,
        debug: debug,

        stats: {
            colors: true,
            reasons: debug
        },

        devtool: debug ? 'source-map' : false,

        plugins: [
            new webpack.optimize.OccurenceOrderPlugin(),
            new DefinePlugin({
                'process.env.NODE_ENV': debug ? '"development"' : '"production"',
                '__DEV__': debug,
                '__SERVER__': false,
            }),
        ].concat(debug ? [] : [
            new webpack.optimize.DedupePlugin(),
            new webpack.optimize.UglifyJsPlugin(),
            new webpack.optimize.AggressiveMergingPlugin(),
        ]).concat([
            function() {
                this.plugin('done', function(stats) {
                    fs.writeFileSync(
                        path.join('./build/stats.json'),
                        JSON.stringify(_.pick(stats.toJson(), ['hash', 'assets'])));
                    if (notifier) {
                        notifier.notify({
                            'title': 'vili',
                            'message': 'jsx built'
                        });
                    }
                });
            }
        ]),

        resolve: {
            extensions: ['', '.js', '.jsx']
        },

        module: {
            loaders: [
                {
                    test: /\.css$/,
                    loader: STYLE_LOADER + '!' + cssLoader + '!postcss-loader'
                },
                {
                    test: /\.less$/,
                    loader: STYLE_LOADER + '!' + cssLoader + '!postcss-loader!less-loader'
                },
                {
                    test: /\.jsx?$/,
                    exclude: /node_modules/,
                    loader: 'babel-loader'
                }
            ]
        },
        node: {
            fs: 'empty'
        }
    };
}

function webpackBuild(callback, debug) {
    // run webpack
    webpack(makeConfig(debug), function(err, stats) {
        if(err) throw new gutil.PluginError('webpack', err);
        gutil.log('[webpack]', stats.toString({
            // output options
        }));
        callback();
    });
}

gulp.task('webpack', function(callback) {
    webpackBuild(callback, false);
});

gulp.task('webpackdev', function(callback) {
    webpackBuild(callback, true);
});

gulp.task('build', function() {
    runSequence('clean', ['webpack']);
});
gulp.task('devbuild', function() {
    runSequence('clean', ['webpackdev']);
});

gulp.task('watch', ['devbuild'], function() {
    gulp.watch('./src/**/*.{js,jsx,less,css}', ['webpackdev']);
});

gulp.task('develop', ['watch']);

gulp.task('clean', function(done) {
    del(['./build'], done);
});
