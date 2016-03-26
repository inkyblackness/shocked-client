/* global ko */
"use strict";

var unifier = require("../../util/unifier.js");

var updateTileProperties = function(tile, tileData) {
   tile.tileType(tileData.properties.type);
   tile.floorHeight(tileData.properties.floorHeight);
   tile.ceilingHeight(tileData.properties.ceilingHeight);
   tile.slopeHeight(tileData.properties.slopeHeight);
   tile.slopeControl(tileData.properties.slopeControl);

   tile.northWallHeight(tileData.properties.calculatedWallHeights.north);
   tile.eastWallHeight(tileData.properties.calculatedWallHeights.east);
   tile.southWallHeight(tileData.properties.calculatedWallHeights.south);
   tile.westWallHeight(tileData.properties.calculatedWallHeights.west);

   if (tileData.realWorld) {
      tile.floorTextureIndex(tileData.properties.realWorld.floorTexture);
      tile.floorTextureRotations(tileData.properties.realWorld.floorTextureRotations);
      tile.ceilingTextureIndex(tileData.properties.realWorld.ceilingTexture);
      tile.ceilingTextureRotations(tileData.properties.realWorld.ceilingTextureRotations);
      tile.wallTextureIndex(tileData.properties.realWorld.wallTexture);
      tile.wallTextureOffset(tileData.properties.realWorld.wallTextureOffset);
      tile.useAdjacentWallTexture(tileData.properties.realWorld.useAdjacentWallTexture);
   }
};

function MapAdapter() {
   this.projectsAdapter = null;
   this.texturesAdapter = null;

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
   var self = this;
   var rest = this.rest;
   var vmMap = {
      tileTypes: ["", "open", "solid",
         "diagonalOpenSouthEast", "diagonalOpenSouthWest", "diagonalOpenNorthWest", "diagonalOpenNorthEast",
         "slopeSouthToNorth", "slopeWestToEast", "slopeNorthToSouth", "slopeEastToWest",
         "valleySouthEastToNorthWest", "valleySouthWestToNorthEast", "valleyNorthWestToSouthEast", "valleyNorthEastToSouthWest",
         "ridgeNorthWestToSouthEast", "ridgeNorthEastToSouthWest", "ridgeSouthEastToNorthWest", "ridgeSouthWestToNorthEast"
      ],
      slopeControls: ["", "ceilingInverted", "ceilingMirrored", "ceilingFlat", "floorFlat"],
      heightValues: ["*", 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
         16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31
      ],
      depthValues: ["*", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
         16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32
      ],
      maybeValues: ["*", "no", "yes"],

      mapSection: ko.observableArray(["control", "tiles", "objects"]),
      selectedMapSection: ko.observable("control"),

      selectedLevel: ko.observable(),

      levelTextures: ko.observableArray(),
      selectedLevelTextureIndex: ko.observable(-1),
      selectedLevelTexture: ko.observable(),

      levelObjects: ko.observableArray(),

      tileRows: ko.observableArray(),

      textureDisplay: ko.observableArray(["Floor", "Ceiling"]),
      selectedTextureDisplay: ko.observable("Floor"),

      selectedTiles: ko.observableArray(),

      selectedTileType: ko.observable(""),
      selectedSlopeControl: ko.observable(""),
      selectedFloorHeight: ko.observable("*"),
      selectedCeilingHeight: ko.observable("*"),
      selectedSlopeHeight: ko.observable("*"),
      selectedTileFloorTextureIndex: ko.observable(-1),
      selectedTileCeilingTextureIndex: ko.observable(-1),
      selectedTileWallTextureIndex: ko.observable(-1),
      selectedWallTextureOffset: ko.observable("*"),
      selectedUseAdjacentWallTexture: ko.observable("*")
   };

   this.vm.map = vmMap;
   vmMap.onTileClicked = this.getTileClickedHandler();
   vmMap.changeLevelTexture = this.changeLevelTexture.bind(this);

   vmMap.isCyberspace = ko.computed(function() {
      var level = vmMap.selectedLevel();

      return !!(level && level.properties.cyberspaceFlag);
   });

   var vmTextures = this.vm.textures;
   vmMap.shouldShowFloorTexture = ko.computed(function() {
      return !vmMap.isCyberspace() && (vmMap.selectedTextureDisplay() === "Floor");
   });
   vmMap.shouldShowCeilingTexture = ko.computed(function() {
      return !vmMap.isCyberspace() && (vmMap.selectedTextureDisplay() === "Ceiling");
   });
   vmMap.selectedTileFloorTextureUrl = ko.computed(this.computeTextureUrl(vmMap.selectedTileFloorTextureIndex));
   vmMap.selectedTileCeilingTextureUrl = ko.computed(this.computeTextureUrl(vmMap.selectedTileCeilingTextureIndex));
   vmMap.selectedTileWallTextureUrl = ko.computed(this.computeTextureUrl(vmMap.selectedTileWallTextureIndex));

   var updateSelectedLevelTexture = function() {
      var textures = vmMap.levelTextures();
      var index = vmMap.selectedLevelTextureIndex();

      if ((index >= 0) && (index < textures.length)) {
         vmMap.selectedLevelTexture(textures[index]);
      } else {
         vmMap.selectedLevelTexture(null);
      }
   };

   vmMap.levelTextures.subscribe(function(newTextures) {
      updateSelectedLevelTexture();
   });
   vmMap.selectedLevelTextureIndex.subscribe(function(newIndex) {
      updateSelectedLevelTexture();
   });

   vmMap.selectedTiles.subscribe(function(newList) {
      var tileTypeUnifier = unifier.withResetValue("");
      var slopeControlUnifier = unifier.withResetValue("");
      var slopeHeightUnifier = unifier.withResetValue("*");
      var floorHeightUnifier = unifier.withResetValue("*");
      var ceilingHeightUnifier = unifier.withResetValue("*");
      var floorTextureIndexUnifier = unifier.withResetValue(-1);
      var ceilingTextureIndexUnifier = unifier.withResetValue(-1);
      var wallTextureIndexUnifier = unifier.withResetValue(-1);
      var wallTextureOffsetUnifier = unifier.withResetValue("*");
      var useAdjacentWallTextureUnifier = unifier.withResetValue("*");

      newList.forEach(function(tile) {
         tileTypeUnifier.add(tile.tileType());
         slopeControlUnifier.add(tile.slopeControl());
         slopeHeightUnifier.add(tile.slopeHeight());
         floorHeightUnifier.add(tile.floorHeight());
         ceilingHeightUnifier.add(32 - tile.ceilingHeight());
      });
      if (!vmMap.isCyberspace()) {
         newList.forEach(function(tile) {
            floorTextureIndexUnifier.add(tile.floorTextureIndex());
            ceilingTextureIndexUnifier.add(tile.ceilingTextureIndex());
            wallTextureIndexUnifier.add(tile.wallTextureIndex());
            wallTextureOffsetUnifier.add(tile.wallTextureOffset());
            useAdjacentWallTextureUnifier.add(tile.useAdjacentWallTexture() ? "yes" : "no");
         });
      }
      vmMap.selectedTileType(tileTypeUnifier.get());
      vmMap.selectedSlopeControl(slopeControlUnifier.get());
      vmMap.selectedSlopeHeight(slopeHeightUnifier.get());
      vmMap.selectedFloorHeight(floorHeightUnifier.get());
      vmMap.selectedCeilingHeight(ceilingHeightUnifier.get());
      vmMap.selectedTileFloorTextureIndex(floorTextureIndexUnifier.get());
      vmMap.selectedTileCeilingTextureIndex(ceilingTextureIndexUnifier.get());
      vmMap.selectedTileWallTextureIndex(wallTextureIndexUnifier.get());
      vmMap.selectedWallTextureOffset(wallTextureOffsetUnifier.get());
      vmMap.selectedUseAdjacentWallTexture(useAdjacentWallTextureUnifier.get());
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

   vmMap.selectedLevel.subscribe(function(level) {
      if (level) {
         rest.getResource(level.href + "/textures", function(levelTextures) {
            self.onLevelTexturesLoaded(levelTextures);
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
   var levelTextures = this.vm.map.levelTextures;

   return function() {
      var textureIndex = indexObservable();
      var textures = levelTextures();
      var url = "";

      if ((textureIndex >= 0) && (textureIndex < textures.length)) {
         url = textures[textureIndex].largeTextureUrl();
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

MapAdapter.prototype.changeLevelTexture = function() {
   var selectedIndex = this.vm.map.selectedLevelTextureIndex();
   var selectedTexture = this.vm.map.selectedLevelTexture();
   var textureIds = this.vm.map.levelTextures().map(function(texture) {
      return texture.id;
   });
   var level = this.vm.map.selectedLevel();
   var self = this;

   if (selectedTexture && (selectedIndex >= 0) && (selectedIndex < textureIds.length)) {
      textureIds[selectedIndex] = selectedTexture.id;
      this.rest.putResource(level.href + "/textures", textureIds, function(levelTextures) {
         self.onLevelTexturesLoaded(levelTextures);
      }, function() {});
   }
};

MapAdapter.prototype.onLevelTexturesLoaded = function(levelTextures) {
   var vmTextures = this.vm.textures;
   var textures = levelTextures.ids.map(function(id) {
      return vmTextures.getTexture(id);
   });
   this.vm.map.levelTextures(textures);
};

MapAdapter.prototype.onMapLoaded = function(tileMap) {
   var self = this;
   var rows = [];

   tileMap.Table.reverse();
   tileMap.Table.forEach(function(row, y) {
      var tileRow = {
         y: 63 - y,
         tileColumns: []
      };

      row.forEach(function(tileData, x) {
         var tile = self.createTile(x, 63 - y, tileData);

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
      slopeControl: ko.observable(tileData.properties.slopeControl),

      northWallHeight: ko.observable(tileData.properties.calculatedWallHeights.north),
      eastWallHeight: ko.observable(tileData.properties.calculatedWallHeights.east),
      southWallHeight: ko.observable(tileData.properties.calculatedWallHeights.south),
      westWallHeight: ko.observable(tileData.properties.calculatedWallHeights.west),

      isSelected: ko.observable(false)
   };
   if (tileData.properties.realWorld) {
      tile.floorTextureIndex = ko.observable(tileData.properties.realWorld.floorTexture);
      tile.floorTextureRotations = ko.observable(tileData.properties.realWorld.floorTextureRotations);

      tile.ceilingTextureIndex = ko.observable(tileData.properties.realWorld.ceilingTexture);
      tile.ceilingTextureRotations = ko.observable(tileData.properties.realWorld.ceilingTextureRotations);

      tile.wallTextureIndex = ko.observable(tileData.properties.realWorld.wallTexture);
      tile.wallTextureOffset = ko.observable(tileData.properties.realWorld.wallTextureOffset);
      tile.useAdjacentWallTexture = ko.observable(tileData.properties.realWorld.useAdjacentWallTexture);
   }

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
