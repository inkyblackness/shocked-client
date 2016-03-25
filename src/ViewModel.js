/* global ko */
"use strict";

ko.options.deferUpdates = true;

function ViewModel() {}

ViewModel.prototype.postConstruct = function() {
   this.mainSections = ko.observableArray(["project", "map", "textures"]);
   this.selectedMainSection = ko.observable("project");

   this.languages = ["ENG", "FRA", "GER"];
};

module.exports = ViewModel;
