/* global window */
"use strict";

var defer = function(cb) {
   window.setTimeout(cb, 0);
};

module.exports = defer;
