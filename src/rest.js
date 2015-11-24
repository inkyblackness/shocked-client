/* global $ */
"use strict";

var rest = {};

rest.getResource = function(url, onSuccess, onFailure) {
   var options = {
      method: "GET",
      url: url,
      dataType: "json",
      jsonp: false,
      success: onSuccess,
      error: onFailure
   };

   $.ajax(options);
};

rest.putResource = function(url, data, onSuccess, onFailure) {
   var options = {
      method: "PUT",
      url: url,
      dataType: "json",
      contentType: "application/json",
      data: JSON.stringify(data),
      jsonp: false,
      processData: false,
      success: onSuccess,
      error: onFailure
   };

   $.ajax(options);
};

module.exports = rest;
