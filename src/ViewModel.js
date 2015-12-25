/* global ko */
"use strict";

ko.options.deferUpdates = true;

function ViewModel() {}

ViewModel.prototype.postConstruct = function() {
   this.mainSections = ko.observableArray(["project", "map"]);
   this.selectedMainSection = ko.observable("project");
};

module.exports = ViewModel;
