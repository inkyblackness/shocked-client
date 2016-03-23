/* global ko */
"use strict";

function ProjectsAdapter() {
   this.vm = null;
   this.rest = null;
}

ProjectsAdapter.prototype.postConstruct = function() {
   var self = this;
   var vmProjects = {
      available: ko.observableArray(),
      selected: ko.observable(),

      newProjectId: ko.observable("")
   };

   vmProjects.canCreateNewProject = ko.computed(function() {
      var newProjectId = vmProjects.newProjectId();
      var available = vmProjects.available();
      var existing = false;

      available.forEach(function(project) {
         if (project.id === newProjectId) {
            existing = true;
         }
      });

      return (newProjectId !== "") && !existing;
   });

   vmProjects.createProject = function() {
      var projectTemplate = {
         id: vmProjects.newProjectId()
      };

      self.rest.postResource("/projects", projectTemplate, function(projectData) {
         vmProjects.available.push(projectData);
      }, function() {});
   };

   this.vm.projects = vmProjects;
   this.rest.getResource("/projects", function(projects) {
      self.onProjectsLoaded(projects);
   }, function() {});
};

ProjectsAdapter.prototype.onProjectsLoaded = function(projects) {
   var vmProjects = this.vm.projects;

   vmProjects.available(projects.items);
};

module.exports = ProjectsAdapter;
