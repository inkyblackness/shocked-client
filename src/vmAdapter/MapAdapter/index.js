/* global ko */
"use strict";

var unifier = require("../../util/unifier.js");

var updateTileProperties = function(tile, tileData) {
   tile.tileType(tileData.properties.type);
   tile.floorHeight(tileData.properties.floorHeight);
   tile.ceilingHeight(tileData.properties.ceilingHeight);
   tile.slopeHeight(tileData.properties.slopeHeight);

   tile.northWallHeight(tileData.properties.calculatedWallHeights.north);
   tile.eastWallHeight(tileData.properties.calculatedWallHeights.east);
   tile.southWallHeight(tileData.properties.calculatedWallHeights.south);
   tile.westWallHeight(tileData.properties.calculatedWallHeights.west);

   tile.floorTextureIndex(tileData.properties.realWorld.floorTexture);
   tile.floorTextureRotations(tileData.properties.realWorld.floorTextureRotations);
   tile.ceilingTextureIndex(tileData.properties.realWorld.ceilingTexture);
   tile.ceilingTextureRotations(tileData.properties.realWorld.ceilingTextureRotations);
   tile.wallTextureIndex(tileData.properties.realWorld.wallTexture);
};

function MapAdapter() {
   this.projectsAdapter = null;

   this.vm = null;
   this.rest = null;
   this.sys = null;
}

function bytesToString(arr) {
   var result = arr.map(function(entry) {
      return entry.toString(16);
   }).map(function(entry) {
      var temp = "0" + entry;
      return temp.substr(temp.length - 2);
   }).join(", 0x");

   if (result.length > 0) {
      result = "0x" + result;
   }

   return "[" + result + "]";
}

MapAdapter.prototype.postConstruct = function() {
   var rest = this.rest;
   var vmMap = {
      tileTypes: ["", "open", "solid",
         "diagonalOpenSouthEast", "diagonalOpenSouthWest", "diagonalOpenNorthWest", "diagonalOpenNorthEast",
         "slopeSouthToNorth", "slopeWestToEast", "slopeNorthToSouth", "slopeEastToWest",
         "valleySouthEastToNorthWest", "valleySouthWestToNorthEast", "valleyNorthWestToSouthEast", "valleyNorthEastToSouthWest",
         "ridgeNorthWestToSouthEast", "ridgeNorthEastToSouthWest", "ridgeSouthEastToNorthWest", "ridgeSouthWestToNorthEast"
      ],

      selectedLevel: ko.observable(),

      levelTextures: ko.observableArray(),
      levelTextureUrls: ko.observableArray(),

      levelObjects: ko.observableArray(),

      tileRows: ko.observableArray(),

      textureDisplay: ko.observableArray(["Floor", "Ceiling"]),
      selectedTextureDisplay: ko.observable("Floor"),

      selectedTiles: ko.observableArray(),

      selectedTileType: ko.observable(""),
      selectedTileFloorTextureIndex: ko.observable(-1),
      selectedTileCeilingTextureIndex: ko.observable(-1),
      selectedTileWallTextureIndex: ko.observable(-1)
   };

   this.vm.map = vmMap;
   vmMap.onTileClicked = this.getTileClickedHandler();
   vmMap.shouldShowFloorTexture = ko.computed(function() {
      return vmMap.selectedTextureDisplay() === "Floor";
   });
   vmMap.shouldShowCeilingTexture = ko.computed(function() {
      return vmMap.selectedTextureDisplay() === "Ceiling";
   });
   vmMap.selectedTileFloorTextureUrl = ko.computed(this.computeTextureUrl(vmMap.selectedTileFloorTextureIndex));
   vmMap.selectedTileCeilingTextureUrl = ko.computed(this.computeTextureUrl(vmMap.selectedTileCeilingTextureIndex));
   vmMap.selectedTileWallTextureUrl = ko.computed(this.computeTextureUrl(vmMap.selectedTileWallTextureIndex));

   vmMap.selectedTiles.subscribe(function(newList) {
      var tileTypeUnifier = unifier.withResetValue("");
      var floorTextureIndexUnifier = unifier.withResetValue(-1);
      var ceilingTextureIndexUnifier = unifier.withResetValue(-1);
      var wallTextureIndexUnifier = unifier.withResetValue(-1);

      newList.forEach(function(tile) {
         tileTypeUnifier.add(tile.tileType());
         floorTextureIndexUnifier.add(tile.floorTextureIndex());
         ceilingTextureIndexUnifier.add(tile.ceilingTextureIndex());
         wallTextureIndexUnifier.add(tile.wallTextureIndex());
      });
      vmMap.selectedTileType(tileTypeUnifier.get());
      vmMap.selectedTileFloorTextureIndex(floorTextureIndexUnifier.get());
      vmMap.selectedTileCeilingTextureIndex(ceilingTextureIndexUnifier.get());
      vmMap.selectedTileWallTextureIndex(wallTextureIndexUnifier.get());
   });

   vmMap.selectedTileType.subscribe(function(newType) {
      if (newType !== "") {
         vmMap.selectedTiles().forEach(function(tile) {
            var properties = {
               type: newType,
            };

            var tileUrl = vmMap.selectedLevel().href + "/tiles/" + tile.y + "/" + tile.x;
            if (tile.tileType() !== newType) {
               rest.putResource(tileUrl, properties, function(tileData) {
                  updateTileProperties(tile, tileData);
               }, function() {});
            }
         });
      }
   });

   vmMap.filteredLevelObjects = ko.computed(function() {
      var allObjects = vmMap.levelObjects();
      var selectedTiles = vmMap.selectedTiles();
      var result = [];

      if (selectedTiles.length === 0) {
         result = allObjects;
      } else {
         allObjects.forEach(function(object) {
            var isIncluded = false;

            selectedTiles.forEach(function(tile) {
               if ((tile.x === object.raw.properties.tileX) && (tile.y === object.raw.properties.tileY)) {
                  isIncluded = true;
               }
            });
            if (isIncluded) {
               result.push(object);
            }
         });
      }

      return result;
   });

   var self = this;

   vmMap.selectedLevel.subscribe(function(level) {
      if (level) {
         rest.getResource(level.href + "/textures", function(levelTextures) {
            self.onTexturesLoaded(levelTextures);
         }, function() {});

         rest.getResource(level.href + "/objects", function(levelObjects) {
            vmMap.levelObjects.removeAll();
            levelObjects.table.forEach(function(raw) {
               var entry = {
                  raw: raw,
                  hacking: {
                     classDataString: bytesToString(raw.Hacking.ClassData)
                  },
                  name: ko.observable("???")
               };
               // TODO: Get game object info just once (central proxy); Also: query links["static"]
               rest.getResource(raw.links[0].href, function(gameObject) {
                  entry.name(gameObject.properties.longName[0]);
               });
               vmMap.levelObjects.push(entry);
            });
         }, function() {});

         rest.getResource(level.href + "/tiles", function(tileMap) {
            self.onMapLoaded(tileMap);
         }, function() {});
      }
   });

};

MapAdapter.prototype.computeTextureUrl = function(indexObservable) {
   var levelTextureUrls = this.vm.map.levelTextureUrls;

   return function() {
      var textureIndex = indexObservable();
      var urls = levelTextureUrls();
      var url = "";

      if ((textureIndex >= 0) && (textureIndex < urls.length)) {
         url = urls[textureIndex];
      }

      return url;
   };
};

MapAdapter.prototype.getTileClickedHandler = function() {
   var vmMap = this.vm.map;

   return function(tile, event) {
      var newState = !tile.isSelected();

      tile.isSelected(newState);
      if (event.ctrlKey) {
         if (newState) {
            vmMap.selectedTiles.push(tile);
         } else {
            vmMap.selectedTiles.remove(tile);
         }
      } else {
         vmMap.selectedTiles.removeAll().forEach(function(other) {
            other.isSelected(false);
         });
         if (newState) {
            vmMap.selectedTiles.push(tile);
         }
      }
   };
};

MapAdapter.prototype.onTexturesLoaded = function(levelTextures) {
   var newUrls = [];
   var newIds = [];
   var vmMap = this.vm.map;
   var vmProjects = this.vm.projects;

   levelTextures.ids.forEach(function(id) {
      newUrls.push(vmProjects.selected().href + "/textures/" + id + "/large/png");
      newIds.push(id);
   });
   vmMap.levelTextures(newIds);
   vmMap.levelTextureUrls(newUrls);
};

MapAdapter.prototype.onMapLoaded = function(tileMap) {
   var self = this;
   var rows = [];

   tileMap.Table.reverse();
   tileMap.Table.forEach(function(row, y) {
      var tileRow = {
         y: y,
         tileColumns: []
      };

      row.forEach(function(tileData, x) {
         var tile = self.createTile(x, y, tileData);

         tileRow.tileColumns.push(tile);
      });
      rows.push(tileRow);
   });
   this.vm.map.tileRows(rows);
};

MapAdapter.prototype.createTile = function(x, y, tileData) {
   var self = this;
   var tile = {
      x: x,
      y: y,
      tileType: ko.observable(tileData.properties.type),
      floorHeight: ko.observable(tileData.properties.floorHeight),
      ceilingHeight: ko.observable(tileData.properties.ceilingHeight),
      slopeHeight: ko.observable(tileData.properties.slopeHeight),

      northWallHeight: ko.observable(tileData.properties.calculatedWallHeights.north),
      eastWallHeight: ko.observable(tileData.properties.calculatedWallHeights.east),
      southWallHeight: ko.observable(tileData.properties.calculatedWallHeights.south),
      westWallHeight: ko.observable(tileData.properties.calculatedWallHeights.west),

      floorTextureIndex: ko.observable(tileData.properties.realWorld.floorTexture),
      floorTextureRotations: ko.observable(tileData.properties.realWorld.floorTextureRotations),

      ceilingTextureIndex: ko.observable(tileData.properties.realWorld.ceilingTexture),
      ceilingTextureRotations: ko.observable(tileData.properties.realWorld.ceilingTextureRotations),

      wallTextureIndex: ko.observable(tileData.properties.realWorld.wallTexture),

      isSelected: ko.observable(false)
   };

   tile.floorTextureUrl = ko.computed(this.computeTextureUrl(tile.floorTextureIndex));
   tile.ceilingTextureUrl = ko.computed(this.computeTextureUrl(tile.ceilingTextureIndex));
   tile.wallTextureUrl = ko.computed(this.computeTextureUrl(tile.wallTextureIndex));

   tile.hasWallSouthWestToNorthEast = ko.computed(function() {
      var tileType = tile.tileType();

      return tileType === "diagonalOpenNorthWest" || tileType === "diagonalOpenSouthEast";
   });
   tile.hasWallSouthEastToNorthWest = ko.computed(function() {
      var tileType = tile.tileType();

      return tileType === "diagonalOpenNorthEast" || tileType === "diagonalOpenSouthWest";
   });

   return tile;
};

module.exports = MapAdapter;
