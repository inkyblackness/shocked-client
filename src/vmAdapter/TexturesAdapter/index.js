/* global ko */
"use strict";

function TexturesAdapter() {
   this.projectsAdapter = null;

   this.vm = null;
   this.rest = null;

   this.revisionCounter = 0;
}

TexturesAdapter.prototype.postConstruct = function() {
   var self = this;
   var vmTextures = {
      all: ko.observableArray(),

      pageIndex: ko.observable(0),
      pageSideWidth: ko.observable(4),
      selectedTexture: ko.observable()
   };

   vmTextures.pageSize = ko.computed(function() {
      var pageSideWidth = vmTextures.pageSideWidth();

      return pageSideWidth * pageSideWidth;
   });
   vmTextures.page = ko.computed(function() {
      var pageIndex = vmTextures.pageIndex();
      var pageSize = vmTextures.pageSize();
      var pageStart = pageIndex * pageSize;

      return vmTextures.all().slice(pageStart, pageStart + pageSize);
   });
   vmTextures.canShowNextPage = ko.computed(function() {
      var pageIndex = vmTextures.pageIndex();
      var pageSize = vmTextures.pageSize();
      var pageStart = pageIndex * pageSize;

      return (pageStart + pageSize) < vmTextures.all().length;
   });

   vmTextures.nextPage = function() {
      if (vmTextures.canShowNextPage()) {
         vmTextures.pageIndex(vmTextures.pageIndex() + 1);
      }
   };
   vmTextures.previousPage = function() {
      var pageIndex = vmTextures.pageIndex();

      if (pageIndex > 0) {
         vmTextures.pageIndex(pageIndex - 1);
      }
   };

   vmTextures.selectTexture = function(textureEntry) {
      vmTextures.selectedTexture(textureEntry);
   };

   vmTextures.reload = this.reloadTextures.bind(this);
   vmTextures.getTexture = this.getTexture.bind(this);


   this.vm.textures = vmTextures;
   this.vm.projects.selected.subscribe(function(project) {
      if (project) {
         self.reloadTexturesFrom(project);
      } else {
         vmTextures.all([]);
      }
   });
};

TexturesAdapter.prototype.getTexture = function(id) {
   var entry = this.vm.textures.all().find(function(entry) {
      return entry.id === id;
   });

   if (!entry) {
      entry = this.createTextureEntry(id);
      this.queryTextureData(entry);
      this.vm.textures.all.push(entry);
   }

   return entry;
};

TexturesAdapter.prototype.reloadTextures = function() {
   var project = this.vm.projects.selected();

   if (project) {
      this.reloadTexturesFrom(project);
   }
};

TexturesAdapter.prototype.reloadTexturesFrom = function(project) {
   var self = this;

   this.revisionCounter++;
   this.rest.getResource(project.href + "/textures", function(textures) {
      self.setTextures(textures.list);
   }, function() {});
};

TexturesAdapter.prototype.setTextures = function(textureDataList) {
   var vmTextures = this.vm.textures;
   var self = this;
   var newList = textureDataList.map(function(textureData) {
      var textureEntry = self.createTextureEntry(textureData.id);

      self.onTextureDataLoaded(textureEntry, textureData);

      return textureEntry;
   });

   vmTextures.all(newList);
};

TexturesAdapter.prototype.queryTextureData = function(textureEntry) {
   var self = this;
   var project = this.vm.projects.selected();

   this.rest.getResource(project.href + "/textures/" + textureEntry.id, function(textureData) {
      self.onTextureDataLoaded(textureEntry, textureData);
   }, function() {});
};

TexturesAdapter.prototype.createTextureEntry = function(id) {
   var entry = {
      id: id,
      largeTextureUrl: ko.observable(""),
      mediumTextureUrl: ko.observable(""),
      smallTextureUrl: ko.observable(""),
      iconTextureUrl: ko.observable(""),
      texts: ko.observableArray()
   };

   return entry;
};

TexturesAdapter.prototype.onTextureDataLoaded = function(textureEntry, textureData) {
   var self = this;
   var texts = [];

   var textEntryAt = function(index) {
      var entry;

      if (index < texts.length) {
         entry = texts[index];
      } else {
         entry = {
            name: ko.observable(""),
            cantBeUsed: ko.observable("")
         };
         texts.push(entry);
      }

      return entry;
   };

   textureEntry.id = textureData.id;
   textureData.images.forEach(function(link) {
      if (link.rel === "large") {
         textureEntry.largeTextureUrl(link.href + "/png?clientRevision=" + self.revisionCounter);
      }
      if (link.rel === "medium") {
         textureEntry.mediumTextureUrl(link.href + "/png?clientRevision=" + self.revisionCounter);
      }
      if (link.rel === "small") {
         textureEntry.smallTextureUrl(link.href + "/png?clientRevision=" + self.revisionCounter);
      }
      if (link.rel === "icon") {
         textureEntry.iconTextureUrl(link.href + "/png?clientRevision=" + self.revisionCounter);
      }
   });
   textureData.properties.name.forEach(function(text, index) {
      textEntryAt(index).name(text);
   });
   textureData.properties.cantBeUsed.forEach(function(text, index) {
      textEntryAt(index).cantBeUsed(text);
   });
   textureEntry.texts(texts);
};

module.exports = TexturesAdapter;
