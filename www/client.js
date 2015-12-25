(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
/* global window */
"use strict";

var defer = function(cb) {
   window.setTimeout(cb, 0);
};

module.exports = defer;

},{}],2:[function(require,module,exports){
/* global ko */
"use strict";

var rest = require("./rest.js");
var unifier = require("./unifier.js");
var defer = require("./browser/defer.js");

ko.options.deferUpdates = true;

var vm = {
   mainSections: ko.observableArray(["project", "map"]),
   selectedMainSection: ko.observable("project"),

   projects: {
      available: ko.observableArray(),
      selected: ko.observable()
   },

   levels: {
      available: ko.observableArray()
   },

   map: {
      selectedLevel: ko.observable(),

      sizeX: ko.observable(0),
      sizeY: ko.observable(0)

   },

   tileTypes: ["", "open", "solid",
      "diagonalOpenSouthEast", "diagonalOpenSouthWest", "diagonalOpenNorthWest", "diagonalOpenNorthEast",
      "slopeSouthToNorth", "slopeWestToEast", "slopeNorthToSouth", "slopeEastToWest",
      "valleySouthEastToNorthWest", "valleySouthWestToNorthEast", "valleyNorthWestToSouthEast", "valleyNorthEastToSouthWest",
      "ridgeNorthWestToSouthEast", "ridgeNorthEastToSouthWest", "ridgeSouthEastToNorthWest", "ridgeSouthWestToNorthEast"
   ],

   tileRows: ko.observableArray(),

   levelTextures: ko.observableArray(),
   levelTextureUrls: ko.observableArray(),

   textureDisplay: ko.observableArray(["Floor", "Ceiling"]),
   selectedTextureDisplay: ko.observable("Floor"),

   selectedTiles: ko.observableArray(),

   selectedTileType: ko.observable(""),
   selectedTileFloorTextureIndex: ko.observable(-1),
   selectedTileCeilingTextureIndex: ko.observable(-1)
};

vm.projects.selected.subscribe(function(project) {
   if (project) {
      rest.getResource(project.href + "/archive/levels", function(levels) {
         vm.levels.available(levels.items);
      }, function() {});
   } else {
      vm.levels.available([]);
   }
});

var computeTextureUrl = function(indexObservable) {
   return function() {
      var textureIndex = indexObservable();
      var urls = vm.levelTextureUrls();
      var url = "";

      if ((textureIndex >= 0) && (textureIndex < urls.length)) {
         url = urls[textureIndex];
      }

      return url;
   };
};

vm.shouldShowFloorTexture = ko.computed(function() {
   return vm.selectedTextureDisplay() === "Floor";
});
vm.shouldShowCeilingTexture = ko.computed(function() {
   return vm.selectedTextureDisplay() === "Ceiling";
});
vm.selectedTileFloorTextureUrl = ko.computed(computeTextureUrl(vm.selectedTileFloorTextureIndex));
vm.selectedTileCeilingTextureUrl = ko.computed(computeTextureUrl(vm.selectedTileCeilingTextureIndex));

vm.selectTile = function(tile, event) {
   var newState = !tile.isSelected();

   tile.isSelected(newState);
   if (event.ctrlKey) {
      if (newState) {
         vm.selectedTiles.push(tile);
      } else {
         vm.selectedTiles.remove(tile);
      }
   } else {
      vm.selectedTiles.removeAll().forEach(function(other) {
         other.isSelected(false);
      });
      if (newState) {
         vm.selectedTiles.push(tile);
      }
   }
};

vm.selectedTiles.subscribe(function(newList) {
   var tileTypeUnifier = unifier.withResetValue("");
   var floorTextureIndexUnifier = unifier.withResetValue(-1);
   var ceilingTextureIndexUnifier = unifier.withResetValue(-1);

   newList.forEach(function(tile) {
      tileTypeUnifier.add(tile.tileType());
      floorTextureIndexUnifier.add(tile.floorTextureIndex());
      ceilingTextureIndexUnifier.add(tile.ceilingTextureIndex());
   });
   vm.selectedTileType(tileTypeUnifier.get());
   vm.selectedTileFloorTextureIndex(floorTextureIndexUnifier.get());
   vm.selectedTileCeilingTextureIndex(ceilingTextureIndexUnifier.get());
});

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

vm.selectedTileType.subscribe(function(newType) {
   if (newType !== "") {
      vm.selectedTiles().forEach(function(tile) {
         var properties = {
            type: newType,
         };

         var tileUrl = vm.map.selectedLevel().href + "/tiles/" + tile.y + "/" + tile.x;
         if (tile.tileType() !== newType) {
            rest.putResource(tileUrl, properties, function(tileData) {
               updateTileProperties(tile, tileData);
            }, function() {});
         }
      });
   }
});

var getTile = function(x, y) {
   var tileRows = vm.tileRows();
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

var getTileType = function(x, y) {
   var tileType = "solid";
   var tile = getTile(x, y);

   if (tile !== null) {
      tileType = tile.tileType();
   }

   return tileType;
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

var createTile = function(x, y) {
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

   tile.floorTextureUrl = ko.computed(computeTextureUrl(tile.floorTextureIndex));
   tile.ceilingTextureUrl = ko.computed(computeTextureUrl(tile.ceilingTextureIndex));

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

var resizeColumns = function(tileRow, newWidth) {
   var list = tileRow.tileColumns;

   while (list().length > newWidth) {
      list.pop();
   }
   while (list().length < newWidth) {
      list().push(createTile(list().length, tileRow.y));
   }
};

vm.map.sizeX.subscribe(function(newWidth) {
   vm.tileRows().forEach(function(tileRow) {
      resizeColumns(tileRow, newWidth);
   });
});

var createTileRow = function(y) {
   var tileRow = {
      y: y,
      tileColumns: ko.observableArray()
   };

   resizeColumns(tileRow, vm.map.sizeX());

   return tileRow;
};

vm.map.sizeY.subscribe(function(newHeight) {
   while (vm.tileRows().length > newHeight) {
      vm.tileRows.pop();
   }
   while (vm.tileRows().length < newHeight) {
      vm.tileRows.push(createTileRow(newHeight - vm.tileRows().length - 1));
   }
});

vm.map.selectedLevel.subscribe(function(level) {
   if (level) {
      rest.getResource(level.href + "/textures", function(levelTextures) {
         vm.levelTextures.removeAll();
         vm.levelTextureUrls.removeAll();
         levelTextures.ids.forEach(function(id) {
            vm.levelTextureUrls.push(vm.projects.selected().href + "/textures/" + id + "/large/png");
            vm.levelTextures.push(id);
         });
      }, function() {});

      rest.getResource(level.href + "/tiles", function(tileMap) {
         tileMap.Table.forEach(function(row, y) {
            row.forEach(function(tileData, x) {
               defer(function() {
                  var rowIndex = 64 - 1 - y;
                  var tile = vm.tileRows()[rowIndex].tileColumns()[x];

                  updateTileProperties(tile, tileData);
               });
            });
         });
      }, function() {});

      vm.map.sizeX(64);
      vm.map.sizeY(64);
   }
});

ko.applyBindings(vm);

rest.getResource("/projects", function(projects) {
   vm.projects.available(projects.items);
}, function() {});

},{"./browser/defer.js":1,"./rest.js":3,"./unifier.js":4}],3:[function(require,module,exports){
/* global $ */
"use strict";

var rest = {};

rest.getResource = function(url, onSuccess, onFailure) {
   var options = {
      method: "GET",
      url: url,
      dataType: "json",
      jsonp: false,
      success: onSuccess,
      error: onFailure
   };

   $.ajax(options);
};

rest.putResource = function(url, data, onSuccess, onFailure) {
   var options = {
      method: "PUT",
      url: url,
      dataType: "json",
      contentType: "application/json",
      data: JSON.stringify(data),
      jsonp: false,
      processData: false,
      success: onSuccess,
      error: onFailure
   };

   $.ajax(options);
};

module.exports = rest;

},{}],4:[function(require,module,exports){
/* global $ */
"use strict";

var unifier = {};

unifier.withResetValue = function(resetValue) {
   var first = true;
   var resultValue = resetValue;
   var stateObj = {
      add: function(singleValue) {
         if (first) {
            resultValue = singleValue;
            first = false;
         } else if (resultValue !== singleValue) {
            resultValue = resetValue;
         }
      },
      get: function() {
         return resultValue;
      }
   };

   return stateObj;
};

module.exports = unifier;

},{}]},{},[2]);
