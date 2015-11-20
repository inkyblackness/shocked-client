ko.options.deferUpdates = true;

var getResource = function(url, onSuccess, onFailure) {
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

var vm = {
   mapWidth: ko.observable(0),
   mapHeight: ko.observable(0),
   tileRows: ko.observableArray(),
   levels: ko.observableArray(),
   selectedLevel: ko.observable(-1),
   levelTextures: ko.observableArray(),
   levelTextureUrls: ko.observableArray(),

   textureDisplay: ko.observableArray(["Floor", "Ceiling"]),
   selectedTextureDisplay: ko.observable("Floor"),

   selectedTiles: ko.observableArray(),

   selectedTileFloorTextureIndex: ko.observable(-1),
   selectedTileCeilingTextureIndex: ko.observable(-1)
};

var computeTextureUrl = function(indexObservable) {
   return function() {
      var textureIndex = indexObservable();
      var urls = vm.levelTextureUrls();
      var url = "";

      if ((textureIndex >= 0) && (textureIndex < urls.length)) {
         url = urls[textureIndex];
      }

      return url;
   }
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

var unifier = function(resetValue) {
   var first = true;
   var unique = true;
   var resultValue = resetValue;
   var stateObj = {
      add: function(singleValue) {
         if (first) {
            resultValue = singleValue;
            first = false;
         } else if (resultValue !== singleValue) {
            unique = false;
            resultValue = resetValue;
         }
      },
      get: function() {
         return resultValue;
      }
   };

   return stateObj;
};

vm.selectedTiles.subscribe(function(newList) {
   var floorTextureIndexUnifier = unifier(-1);
   var ceilingTextureIndexUnifier = unifier(-1);

   newList.forEach(function(tile) {
      floorTextureIndexUnifier.add(tile.floorTextureIndex());
      ceilingTextureIndexUnifier.add(tile.ceilingTextureIndex());
   });
   vm.selectedTileFloorTextureIndex(floorTextureIndexUnifier.get());
   vm.selectedTileCeilingTextureIndex(ceilingTextureIndexUnifier.get());
});

var getTile = function(x, y) {
   var tileRows = vm.tileRows();
   var rowIndex = 64 - 1 - y;
   var tileColumns;
   var tile = null;

   if ((rowIndex >= 0) && (rowIndex < tileRows.length)) {
      tileColumns = tileRows[rowIndex].tileColumns()
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
}

var isTileOpenNorth = function(tileType) {
   return tileType !== "solid" && tileType !== "diagonalOpenSouthEast" && tileType !== "diagonalOpenSouthWest";
}

var isTileOpenEast = function(tileType) {
   return tileType !== "solid" && tileType !== "diagonalOpenSouthWest" && tileType !== "diagonalOpenNorthWest";
}

var isTileOpenWest = function(tileType) {
   return tileType !== "solid" && tileType !== "diagonalOpenSouthEast" && tileType !== "diagonalOpenNorthEast";
}

var createTile = function(x, y) {
   var tile = {
      x: x,
      tileType: ko.observable("solid"),
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
}

var resizeColumns = function(tileRow, newWidth) {
   var list = tileRow.tileColumns;

   while (list().length > newWidth) {
      list.pop();
   }
   while (list().length < newWidth) {
      list().push(createTile(list().length, tileRow.y));
   }
};

vm.mapWidth.subscribe(function(newWidth) {
   vm.tileRows().forEach(function(tileRow) {
      resizeColumns(tileRow, newWidth);
   });
});

var createTileRow = function(y) {
   var tileRow = {
      y: y,
      tileColumns: ko.observableArray()
   };

   resizeColumns(tileRow, vm.mapWidth());

   return tileRow;
};

vm.mapHeight.subscribe(function(newHeight) {
   while (vm.tileRows().length > newHeight) {
      vm.tileRows.pop();
   }
   while (vm.tileRows().length < newHeight) {
      vm.tileRows.push(createTileRow(newHeight - vm.tileRows().length - 1));
   }
});

var loadLevel = function(levelId) {
   getResource("/projects/test1/archive/level/" + levelId + "/textures", function(levelTextures) {
      vm.levelTextures.removeAll();
      vm.levelTextureUrls.removeAll();
      levelTextures.ids.forEach(function(id) {
         vm.levelTextureUrls.push("/projects/test1/textures/" + id + "/large/png");
         vm.levelTextures.push(id);
      });
   }, function() {});

   getResource("/projects/test1/archive/level/" + levelId + "/tiles", function(tileMap) {
      tileMap.Table.forEach(function(row, y) {
         row.forEach(function(tileData, x) {
            setTimeout(function() {
               var rowIndex = 64 - 1 - y;
               var tile = vm.tileRows()[rowIndex].tileColumns()[x];

               tile.tileType(tileData.properties.type);
               tile.floorTextureIndex(tileData.properties.realWorld.floorTexture);
               tile.floorTextureRotations("rotations" + tileData.properties.realWorld.floorTextureRotations);
               tile.ceilingTextureIndex(tileData.properties.realWorld.ceilingTexture);
               tile.ceilingTextureRotations("rotations" + tileData.properties.realWorld.ceilingTextureRotations);
            }, 0);
         });
      });
   }, function() {});
};

var selectLevel = function(levelId) {
   return function() {
      if (vm.selectedLevel() !== levelId) {
         vm.selectedLevel(levelId);
         loadLevel(levelId)
      }
   };
};

var listLevel = function(levelId) {
   var level = {
      id: levelId,
      isSelected: ko.computed(function() {
         return vm.selectedLevel() === levelId;
      }),
      select: selectLevel(levelId)
   };

   vm.levels.push(level);
}

for (var levelId = 0; levelId < 16; levelId++) {
   listLevel(levelId);
}

ko.applyBindings(vm);

vm.mapWidth(64);
vm.mapHeight(64);
