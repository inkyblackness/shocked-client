/* global __dirname */
"use strict";
var path = require("path");

module.exports = function(grunt) {
   var jsFiles = ["Gruntfile.js", "src/**/*.js"];

   grunt.initConfig({
      pkg: grunt.file.readJSON("package.json"),

      // Run JSHint on all sources
      jshint: {
         options: {
            jshintrc: "./.jshintrc"
         },
         all: jsFiles
      },

      // JSBeautifier on all sources
      jsbeautifier: {
         standard: {
            src: jsFiles,
            options: {
               js: grunt.file.readJSON(".jsbeautifyrc")
            }
         }
      },

      // browserify for packing all commonjs files
      browserify: {
         client: {
            src: ["src/index.js"],
            dest: "www/client.js",
            options: {
               browserifyOptions: {
                  standalone: "client"
               }
            }
         }
      },

      // uglify for compression
      uglify: {
         lib: {
            files: {
               "www/client.min.js": ["www/client.js"]
            },
            options: {
               mangle: {
                  except: []
               }
            }
         }
      },
   });

   grunt.loadNpmTasks("grunt-jsbeautifier");
   grunt.loadNpmTasks("grunt-browserify");
   grunt.loadNpmTasks("grunt-contrib-jshint");
   grunt.loadNpmTasks("grunt-contrib-uglify");

   grunt.registerTask("lint", ["jshint", "jsbeautifier"]);
   grunt.registerTask("compile", ["browserify", "uglify"]);

   grunt.registerTask("default", ["lint", "compile"]);
};
