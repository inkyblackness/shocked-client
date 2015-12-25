/* global ko */
"use strict";

var unifier = require("../../util/unifier.js");

var updateTileProperties = function(tile, tileData) {
   tile.tileType(tileData.properties.type);
   tile.floorHeight(tileData.properties.floorHeight);
   tile.ceilingHeight(tileData.properties.ceilingHeight);
   tile.slopeHeight(tileData.properties.slopeHeight);

   tile.floorTextureIndex(tileData.properties.realWorld.floorTexture);
   tile.floorTextureRotations("rotations" + tileData.properties.realWorld.floorTextureRotations);
   tile.ceilingTextureIndex(tileData.properties.realWorld.ceilingTexture);
   tile.ceilingTextureRotations("rotations" + tileData.properties.realWorld.ceilingTextureRotations);
};

var isTileOpenSouth = function(tileType) {
   return tileType !== "solid" && tileType !== "diagonalOpenNorthEast" && tileType !== "diagonalOpenNorthWest";
};

var isTileOpenNorth = function(tileType) {
   return tileType !== "solid" && tileType !== "diagonalOpenSouthEast" && tileType !== "diagonalOpenSouthWest";
};

var isTileOpenEast = function(tileType) {
   return tileType !== "solid" && tileType !== "diagonalOpenSouthWest" && tileType !== "diagonalOpenNorthWest";
};

var isTileOpenWest = function(tileType) {
   return tileType !== "solid" && tileType !== "diagonalOpenSouthEast" && tileType !== "diagonalOpenNorthEast";
};

function MapAdapter() {
   this.projectsAdapter = null;

   this.vm = null;
   this.rest = null;
   this.sys = null;
}

MapAdapter.prototype.postConstruct = function() {
   var rest = this.rest;
   var vmMap = {
      selectedLevel: ko.observable(),

      levelTextures: ko.observableArray(),
      levelTextureUrls: ko.observableArray(),

      sizeX: ko.observable(0),
      sizeY: ko.observable(0),
      tileRows: ko.observableArray(),

      textureDisplay: ko.observableArray(["Floor", "Ceiling"]),
      selectedTextureDisplay: ko.observable("Floor"),

      selectedTiles: ko.observableArray(),

      selectedTileType: ko.observable(""),
      selectedTileFloorTextureIndex: ko.observable(-1),
      selectedTileCeilingTextureIndex: ko.observable(-1)
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

   vmMap.selectedTiles.subscribe(function(newList) {
      var tileTypeUnifier = unifier.withResetValue("");
      var floorTextureIndexUnifier = unifier.withResetValue(-1);
      var ceilingTextureIndexUnifier = unifier.withResetValue(-1);

      newList.forEach(function(tile) {
         tileTypeUnifier.add(tile.tileType());
         floorTextureIndexUnifier.add(tile.floorTextureIndex());
         ceilingTextureIndexUnifier.add(tile.ceilingTextureIndex());
      });
      vmMap.selectedTileType(tileTypeUnifier.get());
      vmMap.selectedTileFloorTextureIndex(floorTextureIndexUnifier.get());
      vmMap.selectedTileCeilingTextureIndex(ceilingTextureIndexUnifier.get());
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

   var self = this;

   vmMap.sizeX.subscribe(function(newWidth) {
      vmMap.tileRows().forEach(function(tileRow) {
         self.resizeColumns(tileRow, newWidth);
      });
   });

   vmMap.sizeY.subscribe(function(newHeight) {
      while (vmMap.tileRows().length > newHeight) {
         vmMap.tileRows.pop();
      }
      while (vmMap.tileRows().length < newHeight) {
         vmMap.tileRows.push(self.createTileRow(newHeight - vmMap.tileRows().length - 1));
      }
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

MapAdapter.prototype.getTile = function(x, y) {
   var tileRows = this.vm.map.tileRows();
   var rowIndex = 64 - 1 - y;
   var tileColumns;
   var tile = null;

   if ((rowIndex >= 0) && (rowIndex < tileRows.length)) {
      tileColumns = tileRows[rowIndex].tileColumns();
      if ((x >= 0) && (x < tileColumns.length)) {
         tile = tileColumns[x];
      }
   }

   return tile;
};

MapAdapter.prototype.getTileType = function(x, y) {
   var tileType = "solid";
   var tile = this.getTile(x, y);

   if (tile !== null) {
      tileType = tile.tileType();
   }

   return tileType;
};

MapAdapter.prototype.createTile = function(x, y) {
   var self = this;
   var getTileType = function(x, y) {
      return self.getTileType(x, y);
   };
   var tile = {
      x: x,
      y: y,
      tileType: ko.observable("solid"),
      floorHeight: ko.observable(0),
      ceilingHeight: ko.observable(0),
      slopeHeight: ko.observable(0),

      floorTextureIndex: ko.observable(-1),
      floorTextureRotations: ko.observable("rotations0"),

      ceilingTextureIndex: ko.observable(-1),
      ceilingTextureRotations: ko.observable("rotations0"),

      isSelected: ko.observable(false)
   };

   tile.floorTextureUrl = ko.computed(this.computeTextureUrl(tile.floorTextureIndex));
   tile.ceilingTextureUrl = ko.computed(this.computeTextureUrl(tile.ceilingTextureIndex));

   tile.hasWallSouthWestToNorthEast = ko.computed(function() {
      var tileType = tile.tileType();

      return tileType === "diagonalOpenNorthWest" || tileType === "diagonalOpenSouthEast";
   });
   tile.hasWallSouthEastToNorthWest = ko.computed(function() {
      var tileType = tile.tileType();

      return tileType === "diagonalOpenNorthEast" || tileType === "diagonalOpenSouthWest";
   });
   tile.hasWallNorth = ko.computed(function() {
      return isTileOpenNorth(tile.tileType()) && !isTileOpenSouth(getTileType(x, y + 1));
   });
   tile.hasWallSouth = ko.computed(function() {
      return isTileOpenSouth(tile.tileType()) && !isTileOpenNorth(getTileType(x, y - 1));
   });
   tile.hasWallEast = ko.computed(function() {
      return isTileOpenEast(tile.tileType()) && !isTileOpenWest(getTileType(x + 1, y));
   });
   tile.hasWallWest = ko.computed(function() {
      return isTileOpenWest(tile.tileType()) && !isTileOpenEast(getTileType(x - 1, y));
   });

   return tile;
};

MapAdapter.prototype.resizeColumns = function(tileRow, newWidth) {
   var list = tileRow.tileColumns;

   while (list().length > newWidth) {
      list.pop();
   }
   while (list().length < newWidth) {
      list().push(this.createTile(list().length, tileRow.y));
   }
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
