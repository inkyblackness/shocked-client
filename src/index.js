/* global ko */
"use strict";

var infuse = require("infuse.js");
var injector = new infuse.Injector();

injector.mapValue("rest", require("./browser/rest.js"));
injector.mapValue("sys", {
   defer: require("./browser/defer.js")
});

injector.mapClass("vm", require("./ViewModel.js"), true);
injector.mapClass("projectsAdapter", require("./vmAdapter/ProjectsAdapter"), true);
injector.mapClass("texturesAdapter", require("./vmAdapter/TexturesAdapter"), true);
injector.mapClass("levelsAdapter", require("./vmAdapter/LevelsAdapter"), true);
injector.mapClass("mapAdapter", require("./vmAdapter/MapAdapter"), true);

function Application() {
   this.vm = null;
   this.projectsAdapter = null;
   this.levelsAdapter = null;
   this.mapAdapter = null;
   this.texturesAdapter = null;
}

Application.prototype.postConstruct = function() {
   ko.applyBindings(this.vm);
};

injector.mapClass("app", Application, true);

module.exports = {
   Application: Application,
   injector: injector,
   buildAndRun: function() {
      return injector.getValue("app");
   }
};
