/* global ko */
"use strict";

function ProjectsAdapter() {
   this.vm = null;
   this.rest = null;
}

ProjectsAdapter.prototype.postConstruct = function() {
   var vmProjects = {
      available: ko.observableArray(),
      selected: ko.observable()
   };

   this.vm.projects = vmProjects;
   this.rest.getResource("/projects", function(projects) {
      vmProjects.available(projects.items);
   }, function() {});
};

module.exports = ProjectsAdapter;
