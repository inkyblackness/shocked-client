(function(f){if(typeof exports==="object"&&typeof module!=="undefined"){module.exports=f()}else if(typeof define==="function"&&define.amd){define([],f)}else{var g;if(typeof window!=="undefined"){g=window}else if(typeof global!=="undefined"){g=global}else if(typeof self!=="undefined"){g=self}else{g=this}g.client = f()}})(function(){var define,module,exports;return (function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
/*
Copyright (c) | 2016 | infuse.js | Romuald Quantin | www.soundstep.com | romu@soundstep.com

Permission is hereby granted, free of charge, to any person obtaining a copy of this software
and associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial
portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

(function(infuse) {

    'use strict';

    infuse.version = '1.0.0';

    // regex from angular JS (https://github.com/angular/angular.js)
    var FN_ARGS = /^function\s*[^\(]*\(\s*([^\)]*)\)/m;
    var FN_ARG_SPLIT = /,/;
    var FN_ARG = /^\s*(_?)(\S+?)\1\s*$/;
    var STRIP_COMMENTS = /((\/\/.*$)|(\/\*[\s\S]*?\*\/))/mg;

    function contains(arr, value) {
        var i = arr.length;
        while (i--) {
            if (arr[i] === value) {
                return true;
            }
        }
        return false;
    }

    infuse.errors = {
        MAPPING_BAD_PROP: '[Error infuse.Injector.mapClass/mapValue] the first parameter is invalid, a string is expected',
        MAPPING_BAD_VALUE: '[Error infuse.Injector.mapClass/mapValue] the second parameter is invalid, it can\'t null or undefined, with property: ',
        MAPPING_BAD_CLASS: '[Error infuse.Injector.mapClass/mapValue] the second parameter is invalid, a function is expected, with property: ',
        MAPPING_BAD_SINGLETON: '[Error infuse.Injector.mapClass] the third parameter is invalid, a boolean is expected, with property: ',
        MAPPING_ALREADY_EXISTS: '[Error infuse.Injector.mapClass/mapValue] this mapping already exists, with property: ',
        CREATE_INSTANCE_INVALID_PARAM: '[Error infuse.Injector.createInstance] invalid parameter, a function is expected',
        NO_MAPPING_FOUND: '[Error infuse.Injector.getInstance] no mapping found',
        INJECT_INSTANCE_IN_ITSELF_PROPERTY: '[Error infuse.Injector.getInjectedValue] A matching property has been found in the target, you can\'t inject an instance in itself',
        INJECT_INSTANCE_IN_ITSELF_CONSTRUCTOR: '[Error infuse.Injector.getInjectedValue] A matching constructor parameter has been found in the target, you can\'t inject an instance in itself',
        DEPENDENCIES_MISSING_IN_STRICT_MODE: '[Error infuse.Injector.getDependencies] An "inject" property (array) that describes the dependencies is missing in strict mode.'
    };

    var MappingVO = function(prop, value, cl, singleton) {
        this.prop = prop;
        this.value = value;
        this.cl = cl;
        this.singleton = singleton || false;
    };

    var validateProp = function(prop) {
        if (typeof prop !== 'string') {
            throw new Error(infuse.errors.MAPPING_BAD_PROP);
        }
    };

    var validateValue = function(prop, val) {
        if (val === undefined || val === null) {
            throw new Error(infuse.errors.MAPPING_BAD_VALUE + prop);
        }
    };

    var validateClass = function(prop, val) {
        if (typeof val !== 'function') {
            throw new Error(infuse.errors.MAPPING_BAD_CLASS + prop);
        }
    };

    var validateBooleanSingleton = function(prop, singleton) {
        if (typeof singleton !== 'boolean') {
            throw new Error(infuse.errors.MAPPING_BAD_SINGLETON + prop);
        }
    };

    var validateConstructorInjectionLoop = function(name, cl) {
        var params = infuse.getDependencies(cl);
        if (contains(params, name)) {
            throw new Error(infuse.errors.INJECT_INSTANCE_IN_ITSELF_CONSTRUCTOR);
        }
    };

    var validatePropertyInjectionLoop = function(name, target) {
        if (target.hasOwnProperty(name)) {
            throw new Error(infuse.errors.INJECT_INSTANCE_IN_ITSELF_PROPERTY);
        }
    };

    infuse.Injector = function() {
        this.mappings = {};
        this.parent = null;
        this.strictMode = false;
    };

    infuse.getDependencies = function(cl) {
        var args = [];
        var deps;

        function extractName(all, underscore, name) {
            args.push(name);
        }

        if (cl.hasOwnProperty('inject') && Object.prototype.toString.call(cl.inject) === '[object Array]' && cl.inject.length > 0) {
            deps = cl.inject;
        }

        var clStr = cl.toString().replace(STRIP_COMMENTS, '');
        var argsFlat = clStr.match(FN_ARGS);
        var spl = argsFlat[1].split(FN_ARG_SPLIT);

        for (var i=0, l=spl.length; i<l; i++) {
            // Only override arg with non-falsey deps value at same key
            var arg = (deps && deps[i]) ? deps[i] : spl[i];
            arg.replace(FN_ARG, extractName);
        }

        return args;
    };

    infuse.Injector.prototype = {

        createChild: function() {
            var injector = new infuse.Injector();
            injector.parent = this;
            injector.strictMode = this.strictMode;
            return injector;
        },

        getMappingVo: function(prop) {
            if (!this.mappings) {
                return null;
            }
            if (this.mappings[prop]) {
                return this.mappings[prop];
            }
            if (this.parent) {
                return this.parent.getMappingVo(prop);
            }
            return null;
        },

        mapValue: function(prop, val) {
            if (this.mappings[prop]) {
                throw new Error(infuse.errors.MAPPING_ALREADY_EXISTS + prop);
            }
            validateProp(prop);
            validateValue(prop, val);
            this.mappings[prop] = new MappingVO(prop, val, undefined, undefined);
            return this;
        },

        mapClass: function(prop, cl, singleton) {
            if (this.mappings[prop]) {
                throw new Error(infuse.errors.MAPPING_ALREADY_EXISTS + prop);
            }
            validateProp(prop);
            validateClass(prop, cl);
            if (singleton) {
                validateBooleanSingleton(prop, singleton);
            }
            this.mappings[prop] = new MappingVO(prop, null, cl, singleton);
            return this;
        },

        removeMapping: function(prop) {
            this.mappings[prop] = null;
            delete this.mappings[prop];
            return this;
        },

        hasMapping: function(prop) {
            return !!this.mappings[prop];
        },

        hasInheritedMapping: function(prop) {
            return !!this.getMappingVo(prop);
        },

        getMapping: function(value) {
            for (var name in this.mappings) {
                if (this.mappings.hasOwnProperty(name)) {
                    var vo = this.mappings[name];
                    if (vo.value === value || vo.cl === value) {
                        return vo.prop;
                    }
                }
            }
            return undefined;
        },

        getValue: function(prop) {
            var vo = this.mappings[prop];
            if (!vo) {
                if (this.parent) {
                    vo = this.parent.getMappingVo.apply(this.parent, arguments);
                }
                else {
                    throw new Error(infuse.errors.NO_MAPPING_FOUND);
                }
            }
            if (vo.cl) {
                var args = Array.prototype.slice.call(arguments);
                args[0] = vo.cl;
                if (vo.singleton) {
                    if (!vo.value) {
                        vo.value = this.createInstance.apply(this, args);
                    }
                    return vo.value;
                }
                else {
                    return this.createInstance.apply(this, args);
                }
            }
            return vo.value;
        },

        getClass: function(prop) {
            var vo = this.mappings[prop];
            if (!vo) {
                if (this.parent) {
                    vo = this.parent.getMappingVo.apply(this.parent, arguments);
                }
                else {
                    return undefined;
                }
            }
            if (vo.cl) {
                return vo.cl;
            }
            return undefined;
        },

        instantiate: function(TargetClass) {
            if (typeof TargetClass !== 'function') {
                throw new Error(infuse.errors.CREATE_INSTANCE_INVALID_PARAM);
            }
            if (this.strictMode && !TargetClass.hasOwnProperty('inject')) {
                throw new Error(infuse.errors.DEPENDENCIES_MISSING_IN_STRICT_MODE);
            }
            var args = [null];
            var params = infuse.getDependencies(TargetClass);
            for (var i=0, l=params.length; i<l; i++) {
                if (arguments.length > i+1 && arguments[i+1] !== undefined && arguments[i+1] !== null) {
                    // argument found
                    args.push(arguments[i+1]);
                }
                else {
                    var name = params[i];
                    // no argument found
                    var vo = this.getMappingVo(name);
                    if (!!vo) {
                        // found mapping
                        var val = this.getInjectedValue(vo, name);
                        args.push(val);
                    }
                    else {
                        // no mapping found
                        args.push(undefined);
                    }
                }
            }
            return new (Function.prototype.bind.apply(TargetClass, args))();
        },

        inject: function (target, isParent) {
            if (this.parent) {
                this.parent.inject(target, true);
            }
            for (var name in this.mappings) {
                if (this.mappings.hasOwnProperty(name)) {
                    var vo = this.getMappingVo(name);
                    if (target.hasOwnProperty(vo.prop) || (target.constructor && target.constructor.prototype && target.constructor.prototype.hasOwnProperty(vo.prop)) ) {
                        target[name] = this.getInjectedValue(vo, name);
                    }
                }
            }
            if (typeof target.postConstruct === 'function' && !isParent) {
                target.postConstruct();
            }
            return this;
        },

        getInjectedValue: function(vo, name) {
            var val = vo.value;
            var injectee;
            if (vo.cl) {
                if (vo.singleton) {
                    if (!vo.value) {
                        validateConstructorInjectionLoop(name, vo.cl);
                        vo.value = this.instantiate(vo.cl);
                        injectee = vo.value;
                    }
                    val = vo.value;
                }
                else {
                    validateConstructorInjectionLoop(name, vo.cl);
                    val = this.instantiate(vo.cl);
                    injectee = val;
                }
            }
            if (injectee) {
                validatePropertyInjectionLoop(name, injectee);
                this.inject(injectee);
            }
            return val;
        },

        createInstance: function() {
            var instance = this.instantiate.apply(this, arguments);
            this.inject(instance);
            return instance;
        },

        getValueFromClass: function(cl) {
            for (var name in this.mappings) {
                if (this.mappings.hasOwnProperty(name)) {
                    var vo = this.mappings[name];
                    if (vo.cl === cl) {
                        if (vo.singleton) {
                            if (!vo.value) {
                                vo.value = this.createInstance.apply(this, arguments);
                            }
                            return vo.value;
                        }
                        else {
                            return this.createInstance.apply(this, arguments);
                        }
                    }
                }
            }
            if (this.parent) {
                return this.parent.getValueFromClass.apply(this.parent, arguments);
            } else {
                throw new Error(infuse.errors.NO_MAPPING_FOUND);
            }
        },

        dispose: function() {
            this.mappings = {};
        }

    };

    if (!Function.prototype.bind) {
        Function.prototype.bind = function bind(that) {
            var target = this;
            if (typeof target !== 'function') {
                throw new Error('Error, you must bind a function.');
            }
            var args = Array.prototype.slice.call(arguments, 1); // for normal call
            var bound = function () {
                if (this instanceof bound) {
                    var F = function(){};
                    F.prototype = target.prototype;
                    var self = new F();
                    var result = target.apply(
                        self,
                        args.concat(Array.prototype.slice.call(arguments))
                    );
                    if (Object(result) === result) {
                        return result;
                    }
                    return self;
                } else {
                    return target.apply(
                        that,
                        args.concat(Array.prototype.slice.call(arguments))
                    );
                }
            };
            return bound;
        };
    }

    // register for AMD module
    if (typeof define === 'function' && typeof define.amd !== 'undefined') {
        define("infuse", infuse);
    }

    // export for node.js
    if (typeof module !== 'undefined' && typeof module.exports !== 'undefined') {
        module.exports = infuse;
    }
    if (typeof exports !== 'undefined') {
        exports = infuse;
    }

})(this['infuse'] = this['infuse'] || {});

},{}],2:[function(require,module,exports){
/* global ko */
"use strict";

ko.options.deferUpdates = true;

function ViewModel() {}

ViewModel.prototype.postConstruct = function() {
   this.mainSections = ko.observableArray(["project", "map", "textures"]);
   this.selectedMainSection = ko.observable("project");

   this.languages = ["ENG", "FRA", "GER"];
};

module.exports = ViewModel;

},{}],3:[function(require,module,exports){
/* global window */
"use strict";

var defer = function(cb) {
   window.setTimeout(cb, 0);
};

module.exports = defer;

},{}],4:[function(require,module,exports){
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

rest.postResource = function(url, data, onSuccess, onFailure) {
   var options = {
      method: "POST",
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

},{}],5:[function(require,module,exports){
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

},{"./ViewModel.js":2,"./browser/defer.js":3,"./browser/rest.js":4,"./vmAdapter/LevelsAdapter":7,"./vmAdapter/MapAdapter":8,"./vmAdapter/ProjectsAdapter":9,"./vmAdapter/TexturesAdapter":10,"infuse.js":1}],6:[function(require,module,exports){
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

},{}],7:[function(require,module,exports){
/* global ko */
"use strict";

function LevelsAdapter() {
   this.projectsAdapter = null;

   this.vm = null;
   this.rest = null;
}

LevelsAdapter.prototype.postConstruct = function() {
   var rest = this.rest;
   var vmLevels = {
      available: ko.observableArray()
   };

   this.vm.levels = vmLevels;
   this.vm.projects.selected.subscribe(function(project) {
      if (project) {
         rest.getResource(project.href + "/archive/levels", function(levels) {
            vmLevels.available(levels.list);
         }, function() {});
      } else {
         vmLevels.available([]);
      }
   });
};

module.exports = LevelsAdapter;

},{}],8:[function(require,module,exports){
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

   if (tileData.properties.realWorld) {
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
      rotations: ["*", 0, 1, 2, 3],

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
      selectedTileFloorTextureRotations: ko.observable("*"),
      selectedTileCeilingTextureIndex: ko.observable(-1),
      selectedTileCeilingTextureRotations: ko.observable("*"),
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
      var floorTextureRotationsUnifier = unifier.withResetValue("*");
      var ceilingTextureIndexUnifier = unifier.withResetValue(-1);
      var ceilingTextureRotationsUnifier = unifier.withResetValue("*");
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
            floorTextureRotationsUnifier.add(tile.floorTextureRotations());
            ceilingTextureIndexUnifier.add(tile.ceilingTextureIndex());
            ceilingTextureRotationsUnifier.add(tile.ceilingTextureRotations());
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
      vmMap.selectedTileFloorTextureRotations(floorTextureRotationsUnifier.get());
      vmMap.selectedTileCeilingTextureIndex(ceilingTextureIndexUnifier.get());
      vmMap.selectedTileCeilingTextureRotations(ceilingTextureRotationsUnifier.get());
      vmMap.selectedTileWallTextureIndex(wallTextureIndexUnifier.get());
      vmMap.selectedWallTextureOffset(wallTextureOffsetUnifier.get());
      vmMap.selectedUseAdjacentWallTexture(useAdjacentWallTextureUnifier.get());
   });

   var changeBaseProperty = function(dataPropertyName, anyValue, tilePropertyName, updateAdjacent) {
      return function(newValue) {
         var properties = {};
         properties[dataPropertyName] = newValue;

         if (newValue !== anyValue) {
            self.changeSelectedTileProperties(properties, updateAdjacent, function(tile) {
               return tile[tilePropertyName]() !== newValue;
            });
         }
      };
   };

   vmMap.selectedTileType.subscribe(changeBaseProperty("type", "", "tileType", true));
   vmMap.selectedSlopeControl.subscribe(changeBaseProperty("slopeControl", "", "slopeControl", true));
   vmMap.selectedSlopeHeight.subscribe(changeBaseProperty("slopeHeight", "*", "slopeHeight", true));
   vmMap.selectedFloorHeight.subscribe(changeBaseProperty("floorHeight", "*", "floorHeight", true));
   vmMap.selectedCeilingHeight.subscribe(function(newValue) {
      var properties = {
         ceilingHeight: 32 - newValue
      };

      if (newValue !== "*") {
         self.changeSelectedTileProperties(properties, true, function(tile) {
            return tile.ceilingHeight() !== properties.ceilingHeight;
         });
      }
   });

   var changeRealWorldProperty = function(dataPropertyName, anyValue, tilePropertyName) {
      return function(newValue) {
         var properties = {
            realWorld: {}
         };
         properties.realWorld[dataPropertyName] = newValue;

         if (newValue !== anyValue) {
            self.changeSelectedTileProperties(properties, false, function(tile) {
               return tile[tilePropertyName]() !== newValue;
            });
         }
      };
   };

   vmMap.selectedTileFloorTextureIndex.subscribe(changeRealWorldProperty("floorTexture", -1, "floorTextureIndex"));
   vmMap.selectedTileFloorTextureRotations.subscribe(changeRealWorldProperty("floorTextureRotations", "*", "floorTextureRotations"));
   vmMap.selectedTileWallTextureIndex.subscribe(changeRealWorldProperty("wallTexture", -1, "wallTextureIndex"));
   vmMap.selectedTileCeilingTextureIndex.subscribe(changeRealWorldProperty("ceilingTexture", -1, "ceilingTextureIndex"));
   vmMap.selectedTileCeilingTextureRotations.subscribe(changeRealWorldProperty("ceilingTextureRotations", "*", "ceilingTextureRotations"));
   vmMap.selectedWallTextureOffset.subscribe(changeRealWorldProperty("wallTextureOffset", "*", "wallTextureOffset"));
   vmMap.selectedUseAdjacentWallTexture.subscribe(changeRealWorldProperty("useAdjacentWallTexture", "", "useAdjacentWallTexture"));

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

MapAdapter.prototype.changeSelectedTileProperties = function(properties, updateAdjacent, predicate) {
   var vmMap = this.vm.map;
   var level = vmMap.selectedLevel();
   var rest = this.rest;
   var self = this;

   vmMap.selectedTiles().forEach(function(tile) {
      var tileUrl = level.href + "/tiles/" + tile.y + "/" + tile.x;

      if (predicate(tile)) {
         rest.putResource(tileUrl, properties, function(tileData) {
            updateTileProperties(tile, tileData);
            if (updateAdjacent) {
               self.requestAdjacentTileProperties(tile.x, tile.y);
            }
         }, function() {});
      }
   });
};

MapAdapter.prototype.requestAdjacentTileProperties = function(x, y) {
   this.requestTilePropertiesAt(x - 1, y);
   this.requestTilePropertiesAt(x + 1, y);
   this.requestTilePropertiesAt(x, y - 1);
   this.requestTilePropertiesAt(x, y + 1);
};

MapAdapter.prototype.requestTilePropertiesAt = function(x, y) {
   var vmMap = this.vm.map;
   var level = vmMap.selectedLevel();
   var tileUrl = level.href + "/tiles/" + y + "/" + x;
   var rest = this.rest;
   var self = this;

   if ((x >= 0) && (x < 64) && (y >= 0) && (y < 64)) {
      rest.getResource(tileUrl, function(tileData) {
         var tileRows = vmMap.tileRows();
         var tileRow = tileRows[63 - y];
         var tile = tileRow.tileColumns[x];

         updateTileProperties(tile, tileData);
      }, function() {});
   }
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

},{"../../util/unifier.js":6}],9:[function(require,module,exports){
/* global ko */
"use strict";

function ProjectsAdapter() {
   this.vm = null;
   this.rest = null;
}

ProjectsAdapter.prototype.postConstruct = function() {
   var self = this;
   var vmProjects = {
      available: ko.observableArray(),
      selected: ko.observable(),

      newProjectId: ko.observable("")
   };

   vmProjects.canCreateNewProject = ko.computed(function() {
      var newProjectId = vmProjects.newProjectId();
      var available = vmProjects.available();
      var existing = false;

      available.forEach(function(project) {
         if (project.id === newProjectId) {
            existing = true;
         }
      });

      return (newProjectId !== "") && !existing;
   });

   vmProjects.createProject = function() {
      var projectTemplate = {
         id: vmProjects.newProjectId()
      };

      self.rest.postResource("/projects", projectTemplate, function(projectData) {
         vmProjects.available.push(projectData);
      }, function() {});
   };

   this.vm.projects = vmProjects;
   this.rest.getResource("/projects", function(projects) {
      self.onProjectsLoaded(projects);
   }, function() {});
};

ProjectsAdapter.prototype.onProjectsLoaded = function(projects) {
   var vmProjects = this.vm.projects;

   vmProjects.available(projects.items);
};

module.exports = ProjectsAdapter;

},{}],10:[function(require,module,exports){
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
   entry.title = ko.computed(function() {
      var texts = entry.texts();
      var result = entry.id + ": ";

      if (texts.length > 0) {
         result += "\"" + texts[0].name() + "\"";
      } else {
         result += "???";
      }

      return result;
   });

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

},{}]},{},[5])(5)
});