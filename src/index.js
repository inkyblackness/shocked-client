/* global ko */
"use strict";

var unifier = require("./util/unifier.js");
var rest = require("./browser/rest.js");
var defer = require("./browser/defer.js");

var infuse = require("infuse.js");
var injector = new infuse.Injector();
//injector.strictMode = true;

ko.options.deferUpdates = true;

var vm = {
   tileTypes: ["", "open", "solid",
      "diagonalOpenSouthEast", "diagonalOpenSouthWest", "diagonalOpenNorthWest", "diagonalOpenNorthEast",
      "slopeSouthToNorth", "slopeWestToEast", "slopeNorthToSouth", "slopeEastToWest",
      "valleySouthEastToNorthWest", "valleySouthWestToNorthEast", "valleyNorthWestToSouthEast", "valleyNorthEastToSouthWest",
      "ridgeNorthWestToSouthEast", "ridgeNorthEastToSouthWest", "ridgeSouthEastToNorthWest", "ridgeSouthWestToNorthEast"
   ],

   mainSections: ko.observableArray(["project", "map"]),
   selectedMainSection: ko.observable("project")
};

injector.mapValue("rest", rest);
injector.mapValue("sys", {
   defer: defer
});

injector.mapValue("vm", vm);

var ProjectsAdapter = require("./vmAdapter/ProjectsAdapter");
injector.mapClass("projectsAdapter", ProjectsAdapter, true);
var LevelsAdapter = require("./vmAdapter/LevelsAdapter");
injector.mapClass("levelsAdapter", LevelsAdapter, true);
var MapAdapter = require("./vmAdapter/MapAdapter");
injector.mapClass("mapAdapter", MapAdapter, true);

var projectsAdapter = injector.getValue("projectsAdapter");
var levelsAdapter = injector.getValue("levelsAdapter");
var mapAdapter = injector.getValue("mapAdapter");

ko.applyBindings(vm);
