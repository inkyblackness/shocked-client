/* global ko */
"use strict";

function LevelsAdapter() {
   this.projectsAdapter = null;

   this.vm = null;
   this.rest = null;
}

LevelsAdapter.prototype.postConstruct = function() {
   var rest = this.rest;
   var vmLevels = {
      available: ko.observableArray()
   };

   this.vm.levels = vmLevels;
   this.vm.projects.selected.subscribe(function(project) {
      if (project) {
         rest.getResource(project.href + "/archive/levels", function(levels) {
            vmLevels.available(levels.list);
         }, function() {});
      } else {
         vmLevels.available([]);
      }
   });
};

module.exports = LevelsAdapter;
