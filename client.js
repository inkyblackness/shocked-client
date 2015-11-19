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
   selectedTextureDisplay: ko.observable("Floor")
};

vm.shouldShowFloorTexture = ko.computed(function() {
   return vm.selectedTextureDisplay() === "Floor";
});
vm.shouldShowCeilingTexture = ko.computed(function() {
   return vm.selectedTextureDisplay() === "Ceiling";
});

ko.applyBindings(vm);

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

var computeTextureUrl = function(indexObservable) {
   return function() {
      var textureIndex = indexObservable();
      var urls = vm.levelTextureUrls();
      var url = "";

      if ((textureIndex >= 0) && (textureIndex < urls.length)) {
         url = urls[textureIndex];
      }

      return "url(" + url + ")";
   }
};

var createTile = function(x, y) {
   var tile = {
      x: x,
      tileType: ko.observable("solid"),
      floorTextureIndex: ko.observable(-1),
      floorTextureRotations: ko.observable("rotations0"),
      
      ceilingTextureIndex: ko.observable(-1),
      ceilingTextureRotations: ko.observable("rotations0"),
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
   console.log("Updating on new width: " + newWidth);
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
   console.log("Updating on new height: " + newHeight);
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

vm.mapWidth(64);
vm.mapHeight(64);
