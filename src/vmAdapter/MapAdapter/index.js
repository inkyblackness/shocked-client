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

      sizeX: ko.observable(0),
      sizeY: ko.observable(0),
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


   vmMap.sizeX.subscribe(function(newWidth) {
      vmMap.tileRows().forEach(function(tileRow) {
         self.resizeColumns(tileRow, newWidth);
      });
   });

   vmMap.sizeY.subscribe(function(newHeight) {
      var raw = vmMap.tileRows();

      if (raw.length > newHeight) {
         raw = raw.slice(0, newHeight);
      } else {
         raw = raw.slice(0, raw.length);
      }
      while (raw.length < newHeight) {
         raw.push(self.createTileRow(newHeight - raw.length - 1));
      }
      vmMap.tileRows(raw);
   });

   var vmProjects = this.vm.projects;

   vmMap.selectedLevel.subscribe(function(level) {
      if (level) {
         rest.getResource(level.href + "/textures", function(levelTextures) {
            vmMap.levelTextures.removeAll();
            vmMap.levelTextureUrls.removeAll();
            levelTextures.ids.forEach(function(id) {
               vmMap.levelTextureUrls.push(vmProjects.selected().href + "/textures/" + id + "/large/png");
               vmMap.levelTextures.push(id);
            });
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
            tileMap.Table.forEach(function(row, y) {
               row.forEach(function(tileData, x) {
                  self.sys.defer(function() {
                     var rowIndex = 64 - 1 - y;
                     var tile = vmMap.tileRows()[rowIndex].tileColumns()[x];

                     updateTileProperties(tile, tileData);
                  });
               });
            });
         }, function() {});

         vmMap.sizeX(64);
         vmMap.sizeY(64);
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

MapAdapter.prototype.createTile = function(x, y) {
   var self = this;
   var tile = {
      x: x,
      y: y,
      tileType: ko.observable("solid"),
      floorHeight: ko.observable(0),
      ceilingHeight: ko.observable(0),
      slopeHeight: ko.observable(0),

      northWallHeight: ko.observable(0.0),
      eastWallHeight: ko.observable(0.0),
      southWallHeight: ko.observable(0.0),
      westWallHeight: ko.observable(0.0),

      floorTextureIndex: ko.observable(-1),
      floorTextureRotations: ko.observable(0),

      ceilingTextureIndex: ko.observable(-1),
      ceilingTextureRotations: ko.observable(0),

      wallTextureIndex: ko.observable(-1),

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

MapAdapter.prototype.resizeColumns = function(tileRow, newWidth) {
   var list = tileRow.tileColumns;
   var raw = list();

   if (raw.length > newWidth) {
      raw = raw.slice(0, newWidth);
   } else {
      raw = raw.slice(0, raw.length);
   }
   while (raw.length < newWidth) {
      raw.push(this.createTile(raw.length, tileRow.y));
   }
   list(raw);
};

MapAdapter.prototype.createTileRow = function(y) {
   var tileRow = {
      y: y,
      tileColumns: ko.observableArray()
   };

   this.resizeColumns(tileRow, this.vm.map.sizeX());

   return tileRow;
};

module.exports = MapAdapter;
