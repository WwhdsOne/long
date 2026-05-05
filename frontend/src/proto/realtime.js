/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import $protobuf from "protobufjs/minimal.js";

// Common aliases
const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
const $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

export const realtime = $root.realtime = (() => {

    /**
     * Namespace realtime.
     * @exports realtime
     * @namespace
     */
    const realtime = {};

    realtime.ClickRequest = (function() {

        /**
         * Properties of a ClickRequest.
         * @memberof realtime
         * @interface IClickRequest
         * @property {string|null} [slug] ClickRequest slug
         * @property {number|Long|null} [comboCount] ClickRequest comboCount
         */

        /**
         * Constructs a new ClickRequest.
         * @memberof realtime
         * @classdesc Represents a ClickRequest.
         * @implements IClickRequest
         * @constructor
         * @param {realtime.IClickRequest=} [properties] Properties to set
         */
        function ClickRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ClickRequest slug.
         * @member {string} slug
         * @memberof realtime.ClickRequest
         * @instance
         */
        ClickRequest.prototype.slug = "";

        /**
         * ClickRequest comboCount.
         * @member {number|Long} comboCount
         * @memberof realtime.ClickRequest
         * @instance
         */
        ClickRequest.prototype.comboCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new ClickRequest instance using the specified properties.
         * @function create
         * @memberof realtime.ClickRequest
         * @static
         * @param {realtime.IClickRequest=} [properties] Properties to set
         * @returns {realtime.ClickRequest} ClickRequest instance
         */
        ClickRequest.create = function create(properties) {
            return new ClickRequest(properties);
        };

        /**
         * Encodes the specified ClickRequest message. Does not implicitly {@link realtime.ClickRequest.verify|verify} messages.
         * @function encode
         * @memberof realtime.ClickRequest
         * @static
         * @param {realtime.IClickRequest} message ClickRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ClickRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.slug != null && Object.hasOwnProperty.call(message, "slug"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.slug);
            if (message.comboCount != null && Object.hasOwnProperty.call(message, "comboCount"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.comboCount);
            return writer;
        };

        /**
         * Encodes the specified ClickRequest message, length delimited. Does not implicitly {@link realtime.ClickRequest.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.ClickRequest
         * @static
         * @param {realtime.IClickRequest} message ClickRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ClickRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ClickRequest message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.ClickRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.ClickRequest} ClickRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ClickRequest.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.ClickRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.slug = reader.string();
                        break;
                    }
                case 2: {
                        message.comboCount = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ClickRequest message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.ClickRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.ClickRequest} ClickRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ClickRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ClickRequest message.
         * @function verify
         * @memberof realtime.ClickRequest
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ClickRequest.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.slug != null && message.hasOwnProperty("slug"))
                if (!$util.isString(message.slug))
                    return "slug: string expected";
            if (message.comboCount != null && message.hasOwnProperty("comboCount"))
                if (!$util.isInteger(message.comboCount) && !(message.comboCount && $util.isInteger(message.comboCount.low) && $util.isInteger(message.comboCount.high)))
                    return "comboCount: integer|Long expected";
            return null;
        };

        /**
         * Creates a ClickRequest message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.ClickRequest
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.ClickRequest} ClickRequest
         */
        ClickRequest.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.ClickRequest)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.ClickRequest();
            if (object.slug != null)
                message.slug = String(object.slug);
            if (object.comboCount != null)
                if ($util.Long)
                    (message.comboCount = $util.Long.fromValue(object.comboCount)).unsigned = false;
                else if (typeof object.comboCount === "string")
                    message.comboCount = parseInt(object.comboCount, 10);
                else if (typeof object.comboCount === "number")
                    message.comboCount = object.comboCount;
                else if (typeof object.comboCount === "object")
                    message.comboCount = new $util.LongBits(object.comboCount.low >>> 0, object.comboCount.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a ClickRequest message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.ClickRequest
         * @static
         * @param {realtime.ClickRequest} message ClickRequest
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ClickRequest.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.slug = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.comboCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.comboCount = options.longs === String ? "0" : 0;
            }
            if (message.slug != null && message.hasOwnProperty("slug"))
                object.slug = message.slug;
            if (message.comboCount != null && message.hasOwnProperty("comboCount"))
                if (typeof message.comboCount === "number")
                    object.comboCount = options.longs === String ? String(message.comboCount) : message.comboCount;
                else
                    object.comboCount = options.longs === String ? $util.Long.prototype.toString.call(message.comboCount) : options.longs === Number ? new $util.LongBits(message.comboCount.low >>> 0, message.comboCount.high >>> 0).toNumber() : message.comboCount;
            return object;
        };

        /**
         * Converts this ClickRequest to JSON.
         * @function toJSON
         * @memberof realtime.ClickRequest
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ClickRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ClickRequest
         * @function getTypeUrl
         * @memberof realtime.ClickRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ClickRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.ClickRequest";
        };

        return ClickRequest;
    })();

    realtime.ClickAck = (function() {

        /**
         * Properties of a ClickAck.
         * @memberof realtime
         * @interface IClickAck
         * @property {number|Long|null} [delta] ClickAck delta
         * @property {boolean|null} [critical] ClickAck critical
         * @property {number|Long|null} [bossDamage] ClickAck bossDamage
         * @property {number|Long|null} [myBossDamage] ClickAck myBossDamage
         * @property {number|null} [bossLeaderboardCount] ClickAck bossLeaderboardCount
         * @property {string|null} [damageType] ClickAck damageType
         * @property {Array.<realtime.ITalentTriggerEvent>|null} [talentEvents] ClickAck talentEvents
         * @property {Array.<realtime.IBossPartStateDelta>|null} [partStateDeltas] ClickAck partStateDeltas
         * @property {realtime.ITalentCombatState|null} [talentCombatState] ClickAck talentCombatState
         * @property {realtime.IUserDeltaPatch|null} [userDelta] ClickAck userDelta
         * @property {realtime.IButtonRef|null} [button] ClickAck button
         */

        /**
         * Constructs a new ClickAck.
         * @memberof realtime
         * @classdesc Represents a ClickAck.
         * @implements IClickAck
         * @constructor
         * @param {realtime.IClickAck=} [properties] Properties to set
         */
        function ClickAck(properties) {
            this.talentEvents = [];
            this.partStateDeltas = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ClickAck delta.
         * @member {number|Long} delta
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.delta = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ClickAck critical.
         * @member {boolean} critical
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.critical = false;

        /**
         * ClickAck bossDamage.
         * @member {number|Long} bossDamage
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.bossDamage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ClickAck myBossDamage.
         * @member {number|Long} myBossDamage
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.myBossDamage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ClickAck bossLeaderboardCount.
         * @member {number} bossLeaderboardCount
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.bossLeaderboardCount = 0;

        /**
         * ClickAck damageType.
         * @member {string} damageType
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.damageType = "";

        /**
         * ClickAck talentEvents.
         * @member {Array.<realtime.ITalentTriggerEvent>} talentEvents
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.talentEvents = $util.emptyArray;

        /**
         * ClickAck partStateDeltas.
         * @member {Array.<realtime.IBossPartStateDelta>} partStateDeltas
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.partStateDeltas = $util.emptyArray;

        /**
         * ClickAck talentCombatState.
         * @member {realtime.ITalentCombatState|null|undefined} talentCombatState
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.talentCombatState = null;

        /**
         * ClickAck userDelta.
         * @member {realtime.IUserDeltaPatch|null|undefined} userDelta
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.userDelta = null;

        /**
         * ClickAck button.
         * @member {realtime.IButtonRef|null|undefined} button
         * @memberof realtime.ClickAck
         * @instance
         */
        ClickAck.prototype.button = null;

        /**
         * Creates a new ClickAck instance using the specified properties.
         * @function create
         * @memberof realtime.ClickAck
         * @static
         * @param {realtime.IClickAck=} [properties] Properties to set
         * @returns {realtime.ClickAck} ClickAck instance
         */
        ClickAck.create = function create(properties) {
            return new ClickAck(properties);
        };

        /**
         * Encodes the specified ClickAck message. Does not implicitly {@link realtime.ClickAck.verify|verify} messages.
         * @function encode
         * @memberof realtime.ClickAck
         * @static
         * @param {realtime.IClickAck} message ClickAck message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ClickAck.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.delta != null && Object.hasOwnProperty.call(message, "delta"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.delta);
            if (message.critical != null && Object.hasOwnProperty.call(message, "critical"))
                writer.uint32(/* id 2, wireType 0 =*/16).bool(message.critical);
            if (message.bossDamage != null && Object.hasOwnProperty.call(message, "bossDamage"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.bossDamage);
            if (message.myBossDamage != null && Object.hasOwnProperty.call(message, "myBossDamage"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.myBossDamage);
            if (message.bossLeaderboardCount != null && Object.hasOwnProperty.call(message, "bossLeaderboardCount"))
                writer.uint32(/* id 5, wireType 0 =*/40).int32(message.bossLeaderboardCount);
            if (message.damageType != null && Object.hasOwnProperty.call(message, "damageType"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.damageType);
            if (message.talentEvents != null && message.talentEvents.length)
                for (let i = 0; i < message.talentEvents.length; ++i)
                    $root.realtime.TalentTriggerEvent.encode(message.talentEvents[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
            if (message.partStateDeltas != null && message.partStateDeltas.length)
                for (let i = 0; i < message.partStateDeltas.length; ++i)
                    $root.realtime.BossPartStateDelta.encode(message.partStateDeltas[i], writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
            if (message.talentCombatState != null && Object.hasOwnProperty.call(message, "talentCombatState"))
                $root.realtime.TalentCombatState.encode(message.talentCombatState, writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
            if (message.userDelta != null && Object.hasOwnProperty.call(message, "userDelta"))
                $root.realtime.UserDeltaPatch.encode(message.userDelta, writer.uint32(/* id 10, wireType 2 =*/82).fork()).ldelim();
            if (message.button != null && Object.hasOwnProperty.call(message, "button"))
                $root.realtime.ButtonRef.encode(message.button, writer.uint32(/* id 11, wireType 2 =*/90).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified ClickAck message, length delimited. Does not implicitly {@link realtime.ClickAck.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.ClickAck
         * @static
         * @param {realtime.IClickAck} message ClickAck message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ClickAck.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ClickAck message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.ClickAck
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.ClickAck} ClickAck
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ClickAck.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.ClickAck();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.delta = reader.int64();
                        break;
                    }
                case 2: {
                        message.critical = reader.bool();
                        break;
                    }
                case 3: {
                        message.bossDamage = reader.int64();
                        break;
                    }
                case 4: {
                        message.myBossDamage = reader.int64();
                        break;
                    }
                case 5: {
                        message.bossLeaderboardCount = reader.int32();
                        break;
                    }
                case 6: {
                        message.damageType = reader.string();
                        break;
                    }
                case 7: {
                        if (!(message.talentEvents && message.talentEvents.length))
                            message.talentEvents = [];
                        message.talentEvents.push($root.realtime.TalentTriggerEvent.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 8: {
                        if (!(message.partStateDeltas && message.partStateDeltas.length))
                            message.partStateDeltas = [];
                        message.partStateDeltas.push($root.realtime.BossPartStateDelta.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 9: {
                        message.talentCombatState = $root.realtime.TalentCombatState.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 10: {
                        message.userDelta = $root.realtime.UserDeltaPatch.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 11: {
                        message.button = $root.realtime.ButtonRef.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ClickAck message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.ClickAck
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.ClickAck} ClickAck
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ClickAck.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ClickAck message.
         * @function verify
         * @memberof realtime.ClickAck
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ClickAck.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.delta != null && message.hasOwnProperty("delta"))
                if (!$util.isInteger(message.delta) && !(message.delta && $util.isInteger(message.delta.low) && $util.isInteger(message.delta.high)))
                    return "delta: integer|Long expected";
            if (message.critical != null && message.hasOwnProperty("critical"))
                if (typeof message.critical !== "boolean")
                    return "critical: boolean expected";
            if (message.bossDamage != null && message.hasOwnProperty("bossDamage"))
                if (!$util.isInteger(message.bossDamage) && !(message.bossDamage && $util.isInteger(message.bossDamage.low) && $util.isInteger(message.bossDamage.high)))
                    return "bossDamage: integer|Long expected";
            if (message.myBossDamage != null && message.hasOwnProperty("myBossDamage"))
                if (!$util.isInteger(message.myBossDamage) && !(message.myBossDamage && $util.isInteger(message.myBossDamage.low) && $util.isInteger(message.myBossDamage.high)))
                    return "myBossDamage: integer|Long expected";
            if (message.bossLeaderboardCount != null && message.hasOwnProperty("bossLeaderboardCount"))
                if (!$util.isInteger(message.bossLeaderboardCount))
                    return "bossLeaderboardCount: integer expected";
            if (message.damageType != null && message.hasOwnProperty("damageType"))
                if (!$util.isString(message.damageType))
                    return "damageType: string expected";
            if (message.talentEvents != null && message.hasOwnProperty("talentEvents")) {
                if (!Array.isArray(message.talentEvents))
                    return "talentEvents: array expected";
                for (let i = 0; i < message.talentEvents.length; ++i) {
                    let error = $root.realtime.TalentTriggerEvent.verify(message.talentEvents[i], long + 1);
                    if (error)
                        return "talentEvents." + error;
                }
            }
            if (message.partStateDeltas != null && message.hasOwnProperty("partStateDeltas")) {
                if (!Array.isArray(message.partStateDeltas))
                    return "partStateDeltas: array expected";
                for (let i = 0; i < message.partStateDeltas.length; ++i) {
                    let error = $root.realtime.BossPartStateDelta.verify(message.partStateDeltas[i], long + 1);
                    if (error)
                        return "partStateDeltas." + error;
                }
            }
            if (message.talentCombatState != null && message.hasOwnProperty("talentCombatState")) {
                let error = $root.realtime.TalentCombatState.verify(message.talentCombatState, long + 1);
                if (error)
                    return "talentCombatState." + error;
            }
            if (message.userDelta != null && message.hasOwnProperty("userDelta")) {
                let error = $root.realtime.UserDeltaPatch.verify(message.userDelta, long + 1);
                if (error)
                    return "userDelta." + error;
            }
            if (message.button != null && message.hasOwnProperty("button")) {
                let error = $root.realtime.ButtonRef.verify(message.button, long + 1);
                if (error)
                    return "button." + error;
            }
            return null;
        };

        /**
         * Creates a ClickAck message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.ClickAck
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.ClickAck} ClickAck
         */
        ClickAck.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.ClickAck)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.ClickAck();
            if (object.delta != null)
                if ($util.Long)
                    (message.delta = $util.Long.fromValue(object.delta)).unsigned = false;
                else if (typeof object.delta === "string")
                    message.delta = parseInt(object.delta, 10);
                else if (typeof object.delta === "number")
                    message.delta = object.delta;
                else if (typeof object.delta === "object")
                    message.delta = new $util.LongBits(object.delta.low >>> 0, object.delta.high >>> 0).toNumber();
            if (object.critical != null)
                message.critical = Boolean(object.critical);
            if (object.bossDamage != null)
                if ($util.Long)
                    (message.bossDamage = $util.Long.fromValue(object.bossDamage)).unsigned = false;
                else if (typeof object.bossDamage === "string")
                    message.bossDamage = parseInt(object.bossDamage, 10);
                else if (typeof object.bossDamage === "number")
                    message.bossDamage = object.bossDamage;
                else if (typeof object.bossDamage === "object")
                    message.bossDamage = new $util.LongBits(object.bossDamage.low >>> 0, object.bossDamage.high >>> 0).toNumber();
            if (object.myBossDamage != null)
                if ($util.Long)
                    (message.myBossDamage = $util.Long.fromValue(object.myBossDamage)).unsigned = false;
                else if (typeof object.myBossDamage === "string")
                    message.myBossDamage = parseInt(object.myBossDamage, 10);
                else if (typeof object.myBossDamage === "number")
                    message.myBossDamage = object.myBossDamage;
                else if (typeof object.myBossDamage === "object")
                    message.myBossDamage = new $util.LongBits(object.myBossDamage.low >>> 0, object.myBossDamage.high >>> 0).toNumber();
            if (object.bossLeaderboardCount != null)
                message.bossLeaderboardCount = object.bossLeaderboardCount | 0;
            if (object.damageType != null)
                message.damageType = String(object.damageType);
            if (object.talentEvents) {
                if (!Array.isArray(object.talentEvents))
                    throw TypeError(".realtime.ClickAck.talentEvents: array expected");
                message.talentEvents = [];
                for (let i = 0; i < object.talentEvents.length; ++i) {
                    if (typeof object.talentEvents[i] !== "object")
                        throw TypeError(".realtime.ClickAck.talentEvents: object expected");
                    message.talentEvents[i] = $root.realtime.TalentTriggerEvent.fromObject(object.talentEvents[i], long + 1);
                }
            }
            if (object.partStateDeltas) {
                if (!Array.isArray(object.partStateDeltas))
                    throw TypeError(".realtime.ClickAck.partStateDeltas: array expected");
                message.partStateDeltas = [];
                for (let i = 0; i < object.partStateDeltas.length; ++i) {
                    if (typeof object.partStateDeltas[i] !== "object")
                        throw TypeError(".realtime.ClickAck.partStateDeltas: object expected");
                    message.partStateDeltas[i] = $root.realtime.BossPartStateDelta.fromObject(object.partStateDeltas[i], long + 1);
                }
            }
            if (object.talentCombatState != null) {
                if (typeof object.talentCombatState !== "object")
                    throw TypeError(".realtime.ClickAck.talentCombatState: object expected");
                message.talentCombatState = $root.realtime.TalentCombatState.fromObject(object.talentCombatState, long + 1);
            }
            if (object.userDelta != null) {
                if (typeof object.userDelta !== "object")
                    throw TypeError(".realtime.ClickAck.userDelta: object expected");
                message.userDelta = $root.realtime.UserDeltaPatch.fromObject(object.userDelta, long + 1);
            }
            if (object.button != null) {
                if (typeof object.button !== "object")
                    throw TypeError(".realtime.ClickAck.button: object expected");
                message.button = $root.realtime.ButtonRef.fromObject(object.button, long + 1);
            }
            return message;
        };

        /**
         * Creates a plain object from a ClickAck message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.ClickAck
         * @static
         * @param {realtime.ClickAck} message ClickAck
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ClickAck.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.talentEvents = [];
                object.partStateDeltas = [];
            }
            if (options.defaults) {
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.delta = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.delta = options.longs === String ? "0" : 0;
                object.critical = false;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.bossDamage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.bossDamage = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.myBossDamage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.myBossDamage = options.longs === String ? "0" : 0;
                object.bossLeaderboardCount = 0;
                object.damageType = "";
                object.talentCombatState = null;
                object.userDelta = null;
                object.button = null;
            }
            if (message.delta != null && message.hasOwnProperty("delta"))
                if (typeof message.delta === "number")
                    object.delta = options.longs === String ? String(message.delta) : message.delta;
                else
                    object.delta = options.longs === String ? $util.Long.prototype.toString.call(message.delta) : options.longs === Number ? new $util.LongBits(message.delta.low >>> 0, message.delta.high >>> 0).toNumber() : message.delta;
            if (message.critical != null && message.hasOwnProperty("critical"))
                object.critical = message.critical;
            if (message.bossDamage != null && message.hasOwnProperty("bossDamage"))
                if (typeof message.bossDamage === "number")
                    object.bossDamage = options.longs === String ? String(message.bossDamage) : message.bossDamage;
                else
                    object.bossDamage = options.longs === String ? $util.Long.prototype.toString.call(message.bossDamage) : options.longs === Number ? new $util.LongBits(message.bossDamage.low >>> 0, message.bossDamage.high >>> 0).toNumber() : message.bossDamage;
            if (message.myBossDamage != null && message.hasOwnProperty("myBossDamage"))
                if (typeof message.myBossDamage === "number")
                    object.myBossDamage = options.longs === String ? String(message.myBossDamage) : message.myBossDamage;
                else
                    object.myBossDamage = options.longs === String ? $util.Long.prototype.toString.call(message.myBossDamage) : options.longs === Number ? new $util.LongBits(message.myBossDamage.low >>> 0, message.myBossDamage.high >>> 0).toNumber() : message.myBossDamage;
            if (message.bossLeaderboardCount != null && message.hasOwnProperty("bossLeaderboardCount"))
                object.bossLeaderboardCount = message.bossLeaderboardCount;
            if (message.damageType != null && message.hasOwnProperty("damageType"))
                object.damageType = message.damageType;
            if (message.talentEvents && message.talentEvents.length) {
                object.talentEvents = [];
                for (let j = 0; j < message.talentEvents.length; ++j)
                    object.talentEvents[j] = $root.realtime.TalentTriggerEvent.toObject(message.talentEvents[j], options);
            }
            if (message.partStateDeltas && message.partStateDeltas.length) {
                object.partStateDeltas = [];
                for (let j = 0; j < message.partStateDeltas.length; ++j)
                    object.partStateDeltas[j] = $root.realtime.BossPartStateDelta.toObject(message.partStateDeltas[j], options);
            }
            if (message.talentCombatState != null && message.hasOwnProperty("talentCombatState"))
                object.talentCombatState = $root.realtime.TalentCombatState.toObject(message.talentCombatState, options);
            if (message.userDelta != null && message.hasOwnProperty("userDelta"))
                object.userDelta = $root.realtime.UserDeltaPatch.toObject(message.userDelta, options);
            if (message.button != null && message.hasOwnProperty("button"))
                object.button = $root.realtime.ButtonRef.toObject(message.button, options);
            return object;
        };

        /**
         * Converts this ClickAck to JSON.
         * @function toJSON
         * @memberof realtime.ClickAck
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ClickAck.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ClickAck
         * @function getTypeUrl
         * @memberof realtime.ClickAck
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ClickAck.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.ClickAck";
        };

        return ClickAck;
    })();

    realtime.UserDeltaPatch = (function() {

        /**
         * Properties of a UserDeltaPatch.
         * @memberof realtime
         * @interface IUserDeltaPatch
         * @property {number|Long|null} [gold] UserDeltaPatch gold
         * @property {number|Long|null} [stones] UserDeltaPatch stones
         * @property {number|Long|null} [talentPoints] UserDeltaPatch talentPoints
         */

        /**
         * Constructs a new UserDeltaPatch.
         * @memberof realtime
         * @classdesc Represents a UserDeltaPatch.
         * @implements IUserDeltaPatch
         * @constructor
         * @param {realtime.IUserDeltaPatch=} [properties] Properties to set
         */
        function UserDeltaPatch(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * UserDeltaPatch gold.
         * @member {number|Long} gold
         * @memberof realtime.UserDeltaPatch
         * @instance
         */
        UserDeltaPatch.prototype.gold = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserDeltaPatch stones.
         * @member {number|Long} stones
         * @memberof realtime.UserDeltaPatch
         * @instance
         */
        UserDeltaPatch.prototype.stones = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserDeltaPatch talentPoints.
         * @member {number|Long} talentPoints
         * @memberof realtime.UserDeltaPatch
         * @instance
         */
        UserDeltaPatch.prototype.talentPoints = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new UserDeltaPatch instance using the specified properties.
         * @function create
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {realtime.IUserDeltaPatch=} [properties] Properties to set
         * @returns {realtime.UserDeltaPatch} UserDeltaPatch instance
         */
        UserDeltaPatch.create = function create(properties) {
            return new UserDeltaPatch(properties);
        };

        /**
         * Encodes the specified UserDeltaPatch message. Does not implicitly {@link realtime.UserDeltaPatch.verify|verify} messages.
         * @function encode
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {realtime.IUserDeltaPatch} message UserDeltaPatch message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserDeltaPatch.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.gold != null && Object.hasOwnProperty.call(message, "gold"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.gold);
            if (message.stones != null && Object.hasOwnProperty.call(message, "stones"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.stones);
            if (message.talentPoints != null && Object.hasOwnProperty.call(message, "talentPoints"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.talentPoints);
            return writer;
        };

        /**
         * Encodes the specified UserDeltaPatch message, length delimited. Does not implicitly {@link realtime.UserDeltaPatch.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {realtime.IUserDeltaPatch} message UserDeltaPatch message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserDeltaPatch.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a UserDeltaPatch message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.UserDeltaPatch} UserDeltaPatch
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserDeltaPatch.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.UserDeltaPatch();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.gold = reader.int64();
                        break;
                    }
                case 2: {
                        message.stones = reader.int64();
                        break;
                    }
                case 3: {
                        message.talentPoints = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a UserDeltaPatch message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.UserDeltaPatch} UserDeltaPatch
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserDeltaPatch.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a UserDeltaPatch message.
         * @function verify
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        UserDeltaPatch.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.gold != null && message.hasOwnProperty("gold"))
                if (!$util.isInteger(message.gold) && !(message.gold && $util.isInteger(message.gold.low) && $util.isInteger(message.gold.high)))
                    return "gold: integer|Long expected";
            if (message.stones != null && message.hasOwnProperty("stones"))
                if (!$util.isInteger(message.stones) && !(message.stones && $util.isInteger(message.stones.low) && $util.isInteger(message.stones.high)))
                    return "stones: integer|Long expected";
            if (message.talentPoints != null && message.hasOwnProperty("talentPoints"))
                if (!$util.isInteger(message.talentPoints) && !(message.talentPoints && $util.isInteger(message.talentPoints.low) && $util.isInteger(message.talentPoints.high)))
                    return "talentPoints: integer|Long expected";
            return null;
        };

        /**
         * Creates a UserDeltaPatch message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.UserDeltaPatch} UserDeltaPatch
         */
        UserDeltaPatch.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.UserDeltaPatch)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.UserDeltaPatch();
            if (object.gold != null)
                if ($util.Long)
                    (message.gold = $util.Long.fromValue(object.gold)).unsigned = false;
                else if (typeof object.gold === "string")
                    message.gold = parseInt(object.gold, 10);
                else if (typeof object.gold === "number")
                    message.gold = object.gold;
                else if (typeof object.gold === "object")
                    message.gold = new $util.LongBits(object.gold.low >>> 0, object.gold.high >>> 0).toNumber();
            if (object.stones != null)
                if ($util.Long)
                    (message.stones = $util.Long.fromValue(object.stones)).unsigned = false;
                else if (typeof object.stones === "string")
                    message.stones = parseInt(object.stones, 10);
                else if (typeof object.stones === "number")
                    message.stones = object.stones;
                else if (typeof object.stones === "object")
                    message.stones = new $util.LongBits(object.stones.low >>> 0, object.stones.high >>> 0).toNumber();
            if (object.talentPoints != null)
                if ($util.Long)
                    (message.talentPoints = $util.Long.fromValue(object.talentPoints)).unsigned = false;
                else if (typeof object.talentPoints === "string")
                    message.talentPoints = parseInt(object.talentPoints, 10);
                else if (typeof object.talentPoints === "number")
                    message.talentPoints = object.talentPoints;
                else if (typeof object.talentPoints === "object")
                    message.talentPoints = new $util.LongBits(object.talentPoints.low >>> 0, object.talentPoints.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a UserDeltaPatch message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {realtime.UserDeltaPatch} message UserDeltaPatch
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        UserDeltaPatch.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.gold = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.gold = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.stones = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.stones = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.talentPoints = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.talentPoints = options.longs === String ? "0" : 0;
            }
            if (message.gold != null && message.hasOwnProperty("gold"))
                if (typeof message.gold === "number")
                    object.gold = options.longs === String ? String(message.gold) : message.gold;
                else
                    object.gold = options.longs === String ? $util.Long.prototype.toString.call(message.gold) : options.longs === Number ? new $util.LongBits(message.gold.low >>> 0, message.gold.high >>> 0).toNumber() : message.gold;
            if (message.stones != null && message.hasOwnProperty("stones"))
                if (typeof message.stones === "number")
                    object.stones = options.longs === String ? String(message.stones) : message.stones;
                else
                    object.stones = options.longs === String ? $util.Long.prototype.toString.call(message.stones) : options.longs === Number ? new $util.LongBits(message.stones.low >>> 0, message.stones.high >>> 0).toNumber() : message.stones;
            if (message.talentPoints != null && message.hasOwnProperty("talentPoints"))
                if (typeof message.talentPoints === "number")
                    object.talentPoints = options.longs === String ? String(message.talentPoints) : message.talentPoints;
                else
                    object.talentPoints = options.longs === String ? $util.Long.prototype.toString.call(message.talentPoints) : options.longs === Number ? new $util.LongBits(message.talentPoints.low >>> 0, message.talentPoints.high >>> 0).toNumber() : message.talentPoints;
            return object;
        };

        /**
         * Converts this UserDeltaPatch to JSON.
         * @function toJSON
         * @memberof realtime.UserDeltaPatch
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        UserDeltaPatch.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for UserDeltaPatch
         * @function getTypeUrl
         * @memberof realtime.UserDeltaPatch
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        UserDeltaPatch.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.UserDeltaPatch";
        };

        return UserDeltaPatch;
    })();

    realtime.ButtonRef = (function() {

        /**
         * Properties of a ButtonRef.
         * @memberof realtime
         * @interface IButtonRef
         * @property {string|null} [key] ButtonRef key
         */

        /**
         * Constructs a new ButtonRef.
         * @memberof realtime
         * @classdesc Represents a ButtonRef.
         * @implements IButtonRef
         * @constructor
         * @param {realtime.IButtonRef=} [properties] Properties to set
         */
        function ButtonRef(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ButtonRef key.
         * @member {string} key
         * @memberof realtime.ButtonRef
         * @instance
         */
        ButtonRef.prototype.key = "";

        /**
         * Creates a new ButtonRef instance using the specified properties.
         * @function create
         * @memberof realtime.ButtonRef
         * @static
         * @param {realtime.IButtonRef=} [properties] Properties to set
         * @returns {realtime.ButtonRef} ButtonRef instance
         */
        ButtonRef.create = function create(properties) {
            return new ButtonRef(properties);
        };

        /**
         * Encodes the specified ButtonRef message. Does not implicitly {@link realtime.ButtonRef.verify|verify} messages.
         * @function encode
         * @memberof realtime.ButtonRef
         * @static
         * @param {realtime.IButtonRef} message ButtonRef message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ButtonRef.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.key != null && Object.hasOwnProperty.call(message, "key"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.key);
            return writer;
        };

        /**
         * Encodes the specified ButtonRef message, length delimited. Does not implicitly {@link realtime.ButtonRef.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.ButtonRef
         * @static
         * @param {realtime.IButtonRef} message ButtonRef message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ButtonRef.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ButtonRef message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.ButtonRef
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.ButtonRef} ButtonRef
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ButtonRef.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.ButtonRef();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.key = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ButtonRef message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.ButtonRef
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.ButtonRef} ButtonRef
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ButtonRef.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ButtonRef message.
         * @function verify
         * @memberof realtime.ButtonRef
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ButtonRef.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.key != null && message.hasOwnProperty("key"))
                if (!$util.isString(message.key))
                    return "key: string expected";
            return null;
        };

        /**
         * Creates a ButtonRef message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.ButtonRef
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.ButtonRef} ButtonRef
         */
        ButtonRef.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.ButtonRef)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.ButtonRef();
            if (object.key != null)
                message.key = String(object.key);
            return message;
        };

        /**
         * Creates a plain object from a ButtonRef message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.ButtonRef
         * @static
         * @param {realtime.ButtonRef} message ButtonRef
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ButtonRef.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults)
                object.key = "";
            if (message.key != null && message.hasOwnProperty("key"))
                object.key = message.key;
            return object;
        };

        /**
         * Converts this ButtonRef to JSON.
         * @function toJSON
         * @memberof realtime.ButtonRef
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ButtonRef.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ButtonRef
         * @function getTypeUrl
         * @memberof realtime.ButtonRef
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ButtonRef.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.ButtonRef";
        };

        return ButtonRef;
    })();

    realtime.PublicDelta = (function() {

        /**
         * Properties of a PublicDelta.
         * @memberof realtime
         * @interface IPublicDelta
         * @property {number|Long|null} [totalVotes] PublicDelta totalVotes
         * @property {Array.<realtime.ILeaderboardEntry>|null} [leaderboard] PublicDelta leaderboard
         * @property {string|null} [roomId] PublicDelta roomId
         * @property {realtime.IBoss|null} [boss] PublicDelta boss
         * @property {Array.<realtime.IBossLeaderboardEntry>|null} [bossLeaderboard] PublicDelta bossLeaderboard
         * @property {string|null} [announcementVersion] PublicDelta announcementVersion
         */

        /**
         * Constructs a new PublicDelta.
         * @memberof realtime
         * @classdesc Represents a PublicDelta.
         * @implements IPublicDelta
         * @constructor
         * @param {realtime.IPublicDelta=} [properties] Properties to set
         */
        function PublicDelta(properties) {
            this.leaderboard = [];
            this.bossLeaderboard = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * PublicDelta totalVotes.
         * @member {number|Long} totalVotes
         * @memberof realtime.PublicDelta
         * @instance
         */
        PublicDelta.prototype.totalVotes = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * PublicDelta leaderboard.
         * @member {Array.<realtime.ILeaderboardEntry>} leaderboard
         * @memberof realtime.PublicDelta
         * @instance
         */
        PublicDelta.prototype.leaderboard = $util.emptyArray;

        /**
         * PublicDelta roomId.
         * @member {string} roomId
         * @memberof realtime.PublicDelta
         * @instance
         */
        PublicDelta.prototype.roomId = "";

        /**
         * PublicDelta boss.
         * @member {realtime.IBoss|null|undefined} boss
         * @memberof realtime.PublicDelta
         * @instance
         */
        PublicDelta.prototype.boss = null;

        /**
         * PublicDelta bossLeaderboard.
         * @member {Array.<realtime.IBossLeaderboardEntry>} bossLeaderboard
         * @memberof realtime.PublicDelta
         * @instance
         */
        PublicDelta.prototype.bossLeaderboard = $util.emptyArray;

        /**
         * PublicDelta announcementVersion.
         * @member {string} announcementVersion
         * @memberof realtime.PublicDelta
         * @instance
         */
        PublicDelta.prototype.announcementVersion = "";

        /**
         * Creates a new PublicDelta instance using the specified properties.
         * @function create
         * @memberof realtime.PublicDelta
         * @static
         * @param {realtime.IPublicDelta=} [properties] Properties to set
         * @returns {realtime.PublicDelta} PublicDelta instance
         */
        PublicDelta.create = function create(properties) {
            return new PublicDelta(properties);
        };

        /**
         * Encodes the specified PublicDelta message. Does not implicitly {@link realtime.PublicDelta.verify|verify} messages.
         * @function encode
         * @memberof realtime.PublicDelta
         * @static
         * @param {realtime.IPublicDelta} message PublicDelta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        PublicDelta.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.totalVotes != null && Object.hasOwnProperty.call(message, "totalVotes"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.totalVotes);
            if (message.leaderboard != null && message.leaderboard.length)
                for (let i = 0; i < message.leaderboard.length; ++i)
                    $root.realtime.LeaderboardEntry.encode(message.leaderboard[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.roomId != null && Object.hasOwnProperty.call(message, "roomId"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.roomId);
            if (message.boss != null && Object.hasOwnProperty.call(message, "boss"))
                $root.realtime.Boss.encode(message.boss, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            if (message.bossLeaderboard != null && message.bossLeaderboard.length)
                for (let i = 0; i < message.bossLeaderboard.length; ++i)
                    $root.realtime.BossLeaderboardEntry.encode(message.bossLeaderboard[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
            if (message.announcementVersion != null && Object.hasOwnProperty.call(message, "announcementVersion"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.announcementVersion);
            return writer;
        };

        /**
         * Encodes the specified PublicDelta message, length delimited. Does not implicitly {@link realtime.PublicDelta.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.PublicDelta
         * @static
         * @param {realtime.IPublicDelta} message PublicDelta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        PublicDelta.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a PublicDelta message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.PublicDelta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.PublicDelta} PublicDelta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        PublicDelta.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.PublicDelta();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.totalVotes = reader.int64();
                        break;
                    }
                case 2: {
                        if (!(message.leaderboard && message.leaderboard.length))
                            message.leaderboard = [];
                        message.leaderboard.push($root.realtime.LeaderboardEntry.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 3: {
                        message.roomId = reader.string();
                        break;
                    }
                case 4: {
                        message.boss = $root.realtime.Boss.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 5: {
                        if (!(message.bossLeaderboard && message.bossLeaderboard.length))
                            message.bossLeaderboard = [];
                        message.bossLeaderboard.push($root.realtime.BossLeaderboardEntry.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 6: {
                        message.announcementVersion = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a PublicDelta message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.PublicDelta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.PublicDelta} PublicDelta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        PublicDelta.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a PublicDelta message.
         * @function verify
         * @memberof realtime.PublicDelta
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        PublicDelta.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.totalVotes != null && message.hasOwnProperty("totalVotes"))
                if (!$util.isInteger(message.totalVotes) && !(message.totalVotes && $util.isInteger(message.totalVotes.low) && $util.isInteger(message.totalVotes.high)))
                    return "totalVotes: integer|Long expected";
            if (message.leaderboard != null && message.hasOwnProperty("leaderboard")) {
                if (!Array.isArray(message.leaderboard))
                    return "leaderboard: array expected";
                for (let i = 0; i < message.leaderboard.length; ++i) {
                    let error = $root.realtime.LeaderboardEntry.verify(message.leaderboard[i], long + 1);
                    if (error)
                        return "leaderboard." + error;
                }
            }
            if (message.roomId != null && message.hasOwnProperty("roomId"))
                if (!$util.isString(message.roomId))
                    return "roomId: string expected";
            if (message.boss != null && message.hasOwnProperty("boss")) {
                let error = $root.realtime.Boss.verify(message.boss, long + 1);
                if (error)
                    return "boss." + error;
            }
            if (message.bossLeaderboard != null && message.hasOwnProperty("bossLeaderboard")) {
                if (!Array.isArray(message.bossLeaderboard))
                    return "bossLeaderboard: array expected";
                for (let i = 0; i < message.bossLeaderboard.length; ++i) {
                    let error = $root.realtime.BossLeaderboardEntry.verify(message.bossLeaderboard[i], long + 1);
                    if (error)
                        return "bossLeaderboard." + error;
                }
            }
            if (message.announcementVersion != null && message.hasOwnProperty("announcementVersion"))
                if (!$util.isString(message.announcementVersion))
                    return "announcementVersion: string expected";
            return null;
        };

        /**
         * Creates a PublicDelta message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.PublicDelta
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.PublicDelta} PublicDelta
         */
        PublicDelta.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.PublicDelta)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.PublicDelta();
            if (object.totalVotes != null)
                if ($util.Long)
                    (message.totalVotes = $util.Long.fromValue(object.totalVotes)).unsigned = false;
                else if (typeof object.totalVotes === "string")
                    message.totalVotes = parseInt(object.totalVotes, 10);
                else if (typeof object.totalVotes === "number")
                    message.totalVotes = object.totalVotes;
                else if (typeof object.totalVotes === "object")
                    message.totalVotes = new $util.LongBits(object.totalVotes.low >>> 0, object.totalVotes.high >>> 0).toNumber();
            if (object.leaderboard) {
                if (!Array.isArray(object.leaderboard))
                    throw TypeError(".realtime.PublicDelta.leaderboard: array expected");
                message.leaderboard = [];
                for (let i = 0; i < object.leaderboard.length; ++i) {
                    if (typeof object.leaderboard[i] !== "object")
                        throw TypeError(".realtime.PublicDelta.leaderboard: object expected");
                    message.leaderboard[i] = $root.realtime.LeaderboardEntry.fromObject(object.leaderboard[i], long + 1);
                }
            }
            if (object.roomId != null)
                message.roomId = String(object.roomId);
            if (object.boss != null) {
                if (typeof object.boss !== "object")
                    throw TypeError(".realtime.PublicDelta.boss: object expected");
                message.boss = $root.realtime.Boss.fromObject(object.boss, long + 1);
            }
            if (object.bossLeaderboard) {
                if (!Array.isArray(object.bossLeaderboard))
                    throw TypeError(".realtime.PublicDelta.bossLeaderboard: array expected");
                message.bossLeaderboard = [];
                for (let i = 0; i < object.bossLeaderboard.length; ++i) {
                    if (typeof object.bossLeaderboard[i] !== "object")
                        throw TypeError(".realtime.PublicDelta.bossLeaderboard: object expected");
                    message.bossLeaderboard[i] = $root.realtime.BossLeaderboardEntry.fromObject(object.bossLeaderboard[i], long + 1);
                }
            }
            if (object.announcementVersion != null)
                message.announcementVersion = String(object.announcementVersion);
            return message;
        };

        /**
         * Creates a plain object from a PublicDelta message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.PublicDelta
         * @static
         * @param {realtime.PublicDelta} message PublicDelta
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        PublicDelta.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.leaderboard = [];
                object.bossLeaderboard = [];
            }
            if (options.defaults) {
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.totalVotes = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.totalVotes = options.longs === String ? "0" : 0;
                object.roomId = "";
                object.boss = null;
                object.announcementVersion = "";
            }
            if (message.totalVotes != null && message.hasOwnProperty("totalVotes"))
                if (typeof message.totalVotes === "number")
                    object.totalVotes = options.longs === String ? String(message.totalVotes) : message.totalVotes;
                else
                    object.totalVotes = options.longs === String ? $util.Long.prototype.toString.call(message.totalVotes) : options.longs === Number ? new $util.LongBits(message.totalVotes.low >>> 0, message.totalVotes.high >>> 0).toNumber() : message.totalVotes;
            if (message.leaderboard && message.leaderboard.length) {
                object.leaderboard = [];
                for (let j = 0; j < message.leaderboard.length; ++j)
                    object.leaderboard[j] = $root.realtime.LeaderboardEntry.toObject(message.leaderboard[j], options);
            }
            if (message.roomId != null && message.hasOwnProperty("roomId"))
                object.roomId = message.roomId;
            if (message.boss != null && message.hasOwnProperty("boss"))
                object.boss = $root.realtime.Boss.toObject(message.boss, options);
            if (message.bossLeaderboard && message.bossLeaderboard.length) {
                object.bossLeaderboard = [];
                for (let j = 0; j < message.bossLeaderboard.length; ++j)
                    object.bossLeaderboard[j] = $root.realtime.BossLeaderboardEntry.toObject(message.bossLeaderboard[j], options);
            }
            if (message.announcementVersion != null && message.hasOwnProperty("announcementVersion"))
                object.announcementVersion = message.announcementVersion;
            return object;
        };

        /**
         * Converts this PublicDelta to JSON.
         * @function toJSON
         * @memberof realtime.PublicDelta
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        PublicDelta.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for PublicDelta
         * @function getTypeUrl
         * @memberof realtime.PublicDelta
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        PublicDelta.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.PublicDelta";
        };

        return PublicDelta;
    })();

    realtime.PublicMeta = (function() {

        /**
         * Properties of a PublicMeta.
         * @memberof realtime
         * @interface IPublicMeta
         * @property {Array.<realtime.ILeaderboardEntry>|null} [leaderboard] PublicMeta leaderboard
         * @property {Array.<realtime.IBossLeaderboardEntry>|null} [bossLeaderboard] PublicMeta bossLeaderboard
         * @property {string|null} [announcementVersion] PublicMeta announcementVersion
         */

        /**
         * Constructs a new PublicMeta.
         * @memberof realtime
         * @classdesc Represents a PublicMeta.
         * @implements IPublicMeta
         * @constructor
         * @param {realtime.IPublicMeta=} [properties] Properties to set
         */
        function PublicMeta(properties) {
            this.leaderboard = [];
            this.bossLeaderboard = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * PublicMeta leaderboard.
         * @member {Array.<realtime.ILeaderboardEntry>} leaderboard
         * @memberof realtime.PublicMeta
         * @instance
         */
        PublicMeta.prototype.leaderboard = $util.emptyArray;

        /**
         * PublicMeta bossLeaderboard.
         * @member {Array.<realtime.IBossLeaderboardEntry>} bossLeaderboard
         * @memberof realtime.PublicMeta
         * @instance
         */
        PublicMeta.prototype.bossLeaderboard = $util.emptyArray;

        /**
         * PublicMeta announcementVersion.
         * @member {string} announcementVersion
         * @memberof realtime.PublicMeta
         * @instance
         */
        PublicMeta.prototype.announcementVersion = "";

        /**
         * Creates a new PublicMeta instance using the specified properties.
         * @function create
         * @memberof realtime.PublicMeta
         * @static
         * @param {realtime.IPublicMeta=} [properties] Properties to set
         * @returns {realtime.PublicMeta} PublicMeta instance
         */
        PublicMeta.create = function create(properties) {
            return new PublicMeta(properties);
        };

        /**
         * Encodes the specified PublicMeta message. Does not implicitly {@link realtime.PublicMeta.verify|verify} messages.
         * @function encode
         * @memberof realtime.PublicMeta
         * @static
         * @param {realtime.IPublicMeta} message PublicMeta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        PublicMeta.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.leaderboard != null && message.leaderboard.length)
                for (let i = 0; i < message.leaderboard.length; ++i)
                    $root.realtime.LeaderboardEntry.encode(message.leaderboard[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.bossLeaderboard != null && message.bossLeaderboard.length)
                for (let i = 0; i < message.bossLeaderboard.length; ++i)
                    $root.realtime.BossLeaderboardEntry.encode(message.bossLeaderboard[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.announcementVersion != null && Object.hasOwnProperty.call(message, "announcementVersion"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.announcementVersion);
            return writer;
        };

        /**
         * Encodes the specified PublicMeta message, length delimited. Does not implicitly {@link realtime.PublicMeta.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.PublicMeta
         * @static
         * @param {realtime.IPublicMeta} message PublicMeta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        PublicMeta.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a PublicMeta message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.PublicMeta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.PublicMeta} PublicMeta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        PublicMeta.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.PublicMeta();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.leaderboard && message.leaderboard.length))
                            message.leaderboard = [];
                        message.leaderboard.push($root.realtime.LeaderboardEntry.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 2: {
                        if (!(message.bossLeaderboard && message.bossLeaderboard.length))
                            message.bossLeaderboard = [];
                        message.bossLeaderboard.push($root.realtime.BossLeaderboardEntry.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 3: {
                        message.announcementVersion = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a PublicMeta message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.PublicMeta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.PublicMeta} PublicMeta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        PublicMeta.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a PublicMeta message.
         * @function verify
         * @memberof realtime.PublicMeta
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        PublicMeta.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.leaderboard != null && message.hasOwnProperty("leaderboard")) {
                if (!Array.isArray(message.leaderboard))
                    return "leaderboard: array expected";
                for (let i = 0; i < message.leaderboard.length; ++i) {
                    let error = $root.realtime.LeaderboardEntry.verify(message.leaderboard[i], long + 1);
                    if (error)
                        return "leaderboard." + error;
                }
            }
            if (message.bossLeaderboard != null && message.hasOwnProperty("bossLeaderboard")) {
                if (!Array.isArray(message.bossLeaderboard))
                    return "bossLeaderboard: array expected";
                for (let i = 0; i < message.bossLeaderboard.length; ++i) {
                    let error = $root.realtime.BossLeaderboardEntry.verify(message.bossLeaderboard[i], long + 1);
                    if (error)
                        return "bossLeaderboard." + error;
                }
            }
            if (message.announcementVersion != null && message.hasOwnProperty("announcementVersion"))
                if (!$util.isString(message.announcementVersion))
                    return "announcementVersion: string expected";
            return null;
        };

        /**
         * Creates a PublicMeta message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.PublicMeta
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.PublicMeta} PublicMeta
         */
        PublicMeta.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.PublicMeta)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.PublicMeta();
            if (object.leaderboard) {
                if (!Array.isArray(object.leaderboard))
                    throw TypeError(".realtime.PublicMeta.leaderboard: array expected");
                message.leaderboard = [];
                for (let i = 0; i < object.leaderboard.length; ++i) {
                    if (typeof object.leaderboard[i] !== "object")
                        throw TypeError(".realtime.PublicMeta.leaderboard: object expected");
                    message.leaderboard[i] = $root.realtime.LeaderboardEntry.fromObject(object.leaderboard[i], long + 1);
                }
            }
            if (object.bossLeaderboard) {
                if (!Array.isArray(object.bossLeaderboard))
                    throw TypeError(".realtime.PublicMeta.bossLeaderboard: array expected");
                message.bossLeaderboard = [];
                for (let i = 0; i < object.bossLeaderboard.length; ++i) {
                    if (typeof object.bossLeaderboard[i] !== "object")
                        throw TypeError(".realtime.PublicMeta.bossLeaderboard: object expected");
                    message.bossLeaderboard[i] = $root.realtime.BossLeaderboardEntry.fromObject(object.bossLeaderboard[i], long + 1);
                }
            }
            if (object.announcementVersion != null)
                message.announcementVersion = String(object.announcementVersion);
            return message;
        };

        /**
         * Creates a plain object from a PublicMeta message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.PublicMeta
         * @static
         * @param {realtime.PublicMeta} message PublicMeta
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        PublicMeta.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.leaderboard = [];
                object.bossLeaderboard = [];
            }
            if (options.defaults)
                object.announcementVersion = "";
            if (message.leaderboard && message.leaderboard.length) {
                object.leaderboard = [];
                for (let j = 0; j < message.leaderboard.length; ++j)
                    object.leaderboard[j] = $root.realtime.LeaderboardEntry.toObject(message.leaderboard[j], options);
            }
            if (message.bossLeaderboard && message.bossLeaderboard.length) {
                object.bossLeaderboard = [];
                for (let j = 0; j < message.bossLeaderboard.length; ++j)
                    object.bossLeaderboard[j] = $root.realtime.BossLeaderboardEntry.toObject(message.bossLeaderboard[j], options);
            }
            if (message.announcementVersion != null && message.hasOwnProperty("announcementVersion"))
                object.announcementVersion = message.announcementVersion;
            return object;
        };

        /**
         * Converts this PublicMeta to JSON.
         * @function toJSON
         * @memberof realtime.PublicMeta
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        PublicMeta.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for PublicMeta
         * @function getTypeUrl
         * @memberof realtime.PublicMeta
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        PublicMeta.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.PublicMeta";
        };

        return PublicMeta;
    })();

    realtime.RoomInfo = (function() {

        /**
         * Properties of a RoomInfo.
         * @memberof realtime
         * @interface IRoomInfo
         * @property {string|null} [id] RoomInfo id
         * @property {string|null} [displayName] RoomInfo displayName
         * @property {boolean|null} [current] RoomInfo current
         * @property {boolean|null} [joinable] RoomInfo joinable
         * @property {number|null} [onlineCount] RoomInfo onlineCount
         * @property {boolean|null} [cycleEnabled] RoomInfo cycleEnabled
         * @property {string|null} [queueId] RoomInfo queueId
         * @property {string|null} [currentBossId] RoomInfo currentBossId
         * @property {string|null} [currentBossName] RoomInfo currentBossName
         * @property {string|null} [currentBossStatus] RoomInfo currentBossStatus
         * @property {number|Long|null} [currentBossHp] RoomInfo currentBossHp
         * @property {number|Long|null} [currentBossMaxHp] RoomInfo currentBossMaxHp
         * @property {number|Long|null} [currentBossAvgHp] RoomInfo currentBossAvgHp
         * @property {number|Long|null} [cooldownRemainingSeconds] RoomInfo cooldownRemainingSeconds
         */

        /**
         * Constructs a new RoomInfo.
         * @memberof realtime
         * @classdesc Represents a RoomInfo.
         * @implements IRoomInfo
         * @constructor
         * @param {realtime.IRoomInfo=} [properties] Properties to set
         */
        function RoomInfo(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * RoomInfo id.
         * @member {string} id
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.id = "";

        /**
         * RoomInfo displayName.
         * @member {string} displayName
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.displayName = "";

        /**
         * RoomInfo current.
         * @member {boolean} current
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.current = false;

        /**
         * RoomInfo joinable.
         * @member {boolean} joinable
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.joinable = false;

        /**
         * RoomInfo onlineCount.
         * @member {number} onlineCount
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.onlineCount = 0;

        /**
         * RoomInfo cycleEnabled.
         * @member {boolean} cycleEnabled
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.cycleEnabled = false;

        /**
         * RoomInfo queueId.
         * @member {string} queueId
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.queueId = "";

        /**
         * RoomInfo currentBossId.
         * @member {string} currentBossId
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.currentBossId = "";

        /**
         * RoomInfo currentBossName.
         * @member {string} currentBossName
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.currentBossName = "";

        /**
         * RoomInfo currentBossStatus.
         * @member {string} currentBossStatus
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.currentBossStatus = "";

        /**
         * RoomInfo currentBossHp.
         * @member {number|Long} currentBossHp
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.currentBossHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * RoomInfo currentBossMaxHp.
         * @member {number|Long} currentBossMaxHp
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.currentBossMaxHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * RoomInfo currentBossAvgHp.
         * @member {number|Long} currentBossAvgHp
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.currentBossAvgHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * RoomInfo cooldownRemainingSeconds.
         * @member {number|Long} cooldownRemainingSeconds
         * @memberof realtime.RoomInfo
         * @instance
         */
        RoomInfo.prototype.cooldownRemainingSeconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new RoomInfo instance using the specified properties.
         * @function create
         * @memberof realtime.RoomInfo
         * @static
         * @param {realtime.IRoomInfo=} [properties] Properties to set
         * @returns {realtime.RoomInfo} RoomInfo instance
         */
        RoomInfo.create = function create(properties) {
            return new RoomInfo(properties);
        };

        /**
         * Encodes the specified RoomInfo message. Does not implicitly {@link realtime.RoomInfo.verify|verify} messages.
         * @function encode
         * @memberof realtime.RoomInfo
         * @static
         * @param {realtime.IRoomInfo} message RoomInfo message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        RoomInfo.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.displayName != null && Object.hasOwnProperty.call(message, "displayName"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.displayName);
            if (message.current != null && Object.hasOwnProperty.call(message, "current"))
                writer.uint32(/* id 3, wireType 0 =*/24).bool(message.current);
            if (message.joinable != null && Object.hasOwnProperty.call(message, "joinable"))
                writer.uint32(/* id 4, wireType 0 =*/32).bool(message.joinable);
            if (message.onlineCount != null && Object.hasOwnProperty.call(message, "onlineCount"))
                writer.uint32(/* id 5, wireType 0 =*/40).int32(message.onlineCount);
            if (message.cycleEnabled != null && Object.hasOwnProperty.call(message, "cycleEnabled"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.cycleEnabled);
            if (message.queueId != null && Object.hasOwnProperty.call(message, "queueId"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.queueId);
            if (message.currentBossId != null && Object.hasOwnProperty.call(message, "currentBossId"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.currentBossId);
            if (message.currentBossName != null && Object.hasOwnProperty.call(message, "currentBossName"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.currentBossName);
            if (message.currentBossStatus != null && Object.hasOwnProperty.call(message, "currentBossStatus"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.currentBossStatus);
            if (message.currentBossHp != null && Object.hasOwnProperty.call(message, "currentBossHp"))
                writer.uint32(/* id 11, wireType 0 =*/88).int64(message.currentBossHp);
            if (message.currentBossMaxHp != null && Object.hasOwnProperty.call(message, "currentBossMaxHp"))
                writer.uint32(/* id 12, wireType 0 =*/96).int64(message.currentBossMaxHp);
            if (message.currentBossAvgHp != null && Object.hasOwnProperty.call(message, "currentBossAvgHp"))
                writer.uint32(/* id 13, wireType 0 =*/104).int64(message.currentBossAvgHp);
            if (message.cooldownRemainingSeconds != null && Object.hasOwnProperty.call(message, "cooldownRemainingSeconds"))
                writer.uint32(/* id 14, wireType 0 =*/112).int64(message.cooldownRemainingSeconds);
            return writer;
        };

        /**
         * Encodes the specified RoomInfo message, length delimited. Does not implicitly {@link realtime.RoomInfo.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.RoomInfo
         * @static
         * @param {realtime.IRoomInfo} message RoomInfo message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        RoomInfo.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a RoomInfo message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.RoomInfo
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.RoomInfo} RoomInfo
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        RoomInfo.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.RoomInfo();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.displayName = reader.string();
                        break;
                    }
                case 3: {
                        message.current = reader.bool();
                        break;
                    }
                case 4: {
                        message.joinable = reader.bool();
                        break;
                    }
                case 5: {
                        message.onlineCount = reader.int32();
                        break;
                    }
                case 6: {
                        message.cycleEnabled = reader.bool();
                        break;
                    }
                case 7: {
                        message.queueId = reader.string();
                        break;
                    }
                case 8: {
                        message.currentBossId = reader.string();
                        break;
                    }
                case 9: {
                        message.currentBossName = reader.string();
                        break;
                    }
                case 10: {
                        message.currentBossStatus = reader.string();
                        break;
                    }
                case 11: {
                        message.currentBossHp = reader.int64();
                        break;
                    }
                case 12: {
                        message.currentBossMaxHp = reader.int64();
                        break;
                    }
                case 13: {
                        message.currentBossAvgHp = reader.int64();
                        break;
                    }
                case 14: {
                        message.cooldownRemainingSeconds = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a RoomInfo message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.RoomInfo
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.RoomInfo} RoomInfo
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        RoomInfo.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a RoomInfo message.
         * @function verify
         * @memberof realtime.RoomInfo
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        RoomInfo.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.displayName != null && message.hasOwnProperty("displayName"))
                if (!$util.isString(message.displayName))
                    return "displayName: string expected";
            if (message.current != null && message.hasOwnProperty("current"))
                if (typeof message.current !== "boolean")
                    return "current: boolean expected";
            if (message.joinable != null && message.hasOwnProperty("joinable"))
                if (typeof message.joinable !== "boolean")
                    return "joinable: boolean expected";
            if (message.onlineCount != null && message.hasOwnProperty("onlineCount"))
                if (!$util.isInteger(message.onlineCount))
                    return "onlineCount: integer expected";
            if (message.cycleEnabled != null && message.hasOwnProperty("cycleEnabled"))
                if (typeof message.cycleEnabled !== "boolean")
                    return "cycleEnabled: boolean expected";
            if (message.queueId != null && message.hasOwnProperty("queueId"))
                if (!$util.isString(message.queueId))
                    return "queueId: string expected";
            if (message.currentBossId != null && message.hasOwnProperty("currentBossId"))
                if (!$util.isString(message.currentBossId))
                    return "currentBossId: string expected";
            if (message.currentBossName != null && message.hasOwnProperty("currentBossName"))
                if (!$util.isString(message.currentBossName))
                    return "currentBossName: string expected";
            if (message.currentBossStatus != null && message.hasOwnProperty("currentBossStatus"))
                if (!$util.isString(message.currentBossStatus))
                    return "currentBossStatus: string expected";
            if (message.currentBossHp != null && message.hasOwnProperty("currentBossHp"))
                if (!$util.isInteger(message.currentBossHp) && !(message.currentBossHp && $util.isInteger(message.currentBossHp.low) && $util.isInteger(message.currentBossHp.high)))
                    return "currentBossHp: integer|Long expected";
            if (message.currentBossMaxHp != null && message.hasOwnProperty("currentBossMaxHp"))
                if (!$util.isInteger(message.currentBossMaxHp) && !(message.currentBossMaxHp && $util.isInteger(message.currentBossMaxHp.low) && $util.isInteger(message.currentBossMaxHp.high)))
                    return "currentBossMaxHp: integer|Long expected";
            if (message.currentBossAvgHp != null && message.hasOwnProperty("currentBossAvgHp"))
                if (!$util.isInteger(message.currentBossAvgHp) && !(message.currentBossAvgHp && $util.isInteger(message.currentBossAvgHp.low) && $util.isInteger(message.currentBossAvgHp.high)))
                    return "currentBossAvgHp: integer|Long expected";
            if (message.cooldownRemainingSeconds != null && message.hasOwnProperty("cooldownRemainingSeconds"))
                if (!$util.isInteger(message.cooldownRemainingSeconds) && !(message.cooldownRemainingSeconds && $util.isInteger(message.cooldownRemainingSeconds.low) && $util.isInteger(message.cooldownRemainingSeconds.high)))
                    return "cooldownRemainingSeconds: integer|Long expected";
            return null;
        };

        /**
         * Creates a RoomInfo message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.RoomInfo
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.RoomInfo} RoomInfo
         */
        RoomInfo.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.RoomInfo)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.RoomInfo();
            if (object.id != null)
                message.id = String(object.id);
            if (object.displayName != null)
                message.displayName = String(object.displayName);
            if (object.current != null)
                message.current = Boolean(object.current);
            if (object.joinable != null)
                message.joinable = Boolean(object.joinable);
            if (object.onlineCount != null)
                message.onlineCount = object.onlineCount | 0;
            if (object.cycleEnabled != null)
                message.cycleEnabled = Boolean(object.cycleEnabled);
            if (object.queueId != null)
                message.queueId = String(object.queueId);
            if (object.currentBossId != null)
                message.currentBossId = String(object.currentBossId);
            if (object.currentBossName != null)
                message.currentBossName = String(object.currentBossName);
            if (object.currentBossStatus != null)
                message.currentBossStatus = String(object.currentBossStatus);
            if (object.currentBossHp != null)
                if ($util.Long)
                    (message.currentBossHp = $util.Long.fromValue(object.currentBossHp)).unsigned = false;
                else if (typeof object.currentBossHp === "string")
                    message.currentBossHp = parseInt(object.currentBossHp, 10);
                else if (typeof object.currentBossHp === "number")
                    message.currentBossHp = object.currentBossHp;
                else if (typeof object.currentBossHp === "object")
                    message.currentBossHp = new $util.LongBits(object.currentBossHp.low >>> 0, object.currentBossHp.high >>> 0).toNumber();
            if (object.currentBossMaxHp != null)
                if ($util.Long)
                    (message.currentBossMaxHp = $util.Long.fromValue(object.currentBossMaxHp)).unsigned = false;
                else if (typeof object.currentBossMaxHp === "string")
                    message.currentBossMaxHp = parseInt(object.currentBossMaxHp, 10);
                else if (typeof object.currentBossMaxHp === "number")
                    message.currentBossMaxHp = object.currentBossMaxHp;
                else if (typeof object.currentBossMaxHp === "object")
                    message.currentBossMaxHp = new $util.LongBits(object.currentBossMaxHp.low >>> 0, object.currentBossMaxHp.high >>> 0).toNumber();
            if (object.currentBossAvgHp != null)
                if ($util.Long)
                    (message.currentBossAvgHp = $util.Long.fromValue(object.currentBossAvgHp)).unsigned = false;
                else if (typeof object.currentBossAvgHp === "string")
                    message.currentBossAvgHp = parseInt(object.currentBossAvgHp, 10);
                else if (typeof object.currentBossAvgHp === "number")
                    message.currentBossAvgHp = object.currentBossAvgHp;
                else if (typeof object.currentBossAvgHp === "object")
                    message.currentBossAvgHp = new $util.LongBits(object.currentBossAvgHp.low >>> 0, object.currentBossAvgHp.high >>> 0).toNumber();
            if (object.cooldownRemainingSeconds != null)
                if ($util.Long)
                    (message.cooldownRemainingSeconds = $util.Long.fromValue(object.cooldownRemainingSeconds)).unsigned = false;
                else if (typeof object.cooldownRemainingSeconds === "string")
                    message.cooldownRemainingSeconds = parseInt(object.cooldownRemainingSeconds, 10);
                else if (typeof object.cooldownRemainingSeconds === "number")
                    message.cooldownRemainingSeconds = object.cooldownRemainingSeconds;
                else if (typeof object.cooldownRemainingSeconds === "object")
                    message.cooldownRemainingSeconds = new $util.LongBits(object.cooldownRemainingSeconds.low >>> 0, object.cooldownRemainingSeconds.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a RoomInfo message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.RoomInfo
         * @static
         * @param {realtime.RoomInfo} message RoomInfo
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        RoomInfo.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                object.displayName = "";
                object.current = false;
                object.joinable = false;
                object.onlineCount = 0;
                object.cycleEnabled = false;
                object.queueId = "";
                object.currentBossId = "";
                object.currentBossName = "";
                object.currentBossStatus = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.currentBossHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.currentBossHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.currentBossMaxHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.currentBossMaxHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.currentBossAvgHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.currentBossAvgHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.cooldownRemainingSeconds = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.cooldownRemainingSeconds = options.longs === String ? "0" : 0;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.displayName != null && message.hasOwnProperty("displayName"))
                object.displayName = message.displayName;
            if (message.current != null && message.hasOwnProperty("current"))
                object.current = message.current;
            if (message.joinable != null && message.hasOwnProperty("joinable"))
                object.joinable = message.joinable;
            if (message.onlineCount != null && message.hasOwnProperty("onlineCount"))
                object.onlineCount = message.onlineCount;
            if (message.cycleEnabled != null && message.hasOwnProperty("cycleEnabled"))
                object.cycleEnabled = message.cycleEnabled;
            if (message.queueId != null && message.hasOwnProperty("queueId"))
                object.queueId = message.queueId;
            if (message.currentBossId != null && message.hasOwnProperty("currentBossId"))
                object.currentBossId = message.currentBossId;
            if (message.currentBossName != null && message.hasOwnProperty("currentBossName"))
                object.currentBossName = message.currentBossName;
            if (message.currentBossStatus != null && message.hasOwnProperty("currentBossStatus"))
                object.currentBossStatus = message.currentBossStatus;
            if (message.currentBossHp != null && message.hasOwnProperty("currentBossHp"))
                if (typeof message.currentBossHp === "number")
                    object.currentBossHp = options.longs === String ? String(message.currentBossHp) : message.currentBossHp;
                else
                    object.currentBossHp = options.longs === String ? $util.Long.prototype.toString.call(message.currentBossHp) : options.longs === Number ? new $util.LongBits(message.currentBossHp.low >>> 0, message.currentBossHp.high >>> 0).toNumber() : message.currentBossHp;
            if (message.currentBossMaxHp != null && message.hasOwnProperty("currentBossMaxHp"))
                if (typeof message.currentBossMaxHp === "number")
                    object.currentBossMaxHp = options.longs === String ? String(message.currentBossMaxHp) : message.currentBossMaxHp;
                else
                    object.currentBossMaxHp = options.longs === String ? $util.Long.prototype.toString.call(message.currentBossMaxHp) : options.longs === Number ? new $util.LongBits(message.currentBossMaxHp.low >>> 0, message.currentBossMaxHp.high >>> 0).toNumber() : message.currentBossMaxHp;
            if (message.currentBossAvgHp != null && message.hasOwnProperty("currentBossAvgHp"))
                if (typeof message.currentBossAvgHp === "number")
                    object.currentBossAvgHp = options.longs === String ? String(message.currentBossAvgHp) : message.currentBossAvgHp;
                else
                    object.currentBossAvgHp = options.longs === String ? $util.Long.prototype.toString.call(message.currentBossAvgHp) : options.longs === Number ? new $util.LongBits(message.currentBossAvgHp.low >>> 0, message.currentBossAvgHp.high >>> 0).toNumber() : message.currentBossAvgHp;
            if (message.cooldownRemainingSeconds != null && message.hasOwnProperty("cooldownRemainingSeconds"))
                if (typeof message.cooldownRemainingSeconds === "number")
                    object.cooldownRemainingSeconds = options.longs === String ? String(message.cooldownRemainingSeconds) : message.cooldownRemainingSeconds;
                else
                    object.cooldownRemainingSeconds = options.longs === String ? $util.Long.prototype.toString.call(message.cooldownRemainingSeconds) : options.longs === Number ? new $util.LongBits(message.cooldownRemainingSeconds.low >>> 0, message.cooldownRemainingSeconds.high >>> 0).toNumber() : message.cooldownRemainingSeconds;
            return object;
        };

        /**
         * Converts this RoomInfo to JSON.
         * @function toJSON
         * @memberof realtime.RoomInfo
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        RoomInfo.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for RoomInfo
         * @function getTypeUrl
         * @memberof realtime.RoomInfo
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        RoomInfo.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.RoomInfo";
        };

        return RoomInfo;
    })();

    realtime.RoomState = (function() {

        /**
         * Properties of a RoomState.
         * @memberof realtime
         * @interface IRoomState
         * @property {string|null} [currentRoomId] RoomState currentRoomId
         * @property {number|Long|null} [switchCooldownRemainingSeconds] RoomState switchCooldownRemainingSeconds
         * @property {Array.<realtime.IRoomInfo>|null} [rooms] RoomState rooms
         */

        /**
         * Constructs a new RoomState.
         * @memberof realtime
         * @classdesc Represents a RoomState.
         * @implements IRoomState
         * @constructor
         * @param {realtime.IRoomState=} [properties] Properties to set
         */
        function RoomState(properties) {
            this.rooms = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * RoomState currentRoomId.
         * @member {string} currentRoomId
         * @memberof realtime.RoomState
         * @instance
         */
        RoomState.prototype.currentRoomId = "";

        /**
         * RoomState switchCooldownRemainingSeconds.
         * @member {number|Long} switchCooldownRemainingSeconds
         * @memberof realtime.RoomState
         * @instance
         */
        RoomState.prototype.switchCooldownRemainingSeconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * RoomState rooms.
         * @member {Array.<realtime.IRoomInfo>} rooms
         * @memberof realtime.RoomState
         * @instance
         */
        RoomState.prototype.rooms = $util.emptyArray;

        /**
         * Creates a new RoomState instance using the specified properties.
         * @function create
         * @memberof realtime.RoomState
         * @static
         * @param {realtime.IRoomState=} [properties] Properties to set
         * @returns {realtime.RoomState} RoomState instance
         */
        RoomState.create = function create(properties) {
            return new RoomState(properties);
        };

        /**
         * Encodes the specified RoomState message. Does not implicitly {@link realtime.RoomState.verify|verify} messages.
         * @function encode
         * @memberof realtime.RoomState
         * @static
         * @param {realtime.IRoomState} message RoomState message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        RoomState.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.currentRoomId != null && Object.hasOwnProperty.call(message, "currentRoomId"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.currentRoomId);
            if (message.switchCooldownRemainingSeconds != null && Object.hasOwnProperty.call(message, "switchCooldownRemainingSeconds"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.switchCooldownRemainingSeconds);
            if (message.rooms != null && message.rooms.length)
                for (let i = 0; i < message.rooms.length; ++i)
                    $root.realtime.RoomInfo.encode(message.rooms[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified RoomState message, length delimited. Does not implicitly {@link realtime.RoomState.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.RoomState
         * @static
         * @param {realtime.IRoomState} message RoomState message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        RoomState.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a RoomState message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.RoomState
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.RoomState} RoomState
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        RoomState.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.RoomState();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.currentRoomId = reader.string();
                        break;
                    }
                case 2: {
                        message.switchCooldownRemainingSeconds = reader.int64();
                        break;
                    }
                case 3: {
                        if (!(message.rooms && message.rooms.length))
                            message.rooms = [];
                        message.rooms.push($root.realtime.RoomInfo.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a RoomState message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.RoomState
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.RoomState} RoomState
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        RoomState.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a RoomState message.
         * @function verify
         * @memberof realtime.RoomState
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        RoomState.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.currentRoomId != null && message.hasOwnProperty("currentRoomId"))
                if (!$util.isString(message.currentRoomId))
                    return "currentRoomId: string expected";
            if (message.switchCooldownRemainingSeconds != null && message.hasOwnProperty("switchCooldownRemainingSeconds"))
                if (!$util.isInteger(message.switchCooldownRemainingSeconds) && !(message.switchCooldownRemainingSeconds && $util.isInteger(message.switchCooldownRemainingSeconds.low) && $util.isInteger(message.switchCooldownRemainingSeconds.high)))
                    return "switchCooldownRemainingSeconds: integer|Long expected";
            if (message.rooms != null && message.hasOwnProperty("rooms")) {
                if (!Array.isArray(message.rooms))
                    return "rooms: array expected";
                for (let i = 0; i < message.rooms.length; ++i) {
                    let error = $root.realtime.RoomInfo.verify(message.rooms[i], long + 1);
                    if (error)
                        return "rooms." + error;
                }
            }
            return null;
        };

        /**
         * Creates a RoomState message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.RoomState
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.RoomState} RoomState
         */
        RoomState.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.RoomState)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.RoomState();
            if (object.currentRoomId != null)
                message.currentRoomId = String(object.currentRoomId);
            if (object.switchCooldownRemainingSeconds != null)
                if ($util.Long)
                    (message.switchCooldownRemainingSeconds = $util.Long.fromValue(object.switchCooldownRemainingSeconds)).unsigned = false;
                else if (typeof object.switchCooldownRemainingSeconds === "string")
                    message.switchCooldownRemainingSeconds = parseInt(object.switchCooldownRemainingSeconds, 10);
                else if (typeof object.switchCooldownRemainingSeconds === "number")
                    message.switchCooldownRemainingSeconds = object.switchCooldownRemainingSeconds;
                else if (typeof object.switchCooldownRemainingSeconds === "object")
                    message.switchCooldownRemainingSeconds = new $util.LongBits(object.switchCooldownRemainingSeconds.low >>> 0, object.switchCooldownRemainingSeconds.high >>> 0).toNumber();
            if (object.rooms) {
                if (!Array.isArray(object.rooms))
                    throw TypeError(".realtime.RoomState.rooms: array expected");
                message.rooms = [];
                for (let i = 0; i < object.rooms.length; ++i) {
                    if (typeof object.rooms[i] !== "object")
                        throw TypeError(".realtime.RoomState.rooms: object expected");
                    message.rooms[i] = $root.realtime.RoomInfo.fromObject(object.rooms[i], long + 1);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a RoomState message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.RoomState
         * @static
         * @param {realtime.RoomState} message RoomState
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        RoomState.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.rooms = [];
            if (options.defaults) {
                object.currentRoomId = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.switchCooldownRemainingSeconds = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.switchCooldownRemainingSeconds = options.longs === String ? "0" : 0;
            }
            if (message.currentRoomId != null && message.hasOwnProperty("currentRoomId"))
                object.currentRoomId = message.currentRoomId;
            if (message.switchCooldownRemainingSeconds != null && message.hasOwnProperty("switchCooldownRemainingSeconds"))
                if (typeof message.switchCooldownRemainingSeconds === "number")
                    object.switchCooldownRemainingSeconds = options.longs === String ? String(message.switchCooldownRemainingSeconds) : message.switchCooldownRemainingSeconds;
                else
                    object.switchCooldownRemainingSeconds = options.longs === String ? $util.Long.prototype.toString.call(message.switchCooldownRemainingSeconds) : options.longs === Number ? new $util.LongBits(message.switchCooldownRemainingSeconds.low >>> 0, message.switchCooldownRemainingSeconds.high >>> 0).toNumber() : message.switchCooldownRemainingSeconds;
            if (message.rooms && message.rooms.length) {
                object.rooms = [];
                for (let j = 0; j < message.rooms.length; ++j)
                    object.rooms[j] = $root.realtime.RoomInfo.toObject(message.rooms[j], options);
            }
            return object;
        };

        /**
         * Converts this RoomState to JSON.
         * @function toJSON
         * @memberof realtime.RoomState
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        RoomState.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for RoomState
         * @function getTypeUrl
         * @memberof realtime.RoomState
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        RoomState.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.RoomState";
        };

        return RoomState;
    })();

    realtime.UserDelta = (function() {

        /**
         * Properties of a UserDelta.
         * @memberof realtime
         * @interface IUserDelta
         * @property {realtime.IUserStats|null} [userStats] UserDelta userStats
         * @property {realtime.IBossUserStats|null} [myBossStats] UserDelta myBossStats
         * @property {number|Long|null} [myBossKills] UserDelta myBossKills
         * @property {number|Long|null} [totalBossKills] UserDelta totalBossKills
         * @property {string|null} [roomId] UserDelta roomId
         * @property {realtime.ILoadout|null} [loadout] UserDelta loadout
         * @property {realtime.ICombatStats|null} [combatStats] UserDelta combatStats
         * @property {number|Long|null} [gold] UserDelta gold
         * @property {number|Long|null} [stones] UserDelta stones
         * @property {number|Long|null} [talentPoints] UserDelta talentPoints
         * @property {Array.<realtime.IReward>|null} [recentRewards] UserDelta recentRewards
         * @property {Array.<realtime.ITalentTriggerEvent>|null} [talentEvents] UserDelta talentEvents
         * @property {realtime.ITalentCombatState|null} [talentCombatState] UserDelta talentCombatState
         * @property {string|null} [equippedBattleClickSkinId] UserDelta equippedBattleClickSkinId
         * @property {string|null} [equippedBattleClickCursorImagePath] UserDelta equippedBattleClickCursorImagePath
         */

        /**
         * Constructs a new UserDelta.
         * @memberof realtime
         * @classdesc Represents a UserDelta.
         * @implements IUserDelta
         * @constructor
         * @param {realtime.IUserDelta=} [properties] Properties to set
         */
        function UserDelta(properties) {
            this.recentRewards = [];
            this.talentEvents = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * UserDelta userStats.
         * @member {realtime.IUserStats|null|undefined} userStats
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.userStats = null;

        /**
         * UserDelta myBossStats.
         * @member {realtime.IBossUserStats|null|undefined} myBossStats
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.myBossStats = null;

        /**
         * UserDelta myBossKills.
         * @member {number|Long} myBossKills
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.myBossKills = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserDelta totalBossKills.
         * @member {number|Long} totalBossKills
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.totalBossKills = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserDelta roomId.
         * @member {string} roomId
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.roomId = "";

        /**
         * UserDelta loadout.
         * @member {realtime.ILoadout|null|undefined} loadout
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.loadout = null;

        /**
         * UserDelta combatStats.
         * @member {realtime.ICombatStats|null|undefined} combatStats
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.combatStats = null;

        /**
         * UserDelta gold.
         * @member {number|Long} gold
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.gold = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserDelta stones.
         * @member {number|Long} stones
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.stones = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserDelta talentPoints.
         * @member {number|Long} talentPoints
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.talentPoints = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserDelta recentRewards.
         * @member {Array.<realtime.IReward>} recentRewards
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.recentRewards = $util.emptyArray;

        /**
         * UserDelta talentEvents.
         * @member {Array.<realtime.ITalentTriggerEvent>} talentEvents
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.talentEvents = $util.emptyArray;

        /**
         * UserDelta talentCombatState.
         * @member {realtime.ITalentCombatState|null|undefined} talentCombatState
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.talentCombatState = null;

        /**
         * UserDelta equippedBattleClickSkinId.
         * @member {string} equippedBattleClickSkinId
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.equippedBattleClickSkinId = "";

        /**
         * UserDelta equippedBattleClickCursorImagePath.
         * @member {string} equippedBattleClickCursorImagePath
         * @memberof realtime.UserDelta
         * @instance
         */
        UserDelta.prototype.equippedBattleClickCursorImagePath = "";

        /**
         * Creates a new UserDelta instance using the specified properties.
         * @function create
         * @memberof realtime.UserDelta
         * @static
         * @param {realtime.IUserDelta=} [properties] Properties to set
         * @returns {realtime.UserDelta} UserDelta instance
         */
        UserDelta.create = function create(properties) {
            return new UserDelta(properties);
        };

        /**
         * Encodes the specified UserDelta message. Does not implicitly {@link realtime.UserDelta.verify|verify} messages.
         * @function encode
         * @memberof realtime.UserDelta
         * @static
         * @param {realtime.IUserDelta} message UserDelta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserDelta.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.userStats != null && Object.hasOwnProperty.call(message, "userStats"))
                $root.realtime.UserStats.encode(message.userStats, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.myBossStats != null && Object.hasOwnProperty.call(message, "myBossStats"))
                $root.realtime.BossUserStats.encode(message.myBossStats, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.myBossKills != null && Object.hasOwnProperty.call(message, "myBossKills"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.myBossKills);
            if (message.totalBossKills != null && Object.hasOwnProperty.call(message, "totalBossKills"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.totalBossKills);
            if (message.roomId != null && Object.hasOwnProperty.call(message, "roomId"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.roomId);
            if (message.loadout != null && Object.hasOwnProperty.call(message, "loadout"))
                $root.realtime.Loadout.encode(message.loadout, writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
            if (message.combatStats != null && Object.hasOwnProperty.call(message, "combatStats"))
                $root.realtime.CombatStats.encode(message.combatStats, writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
            if (message.gold != null && Object.hasOwnProperty.call(message, "gold"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.gold);
            if (message.stones != null && Object.hasOwnProperty.call(message, "stones"))
                writer.uint32(/* id 9, wireType 0 =*/72).int64(message.stones);
            if (message.talentPoints != null && Object.hasOwnProperty.call(message, "talentPoints"))
                writer.uint32(/* id 10, wireType 0 =*/80).int64(message.talentPoints);
            if (message.recentRewards != null && message.recentRewards.length)
                for (let i = 0; i < message.recentRewards.length; ++i)
                    $root.realtime.Reward.encode(message.recentRewards[i], writer.uint32(/* id 11, wireType 2 =*/90).fork()).ldelim();
            if (message.talentEvents != null && message.talentEvents.length)
                for (let i = 0; i < message.talentEvents.length; ++i)
                    $root.realtime.TalentTriggerEvent.encode(message.talentEvents[i], writer.uint32(/* id 12, wireType 2 =*/98).fork()).ldelim();
            if (message.talentCombatState != null && Object.hasOwnProperty.call(message, "talentCombatState"))
                $root.realtime.TalentCombatState.encode(message.talentCombatState, writer.uint32(/* id 13, wireType 2 =*/106).fork()).ldelim();
            if (message.equippedBattleClickSkinId != null && Object.hasOwnProperty.call(message, "equippedBattleClickSkinId"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.equippedBattleClickSkinId);
            if (message.equippedBattleClickCursorImagePath != null && Object.hasOwnProperty.call(message, "equippedBattleClickCursorImagePath"))
                writer.uint32(/* id 15, wireType 2 =*/122).string(message.equippedBattleClickCursorImagePath);
            return writer;
        };

        /**
         * Encodes the specified UserDelta message, length delimited. Does not implicitly {@link realtime.UserDelta.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.UserDelta
         * @static
         * @param {realtime.IUserDelta} message UserDelta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserDelta.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a UserDelta message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.UserDelta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.UserDelta} UserDelta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserDelta.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.UserDelta();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.userStats = $root.realtime.UserStats.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 2: {
                        message.myBossStats = $root.realtime.BossUserStats.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 3: {
                        message.myBossKills = reader.int64();
                        break;
                    }
                case 4: {
                        message.totalBossKills = reader.int64();
                        break;
                    }
                case 5: {
                        message.roomId = reader.string();
                        break;
                    }
                case 6: {
                        message.loadout = $root.realtime.Loadout.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 7: {
                        message.combatStats = $root.realtime.CombatStats.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 8: {
                        message.gold = reader.int64();
                        break;
                    }
                case 9: {
                        message.stones = reader.int64();
                        break;
                    }
                case 10: {
                        message.talentPoints = reader.int64();
                        break;
                    }
                case 11: {
                        if (!(message.recentRewards && message.recentRewards.length))
                            message.recentRewards = [];
                        message.recentRewards.push($root.realtime.Reward.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 12: {
                        if (!(message.talentEvents && message.talentEvents.length))
                            message.talentEvents = [];
                        message.talentEvents.push($root.realtime.TalentTriggerEvent.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 13: {
                        message.talentCombatState = $root.realtime.TalentCombatState.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 14: {
                        message.equippedBattleClickSkinId = reader.string();
                        break;
                    }
                case 15: {
                        message.equippedBattleClickCursorImagePath = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a UserDelta message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.UserDelta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.UserDelta} UserDelta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserDelta.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a UserDelta message.
         * @function verify
         * @memberof realtime.UserDelta
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        UserDelta.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.userStats != null && message.hasOwnProperty("userStats")) {
                let error = $root.realtime.UserStats.verify(message.userStats, long + 1);
                if (error)
                    return "userStats." + error;
            }
            if (message.myBossStats != null && message.hasOwnProperty("myBossStats")) {
                let error = $root.realtime.BossUserStats.verify(message.myBossStats, long + 1);
                if (error)
                    return "myBossStats." + error;
            }
            if (message.myBossKills != null && message.hasOwnProperty("myBossKills"))
                if (!$util.isInteger(message.myBossKills) && !(message.myBossKills && $util.isInteger(message.myBossKills.low) && $util.isInteger(message.myBossKills.high)))
                    return "myBossKills: integer|Long expected";
            if (message.totalBossKills != null && message.hasOwnProperty("totalBossKills"))
                if (!$util.isInteger(message.totalBossKills) && !(message.totalBossKills && $util.isInteger(message.totalBossKills.low) && $util.isInteger(message.totalBossKills.high)))
                    return "totalBossKills: integer|Long expected";
            if (message.roomId != null && message.hasOwnProperty("roomId"))
                if (!$util.isString(message.roomId))
                    return "roomId: string expected";
            if (message.loadout != null && message.hasOwnProperty("loadout")) {
                let error = $root.realtime.Loadout.verify(message.loadout, long + 1);
                if (error)
                    return "loadout." + error;
            }
            if (message.combatStats != null && message.hasOwnProperty("combatStats")) {
                let error = $root.realtime.CombatStats.verify(message.combatStats, long + 1);
                if (error)
                    return "combatStats." + error;
            }
            if (message.gold != null && message.hasOwnProperty("gold"))
                if (!$util.isInteger(message.gold) && !(message.gold && $util.isInteger(message.gold.low) && $util.isInteger(message.gold.high)))
                    return "gold: integer|Long expected";
            if (message.stones != null && message.hasOwnProperty("stones"))
                if (!$util.isInteger(message.stones) && !(message.stones && $util.isInteger(message.stones.low) && $util.isInteger(message.stones.high)))
                    return "stones: integer|Long expected";
            if (message.talentPoints != null && message.hasOwnProperty("talentPoints"))
                if (!$util.isInteger(message.talentPoints) && !(message.talentPoints && $util.isInteger(message.talentPoints.low) && $util.isInteger(message.talentPoints.high)))
                    return "talentPoints: integer|Long expected";
            if (message.recentRewards != null && message.hasOwnProperty("recentRewards")) {
                if (!Array.isArray(message.recentRewards))
                    return "recentRewards: array expected";
                for (let i = 0; i < message.recentRewards.length; ++i) {
                    let error = $root.realtime.Reward.verify(message.recentRewards[i], long + 1);
                    if (error)
                        return "recentRewards." + error;
                }
            }
            if (message.talentEvents != null && message.hasOwnProperty("talentEvents")) {
                if (!Array.isArray(message.talentEvents))
                    return "talentEvents: array expected";
                for (let i = 0; i < message.talentEvents.length; ++i) {
                    let error = $root.realtime.TalentTriggerEvent.verify(message.talentEvents[i], long + 1);
                    if (error)
                        return "talentEvents." + error;
                }
            }
            if (message.talentCombatState != null && message.hasOwnProperty("talentCombatState")) {
                let error = $root.realtime.TalentCombatState.verify(message.talentCombatState, long + 1);
                if (error)
                    return "talentCombatState." + error;
            }
            if (message.equippedBattleClickSkinId != null && message.hasOwnProperty("equippedBattleClickSkinId"))
                if (!$util.isString(message.equippedBattleClickSkinId))
                    return "equippedBattleClickSkinId: string expected";
            if (message.equippedBattleClickCursorImagePath != null && message.hasOwnProperty("equippedBattleClickCursorImagePath"))
                if (!$util.isString(message.equippedBattleClickCursorImagePath))
                    return "equippedBattleClickCursorImagePath: string expected";
            return null;
        };

        /**
         * Creates a UserDelta message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.UserDelta
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.UserDelta} UserDelta
         */
        UserDelta.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.UserDelta)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.UserDelta();
            if (object.userStats != null) {
                if (typeof object.userStats !== "object")
                    throw TypeError(".realtime.UserDelta.userStats: object expected");
                message.userStats = $root.realtime.UserStats.fromObject(object.userStats, long + 1);
            }
            if (object.myBossStats != null) {
                if (typeof object.myBossStats !== "object")
                    throw TypeError(".realtime.UserDelta.myBossStats: object expected");
                message.myBossStats = $root.realtime.BossUserStats.fromObject(object.myBossStats, long + 1);
            }
            if (object.myBossKills != null)
                if ($util.Long)
                    (message.myBossKills = $util.Long.fromValue(object.myBossKills)).unsigned = false;
                else if (typeof object.myBossKills === "string")
                    message.myBossKills = parseInt(object.myBossKills, 10);
                else if (typeof object.myBossKills === "number")
                    message.myBossKills = object.myBossKills;
                else if (typeof object.myBossKills === "object")
                    message.myBossKills = new $util.LongBits(object.myBossKills.low >>> 0, object.myBossKills.high >>> 0).toNumber();
            if (object.totalBossKills != null)
                if ($util.Long)
                    (message.totalBossKills = $util.Long.fromValue(object.totalBossKills)).unsigned = false;
                else if (typeof object.totalBossKills === "string")
                    message.totalBossKills = parseInt(object.totalBossKills, 10);
                else if (typeof object.totalBossKills === "number")
                    message.totalBossKills = object.totalBossKills;
                else if (typeof object.totalBossKills === "object")
                    message.totalBossKills = new $util.LongBits(object.totalBossKills.low >>> 0, object.totalBossKills.high >>> 0).toNumber();
            if (object.roomId != null)
                message.roomId = String(object.roomId);
            if (object.loadout != null) {
                if (typeof object.loadout !== "object")
                    throw TypeError(".realtime.UserDelta.loadout: object expected");
                message.loadout = $root.realtime.Loadout.fromObject(object.loadout, long + 1);
            }
            if (object.combatStats != null) {
                if (typeof object.combatStats !== "object")
                    throw TypeError(".realtime.UserDelta.combatStats: object expected");
                message.combatStats = $root.realtime.CombatStats.fromObject(object.combatStats, long + 1);
            }
            if (object.gold != null)
                if ($util.Long)
                    (message.gold = $util.Long.fromValue(object.gold)).unsigned = false;
                else if (typeof object.gold === "string")
                    message.gold = parseInt(object.gold, 10);
                else if (typeof object.gold === "number")
                    message.gold = object.gold;
                else if (typeof object.gold === "object")
                    message.gold = new $util.LongBits(object.gold.low >>> 0, object.gold.high >>> 0).toNumber();
            if (object.stones != null)
                if ($util.Long)
                    (message.stones = $util.Long.fromValue(object.stones)).unsigned = false;
                else if (typeof object.stones === "string")
                    message.stones = parseInt(object.stones, 10);
                else if (typeof object.stones === "number")
                    message.stones = object.stones;
                else if (typeof object.stones === "object")
                    message.stones = new $util.LongBits(object.stones.low >>> 0, object.stones.high >>> 0).toNumber();
            if (object.talentPoints != null)
                if ($util.Long)
                    (message.talentPoints = $util.Long.fromValue(object.talentPoints)).unsigned = false;
                else if (typeof object.talentPoints === "string")
                    message.talentPoints = parseInt(object.talentPoints, 10);
                else if (typeof object.talentPoints === "number")
                    message.talentPoints = object.talentPoints;
                else if (typeof object.talentPoints === "object")
                    message.talentPoints = new $util.LongBits(object.talentPoints.low >>> 0, object.talentPoints.high >>> 0).toNumber();
            if (object.recentRewards) {
                if (!Array.isArray(object.recentRewards))
                    throw TypeError(".realtime.UserDelta.recentRewards: array expected");
                message.recentRewards = [];
                for (let i = 0; i < object.recentRewards.length; ++i) {
                    if (typeof object.recentRewards[i] !== "object")
                        throw TypeError(".realtime.UserDelta.recentRewards: object expected");
                    message.recentRewards[i] = $root.realtime.Reward.fromObject(object.recentRewards[i], long + 1);
                }
            }
            if (object.talentEvents) {
                if (!Array.isArray(object.talentEvents))
                    throw TypeError(".realtime.UserDelta.talentEvents: array expected");
                message.talentEvents = [];
                for (let i = 0; i < object.talentEvents.length; ++i) {
                    if (typeof object.talentEvents[i] !== "object")
                        throw TypeError(".realtime.UserDelta.talentEvents: object expected");
                    message.talentEvents[i] = $root.realtime.TalentTriggerEvent.fromObject(object.talentEvents[i], long + 1);
                }
            }
            if (object.talentCombatState != null) {
                if (typeof object.talentCombatState !== "object")
                    throw TypeError(".realtime.UserDelta.talentCombatState: object expected");
                message.talentCombatState = $root.realtime.TalentCombatState.fromObject(object.talentCombatState, long + 1);
            }
            if (object.equippedBattleClickSkinId != null)
                message.equippedBattleClickSkinId = String(object.equippedBattleClickSkinId);
            if (object.equippedBattleClickCursorImagePath != null)
                message.equippedBattleClickCursorImagePath = String(object.equippedBattleClickCursorImagePath);
            return message;
        };

        /**
         * Creates a plain object from a UserDelta message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.UserDelta
         * @static
         * @param {realtime.UserDelta} message UserDelta
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        UserDelta.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.recentRewards = [];
                object.talentEvents = [];
            }
            if (options.defaults) {
                object.userStats = null;
                object.myBossStats = null;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.myBossKills = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.myBossKills = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.totalBossKills = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.totalBossKills = options.longs === String ? "0" : 0;
                object.roomId = "";
                object.loadout = null;
                object.combatStats = null;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.gold = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.gold = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.stones = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.stones = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.talentPoints = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.talentPoints = options.longs === String ? "0" : 0;
                object.talentCombatState = null;
                object.equippedBattleClickSkinId = "";
                object.equippedBattleClickCursorImagePath = "";
            }
            if (message.userStats != null && message.hasOwnProperty("userStats"))
                object.userStats = $root.realtime.UserStats.toObject(message.userStats, options);
            if (message.myBossStats != null && message.hasOwnProperty("myBossStats"))
                object.myBossStats = $root.realtime.BossUserStats.toObject(message.myBossStats, options);
            if (message.myBossKills != null && message.hasOwnProperty("myBossKills"))
                if (typeof message.myBossKills === "number")
                    object.myBossKills = options.longs === String ? String(message.myBossKills) : message.myBossKills;
                else
                    object.myBossKills = options.longs === String ? $util.Long.prototype.toString.call(message.myBossKills) : options.longs === Number ? new $util.LongBits(message.myBossKills.low >>> 0, message.myBossKills.high >>> 0).toNumber() : message.myBossKills;
            if (message.totalBossKills != null && message.hasOwnProperty("totalBossKills"))
                if (typeof message.totalBossKills === "number")
                    object.totalBossKills = options.longs === String ? String(message.totalBossKills) : message.totalBossKills;
                else
                    object.totalBossKills = options.longs === String ? $util.Long.prototype.toString.call(message.totalBossKills) : options.longs === Number ? new $util.LongBits(message.totalBossKills.low >>> 0, message.totalBossKills.high >>> 0).toNumber() : message.totalBossKills;
            if (message.roomId != null && message.hasOwnProperty("roomId"))
                object.roomId = message.roomId;
            if (message.loadout != null && message.hasOwnProperty("loadout"))
                object.loadout = $root.realtime.Loadout.toObject(message.loadout, options);
            if (message.combatStats != null && message.hasOwnProperty("combatStats"))
                object.combatStats = $root.realtime.CombatStats.toObject(message.combatStats, options);
            if (message.gold != null && message.hasOwnProperty("gold"))
                if (typeof message.gold === "number")
                    object.gold = options.longs === String ? String(message.gold) : message.gold;
                else
                    object.gold = options.longs === String ? $util.Long.prototype.toString.call(message.gold) : options.longs === Number ? new $util.LongBits(message.gold.low >>> 0, message.gold.high >>> 0).toNumber() : message.gold;
            if (message.stones != null && message.hasOwnProperty("stones"))
                if (typeof message.stones === "number")
                    object.stones = options.longs === String ? String(message.stones) : message.stones;
                else
                    object.stones = options.longs === String ? $util.Long.prototype.toString.call(message.stones) : options.longs === Number ? new $util.LongBits(message.stones.low >>> 0, message.stones.high >>> 0).toNumber() : message.stones;
            if (message.talentPoints != null && message.hasOwnProperty("talentPoints"))
                if (typeof message.talentPoints === "number")
                    object.talentPoints = options.longs === String ? String(message.talentPoints) : message.talentPoints;
                else
                    object.talentPoints = options.longs === String ? $util.Long.prototype.toString.call(message.talentPoints) : options.longs === Number ? new $util.LongBits(message.talentPoints.low >>> 0, message.talentPoints.high >>> 0).toNumber() : message.talentPoints;
            if (message.recentRewards && message.recentRewards.length) {
                object.recentRewards = [];
                for (let j = 0; j < message.recentRewards.length; ++j)
                    object.recentRewards[j] = $root.realtime.Reward.toObject(message.recentRewards[j], options);
            }
            if (message.talentEvents && message.talentEvents.length) {
                object.talentEvents = [];
                for (let j = 0; j < message.talentEvents.length; ++j)
                    object.talentEvents[j] = $root.realtime.TalentTriggerEvent.toObject(message.talentEvents[j], options);
            }
            if (message.talentCombatState != null && message.hasOwnProperty("talentCombatState"))
                object.talentCombatState = $root.realtime.TalentCombatState.toObject(message.talentCombatState, options);
            if (message.equippedBattleClickSkinId != null && message.hasOwnProperty("equippedBattleClickSkinId"))
                object.equippedBattleClickSkinId = message.equippedBattleClickSkinId;
            if (message.equippedBattleClickCursorImagePath != null && message.hasOwnProperty("equippedBattleClickCursorImagePath"))
                object.equippedBattleClickCursorImagePath = message.equippedBattleClickCursorImagePath;
            return object;
        };

        /**
         * Converts this UserDelta to JSON.
         * @function toJSON
         * @memberof realtime.UserDelta
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        UserDelta.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for UserDelta
         * @function getTypeUrl
         * @memberof realtime.UserDelta
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        UserDelta.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.UserDelta";
        };

        return UserDelta;
    })();

    realtime.UserStats = (function() {

        /**
         * Properties of a UserStats.
         * @memberof realtime
         * @interface IUserStats
         * @property {string|null} [nickname] UserStats nickname
         * @property {number|Long|null} [clickCount] UserStats clickCount
         */

        /**
         * Constructs a new UserStats.
         * @memberof realtime
         * @classdesc Represents a UserStats.
         * @implements IUserStats
         * @constructor
         * @param {realtime.IUserStats=} [properties] Properties to set
         */
        function UserStats(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * UserStats nickname.
         * @member {string} nickname
         * @memberof realtime.UserStats
         * @instance
         */
        UserStats.prototype.nickname = "";

        /**
         * UserStats clickCount.
         * @member {number|Long} clickCount
         * @memberof realtime.UserStats
         * @instance
         */
        UserStats.prototype.clickCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new UserStats instance using the specified properties.
         * @function create
         * @memberof realtime.UserStats
         * @static
         * @param {realtime.IUserStats=} [properties] Properties to set
         * @returns {realtime.UserStats} UserStats instance
         */
        UserStats.create = function create(properties) {
            return new UserStats(properties);
        };

        /**
         * Encodes the specified UserStats message. Does not implicitly {@link realtime.UserStats.verify|verify} messages.
         * @function encode
         * @memberof realtime.UserStats
         * @static
         * @param {realtime.IUserStats} message UserStats message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserStats.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.nickname != null && Object.hasOwnProperty.call(message, "nickname"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.nickname);
            if (message.clickCount != null && Object.hasOwnProperty.call(message, "clickCount"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.clickCount);
            return writer;
        };

        /**
         * Encodes the specified UserStats message, length delimited. Does not implicitly {@link realtime.UserStats.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.UserStats
         * @static
         * @param {realtime.IUserStats} message UserStats message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserStats.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a UserStats message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.UserStats
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.UserStats} UserStats
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserStats.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.UserStats();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.nickname = reader.string();
                        break;
                    }
                case 2: {
                        message.clickCount = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a UserStats message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.UserStats
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.UserStats} UserStats
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserStats.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a UserStats message.
         * @function verify
         * @memberof realtime.UserStats
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        UserStats.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                if (!$util.isString(message.nickname))
                    return "nickname: string expected";
            if (message.clickCount != null && message.hasOwnProperty("clickCount"))
                if (!$util.isInteger(message.clickCount) && !(message.clickCount && $util.isInteger(message.clickCount.low) && $util.isInteger(message.clickCount.high)))
                    return "clickCount: integer|Long expected";
            return null;
        };

        /**
         * Creates a UserStats message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.UserStats
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.UserStats} UserStats
         */
        UserStats.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.UserStats)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.UserStats();
            if (object.nickname != null)
                message.nickname = String(object.nickname);
            if (object.clickCount != null)
                if ($util.Long)
                    (message.clickCount = $util.Long.fromValue(object.clickCount)).unsigned = false;
                else if (typeof object.clickCount === "string")
                    message.clickCount = parseInt(object.clickCount, 10);
                else if (typeof object.clickCount === "number")
                    message.clickCount = object.clickCount;
                else if (typeof object.clickCount === "object")
                    message.clickCount = new $util.LongBits(object.clickCount.low >>> 0, object.clickCount.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a UserStats message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.UserStats
         * @static
         * @param {realtime.UserStats} message UserStats
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        UserStats.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.nickname = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.clickCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.clickCount = options.longs === String ? "0" : 0;
            }
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                object.nickname = message.nickname;
            if (message.clickCount != null && message.hasOwnProperty("clickCount"))
                if (typeof message.clickCount === "number")
                    object.clickCount = options.longs === String ? String(message.clickCount) : message.clickCount;
                else
                    object.clickCount = options.longs === String ? $util.Long.prototype.toString.call(message.clickCount) : options.longs === Number ? new $util.LongBits(message.clickCount.low >>> 0, message.clickCount.high >>> 0).toNumber() : message.clickCount;
            return object;
        };

        /**
         * Converts this UserStats to JSON.
         * @function toJSON
         * @memberof realtime.UserStats
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        UserStats.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for UserStats
         * @function getTypeUrl
         * @memberof realtime.UserStats
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        UserStats.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.UserStats";
        };

        return UserStats;
    })();

    realtime.LeaderboardEntry = (function() {

        /**
         * Properties of a LeaderboardEntry.
         * @memberof realtime
         * @interface ILeaderboardEntry
         * @property {number|null} [rank] LeaderboardEntry rank
         * @property {string|null} [nickname] LeaderboardEntry nickname
         * @property {number|Long|null} [clickCount] LeaderboardEntry clickCount
         */

        /**
         * Constructs a new LeaderboardEntry.
         * @memberof realtime
         * @classdesc Represents a LeaderboardEntry.
         * @implements ILeaderboardEntry
         * @constructor
         * @param {realtime.ILeaderboardEntry=} [properties] Properties to set
         */
        function LeaderboardEntry(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * LeaderboardEntry rank.
         * @member {number} rank
         * @memberof realtime.LeaderboardEntry
         * @instance
         */
        LeaderboardEntry.prototype.rank = 0;

        /**
         * LeaderboardEntry nickname.
         * @member {string} nickname
         * @memberof realtime.LeaderboardEntry
         * @instance
         */
        LeaderboardEntry.prototype.nickname = "";

        /**
         * LeaderboardEntry clickCount.
         * @member {number|Long} clickCount
         * @memberof realtime.LeaderboardEntry
         * @instance
         */
        LeaderboardEntry.prototype.clickCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new LeaderboardEntry instance using the specified properties.
         * @function create
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {realtime.ILeaderboardEntry=} [properties] Properties to set
         * @returns {realtime.LeaderboardEntry} LeaderboardEntry instance
         */
        LeaderboardEntry.create = function create(properties) {
            return new LeaderboardEntry(properties);
        };

        /**
         * Encodes the specified LeaderboardEntry message. Does not implicitly {@link realtime.LeaderboardEntry.verify|verify} messages.
         * @function encode
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {realtime.ILeaderboardEntry} message LeaderboardEntry message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        LeaderboardEntry.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.rank != null && Object.hasOwnProperty.call(message, "rank"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.rank);
            if (message.nickname != null && Object.hasOwnProperty.call(message, "nickname"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.nickname);
            if (message.clickCount != null && Object.hasOwnProperty.call(message, "clickCount"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.clickCount);
            return writer;
        };

        /**
         * Encodes the specified LeaderboardEntry message, length delimited. Does not implicitly {@link realtime.LeaderboardEntry.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {realtime.ILeaderboardEntry} message LeaderboardEntry message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        LeaderboardEntry.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a LeaderboardEntry message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.LeaderboardEntry} LeaderboardEntry
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LeaderboardEntry.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.LeaderboardEntry();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.rank = reader.int32();
                        break;
                    }
                case 2: {
                        message.nickname = reader.string();
                        break;
                    }
                case 3: {
                        message.clickCount = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a LeaderboardEntry message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.LeaderboardEntry} LeaderboardEntry
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LeaderboardEntry.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a LeaderboardEntry message.
         * @function verify
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        LeaderboardEntry.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.rank != null && message.hasOwnProperty("rank"))
                if (!$util.isInteger(message.rank))
                    return "rank: integer expected";
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                if (!$util.isString(message.nickname))
                    return "nickname: string expected";
            if (message.clickCount != null && message.hasOwnProperty("clickCount"))
                if (!$util.isInteger(message.clickCount) && !(message.clickCount && $util.isInteger(message.clickCount.low) && $util.isInteger(message.clickCount.high)))
                    return "clickCount: integer|Long expected";
            return null;
        };

        /**
         * Creates a LeaderboardEntry message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.LeaderboardEntry} LeaderboardEntry
         */
        LeaderboardEntry.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.LeaderboardEntry)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.LeaderboardEntry();
            if (object.rank != null)
                message.rank = object.rank | 0;
            if (object.nickname != null)
                message.nickname = String(object.nickname);
            if (object.clickCount != null)
                if ($util.Long)
                    (message.clickCount = $util.Long.fromValue(object.clickCount)).unsigned = false;
                else if (typeof object.clickCount === "string")
                    message.clickCount = parseInt(object.clickCount, 10);
                else if (typeof object.clickCount === "number")
                    message.clickCount = object.clickCount;
                else if (typeof object.clickCount === "object")
                    message.clickCount = new $util.LongBits(object.clickCount.low >>> 0, object.clickCount.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a LeaderboardEntry message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {realtime.LeaderboardEntry} message LeaderboardEntry
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        LeaderboardEntry.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.rank = 0;
                object.nickname = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.clickCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.clickCount = options.longs === String ? "0" : 0;
            }
            if (message.rank != null && message.hasOwnProperty("rank"))
                object.rank = message.rank;
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                object.nickname = message.nickname;
            if (message.clickCount != null && message.hasOwnProperty("clickCount"))
                if (typeof message.clickCount === "number")
                    object.clickCount = options.longs === String ? String(message.clickCount) : message.clickCount;
                else
                    object.clickCount = options.longs === String ? $util.Long.prototype.toString.call(message.clickCount) : options.longs === Number ? new $util.LongBits(message.clickCount.low >>> 0, message.clickCount.high >>> 0).toNumber() : message.clickCount;
            return object;
        };

        /**
         * Converts this LeaderboardEntry to JSON.
         * @function toJSON
         * @memberof realtime.LeaderboardEntry
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        LeaderboardEntry.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for LeaderboardEntry
         * @function getTypeUrl
         * @memberof realtime.LeaderboardEntry
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        LeaderboardEntry.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.LeaderboardEntry";
        };

        return LeaderboardEntry;
    })();

    realtime.BossPart = (function() {

        /**
         * Properties of a BossPart.
         * @memberof realtime
         * @interface IBossPart
         * @property {number|null} [x] BossPart x
         * @property {number|null} [y] BossPart y
         * @property {string|null} [type] BossPart type
         * @property {string|null} [displayName] BossPart displayName
         * @property {string|null} [imagePath] BossPart imagePath
         * @property {number|Long|null} [maxHp] BossPart maxHp
         * @property {number|Long|null} [currentHp] BossPart currentHp
         * @property {number|Long|null} [armor] BossPart armor
         * @property {boolean|null} [alive] BossPart alive
         */

        /**
         * Constructs a new BossPart.
         * @memberof realtime
         * @classdesc Represents a BossPart.
         * @implements IBossPart
         * @constructor
         * @param {realtime.IBossPart=} [properties] Properties to set
         */
        function BossPart(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * BossPart x.
         * @member {number} x
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.x = 0;

        /**
         * BossPart y.
         * @member {number} y
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.y = 0;

        /**
         * BossPart type.
         * @member {string} type
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.type = "";

        /**
         * BossPart displayName.
         * @member {string} displayName
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.displayName = "";

        /**
         * BossPart imagePath.
         * @member {string} imagePath
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.imagePath = "";

        /**
         * BossPart maxHp.
         * @member {number|Long} maxHp
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.maxHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * BossPart currentHp.
         * @member {number|Long} currentHp
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.currentHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * BossPart armor.
         * @member {number|Long} armor
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.armor = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * BossPart alive.
         * @member {boolean} alive
         * @memberof realtime.BossPart
         * @instance
         */
        BossPart.prototype.alive = false;

        /**
         * Creates a new BossPart instance using the specified properties.
         * @function create
         * @memberof realtime.BossPart
         * @static
         * @param {realtime.IBossPart=} [properties] Properties to set
         * @returns {realtime.BossPart} BossPart instance
         */
        BossPart.create = function create(properties) {
            return new BossPart(properties);
        };

        /**
         * Encodes the specified BossPart message. Does not implicitly {@link realtime.BossPart.verify|verify} messages.
         * @function encode
         * @memberof realtime.BossPart
         * @static
         * @param {realtime.IBossPart} message BossPart message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossPart.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.x != null && Object.hasOwnProperty.call(message, "x"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.x);
            if (message.y != null && Object.hasOwnProperty.call(message, "y"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.y);
            if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.type);
            if (message.displayName != null && Object.hasOwnProperty.call(message, "displayName"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.displayName);
            if (message.imagePath != null && Object.hasOwnProperty.call(message, "imagePath"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.imagePath);
            if (message.maxHp != null && Object.hasOwnProperty.call(message, "maxHp"))
                writer.uint32(/* id 6, wireType 0 =*/48).int64(message.maxHp);
            if (message.currentHp != null && Object.hasOwnProperty.call(message, "currentHp"))
                writer.uint32(/* id 7, wireType 0 =*/56).int64(message.currentHp);
            if (message.armor != null && Object.hasOwnProperty.call(message, "armor"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.armor);
            if (message.alive != null && Object.hasOwnProperty.call(message, "alive"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.alive);
            return writer;
        };

        /**
         * Encodes the specified BossPart message, length delimited. Does not implicitly {@link realtime.BossPart.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.BossPart
         * @static
         * @param {realtime.IBossPart} message BossPart message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossPart.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a BossPart message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.BossPart
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.BossPart} BossPart
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossPart.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.BossPart();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.x = reader.int32();
                        break;
                    }
                case 2: {
                        message.y = reader.int32();
                        break;
                    }
                case 3: {
                        message.type = reader.string();
                        break;
                    }
                case 4: {
                        message.displayName = reader.string();
                        break;
                    }
                case 5: {
                        message.imagePath = reader.string();
                        break;
                    }
                case 6: {
                        message.maxHp = reader.int64();
                        break;
                    }
                case 7: {
                        message.currentHp = reader.int64();
                        break;
                    }
                case 8: {
                        message.armor = reader.int64();
                        break;
                    }
                case 9: {
                        message.alive = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a BossPart message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.BossPart
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.BossPart} BossPart
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossPart.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a BossPart message.
         * @function verify
         * @memberof realtime.BossPart
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        BossPart.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.x != null && message.hasOwnProperty("x"))
                if (!$util.isInteger(message.x))
                    return "x: integer expected";
            if (message.y != null && message.hasOwnProperty("y"))
                if (!$util.isInteger(message.y))
                    return "y: integer expected";
            if (message.type != null && message.hasOwnProperty("type"))
                if (!$util.isString(message.type))
                    return "type: string expected";
            if (message.displayName != null && message.hasOwnProperty("displayName"))
                if (!$util.isString(message.displayName))
                    return "displayName: string expected";
            if (message.imagePath != null && message.hasOwnProperty("imagePath"))
                if (!$util.isString(message.imagePath))
                    return "imagePath: string expected";
            if (message.maxHp != null && message.hasOwnProperty("maxHp"))
                if (!$util.isInteger(message.maxHp) && !(message.maxHp && $util.isInteger(message.maxHp.low) && $util.isInteger(message.maxHp.high)))
                    return "maxHp: integer|Long expected";
            if (message.currentHp != null && message.hasOwnProperty("currentHp"))
                if (!$util.isInteger(message.currentHp) && !(message.currentHp && $util.isInteger(message.currentHp.low) && $util.isInteger(message.currentHp.high)))
                    return "currentHp: integer|Long expected";
            if (message.armor != null && message.hasOwnProperty("armor"))
                if (!$util.isInteger(message.armor) && !(message.armor && $util.isInteger(message.armor.low) && $util.isInteger(message.armor.high)))
                    return "armor: integer|Long expected";
            if (message.alive != null && message.hasOwnProperty("alive"))
                if (typeof message.alive !== "boolean")
                    return "alive: boolean expected";
            return null;
        };

        /**
         * Creates a BossPart message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.BossPart
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.BossPart} BossPart
         */
        BossPart.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.BossPart)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.BossPart();
            if (object.x != null)
                message.x = object.x | 0;
            if (object.y != null)
                message.y = object.y | 0;
            if (object.type != null)
                message.type = String(object.type);
            if (object.displayName != null)
                message.displayName = String(object.displayName);
            if (object.imagePath != null)
                message.imagePath = String(object.imagePath);
            if (object.maxHp != null)
                if ($util.Long)
                    (message.maxHp = $util.Long.fromValue(object.maxHp)).unsigned = false;
                else if (typeof object.maxHp === "string")
                    message.maxHp = parseInt(object.maxHp, 10);
                else if (typeof object.maxHp === "number")
                    message.maxHp = object.maxHp;
                else if (typeof object.maxHp === "object")
                    message.maxHp = new $util.LongBits(object.maxHp.low >>> 0, object.maxHp.high >>> 0).toNumber();
            if (object.currentHp != null)
                if ($util.Long)
                    (message.currentHp = $util.Long.fromValue(object.currentHp)).unsigned = false;
                else if (typeof object.currentHp === "string")
                    message.currentHp = parseInt(object.currentHp, 10);
                else if (typeof object.currentHp === "number")
                    message.currentHp = object.currentHp;
                else if (typeof object.currentHp === "object")
                    message.currentHp = new $util.LongBits(object.currentHp.low >>> 0, object.currentHp.high >>> 0).toNumber();
            if (object.armor != null)
                if ($util.Long)
                    (message.armor = $util.Long.fromValue(object.armor)).unsigned = false;
                else if (typeof object.armor === "string")
                    message.armor = parseInt(object.armor, 10);
                else if (typeof object.armor === "number")
                    message.armor = object.armor;
                else if (typeof object.armor === "object")
                    message.armor = new $util.LongBits(object.armor.low >>> 0, object.armor.high >>> 0).toNumber();
            if (object.alive != null)
                message.alive = Boolean(object.alive);
            return message;
        };

        /**
         * Creates a plain object from a BossPart message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.BossPart
         * @static
         * @param {realtime.BossPart} message BossPart
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        BossPart.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.x = 0;
                object.y = 0;
                object.type = "";
                object.displayName = "";
                object.imagePath = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.maxHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.maxHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.currentHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.currentHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.armor = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.armor = options.longs === String ? "0" : 0;
                object.alive = false;
            }
            if (message.x != null && message.hasOwnProperty("x"))
                object.x = message.x;
            if (message.y != null && message.hasOwnProperty("y"))
                object.y = message.y;
            if (message.type != null && message.hasOwnProperty("type"))
                object.type = message.type;
            if (message.displayName != null && message.hasOwnProperty("displayName"))
                object.displayName = message.displayName;
            if (message.imagePath != null && message.hasOwnProperty("imagePath"))
                object.imagePath = message.imagePath;
            if (message.maxHp != null && message.hasOwnProperty("maxHp"))
                if (typeof message.maxHp === "number")
                    object.maxHp = options.longs === String ? String(message.maxHp) : message.maxHp;
                else
                    object.maxHp = options.longs === String ? $util.Long.prototype.toString.call(message.maxHp) : options.longs === Number ? new $util.LongBits(message.maxHp.low >>> 0, message.maxHp.high >>> 0).toNumber() : message.maxHp;
            if (message.currentHp != null && message.hasOwnProperty("currentHp"))
                if (typeof message.currentHp === "number")
                    object.currentHp = options.longs === String ? String(message.currentHp) : message.currentHp;
                else
                    object.currentHp = options.longs === String ? $util.Long.prototype.toString.call(message.currentHp) : options.longs === Number ? new $util.LongBits(message.currentHp.low >>> 0, message.currentHp.high >>> 0).toNumber() : message.currentHp;
            if (message.armor != null && message.hasOwnProperty("armor"))
                if (typeof message.armor === "number")
                    object.armor = options.longs === String ? String(message.armor) : message.armor;
                else
                    object.armor = options.longs === String ? $util.Long.prototype.toString.call(message.armor) : options.longs === Number ? new $util.LongBits(message.armor.low >>> 0, message.armor.high >>> 0).toNumber() : message.armor;
            if (message.alive != null && message.hasOwnProperty("alive"))
                object.alive = message.alive;
            return object;
        };

        /**
         * Converts this BossPart to JSON.
         * @function toJSON
         * @memberof realtime.BossPart
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        BossPart.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for BossPart
         * @function getTypeUrl
         * @memberof realtime.BossPart
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BossPart.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.BossPart";
        };

        return BossPart;
    })();

    realtime.Boss = (function() {

        /**
         * Properties of a Boss.
         * @memberof realtime
         * @interface IBoss
         * @property {string|null} [id] Boss id
         * @property {string|null} [templateId] Boss templateId
         * @property {string|null} [roomId] Boss roomId
         * @property {string|null} [queueId] Boss queueId
         * @property {string|null} [name] Boss name
         * @property {string|null} [status] Boss status
         * @property {number|Long|null} [maxHp] Boss maxHp
         * @property {number|Long|null} [currentHp] Boss currentHp
         * @property {number|Long|null} [goldOnKill] Boss goldOnKill
         * @property {number|Long|null} [stoneOnKill] Boss stoneOnKill
         * @property {number|Long|null} [talentPointsOnKill] Boss talentPointsOnKill
         * @property {Array.<realtime.IBossPart>|null} [parts] Boss parts
         * @property {number|Long|null} [startedAt] Boss startedAt
         * @property {number|Long|null} [defeatedAt] Boss defeatedAt
         */

        /**
         * Constructs a new Boss.
         * @memberof realtime
         * @classdesc Represents a Boss.
         * @implements IBoss
         * @constructor
         * @param {realtime.IBoss=} [properties] Properties to set
         */
        function Boss(properties) {
            this.parts = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Boss id.
         * @member {string} id
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.id = "";

        /**
         * Boss templateId.
         * @member {string} templateId
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.templateId = "";

        /**
         * Boss roomId.
         * @member {string} roomId
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.roomId = "";

        /**
         * Boss queueId.
         * @member {string} queueId
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.queueId = "";

        /**
         * Boss name.
         * @member {string} name
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.name = "";

        /**
         * Boss status.
         * @member {string} status
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.status = "";

        /**
         * Boss maxHp.
         * @member {number|Long} maxHp
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.maxHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Boss currentHp.
         * @member {number|Long} currentHp
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.currentHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Boss goldOnKill.
         * @member {number|Long} goldOnKill
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.goldOnKill = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Boss stoneOnKill.
         * @member {number|Long} stoneOnKill
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.stoneOnKill = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Boss talentPointsOnKill.
         * @member {number|Long} talentPointsOnKill
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.talentPointsOnKill = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Boss parts.
         * @member {Array.<realtime.IBossPart>} parts
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.parts = $util.emptyArray;

        /**
         * Boss startedAt.
         * @member {number|Long} startedAt
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.startedAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Boss defeatedAt.
         * @member {number|Long} defeatedAt
         * @memberof realtime.Boss
         * @instance
         */
        Boss.prototype.defeatedAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new Boss instance using the specified properties.
         * @function create
         * @memberof realtime.Boss
         * @static
         * @param {realtime.IBoss=} [properties] Properties to set
         * @returns {realtime.Boss} Boss instance
         */
        Boss.create = function create(properties) {
            return new Boss(properties);
        };

        /**
         * Encodes the specified Boss message. Does not implicitly {@link realtime.Boss.verify|verify} messages.
         * @function encode
         * @memberof realtime.Boss
         * @static
         * @param {realtime.IBoss} message Boss message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Boss.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.templateId != null && Object.hasOwnProperty.call(message, "templateId"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.templateId);
            if (message.roomId != null && Object.hasOwnProperty.call(message, "roomId"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.roomId);
            if (message.queueId != null && Object.hasOwnProperty.call(message, "queueId"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.queueId);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.name);
            if (message.status != null && Object.hasOwnProperty.call(message, "status"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.status);
            if (message.maxHp != null && Object.hasOwnProperty.call(message, "maxHp"))
                writer.uint32(/* id 7, wireType 0 =*/56).int64(message.maxHp);
            if (message.currentHp != null && Object.hasOwnProperty.call(message, "currentHp"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.currentHp);
            if (message.goldOnKill != null && Object.hasOwnProperty.call(message, "goldOnKill"))
                writer.uint32(/* id 9, wireType 0 =*/72).int64(message.goldOnKill);
            if (message.stoneOnKill != null && Object.hasOwnProperty.call(message, "stoneOnKill"))
                writer.uint32(/* id 10, wireType 0 =*/80).int64(message.stoneOnKill);
            if (message.talentPointsOnKill != null && Object.hasOwnProperty.call(message, "talentPointsOnKill"))
                writer.uint32(/* id 11, wireType 0 =*/88).int64(message.talentPointsOnKill);
            if (message.parts != null && message.parts.length)
                for (let i = 0; i < message.parts.length; ++i)
                    $root.realtime.BossPart.encode(message.parts[i], writer.uint32(/* id 12, wireType 2 =*/98).fork()).ldelim();
            if (message.startedAt != null && Object.hasOwnProperty.call(message, "startedAt"))
                writer.uint32(/* id 13, wireType 0 =*/104).int64(message.startedAt);
            if (message.defeatedAt != null && Object.hasOwnProperty.call(message, "defeatedAt"))
                writer.uint32(/* id 14, wireType 0 =*/112).int64(message.defeatedAt);
            return writer;
        };

        /**
         * Encodes the specified Boss message, length delimited. Does not implicitly {@link realtime.Boss.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.Boss
         * @static
         * @param {realtime.IBoss} message Boss message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Boss.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Boss message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.Boss
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.Boss} Boss
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Boss.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.Boss();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.templateId = reader.string();
                        break;
                    }
                case 3: {
                        message.roomId = reader.string();
                        break;
                    }
                case 4: {
                        message.queueId = reader.string();
                        break;
                    }
                case 5: {
                        message.name = reader.string();
                        break;
                    }
                case 6: {
                        message.status = reader.string();
                        break;
                    }
                case 7: {
                        message.maxHp = reader.int64();
                        break;
                    }
                case 8: {
                        message.currentHp = reader.int64();
                        break;
                    }
                case 9: {
                        message.goldOnKill = reader.int64();
                        break;
                    }
                case 10: {
                        message.stoneOnKill = reader.int64();
                        break;
                    }
                case 11: {
                        message.talentPointsOnKill = reader.int64();
                        break;
                    }
                case 12: {
                        if (!(message.parts && message.parts.length))
                            message.parts = [];
                        message.parts.push($root.realtime.BossPart.decode(reader, reader.uint32(), undefined, long + 1));
                        break;
                    }
                case 13: {
                        message.startedAt = reader.int64();
                        break;
                    }
                case 14: {
                        message.defeatedAt = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Boss message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.Boss
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.Boss} Boss
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Boss.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Boss message.
         * @function verify
         * @memberof realtime.Boss
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Boss.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.templateId != null && message.hasOwnProperty("templateId"))
                if (!$util.isString(message.templateId))
                    return "templateId: string expected";
            if (message.roomId != null && message.hasOwnProperty("roomId"))
                if (!$util.isString(message.roomId))
                    return "roomId: string expected";
            if (message.queueId != null && message.hasOwnProperty("queueId"))
                if (!$util.isString(message.queueId))
                    return "queueId: string expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.status != null && message.hasOwnProperty("status"))
                if (!$util.isString(message.status))
                    return "status: string expected";
            if (message.maxHp != null && message.hasOwnProperty("maxHp"))
                if (!$util.isInteger(message.maxHp) && !(message.maxHp && $util.isInteger(message.maxHp.low) && $util.isInteger(message.maxHp.high)))
                    return "maxHp: integer|Long expected";
            if (message.currentHp != null && message.hasOwnProperty("currentHp"))
                if (!$util.isInteger(message.currentHp) && !(message.currentHp && $util.isInteger(message.currentHp.low) && $util.isInteger(message.currentHp.high)))
                    return "currentHp: integer|Long expected";
            if (message.goldOnKill != null && message.hasOwnProperty("goldOnKill"))
                if (!$util.isInteger(message.goldOnKill) && !(message.goldOnKill && $util.isInteger(message.goldOnKill.low) && $util.isInteger(message.goldOnKill.high)))
                    return "goldOnKill: integer|Long expected";
            if (message.stoneOnKill != null && message.hasOwnProperty("stoneOnKill"))
                if (!$util.isInteger(message.stoneOnKill) && !(message.stoneOnKill && $util.isInteger(message.stoneOnKill.low) && $util.isInteger(message.stoneOnKill.high)))
                    return "stoneOnKill: integer|Long expected";
            if (message.talentPointsOnKill != null && message.hasOwnProperty("talentPointsOnKill"))
                if (!$util.isInteger(message.talentPointsOnKill) && !(message.talentPointsOnKill && $util.isInteger(message.talentPointsOnKill.low) && $util.isInteger(message.talentPointsOnKill.high)))
                    return "talentPointsOnKill: integer|Long expected";
            if (message.parts != null && message.hasOwnProperty("parts")) {
                if (!Array.isArray(message.parts))
                    return "parts: array expected";
                for (let i = 0; i < message.parts.length; ++i) {
                    let error = $root.realtime.BossPart.verify(message.parts[i], long + 1);
                    if (error)
                        return "parts." + error;
                }
            }
            if (message.startedAt != null && message.hasOwnProperty("startedAt"))
                if (!$util.isInteger(message.startedAt) && !(message.startedAt && $util.isInteger(message.startedAt.low) && $util.isInteger(message.startedAt.high)))
                    return "startedAt: integer|Long expected";
            if (message.defeatedAt != null && message.hasOwnProperty("defeatedAt"))
                if (!$util.isInteger(message.defeatedAt) && !(message.defeatedAt && $util.isInteger(message.defeatedAt.low) && $util.isInteger(message.defeatedAt.high)))
                    return "defeatedAt: integer|Long expected";
            return null;
        };

        /**
         * Creates a Boss message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.Boss
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.Boss} Boss
         */
        Boss.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.Boss)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.Boss();
            if (object.id != null)
                message.id = String(object.id);
            if (object.templateId != null)
                message.templateId = String(object.templateId);
            if (object.roomId != null)
                message.roomId = String(object.roomId);
            if (object.queueId != null)
                message.queueId = String(object.queueId);
            if (object.name != null)
                message.name = String(object.name);
            if (object.status != null)
                message.status = String(object.status);
            if (object.maxHp != null)
                if ($util.Long)
                    (message.maxHp = $util.Long.fromValue(object.maxHp)).unsigned = false;
                else if (typeof object.maxHp === "string")
                    message.maxHp = parseInt(object.maxHp, 10);
                else if (typeof object.maxHp === "number")
                    message.maxHp = object.maxHp;
                else if (typeof object.maxHp === "object")
                    message.maxHp = new $util.LongBits(object.maxHp.low >>> 0, object.maxHp.high >>> 0).toNumber();
            if (object.currentHp != null)
                if ($util.Long)
                    (message.currentHp = $util.Long.fromValue(object.currentHp)).unsigned = false;
                else if (typeof object.currentHp === "string")
                    message.currentHp = parseInt(object.currentHp, 10);
                else if (typeof object.currentHp === "number")
                    message.currentHp = object.currentHp;
                else if (typeof object.currentHp === "object")
                    message.currentHp = new $util.LongBits(object.currentHp.low >>> 0, object.currentHp.high >>> 0).toNumber();
            if (object.goldOnKill != null)
                if ($util.Long)
                    (message.goldOnKill = $util.Long.fromValue(object.goldOnKill)).unsigned = false;
                else if (typeof object.goldOnKill === "string")
                    message.goldOnKill = parseInt(object.goldOnKill, 10);
                else if (typeof object.goldOnKill === "number")
                    message.goldOnKill = object.goldOnKill;
                else if (typeof object.goldOnKill === "object")
                    message.goldOnKill = new $util.LongBits(object.goldOnKill.low >>> 0, object.goldOnKill.high >>> 0).toNumber();
            if (object.stoneOnKill != null)
                if ($util.Long)
                    (message.stoneOnKill = $util.Long.fromValue(object.stoneOnKill)).unsigned = false;
                else if (typeof object.stoneOnKill === "string")
                    message.stoneOnKill = parseInt(object.stoneOnKill, 10);
                else if (typeof object.stoneOnKill === "number")
                    message.stoneOnKill = object.stoneOnKill;
                else if (typeof object.stoneOnKill === "object")
                    message.stoneOnKill = new $util.LongBits(object.stoneOnKill.low >>> 0, object.stoneOnKill.high >>> 0).toNumber();
            if (object.talentPointsOnKill != null)
                if ($util.Long)
                    (message.talentPointsOnKill = $util.Long.fromValue(object.talentPointsOnKill)).unsigned = false;
                else if (typeof object.talentPointsOnKill === "string")
                    message.talentPointsOnKill = parseInt(object.talentPointsOnKill, 10);
                else if (typeof object.talentPointsOnKill === "number")
                    message.talentPointsOnKill = object.talentPointsOnKill;
                else if (typeof object.talentPointsOnKill === "object")
                    message.talentPointsOnKill = new $util.LongBits(object.talentPointsOnKill.low >>> 0, object.talentPointsOnKill.high >>> 0).toNumber();
            if (object.parts) {
                if (!Array.isArray(object.parts))
                    throw TypeError(".realtime.Boss.parts: array expected");
                message.parts = [];
                for (let i = 0; i < object.parts.length; ++i) {
                    if (typeof object.parts[i] !== "object")
                        throw TypeError(".realtime.Boss.parts: object expected");
                    message.parts[i] = $root.realtime.BossPart.fromObject(object.parts[i], long + 1);
                }
            }
            if (object.startedAt != null)
                if ($util.Long)
                    (message.startedAt = $util.Long.fromValue(object.startedAt)).unsigned = false;
                else if (typeof object.startedAt === "string")
                    message.startedAt = parseInt(object.startedAt, 10);
                else if (typeof object.startedAt === "number")
                    message.startedAt = object.startedAt;
                else if (typeof object.startedAt === "object")
                    message.startedAt = new $util.LongBits(object.startedAt.low >>> 0, object.startedAt.high >>> 0).toNumber();
            if (object.defeatedAt != null)
                if ($util.Long)
                    (message.defeatedAt = $util.Long.fromValue(object.defeatedAt)).unsigned = false;
                else if (typeof object.defeatedAt === "string")
                    message.defeatedAt = parseInt(object.defeatedAt, 10);
                else if (typeof object.defeatedAt === "number")
                    message.defeatedAt = object.defeatedAt;
                else if (typeof object.defeatedAt === "object")
                    message.defeatedAt = new $util.LongBits(object.defeatedAt.low >>> 0, object.defeatedAt.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a Boss message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.Boss
         * @static
         * @param {realtime.Boss} message Boss
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Boss.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.parts = [];
            if (options.defaults) {
                object.id = "";
                object.templateId = "";
                object.roomId = "";
                object.queueId = "";
                object.name = "";
                object.status = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.maxHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.maxHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.currentHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.currentHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.goldOnKill = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.goldOnKill = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.stoneOnKill = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.stoneOnKill = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.talentPointsOnKill = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.talentPointsOnKill = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.startedAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.startedAt = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.defeatedAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.defeatedAt = options.longs === String ? "0" : 0;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.templateId != null && message.hasOwnProperty("templateId"))
                object.templateId = message.templateId;
            if (message.roomId != null && message.hasOwnProperty("roomId"))
                object.roomId = message.roomId;
            if (message.queueId != null && message.hasOwnProperty("queueId"))
                object.queueId = message.queueId;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.status != null && message.hasOwnProperty("status"))
                object.status = message.status;
            if (message.maxHp != null && message.hasOwnProperty("maxHp"))
                if (typeof message.maxHp === "number")
                    object.maxHp = options.longs === String ? String(message.maxHp) : message.maxHp;
                else
                    object.maxHp = options.longs === String ? $util.Long.prototype.toString.call(message.maxHp) : options.longs === Number ? new $util.LongBits(message.maxHp.low >>> 0, message.maxHp.high >>> 0).toNumber() : message.maxHp;
            if (message.currentHp != null && message.hasOwnProperty("currentHp"))
                if (typeof message.currentHp === "number")
                    object.currentHp = options.longs === String ? String(message.currentHp) : message.currentHp;
                else
                    object.currentHp = options.longs === String ? $util.Long.prototype.toString.call(message.currentHp) : options.longs === Number ? new $util.LongBits(message.currentHp.low >>> 0, message.currentHp.high >>> 0).toNumber() : message.currentHp;
            if (message.goldOnKill != null && message.hasOwnProperty("goldOnKill"))
                if (typeof message.goldOnKill === "number")
                    object.goldOnKill = options.longs === String ? String(message.goldOnKill) : message.goldOnKill;
                else
                    object.goldOnKill = options.longs === String ? $util.Long.prototype.toString.call(message.goldOnKill) : options.longs === Number ? new $util.LongBits(message.goldOnKill.low >>> 0, message.goldOnKill.high >>> 0).toNumber() : message.goldOnKill;
            if (message.stoneOnKill != null && message.hasOwnProperty("stoneOnKill"))
                if (typeof message.stoneOnKill === "number")
                    object.stoneOnKill = options.longs === String ? String(message.stoneOnKill) : message.stoneOnKill;
                else
                    object.stoneOnKill = options.longs === String ? $util.Long.prototype.toString.call(message.stoneOnKill) : options.longs === Number ? new $util.LongBits(message.stoneOnKill.low >>> 0, message.stoneOnKill.high >>> 0).toNumber() : message.stoneOnKill;
            if (message.talentPointsOnKill != null && message.hasOwnProperty("talentPointsOnKill"))
                if (typeof message.talentPointsOnKill === "number")
                    object.talentPointsOnKill = options.longs === String ? String(message.talentPointsOnKill) : message.talentPointsOnKill;
                else
                    object.talentPointsOnKill = options.longs === String ? $util.Long.prototype.toString.call(message.talentPointsOnKill) : options.longs === Number ? new $util.LongBits(message.talentPointsOnKill.low >>> 0, message.talentPointsOnKill.high >>> 0).toNumber() : message.talentPointsOnKill;
            if (message.parts && message.parts.length) {
                object.parts = [];
                for (let j = 0; j < message.parts.length; ++j)
                    object.parts[j] = $root.realtime.BossPart.toObject(message.parts[j], options);
            }
            if (message.startedAt != null && message.hasOwnProperty("startedAt"))
                if (typeof message.startedAt === "number")
                    object.startedAt = options.longs === String ? String(message.startedAt) : message.startedAt;
                else
                    object.startedAt = options.longs === String ? $util.Long.prototype.toString.call(message.startedAt) : options.longs === Number ? new $util.LongBits(message.startedAt.low >>> 0, message.startedAt.high >>> 0).toNumber() : message.startedAt;
            if (message.defeatedAt != null && message.hasOwnProperty("defeatedAt"))
                if (typeof message.defeatedAt === "number")
                    object.defeatedAt = options.longs === String ? String(message.defeatedAt) : message.defeatedAt;
                else
                    object.defeatedAt = options.longs === String ? $util.Long.prototype.toString.call(message.defeatedAt) : options.longs === Number ? new $util.LongBits(message.defeatedAt.low >>> 0, message.defeatedAt.high >>> 0).toNumber() : message.defeatedAt;
            return object;
        };

        /**
         * Converts this Boss to JSON.
         * @function toJSON
         * @memberof realtime.Boss
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Boss.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Boss
         * @function getTypeUrl
         * @memberof realtime.Boss
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Boss.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.Boss";
        };

        return Boss;
    })();

    realtime.BossLeaderboardEntry = (function() {

        /**
         * Properties of a BossLeaderboardEntry.
         * @memberof realtime
         * @interface IBossLeaderboardEntry
         * @property {number|null} [rank] BossLeaderboardEntry rank
         * @property {string|null} [nickname] BossLeaderboardEntry nickname
         * @property {number|Long|null} [damage] BossLeaderboardEntry damage
         */

        /**
         * Constructs a new BossLeaderboardEntry.
         * @memberof realtime
         * @classdesc Represents a BossLeaderboardEntry.
         * @implements IBossLeaderboardEntry
         * @constructor
         * @param {realtime.IBossLeaderboardEntry=} [properties] Properties to set
         */
        function BossLeaderboardEntry(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * BossLeaderboardEntry rank.
         * @member {number} rank
         * @memberof realtime.BossLeaderboardEntry
         * @instance
         */
        BossLeaderboardEntry.prototype.rank = 0;

        /**
         * BossLeaderboardEntry nickname.
         * @member {string} nickname
         * @memberof realtime.BossLeaderboardEntry
         * @instance
         */
        BossLeaderboardEntry.prototype.nickname = "";

        /**
         * BossLeaderboardEntry damage.
         * @member {number|Long} damage
         * @memberof realtime.BossLeaderboardEntry
         * @instance
         */
        BossLeaderboardEntry.prototype.damage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new BossLeaderboardEntry instance using the specified properties.
         * @function create
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {realtime.IBossLeaderboardEntry=} [properties] Properties to set
         * @returns {realtime.BossLeaderboardEntry} BossLeaderboardEntry instance
         */
        BossLeaderboardEntry.create = function create(properties) {
            return new BossLeaderboardEntry(properties);
        };

        /**
         * Encodes the specified BossLeaderboardEntry message. Does not implicitly {@link realtime.BossLeaderboardEntry.verify|verify} messages.
         * @function encode
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {realtime.IBossLeaderboardEntry} message BossLeaderboardEntry message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossLeaderboardEntry.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.rank != null && Object.hasOwnProperty.call(message, "rank"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.rank);
            if (message.nickname != null && Object.hasOwnProperty.call(message, "nickname"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.nickname);
            if (message.damage != null && Object.hasOwnProperty.call(message, "damage"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.damage);
            return writer;
        };

        /**
         * Encodes the specified BossLeaderboardEntry message, length delimited. Does not implicitly {@link realtime.BossLeaderboardEntry.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {realtime.IBossLeaderboardEntry} message BossLeaderboardEntry message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossLeaderboardEntry.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a BossLeaderboardEntry message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.BossLeaderboardEntry} BossLeaderboardEntry
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossLeaderboardEntry.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.BossLeaderboardEntry();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.rank = reader.int32();
                        break;
                    }
                case 2: {
                        message.nickname = reader.string();
                        break;
                    }
                case 3: {
                        message.damage = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a BossLeaderboardEntry message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.BossLeaderboardEntry} BossLeaderboardEntry
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossLeaderboardEntry.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a BossLeaderboardEntry message.
         * @function verify
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        BossLeaderboardEntry.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.rank != null && message.hasOwnProperty("rank"))
                if (!$util.isInteger(message.rank))
                    return "rank: integer expected";
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                if (!$util.isString(message.nickname))
                    return "nickname: string expected";
            if (message.damage != null && message.hasOwnProperty("damage"))
                if (!$util.isInteger(message.damage) && !(message.damage && $util.isInteger(message.damage.low) && $util.isInteger(message.damage.high)))
                    return "damage: integer|Long expected";
            return null;
        };

        /**
         * Creates a BossLeaderboardEntry message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.BossLeaderboardEntry} BossLeaderboardEntry
         */
        BossLeaderboardEntry.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.BossLeaderboardEntry)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.BossLeaderboardEntry();
            if (object.rank != null)
                message.rank = object.rank | 0;
            if (object.nickname != null)
                message.nickname = String(object.nickname);
            if (object.damage != null)
                if ($util.Long)
                    (message.damage = $util.Long.fromValue(object.damage)).unsigned = false;
                else if (typeof object.damage === "string")
                    message.damage = parseInt(object.damage, 10);
                else if (typeof object.damage === "number")
                    message.damage = object.damage;
                else if (typeof object.damage === "object")
                    message.damage = new $util.LongBits(object.damage.low >>> 0, object.damage.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a BossLeaderboardEntry message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {realtime.BossLeaderboardEntry} message BossLeaderboardEntry
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        BossLeaderboardEntry.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.rank = 0;
                object.nickname = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.damage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.damage = options.longs === String ? "0" : 0;
            }
            if (message.rank != null && message.hasOwnProperty("rank"))
                object.rank = message.rank;
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                object.nickname = message.nickname;
            if (message.damage != null && message.hasOwnProperty("damage"))
                if (typeof message.damage === "number")
                    object.damage = options.longs === String ? String(message.damage) : message.damage;
                else
                    object.damage = options.longs === String ? $util.Long.prototype.toString.call(message.damage) : options.longs === Number ? new $util.LongBits(message.damage.low >>> 0, message.damage.high >>> 0).toNumber() : message.damage;
            return object;
        };

        /**
         * Converts this BossLeaderboardEntry to JSON.
         * @function toJSON
         * @memberof realtime.BossLeaderboardEntry
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        BossLeaderboardEntry.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for BossLeaderboardEntry
         * @function getTypeUrl
         * @memberof realtime.BossLeaderboardEntry
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BossLeaderboardEntry.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.BossLeaderboardEntry";
        };

        return BossLeaderboardEntry;
    })();

    realtime.BossUserStats = (function() {

        /**
         * Properties of a BossUserStats.
         * @memberof realtime
         * @interface IBossUserStats
         * @property {string|null} [nickname] BossUserStats nickname
         * @property {number|Long|null} [damage] BossUserStats damage
         * @property {number|null} [rank] BossUserStats rank
         */

        /**
         * Constructs a new BossUserStats.
         * @memberof realtime
         * @classdesc Represents a BossUserStats.
         * @implements IBossUserStats
         * @constructor
         * @param {realtime.IBossUserStats=} [properties] Properties to set
         */
        function BossUserStats(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * BossUserStats nickname.
         * @member {string} nickname
         * @memberof realtime.BossUserStats
         * @instance
         */
        BossUserStats.prototype.nickname = "";

        /**
         * BossUserStats damage.
         * @member {number|Long} damage
         * @memberof realtime.BossUserStats
         * @instance
         */
        BossUserStats.prototype.damage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * BossUserStats rank.
         * @member {number} rank
         * @memberof realtime.BossUserStats
         * @instance
         */
        BossUserStats.prototype.rank = 0;

        /**
         * Creates a new BossUserStats instance using the specified properties.
         * @function create
         * @memberof realtime.BossUserStats
         * @static
         * @param {realtime.IBossUserStats=} [properties] Properties to set
         * @returns {realtime.BossUserStats} BossUserStats instance
         */
        BossUserStats.create = function create(properties) {
            return new BossUserStats(properties);
        };

        /**
         * Encodes the specified BossUserStats message. Does not implicitly {@link realtime.BossUserStats.verify|verify} messages.
         * @function encode
         * @memberof realtime.BossUserStats
         * @static
         * @param {realtime.IBossUserStats} message BossUserStats message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossUserStats.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.nickname != null && Object.hasOwnProperty.call(message, "nickname"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.nickname);
            if (message.damage != null && Object.hasOwnProperty.call(message, "damage"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.damage);
            if (message.rank != null && Object.hasOwnProperty.call(message, "rank"))
                writer.uint32(/* id 3, wireType 0 =*/24).int32(message.rank);
            return writer;
        };

        /**
         * Encodes the specified BossUserStats message, length delimited. Does not implicitly {@link realtime.BossUserStats.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.BossUserStats
         * @static
         * @param {realtime.IBossUserStats} message BossUserStats message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossUserStats.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a BossUserStats message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.BossUserStats
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.BossUserStats} BossUserStats
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossUserStats.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.BossUserStats();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.nickname = reader.string();
                        break;
                    }
                case 2: {
                        message.damage = reader.int64();
                        break;
                    }
                case 3: {
                        message.rank = reader.int32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a BossUserStats message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.BossUserStats
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.BossUserStats} BossUserStats
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossUserStats.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a BossUserStats message.
         * @function verify
         * @memberof realtime.BossUserStats
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        BossUserStats.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                if (!$util.isString(message.nickname))
                    return "nickname: string expected";
            if (message.damage != null && message.hasOwnProperty("damage"))
                if (!$util.isInteger(message.damage) && !(message.damage && $util.isInteger(message.damage.low) && $util.isInteger(message.damage.high)))
                    return "damage: integer|Long expected";
            if (message.rank != null && message.hasOwnProperty("rank"))
                if (!$util.isInteger(message.rank))
                    return "rank: integer expected";
            return null;
        };

        /**
         * Creates a BossUserStats message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.BossUserStats
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.BossUserStats} BossUserStats
         */
        BossUserStats.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.BossUserStats)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.BossUserStats();
            if (object.nickname != null)
                message.nickname = String(object.nickname);
            if (object.damage != null)
                if ($util.Long)
                    (message.damage = $util.Long.fromValue(object.damage)).unsigned = false;
                else if (typeof object.damage === "string")
                    message.damage = parseInt(object.damage, 10);
                else if (typeof object.damage === "number")
                    message.damage = object.damage;
                else if (typeof object.damage === "object")
                    message.damage = new $util.LongBits(object.damage.low >>> 0, object.damage.high >>> 0).toNumber();
            if (object.rank != null)
                message.rank = object.rank | 0;
            return message;
        };

        /**
         * Creates a plain object from a BossUserStats message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.BossUserStats
         * @static
         * @param {realtime.BossUserStats} message BossUserStats
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        BossUserStats.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.nickname = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.damage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.damage = options.longs === String ? "0" : 0;
                object.rank = 0;
            }
            if (message.nickname != null && message.hasOwnProperty("nickname"))
                object.nickname = message.nickname;
            if (message.damage != null && message.hasOwnProperty("damage"))
                if (typeof message.damage === "number")
                    object.damage = options.longs === String ? String(message.damage) : message.damage;
                else
                    object.damage = options.longs === String ? $util.Long.prototype.toString.call(message.damage) : options.longs === Number ? new $util.LongBits(message.damage.low >>> 0, message.damage.high >>> 0).toNumber() : message.damage;
            if (message.rank != null && message.hasOwnProperty("rank"))
                object.rank = message.rank;
            return object;
        };

        /**
         * Converts this BossUserStats to JSON.
         * @function toJSON
         * @memberof realtime.BossUserStats
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        BossUserStats.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for BossUserStats
         * @function getTypeUrl
         * @memberof realtime.BossUserStats
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BossUserStats.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.BossUserStats";
        };

        return BossUserStats;
    })();

    realtime.InventoryItem = (function() {

        /**
         * Properties of an InventoryItem.
         * @memberof realtime
         * @interface IInventoryItem
         * @property {string|null} [itemId] InventoryItem itemId
         * @property {string|null} [instanceId] InventoryItem instanceId
         * @property {string|null} [name] InventoryItem name
         * @property {string|null} [slot] InventoryItem slot
         * @property {string|null} [rarity] InventoryItem rarity
         * @property {string|null} [imagePath] InventoryItem imagePath
         * @property {string|null} [imageAlt] InventoryItem imageAlt
         * @property {number|Long|null} [quantity] InventoryItem quantity
         * @property {boolean|null} [equipped] InventoryItem equipped
         * @property {number|null} [enhanceLevel] InventoryItem enhanceLevel
         * @property {boolean|null} [bound] InventoryItem bound
         * @property {boolean|null} [locked] InventoryItem locked
         * @property {number|Long|null} [attackPower] InventoryItem attackPower
         * @property {number|null} [armorPenPercent] InventoryItem armorPenPercent
         * @property {number|null} [critRate] InventoryItem critRate
         * @property {number|null} [critDamageMultiplier] InventoryItem critDamageMultiplier
         * @property {number|null} [partTypeDamageSoft] InventoryItem partTypeDamageSoft
         * @property {number|null} [partTypeDamageHeavy] InventoryItem partTypeDamageHeavy
         * @property {number|null} [partTypeDamageWeak] InventoryItem partTypeDamageWeak
         */

        /**
         * Constructs a new InventoryItem.
         * @memberof realtime
         * @classdesc Represents an InventoryItem.
         * @implements IInventoryItem
         * @constructor
         * @param {realtime.IInventoryItem=} [properties] Properties to set
         */
        function InventoryItem(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * InventoryItem itemId.
         * @member {string} itemId
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.itemId = "";

        /**
         * InventoryItem instanceId.
         * @member {string} instanceId
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.instanceId = "";

        /**
         * InventoryItem name.
         * @member {string} name
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.name = "";

        /**
         * InventoryItem slot.
         * @member {string} slot
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.slot = "";

        /**
         * InventoryItem rarity.
         * @member {string} rarity
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.rarity = "";

        /**
         * InventoryItem imagePath.
         * @member {string} imagePath
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.imagePath = "";

        /**
         * InventoryItem imageAlt.
         * @member {string} imageAlt
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.imageAlt = "";

        /**
         * InventoryItem quantity.
         * @member {number|Long} quantity
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.quantity = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * InventoryItem equipped.
         * @member {boolean} equipped
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.equipped = false;

        /**
         * InventoryItem enhanceLevel.
         * @member {number} enhanceLevel
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.enhanceLevel = 0;

        /**
         * InventoryItem bound.
         * @member {boolean} bound
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.bound = false;

        /**
         * InventoryItem locked.
         * @member {boolean} locked
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.locked = false;

        /**
         * InventoryItem attackPower.
         * @member {number|Long} attackPower
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.attackPower = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * InventoryItem armorPenPercent.
         * @member {number} armorPenPercent
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.armorPenPercent = 0;

        /**
         * InventoryItem critRate.
         * @member {number} critRate
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.critRate = 0;

        /**
         * InventoryItem critDamageMultiplier.
         * @member {number} critDamageMultiplier
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.critDamageMultiplier = 0;

        /**
         * InventoryItem partTypeDamageSoft.
         * @member {number} partTypeDamageSoft
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.partTypeDamageSoft = 0;

        /**
         * InventoryItem partTypeDamageHeavy.
         * @member {number} partTypeDamageHeavy
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.partTypeDamageHeavy = 0;

        /**
         * InventoryItem partTypeDamageWeak.
         * @member {number} partTypeDamageWeak
         * @memberof realtime.InventoryItem
         * @instance
         */
        InventoryItem.prototype.partTypeDamageWeak = 0;

        /**
         * Creates a new InventoryItem instance using the specified properties.
         * @function create
         * @memberof realtime.InventoryItem
         * @static
         * @param {realtime.IInventoryItem=} [properties] Properties to set
         * @returns {realtime.InventoryItem} InventoryItem instance
         */
        InventoryItem.create = function create(properties) {
            return new InventoryItem(properties);
        };

        /**
         * Encodes the specified InventoryItem message. Does not implicitly {@link realtime.InventoryItem.verify|verify} messages.
         * @function encode
         * @memberof realtime.InventoryItem
         * @static
         * @param {realtime.IInventoryItem} message InventoryItem message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InventoryItem.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.itemId != null && Object.hasOwnProperty.call(message, "itemId"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.itemId);
            if (message.instanceId != null && Object.hasOwnProperty.call(message, "instanceId"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.instanceId);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.slot != null && Object.hasOwnProperty.call(message, "slot"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.slot);
            if (message.rarity != null && Object.hasOwnProperty.call(message, "rarity"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.rarity);
            if (message.imagePath != null && Object.hasOwnProperty.call(message, "imagePath"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.imagePath);
            if (message.imageAlt != null && Object.hasOwnProperty.call(message, "imageAlt"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.imageAlt);
            if (message.quantity != null && Object.hasOwnProperty.call(message, "quantity"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.quantity);
            if (message.equipped != null && Object.hasOwnProperty.call(message, "equipped"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.equipped);
            if (message.enhanceLevel != null && Object.hasOwnProperty.call(message, "enhanceLevel"))
                writer.uint32(/* id 10, wireType 0 =*/80).int32(message.enhanceLevel);
            if (message.bound != null && Object.hasOwnProperty.call(message, "bound"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.bound);
            if (message.locked != null && Object.hasOwnProperty.call(message, "locked"))
                writer.uint32(/* id 12, wireType 0 =*/96).bool(message.locked);
            if (message.attackPower != null && Object.hasOwnProperty.call(message, "attackPower"))
                writer.uint32(/* id 13, wireType 0 =*/104).int64(message.attackPower);
            if (message.armorPenPercent != null && Object.hasOwnProperty.call(message, "armorPenPercent"))
                writer.uint32(/* id 14, wireType 1 =*/113).double(message.armorPenPercent);
            if (message.critRate != null && Object.hasOwnProperty.call(message, "critRate"))
                writer.uint32(/* id 15, wireType 1 =*/121).double(message.critRate);
            if (message.critDamageMultiplier != null && Object.hasOwnProperty.call(message, "critDamageMultiplier"))
                writer.uint32(/* id 16, wireType 1 =*/129).double(message.critDamageMultiplier);
            if (message.partTypeDamageSoft != null && Object.hasOwnProperty.call(message, "partTypeDamageSoft"))
                writer.uint32(/* id 17, wireType 1 =*/137).double(message.partTypeDamageSoft);
            if (message.partTypeDamageHeavy != null && Object.hasOwnProperty.call(message, "partTypeDamageHeavy"))
                writer.uint32(/* id 18, wireType 1 =*/145).double(message.partTypeDamageHeavy);
            if (message.partTypeDamageWeak != null && Object.hasOwnProperty.call(message, "partTypeDamageWeak"))
                writer.uint32(/* id 19, wireType 1 =*/153).double(message.partTypeDamageWeak);
            return writer;
        };

        /**
         * Encodes the specified InventoryItem message, length delimited. Does not implicitly {@link realtime.InventoryItem.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.InventoryItem
         * @static
         * @param {realtime.IInventoryItem} message InventoryItem message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InventoryItem.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an InventoryItem message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.InventoryItem
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.InventoryItem} InventoryItem
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InventoryItem.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.InventoryItem();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.itemId = reader.string();
                        break;
                    }
                case 2: {
                        message.instanceId = reader.string();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.slot = reader.string();
                        break;
                    }
                case 5: {
                        message.rarity = reader.string();
                        break;
                    }
                case 6: {
                        message.imagePath = reader.string();
                        break;
                    }
                case 7: {
                        message.imageAlt = reader.string();
                        break;
                    }
                case 8: {
                        message.quantity = reader.int64();
                        break;
                    }
                case 9: {
                        message.equipped = reader.bool();
                        break;
                    }
                case 10: {
                        message.enhanceLevel = reader.int32();
                        break;
                    }
                case 11: {
                        message.bound = reader.bool();
                        break;
                    }
                case 12: {
                        message.locked = reader.bool();
                        break;
                    }
                case 13: {
                        message.attackPower = reader.int64();
                        break;
                    }
                case 14: {
                        message.armorPenPercent = reader.double();
                        break;
                    }
                case 15: {
                        message.critRate = reader.double();
                        break;
                    }
                case 16: {
                        message.critDamageMultiplier = reader.double();
                        break;
                    }
                case 17: {
                        message.partTypeDamageSoft = reader.double();
                        break;
                    }
                case 18: {
                        message.partTypeDamageHeavy = reader.double();
                        break;
                    }
                case 19: {
                        message.partTypeDamageWeak = reader.double();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an InventoryItem message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.InventoryItem
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.InventoryItem} InventoryItem
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InventoryItem.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an InventoryItem message.
         * @function verify
         * @memberof realtime.InventoryItem
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        InventoryItem.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.itemId != null && message.hasOwnProperty("itemId"))
                if (!$util.isString(message.itemId))
                    return "itemId: string expected";
            if (message.instanceId != null && message.hasOwnProperty("instanceId"))
                if (!$util.isString(message.instanceId))
                    return "instanceId: string expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.slot != null && message.hasOwnProperty("slot"))
                if (!$util.isString(message.slot))
                    return "slot: string expected";
            if (message.rarity != null && message.hasOwnProperty("rarity"))
                if (!$util.isString(message.rarity))
                    return "rarity: string expected";
            if (message.imagePath != null && message.hasOwnProperty("imagePath"))
                if (!$util.isString(message.imagePath))
                    return "imagePath: string expected";
            if (message.imageAlt != null && message.hasOwnProperty("imageAlt"))
                if (!$util.isString(message.imageAlt))
                    return "imageAlt: string expected";
            if (message.quantity != null && message.hasOwnProperty("quantity"))
                if (!$util.isInteger(message.quantity) && !(message.quantity && $util.isInteger(message.quantity.low) && $util.isInteger(message.quantity.high)))
                    return "quantity: integer|Long expected";
            if (message.equipped != null && message.hasOwnProperty("equipped"))
                if (typeof message.equipped !== "boolean")
                    return "equipped: boolean expected";
            if (message.enhanceLevel != null && message.hasOwnProperty("enhanceLevel"))
                if (!$util.isInteger(message.enhanceLevel))
                    return "enhanceLevel: integer expected";
            if (message.bound != null && message.hasOwnProperty("bound"))
                if (typeof message.bound !== "boolean")
                    return "bound: boolean expected";
            if (message.locked != null && message.hasOwnProperty("locked"))
                if (typeof message.locked !== "boolean")
                    return "locked: boolean expected";
            if (message.attackPower != null && message.hasOwnProperty("attackPower"))
                if (!$util.isInteger(message.attackPower) && !(message.attackPower && $util.isInteger(message.attackPower.low) && $util.isInteger(message.attackPower.high)))
                    return "attackPower: integer|Long expected";
            if (message.armorPenPercent != null && message.hasOwnProperty("armorPenPercent"))
                if (typeof message.armorPenPercent !== "number")
                    return "armorPenPercent: number expected";
            if (message.critRate != null && message.hasOwnProperty("critRate"))
                if (typeof message.critRate !== "number")
                    return "critRate: number expected";
            if (message.critDamageMultiplier != null && message.hasOwnProperty("critDamageMultiplier"))
                if (typeof message.critDamageMultiplier !== "number")
                    return "critDamageMultiplier: number expected";
            if (message.partTypeDamageSoft != null && message.hasOwnProperty("partTypeDamageSoft"))
                if (typeof message.partTypeDamageSoft !== "number")
                    return "partTypeDamageSoft: number expected";
            if (message.partTypeDamageHeavy != null && message.hasOwnProperty("partTypeDamageHeavy"))
                if (typeof message.partTypeDamageHeavy !== "number")
                    return "partTypeDamageHeavy: number expected";
            if (message.partTypeDamageWeak != null && message.hasOwnProperty("partTypeDamageWeak"))
                if (typeof message.partTypeDamageWeak !== "number")
                    return "partTypeDamageWeak: number expected";
            return null;
        };

        /**
         * Creates an InventoryItem message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.InventoryItem
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.InventoryItem} InventoryItem
         */
        InventoryItem.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.InventoryItem)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.InventoryItem();
            if (object.itemId != null)
                message.itemId = String(object.itemId);
            if (object.instanceId != null)
                message.instanceId = String(object.instanceId);
            if (object.name != null)
                message.name = String(object.name);
            if (object.slot != null)
                message.slot = String(object.slot);
            if (object.rarity != null)
                message.rarity = String(object.rarity);
            if (object.imagePath != null)
                message.imagePath = String(object.imagePath);
            if (object.imageAlt != null)
                message.imageAlt = String(object.imageAlt);
            if (object.quantity != null)
                if ($util.Long)
                    (message.quantity = $util.Long.fromValue(object.quantity)).unsigned = false;
                else if (typeof object.quantity === "string")
                    message.quantity = parseInt(object.quantity, 10);
                else if (typeof object.quantity === "number")
                    message.quantity = object.quantity;
                else if (typeof object.quantity === "object")
                    message.quantity = new $util.LongBits(object.quantity.low >>> 0, object.quantity.high >>> 0).toNumber();
            if (object.equipped != null)
                message.equipped = Boolean(object.equipped);
            if (object.enhanceLevel != null)
                message.enhanceLevel = object.enhanceLevel | 0;
            if (object.bound != null)
                message.bound = Boolean(object.bound);
            if (object.locked != null)
                message.locked = Boolean(object.locked);
            if (object.attackPower != null)
                if ($util.Long)
                    (message.attackPower = $util.Long.fromValue(object.attackPower)).unsigned = false;
                else if (typeof object.attackPower === "string")
                    message.attackPower = parseInt(object.attackPower, 10);
                else if (typeof object.attackPower === "number")
                    message.attackPower = object.attackPower;
                else if (typeof object.attackPower === "object")
                    message.attackPower = new $util.LongBits(object.attackPower.low >>> 0, object.attackPower.high >>> 0).toNumber();
            if (object.armorPenPercent != null)
                message.armorPenPercent = Number(object.armorPenPercent);
            if (object.critRate != null)
                message.critRate = Number(object.critRate);
            if (object.critDamageMultiplier != null)
                message.critDamageMultiplier = Number(object.critDamageMultiplier);
            if (object.partTypeDamageSoft != null)
                message.partTypeDamageSoft = Number(object.partTypeDamageSoft);
            if (object.partTypeDamageHeavy != null)
                message.partTypeDamageHeavy = Number(object.partTypeDamageHeavy);
            if (object.partTypeDamageWeak != null)
                message.partTypeDamageWeak = Number(object.partTypeDamageWeak);
            return message;
        };

        /**
         * Creates a plain object from an InventoryItem message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.InventoryItem
         * @static
         * @param {realtime.InventoryItem} message InventoryItem
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        InventoryItem.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.itemId = "";
                object.instanceId = "";
                object.name = "";
                object.slot = "";
                object.rarity = "";
                object.imagePath = "";
                object.imageAlt = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.quantity = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.quantity = options.longs === String ? "0" : 0;
                object.equipped = false;
                object.enhanceLevel = 0;
                object.bound = false;
                object.locked = false;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.attackPower = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.attackPower = options.longs === String ? "0" : 0;
                object.armorPenPercent = 0;
                object.critRate = 0;
                object.critDamageMultiplier = 0;
                object.partTypeDamageSoft = 0;
                object.partTypeDamageHeavy = 0;
                object.partTypeDamageWeak = 0;
            }
            if (message.itemId != null && message.hasOwnProperty("itemId"))
                object.itemId = message.itemId;
            if (message.instanceId != null && message.hasOwnProperty("instanceId"))
                object.instanceId = message.instanceId;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.slot != null && message.hasOwnProperty("slot"))
                object.slot = message.slot;
            if (message.rarity != null && message.hasOwnProperty("rarity"))
                object.rarity = message.rarity;
            if (message.imagePath != null && message.hasOwnProperty("imagePath"))
                object.imagePath = message.imagePath;
            if (message.imageAlt != null && message.hasOwnProperty("imageAlt"))
                object.imageAlt = message.imageAlt;
            if (message.quantity != null && message.hasOwnProperty("quantity"))
                if (typeof message.quantity === "number")
                    object.quantity = options.longs === String ? String(message.quantity) : message.quantity;
                else
                    object.quantity = options.longs === String ? $util.Long.prototype.toString.call(message.quantity) : options.longs === Number ? new $util.LongBits(message.quantity.low >>> 0, message.quantity.high >>> 0).toNumber() : message.quantity;
            if (message.equipped != null && message.hasOwnProperty("equipped"))
                object.equipped = message.equipped;
            if (message.enhanceLevel != null && message.hasOwnProperty("enhanceLevel"))
                object.enhanceLevel = message.enhanceLevel;
            if (message.bound != null && message.hasOwnProperty("bound"))
                object.bound = message.bound;
            if (message.locked != null && message.hasOwnProperty("locked"))
                object.locked = message.locked;
            if (message.attackPower != null && message.hasOwnProperty("attackPower"))
                if (typeof message.attackPower === "number")
                    object.attackPower = options.longs === String ? String(message.attackPower) : message.attackPower;
                else
                    object.attackPower = options.longs === String ? $util.Long.prototype.toString.call(message.attackPower) : options.longs === Number ? new $util.LongBits(message.attackPower.low >>> 0, message.attackPower.high >>> 0).toNumber() : message.attackPower;
            if (message.armorPenPercent != null && message.hasOwnProperty("armorPenPercent"))
                object.armorPenPercent = options.json && !isFinite(message.armorPenPercent) ? String(message.armorPenPercent) : message.armorPenPercent;
            if (message.critRate != null && message.hasOwnProperty("critRate"))
                object.critRate = options.json && !isFinite(message.critRate) ? String(message.critRate) : message.critRate;
            if (message.critDamageMultiplier != null && message.hasOwnProperty("critDamageMultiplier"))
                object.critDamageMultiplier = options.json && !isFinite(message.critDamageMultiplier) ? String(message.critDamageMultiplier) : message.critDamageMultiplier;
            if (message.partTypeDamageSoft != null && message.hasOwnProperty("partTypeDamageSoft"))
                object.partTypeDamageSoft = options.json && !isFinite(message.partTypeDamageSoft) ? String(message.partTypeDamageSoft) : message.partTypeDamageSoft;
            if (message.partTypeDamageHeavy != null && message.hasOwnProperty("partTypeDamageHeavy"))
                object.partTypeDamageHeavy = options.json && !isFinite(message.partTypeDamageHeavy) ? String(message.partTypeDamageHeavy) : message.partTypeDamageHeavy;
            if (message.partTypeDamageWeak != null && message.hasOwnProperty("partTypeDamageWeak"))
                object.partTypeDamageWeak = options.json && !isFinite(message.partTypeDamageWeak) ? String(message.partTypeDamageWeak) : message.partTypeDamageWeak;
            return object;
        };

        /**
         * Converts this InventoryItem to JSON.
         * @function toJSON
         * @memberof realtime.InventoryItem
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        InventoryItem.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for InventoryItem
         * @function getTypeUrl
         * @memberof realtime.InventoryItem
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InventoryItem.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.InventoryItem";
        };

        return InventoryItem;
    })();

    realtime.Loadout = (function() {

        /**
         * Properties of a Loadout.
         * @memberof realtime
         * @interface ILoadout
         * @property {realtime.IInventoryItem|null} [weapon] Loadout weapon
         * @property {realtime.IInventoryItem|null} [helmet] Loadout helmet
         * @property {realtime.IInventoryItem|null} [chest] Loadout chest
         * @property {realtime.IInventoryItem|null} [gloves] Loadout gloves
         * @property {realtime.IInventoryItem|null} [legs] Loadout legs
         * @property {realtime.IInventoryItem|null} [accessory] Loadout accessory
         */

        /**
         * Constructs a new Loadout.
         * @memberof realtime
         * @classdesc Represents a Loadout.
         * @implements ILoadout
         * @constructor
         * @param {realtime.ILoadout=} [properties] Properties to set
         */
        function Loadout(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Loadout weapon.
         * @member {realtime.IInventoryItem|null|undefined} weapon
         * @memberof realtime.Loadout
         * @instance
         */
        Loadout.prototype.weapon = null;

        /**
         * Loadout helmet.
         * @member {realtime.IInventoryItem|null|undefined} helmet
         * @memberof realtime.Loadout
         * @instance
         */
        Loadout.prototype.helmet = null;

        /**
         * Loadout chest.
         * @member {realtime.IInventoryItem|null|undefined} chest
         * @memberof realtime.Loadout
         * @instance
         */
        Loadout.prototype.chest = null;

        /**
         * Loadout gloves.
         * @member {realtime.IInventoryItem|null|undefined} gloves
         * @memberof realtime.Loadout
         * @instance
         */
        Loadout.prototype.gloves = null;

        /**
         * Loadout legs.
         * @member {realtime.IInventoryItem|null|undefined} legs
         * @memberof realtime.Loadout
         * @instance
         */
        Loadout.prototype.legs = null;

        /**
         * Loadout accessory.
         * @member {realtime.IInventoryItem|null|undefined} accessory
         * @memberof realtime.Loadout
         * @instance
         */
        Loadout.prototype.accessory = null;

        /**
         * Creates a new Loadout instance using the specified properties.
         * @function create
         * @memberof realtime.Loadout
         * @static
         * @param {realtime.ILoadout=} [properties] Properties to set
         * @returns {realtime.Loadout} Loadout instance
         */
        Loadout.create = function create(properties) {
            return new Loadout(properties);
        };

        /**
         * Encodes the specified Loadout message. Does not implicitly {@link realtime.Loadout.verify|verify} messages.
         * @function encode
         * @memberof realtime.Loadout
         * @static
         * @param {realtime.ILoadout} message Loadout message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Loadout.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.weapon != null && Object.hasOwnProperty.call(message, "weapon"))
                $root.realtime.InventoryItem.encode(message.weapon, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.helmet != null && Object.hasOwnProperty.call(message, "helmet"))
                $root.realtime.InventoryItem.encode(message.helmet, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.chest != null && Object.hasOwnProperty.call(message, "chest"))
                $root.realtime.InventoryItem.encode(message.chest, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.gloves != null && Object.hasOwnProperty.call(message, "gloves"))
                $root.realtime.InventoryItem.encode(message.gloves, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            if (message.legs != null && Object.hasOwnProperty.call(message, "legs"))
                $root.realtime.InventoryItem.encode(message.legs, writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
            if (message.accessory != null && Object.hasOwnProperty.call(message, "accessory"))
                $root.realtime.InventoryItem.encode(message.accessory, writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified Loadout message, length delimited. Does not implicitly {@link realtime.Loadout.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.Loadout
         * @static
         * @param {realtime.ILoadout} message Loadout message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Loadout.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Loadout message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.Loadout
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.Loadout} Loadout
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Loadout.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.Loadout();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.weapon = $root.realtime.InventoryItem.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 2: {
                        message.helmet = $root.realtime.InventoryItem.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 3: {
                        message.chest = $root.realtime.InventoryItem.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 4: {
                        message.gloves = $root.realtime.InventoryItem.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 5: {
                        message.legs = $root.realtime.InventoryItem.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                case 6: {
                        message.accessory = $root.realtime.InventoryItem.decode(reader, reader.uint32(), undefined, long + 1);
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Loadout message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.Loadout
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.Loadout} Loadout
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Loadout.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Loadout message.
         * @function verify
         * @memberof realtime.Loadout
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Loadout.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.weapon != null && message.hasOwnProperty("weapon")) {
                let error = $root.realtime.InventoryItem.verify(message.weapon, long + 1);
                if (error)
                    return "weapon." + error;
            }
            if (message.helmet != null && message.hasOwnProperty("helmet")) {
                let error = $root.realtime.InventoryItem.verify(message.helmet, long + 1);
                if (error)
                    return "helmet." + error;
            }
            if (message.chest != null && message.hasOwnProperty("chest")) {
                let error = $root.realtime.InventoryItem.verify(message.chest, long + 1);
                if (error)
                    return "chest." + error;
            }
            if (message.gloves != null && message.hasOwnProperty("gloves")) {
                let error = $root.realtime.InventoryItem.verify(message.gloves, long + 1);
                if (error)
                    return "gloves." + error;
            }
            if (message.legs != null && message.hasOwnProperty("legs")) {
                let error = $root.realtime.InventoryItem.verify(message.legs, long + 1);
                if (error)
                    return "legs." + error;
            }
            if (message.accessory != null && message.hasOwnProperty("accessory")) {
                let error = $root.realtime.InventoryItem.verify(message.accessory, long + 1);
                if (error)
                    return "accessory." + error;
            }
            return null;
        };

        /**
         * Creates a Loadout message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.Loadout
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.Loadout} Loadout
         */
        Loadout.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.Loadout)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.Loadout();
            if (object.weapon != null) {
                if (typeof object.weapon !== "object")
                    throw TypeError(".realtime.Loadout.weapon: object expected");
                message.weapon = $root.realtime.InventoryItem.fromObject(object.weapon, long + 1);
            }
            if (object.helmet != null) {
                if (typeof object.helmet !== "object")
                    throw TypeError(".realtime.Loadout.helmet: object expected");
                message.helmet = $root.realtime.InventoryItem.fromObject(object.helmet, long + 1);
            }
            if (object.chest != null) {
                if (typeof object.chest !== "object")
                    throw TypeError(".realtime.Loadout.chest: object expected");
                message.chest = $root.realtime.InventoryItem.fromObject(object.chest, long + 1);
            }
            if (object.gloves != null) {
                if (typeof object.gloves !== "object")
                    throw TypeError(".realtime.Loadout.gloves: object expected");
                message.gloves = $root.realtime.InventoryItem.fromObject(object.gloves, long + 1);
            }
            if (object.legs != null) {
                if (typeof object.legs !== "object")
                    throw TypeError(".realtime.Loadout.legs: object expected");
                message.legs = $root.realtime.InventoryItem.fromObject(object.legs, long + 1);
            }
            if (object.accessory != null) {
                if (typeof object.accessory !== "object")
                    throw TypeError(".realtime.Loadout.accessory: object expected");
                message.accessory = $root.realtime.InventoryItem.fromObject(object.accessory, long + 1);
            }
            return message;
        };

        /**
         * Creates a plain object from a Loadout message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.Loadout
         * @static
         * @param {realtime.Loadout} message Loadout
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Loadout.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.weapon = null;
                object.helmet = null;
                object.chest = null;
                object.gloves = null;
                object.legs = null;
                object.accessory = null;
            }
            if (message.weapon != null && message.hasOwnProperty("weapon"))
                object.weapon = $root.realtime.InventoryItem.toObject(message.weapon, options);
            if (message.helmet != null && message.hasOwnProperty("helmet"))
                object.helmet = $root.realtime.InventoryItem.toObject(message.helmet, options);
            if (message.chest != null && message.hasOwnProperty("chest"))
                object.chest = $root.realtime.InventoryItem.toObject(message.chest, options);
            if (message.gloves != null && message.hasOwnProperty("gloves"))
                object.gloves = $root.realtime.InventoryItem.toObject(message.gloves, options);
            if (message.legs != null && message.hasOwnProperty("legs"))
                object.legs = $root.realtime.InventoryItem.toObject(message.legs, options);
            if (message.accessory != null && message.hasOwnProperty("accessory"))
                object.accessory = $root.realtime.InventoryItem.toObject(message.accessory, options);
            return object;
        };

        /**
         * Converts this Loadout to JSON.
         * @function toJSON
         * @memberof realtime.Loadout
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Loadout.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Loadout
         * @function getTypeUrl
         * @memberof realtime.Loadout
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Loadout.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.Loadout";
        };

        return Loadout;
    })();

    realtime.CombatStats = (function() {

        /**
         * Properties of a CombatStats.
         * @memberof realtime
         * @interface ICombatStats
         * @property {number|Long|null} [effectiveIncrement] CombatStats effectiveIncrement
         * @property {number|Long|null} [normalDamage] CombatStats normalDamage
         * @property {number|null} [criticalChancePercent] CombatStats criticalChancePercent
         * @property {number|Long|null} [criticalDamage] CombatStats criticalDamage
         * @property {number|Long|null} [attackPower] CombatStats attackPower
         * @property {number|null} [armorPenPercent] CombatStats armorPenPercent
         * @property {number|null} [critDamageMultiplier] CombatStats critDamageMultiplier
         * @property {number|null} [allDamageAmplify] CombatStats allDamageAmplify
         * @property {number|null} [partTypeDamageSoft] CombatStats partTypeDamageSoft
         * @property {number|null} [partTypeDamageHeavy] CombatStats partTypeDamageHeavy
         * @property {number|null} [partTypeDamageWeak] CombatStats partTypeDamageWeak
         * @property {number|null} [perPartDamagePercent] CombatStats perPartDamagePercent
         * @property {number|null} [lowHpMultiplier] CombatStats lowHpMultiplier
         * @property {number|null} [lowHpThreshold] CombatStats lowHpThreshold
         */

        /**
         * Constructs a new CombatStats.
         * @memberof realtime
         * @classdesc Represents a CombatStats.
         * @implements ICombatStats
         * @constructor
         * @param {realtime.ICombatStats=} [properties] Properties to set
         */
        function CombatStats(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CombatStats effectiveIncrement.
         * @member {number|Long} effectiveIncrement
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.effectiveIncrement = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CombatStats normalDamage.
         * @member {number|Long} normalDamage
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.normalDamage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CombatStats criticalChancePercent.
         * @member {number} criticalChancePercent
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.criticalChancePercent = 0;

        /**
         * CombatStats criticalDamage.
         * @member {number|Long} criticalDamage
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.criticalDamage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CombatStats attackPower.
         * @member {number|Long} attackPower
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.attackPower = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CombatStats armorPenPercent.
         * @member {number} armorPenPercent
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.armorPenPercent = 0;

        /**
         * CombatStats critDamageMultiplier.
         * @member {number} critDamageMultiplier
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.critDamageMultiplier = 0;

        /**
         * CombatStats allDamageAmplify.
         * @member {number} allDamageAmplify
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.allDamageAmplify = 0;

        /**
         * CombatStats partTypeDamageSoft.
         * @member {number} partTypeDamageSoft
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.partTypeDamageSoft = 0;

        /**
         * CombatStats partTypeDamageHeavy.
         * @member {number} partTypeDamageHeavy
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.partTypeDamageHeavy = 0;

        /**
         * CombatStats partTypeDamageWeak.
         * @member {number} partTypeDamageWeak
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.partTypeDamageWeak = 0;

        /**
         * CombatStats perPartDamagePercent.
         * @member {number} perPartDamagePercent
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.perPartDamagePercent = 0;

        /**
         * CombatStats lowHpMultiplier.
         * @member {number} lowHpMultiplier
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.lowHpMultiplier = 0;

        /**
         * CombatStats lowHpThreshold.
         * @member {number} lowHpThreshold
         * @memberof realtime.CombatStats
         * @instance
         */
        CombatStats.prototype.lowHpThreshold = 0;

        /**
         * Creates a new CombatStats instance using the specified properties.
         * @function create
         * @memberof realtime.CombatStats
         * @static
         * @param {realtime.ICombatStats=} [properties] Properties to set
         * @returns {realtime.CombatStats} CombatStats instance
         */
        CombatStats.create = function create(properties) {
            return new CombatStats(properties);
        };

        /**
         * Encodes the specified CombatStats message. Does not implicitly {@link realtime.CombatStats.verify|verify} messages.
         * @function encode
         * @memberof realtime.CombatStats
         * @static
         * @param {realtime.ICombatStats} message CombatStats message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CombatStats.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.effectiveIncrement != null && Object.hasOwnProperty.call(message, "effectiveIncrement"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.effectiveIncrement);
            if (message.normalDamage != null && Object.hasOwnProperty.call(message, "normalDamage"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.normalDamage);
            if (message.criticalChancePercent != null && Object.hasOwnProperty.call(message, "criticalChancePercent"))
                writer.uint32(/* id 3, wireType 1 =*/25).double(message.criticalChancePercent);
            if (message.criticalDamage != null && Object.hasOwnProperty.call(message, "criticalDamage"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.criticalDamage);
            if (message.attackPower != null && Object.hasOwnProperty.call(message, "attackPower"))
                writer.uint32(/* id 5, wireType 0 =*/40).int64(message.attackPower);
            if (message.armorPenPercent != null && Object.hasOwnProperty.call(message, "armorPenPercent"))
                writer.uint32(/* id 6, wireType 1 =*/49).double(message.armorPenPercent);
            if (message.critDamageMultiplier != null && Object.hasOwnProperty.call(message, "critDamageMultiplier"))
                writer.uint32(/* id 7, wireType 1 =*/57).double(message.critDamageMultiplier);
            if (message.allDamageAmplify != null && Object.hasOwnProperty.call(message, "allDamageAmplify"))
                writer.uint32(/* id 8, wireType 1 =*/65).double(message.allDamageAmplify);
            if (message.partTypeDamageSoft != null && Object.hasOwnProperty.call(message, "partTypeDamageSoft"))
                writer.uint32(/* id 9, wireType 1 =*/73).double(message.partTypeDamageSoft);
            if (message.partTypeDamageHeavy != null && Object.hasOwnProperty.call(message, "partTypeDamageHeavy"))
                writer.uint32(/* id 10, wireType 1 =*/81).double(message.partTypeDamageHeavy);
            if (message.partTypeDamageWeak != null && Object.hasOwnProperty.call(message, "partTypeDamageWeak"))
                writer.uint32(/* id 11, wireType 1 =*/89).double(message.partTypeDamageWeak);
            if (message.perPartDamagePercent != null && Object.hasOwnProperty.call(message, "perPartDamagePercent"))
                writer.uint32(/* id 12, wireType 1 =*/97).double(message.perPartDamagePercent);
            if (message.lowHpMultiplier != null && Object.hasOwnProperty.call(message, "lowHpMultiplier"))
                writer.uint32(/* id 13, wireType 1 =*/105).double(message.lowHpMultiplier);
            if (message.lowHpThreshold != null && Object.hasOwnProperty.call(message, "lowHpThreshold"))
                writer.uint32(/* id 14, wireType 1 =*/113).double(message.lowHpThreshold);
            return writer;
        };

        /**
         * Encodes the specified CombatStats message, length delimited. Does not implicitly {@link realtime.CombatStats.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.CombatStats
         * @static
         * @param {realtime.ICombatStats} message CombatStats message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CombatStats.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CombatStats message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.CombatStats
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.CombatStats} CombatStats
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CombatStats.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.CombatStats();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.effectiveIncrement = reader.int64();
                        break;
                    }
                case 2: {
                        message.normalDamage = reader.int64();
                        break;
                    }
                case 3: {
                        message.criticalChancePercent = reader.double();
                        break;
                    }
                case 4: {
                        message.criticalDamage = reader.int64();
                        break;
                    }
                case 5: {
                        message.attackPower = reader.int64();
                        break;
                    }
                case 6: {
                        message.armorPenPercent = reader.double();
                        break;
                    }
                case 7: {
                        message.critDamageMultiplier = reader.double();
                        break;
                    }
                case 8: {
                        message.allDamageAmplify = reader.double();
                        break;
                    }
                case 9: {
                        message.partTypeDamageSoft = reader.double();
                        break;
                    }
                case 10: {
                        message.partTypeDamageHeavy = reader.double();
                        break;
                    }
                case 11: {
                        message.partTypeDamageWeak = reader.double();
                        break;
                    }
                case 12: {
                        message.perPartDamagePercent = reader.double();
                        break;
                    }
                case 13: {
                        message.lowHpMultiplier = reader.double();
                        break;
                    }
                case 14: {
                        message.lowHpThreshold = reader.double();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a CombatStats message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.CombatStats
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.CombatStats} CombatStats
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CombatStats.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CombatStats message.
         * @function verify
         * @memberof realtime.CombatStats
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CombatStats.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.effectiveIncrement != null && message.hasOwnProperty("effectiveIncrement"))
                if (!$util.isInteger(message.effectiveIncrement) && !(message.effectiveIncrement && $util.isInteger(message.effectiveIncrement.low) && $util.isInteger(message.effectiveIncrement.high)))
                    return "effectiveIncrement: integer|Long expected";
            if (message.normalDamage != null && message.hasOwnProperty("normalDamage"))
                if (!$util.isInteger(message.normalDamage) && !(message.normalDamage && $util.isInteger(message.normalDamage.low) && $util.isInteger(message.normalDamage.high)))
                    return "normalDamage: integer|Long expected";
            if (message.criticalChancePercent != null && message.hasOwnProperty("criticalChancePercent"))
                if (typeof message.criticalChancePercent !== "number")
                    return "criticalChancePercent: number expected";
            if (message.criticalDamage != null && message.hasOwnProperty("criticalDamage"))
                if (!$util.isInteger(message.criticalDamage) && !(message.criticalDamage && $util.isInteger(message.criticalDamage.low) && $util.isInteger(message.criticalDamage.high)))
                    return "criticalDamage: integer|Long expected";
            if (message.attackPower != null && message.hasOwnProperty("attackPower"))
                if (!$util.isInteger(message.attackPower) && !(message.attackPower && $util.isInteger(message.attackPower.low) && $util.isInteger(message.attackPower.high)))
                    return "attackPower: integer|Long expected";
            if (message.armorPenPercent != null && message.hasOwnProperty("armorPenPercent"))
                if (typeof message.armorPenPercent !== "number")
                    return "armorPenPercent: number expected";
            if (message.critDamageMultiplier != null && message.hasOwnProperty("critDamageMultiplier"))
                if (typeof message.critDamageMultiplier !== "number")
                    return "critDamageMultiplier: number expected";
            if (message.allDamageAmplify != null && message.hasOwnProperty("allDamageAmplify"))
                if (typeof message.allDamageAmplify !== "number")
                    return "allDamageAmplify: number expected";
            if (message.partTypeDamageSoft != null && message.hasOwnProperty("partTypeDamageSoft"))
                if (typeof message.partTypeDamageSoft !== "number")
                    return "partTypeDamageSoft: number expected";
            if (message.partTypeDamageHeavy != null && message.hasOwnProperty("partTypeDamageHeavy"))
                if (typeof message.partTypeDamageHeavy !== "number")
                    return "partTypeDamageHeavy: number expected";
            if (message.partTypeDamageWeak != null && message.hasOwnProperty("partTypeDamageWeak"))
                if (typeof message.partTypeDamageWeak !== "number")
                    return "partTypeDamageWeak: number expected";
            if (message.perPartDamagePercent != null && message.hasOwnProperty("perPartDamagePercent"))
                if (typeof message.perPartDamagePercent !== "number")
                    return "perPartDamagePercent: number expected";
            if (message.lowHpMultiplier != null && message.hasOwnProperty("lowHpMultiplier"))
                if (typeof message.lowHpMultiplier !== "number")
                    return "lowHpMultiplier: number expected";
            if (message.lowHpThreshold != null && message.hasOwnProperty("lowHpThreshold"))
                if (typeof message.lowHpThreshold !== "number")
                    return "lowHpThreshold: number expected";
            return null;
        };

        /**
         * Creates a CombatStats message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.CombatStats
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.CombatStats} CombatStats
         */
        CombatStats.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.CombatStats)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.CombatStats();
            if (object.effectiveIncrement != null)
                if ($util.Long)
                    (message.effectiveIncrement = $util.Long.fromValue(object.effectiveIncrement)).unsigned = false;
                else if (typeof object.effectiveIncrement === "string")
                    message.effectiveIncrement = parseInt(object.effectiveIncrement, 10);
                else if (typeof object.effectiveIncrement === "number")
                    message.effectiveIncrement = object.effectiveIncrement;
                else if (typeof object.effectiveIncrement === "object")
                    message.effectiveIncrement = new $util.LongBits(object.effectiveIncrement.low >>> 0, object.effectiveIncrement.high >>> 0).toNumber();
            if (object.normalDamage != null)
                if ($util.Long)
                    (message.normalDamage = $util.Long.fromValue(object.normalDamage)).unsigned = false;
                else if (typeof object.normalDamage === "string")
                    message.normalDamage = parseInt(object.normalDamage, 10);
                else if (typeof object.normalDamage === "number")
                    message.normalDamage = object.normalDamage;
                else if (typeof object.normalDamage === "object")
                    message.normalDamage = new $util.LongBits(object.normalDamage.low >>> 0, object.normalDamage.high >>> 0).toNumber();
            if (object.criticalChancePercent != null)
                message.criticalChancePercent = Number(object.criticalChancePercent);
            if (object.criticalDamage != null)
                if ($util.Long)
                    (message.criticalDamage = $util.Long.fromValue(object.criticalDamage)).unsigned = false;
                else if (typeof object.criticalDamage === "string")
                    message.criticalDamage = parseInt(object.criticalDamage, 10);
                else if (typeof object.criticalDamage === "number")
                    message.criticalDamage = object.criticalDamage;
                else if (typeof object.criticalDamage === "object")
                    message.criticalDamage = new $util.LongBits(object.criticalDamage.low >>> 0, object.criticalDamage.high >>> 0).toNumber();
            if (object.attackPower != null)
                if ($util.Long)
                    (message.attackPower = $util.Long.fromValue(object.attackPower)).unsigned = false;
                else if (typeof object.attackPower === "string")
                    message.attackPower = parseInt(object.attackPower, 10);
                else if (typeof object.attackPower === "number")
                    message.attackPower = object.attackPower;
                else if (typeof object.attackPower === "object")
                    message.attackPower = new $util.LongBits(object.attackPower.low >>> 0, object.attackPower.high >>> 0).toNumber();
            if (object.armorPenPercent != null)
                message.armorPenPercent = Number(object.armorPenPercent);
            if (object.critDamageMultiplier != null)
                message.critDamageMultiplier = Number(object.critDamageMultiplier);
            if (object.allDamageAmplify != null)
                message.allDamageAmplify = Number(object.allDamageAmplify);
            if (object.partTypeDamageSoft != null)
                message.partTypeDamageSoft = Number(object.partTypeDamageSoft);
            if (object.partTypeDamageHeavy != null)
                message.partTypeDamageHeavy = Number(object.partTypeDamageHeavy);
            if (object.partTypeDamageWeak != null)
                message.partTypeDamageWeak = Number(object.partTypeDamageWeak);
            if (object.perPartDamagePercent != null)
                message.perPartDamagePercent = Number(object.perPartDamagePercent);
            if (object.lowHpMultiplier != null)
                message.lowHpMultiplier = Number(object.lowHpMultiplier);
            if (object.lowHpThreshold != null)
                message.lowHpThreshold = Number(object.lowHpThreshold);
            return message;
        };

        /**
         * Creates a plain object from a CombatStats message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.CombatStats
         * @static
         * @param {realtime.CombatStats} message CombatStats
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CombatStats.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.effectiveIncrement = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.effectiveIncrement = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.normalDamage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.normalDamage = options.longs === String ? "0" : 0;
                object.criticalChancePercent = 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.criticalDamage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.criticalDamage = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.attackPower = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.attackPower = options.longs === String ? "0" : 0;
                object.armorPenPercent = 0;
                object.critDamageMultiplier = 0;
                object.allDamageAmplify = 0;
                object.partTypeDamageSoft = 0;
                object.partTypeDamageHeavy = 0;
                object.partTypeDamageWeak = 0;
                object.perPartDamagePercent = 0;
                object.lowHpMultiplier = 0;
                object.lowHpThreshold = 0;
            }
            if (message.effectiveIncrement != null && message.hasOwnProperty("effectiveIncrement"))
                if (typeof message.effectiveIncrement === "number")
                    object.effectiveIncrement = options.longs === String ? String(message.effectiveIncrement) : message.effectiveIncrement;
                else
                    object.effectiveIncrement = options.longs === String ? $util.Long.prototype.toString.call(message.effectiveIncrement) : options.longs === Number ? new $util.LongBits(message.effectiveIncrement.low >>> 0, message.effectiveIncrement.high >>> 0).toNumber() : message.effectiveIncrement;
            if (message.normalDamage != null && message.hasOwnProperty("normalDamage"))
                if (typeof message.normalDamage === "number")
                    object.normalDamage = options.longs === String ? String(message.normalDamage) : message.normalDamage;
                else
                    object.normalDamage = options.longs === String ? $util.Long.prototype.toString.call(message.normalDamage) : options.longs === Number ? new $util.LongBits(message.normalDamage.low >>> 0, message.normalDamage.high >>> 0).toNumber() : message.normalDamage;
            if (message.criticalChancePercent != null && message.hasOwnProperty("criticalChancePercent"))
                object.criticalChancePercent = options.json && !isFinite(message.criticalChancePercent) ? String(message.criticalChancePercent) : message.criticalChancePercent;
            if (message.criticalDamage != null && message.hasOwnProperty("criticalDamage"))
                if (typeof message.criticalDamage === "number")
                    object.criticalDamage = options.longs === String ? String(message.criticalDamage) : message.criticalDamage;
                else
                    object.criticalDamage = options.longs === String ? $util.Long.prototype.toString.call(message.criticalDamage) : options.longs === Number ? new $util.LongBits(message.criticalDamage.low >>> 0, message.criticalDamage.high >>> 0).toNumber() : message.criticalDamage;
            if (message.attackPower != null && message.hasOwnProperty("attackPower"))
                if (typeof message.attackPower === "number")
                    object.attackPower = options.longs === String ? String(message.attackPower) : message.attackPower;
                else
                    object.attackPower = options.longs === String ? $util.Long.prototype.toString.call(message.attackPower) : options.longs === Number ? new $util.LongBits(message.attackPower.low >>> 0, message.attackPower.high >>> 0).toNumber() : message.attackPower;
            if (message.armorPenPercent != null && message.hasOwnProperty("armorPenPercent"))
                object.armorPenPercent = options.json && !isFinite(message.armorPenPercent) ? String(message.armorPenPercent) : message.armorPenPercent;
            if (message.critDamageMultiplier != null && message.hasOwnProperty("critDamageMultiplier"))
                object.critDamageMultiplier = options.json && !isFinite(message.critDamageMultiplier) ? String(message.critDamageMultiplier) : message.critDamageMultiplier;
            if (message.allDamageAmplify != null && message.hasOwnProperty("allDamageAmplify"))
                object.allDamageAmplify = options.json && !isFinite(message.allDamageAmplify) ? String(message.allDamageAmplify) : message.allDamageAmplify;
            if (message.partTypeDamageSoft != null && message.hasOwnProperty("partTypeDamageSoft"))
                object.partTypeDamageSoft = options.json && !isFinite(message.partTypeDamageSoft) ? String(message.partTypeDamageSoft) : message.partTypeDamageSoft;
            if (message.partTypeDamageHeavy != null && message.hasOwnProperty("partTypeDamageHeavy"))
                object.partTypeDamageHeavy = options.json && !isFinite(message.partTypeDamageHeavy) ? String(message.partTypeDamageHeavy) : message.partTypeDamageHeavy;
            if (message.partTypeDamageWeak != null && message.hasOwnProperty("partTypeDamageWeak"))
                object.partTypeDamageWeak = options.json && !isFinite(message.partTypeDamageWeak) ? String(message.partTypeDamageWeak) : message.partTypeDamageWeak;
            if (message.perPartDamagePercent != null && message.hasOwnProperty("perPartDamagePercent"))
                object.perPartDamagePercent = options.json && !isFinite(message.perPartDamagePercent) ? String(message.perPartDamagePercent) : message.perPartDamagePercent;
            if (message.lowHpMultiplier != null && message.hasOwnProperty("lowHpMultiplier"))
                object.lowHpMultiplier = options.json && !isFinite(message.lowHpMultiplier) ? String(message.lowHpMultiplier) : message.lowHpMultiplier;
            if (message.lowHpThreshold != null && message.hasOwnProperty("lowHpThreshold"))
                object.lowHpThreshold = options.json && !isFinite(message.lowHpThreshold) ? String(message.lowHpThreshold) : message.lowHpThreshold;
            return object;
        };

        /**
         * Converts this CombatStats to JSON.
         * @function toJSON
         * @memberof realtime.CombatStats
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CombatStats.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CombatStats
         * @function getTypeUrl
         * @memberof realtime.CombatStats
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CombatStats.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.CombatStats";
        };

        return CombatStats;
    })();

    realtime.Reward = (function() {

        /**
         * Properties of a Reward.
         * @memberof realtime
         * @interface IReward
         * @property {string|null} [bossId] Reward bossId
         * @property {string|null} [bossName] Reward bossName
         * @property {string|null} [itemId] Reward itemId
         * @property {string|null} [itemName] Reward itemName
         * @property {number|Long|null} [grantedAt] Reward grantedAt
         */

        /**
         * Constructs a new Reward.
         * @memberof realtime
         * @classdesc Represents a Reward.
         * @implements IReward
         * @constructor
         * @param {realtime.IReward=} [properties] Properties to set
         */
        function Reward(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Reward bossId.
         * @member {string} bossId
         * @memberof realtime.Reward
         * @instance
         */
        Reward.prototype.bossId = "";

        /**
         * Reward bossName.
         * @member {string} bossName
         * @memberof realtime.Reward
         * @instance
         */
        Reward.prototype.bossName = "";

        /**
         * Reward itemId.
         * @member {string} itemId
         * @memberof realtime.Reward
         * @instance
         */
        Reward.prototype.itemId = "";

        /**
         * Reward itemName.
         * @member {string} itemName
         * @memberof realtime.Reward
         * @instance
         */
        Reward.prototype.itemName = "";

        /**
         * Reward grantedAt.
         * @member {number|Long} grantedAt
         * @memberof realtime.Reward
         * @instance
         */
        Reward.prototype.grantedAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new Reward instance using the specified properties.
         * @function create
         * @memberof realtime.Reward
         * @static
         * @param {realtime.IReward=} [properties] Properties to set
         * @returns {realtime.Reward} Reward instance
         */
        Reward.create = function create(properties) {
            return new Reward(properties);
        };

        /**
         * Encodes the specified Reward message. Does not implicitly {@link realtime.Reward.verify|verify} messages.
         * @function encode
         * @memberof realtime.Reward
         * @static
         * @param {realtime.IReward} message Reward message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Reward.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.bossId != null && Object.hasOwnProperty.call(message, "bossId"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.bossId);
            if (message.bossName != null && Object.hasOwnProperty.call(message, "bossName"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.bossName);
            if (message.itemId != null && Object.hasOwnProperty.call(message, "itemId"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.itemId);
            if (message.itemName != null && Object.hasOwnProperty.call(message, "itemName"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.itemName);
            if (message.grantedAt != null && Object.hasOwnProperty.call(message, "grantedAt"))
                writer.uint32(/* id 5, wireType 0 =*/40).int64(message.grantedAt);
            return writer;
        };

        /**
         * Encodes the specified Reward message, length delimited. Does not implicitly {@link realtime.Reward.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.Reward
         * @static
         * @param {realtime.IReward} message Reward message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Reward.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Reward message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.Reward
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.Reward} Reward
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Reward.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.Reward();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.bossId = reader.string();
                        break;
                    }
                case 2: {
                        message.bossName = reader.string();
                        break;
                    }
                case 3: {
                        message.itemId = reader.string();
                        break;
                    }
                case 4: {
                        message.itemName = reader.string();
                        break;
                    }
                case 5: {
                        message.grantedAt = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Reward message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.Reward
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.Reward} Reward
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Reward.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Reward message.
         * @function verify
         * @memberof realtime.Reward
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Reward.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.bossId != null && message.hasOwnProperty("bossId"))
                if (!$util.isString(message.bossId))
                    return "bossId: string expected";
            if (message.bossName != null && message.hasOwnProperty("bossName"))
                if (!$util.isString(message.bossName))
                    return "bossName: string expected";
            if (message.itemId != null && message.hasOwnProperty("itemId"))
                if (!$util.isString(message.itemId))
                    return "itemId: string expected";
            if (message.itemName != null && message.hasOwnProperty("itemName"))
                if (!$util.isString(message.itemName))
                    return "itemName: string expected";
            if (message.grantedAt != null && message.hasOwnProperty("grantedAt"))
                if (!$util.isInteger(message.grantedAt) && !(message.grantedAt && $util.isInteger(message.grantedAt.low) && $util.isInteger(message.grantedAt.high)))
                    return "grantedAt: integer|Long expected";
            return null;
        };

        /**
         * Creates a Reward message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.Reward
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.Reward} Reward
         */
        Reward.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.Reward)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.Reward();
            if (object.bossId != null)
                message.bossId = String(object.bossId);
            if (object.bossName != null)
                message.bossName = String(object.bossName);
            if (object.itemId != null)
                message.itemId = String(object.itemId);
            if (object.itemName != null)
                message.itemName = String(object.itemName);
            if (object.grantedAt != null)
                if ($util.Long)
                    (message.grantedAt = $util.Long.fromValue(object.grantedAt)).unsigned = false;
                else if (typeof object.grantedAt === "string")
                    message.grantedAt = parseInt(object.grantedAt, 10);
                else if (typeof object.grantedAt === "number")
                    message.grantedAt = object.grantedAt;
                else if (typeof object.grantedAt === "object")
                    message.grantedAt = new $util.LongBits(object.grantedAt.low >>> 0, object.grantedAt.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a Reward message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.Reward
         * @static
         * @param {realtime.Reward} message Reward
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Reward.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.bossId = "";
                object.bossName = "";
                object.itemId = "";
                object.itemName = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.grantedAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.grantedAt = options.longs === String ? "0" : 0;
            }
            if (message.bossId != null && message.hasOwnProperty("bossId"))
                object.bossId = message.bossId;
            if (message.bossName != null && message.hasOwnProperty("bossName"))
                object.bossName = message.bossName;
            if (message.itemId != null && message.hasOwnProperty("itemId"))
                object.itemId = message.itemId;
            if (message.itemName != null && message.hasOwnProperty("itemName"))
                object.itemName = message.itemName;
            if (message.grantedAt != null && message.hasOwnProperty("grantedAt"))
                if (typeof message.grantedAt === "number")
                    object.grantedAt = options.longs === String ? String(message.grantedAt) : message.grantedAt;
                else
                    object.grantedAt = options.longs === String ? $util.Long.prototype.toString.call(message.grantedAt) : options.longs === Number ? new $util.LongBits(message.grantedAt.low >>> 0, message.grantedAt.high >>> 0).toNumber() : message.grantedAt;
            return object;
        };

        /**
         * Converts this Reward to JSON.
         * @function toJSON
         * @memberof realtime.Reward
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Reward.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Reward
         * @function getTypeUrl
         * @memberof realtime.Reward
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Reward.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.Reward";
        };

        return Reward;
    })();

    realtime.TalentTriggerEvent = (function() {

        /**
         * Properties of a TalentTriggerEvent.
         * @memberof realtime
         * @interface ITalentTriggerEvent
         * @property {string|null} [talentId] TalentTriggerEvent talentId
         * @property {string|null} [name] TalentTriggerEvent name
         * @property {string|null} [effectType] TalentTriggerEvent effectType
         * @property {number|Long|null} [extraDamage] TalentTriggerEvent extraDamage
         * @property {string|null} [message] TalentTriggerEvent message
         * @property {number|null} [partX] TalentTriggerEvent partX
         * @property {number|null} [partY] TalentTriggerEvent partY
         */

        /**
         * Constructs a new TalentTriggerEvent.
         * @memberof realtime
         * @classdesc Represents a TalentTriggerEvent.
         * @implements ITalentTriggerEvent
         * @constructor
         * @param {realtime.ITalentTriggerEvent=} [properties] Properties to set
         */
        function TalentTriggerEvent(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TalentTriggerEvent talentId.
         * @member {string} talentId
         * @memberof realtime.TalentTriggerEvent
         * @instance
         */
        TalentTriggerEvent.prototype.talentId = "";

        /**
         * TalentTriggerEvent name.
         * @member {string} name
         * @memberof realtime.TalentTriggerEvent
         * @instance
         */
        TalentTriggerEvent.prototype.name = "";

        /**
         * TalentTriggerEvent effectType.
         * @member {string} effectType
         * @memberof realtime.TalentTriggerEvent
         * @instance
         */
        TalentTriggerEvent.prototype.effectType = "";

        /**
         * TalentTriggerEvent extraDamage.
         * @member {number|Long} extraDamage
         * @memberof realtime.TalentTriggerEvent
         * @instance
         */
        TalentTriggerEvent.prototype.extraDamage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentTriggerEvent message.
         * @member {string} message
         * @memberof realtime.TalentTriggerEvent
         * @instance
         */
        TalentTriggerEvent.prototype.message = "";

        /**
         * TalentTriggerEvent partX.
         * @member {number} partX
         * @memberof realtime.TalentTriggerEvent
         * @instance
         */
        TalentTriggerEvent.prototype.partX = 0;

        /**
         * TalentTriggerEvent partY.
         * @member {number} partY
         * @memberof realtime.TalentTriggerEvent
         * @instance
         */
        TalentTriggerEvent.prototype.partY = 0;

        /**
         * Creates a new TalentTriggerEvent instance using the specified properties.
         * @function create
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {realtime.ITalentTriggerEvent=} [properties] Properties to set
         * @returns {realtime.TalentTriggerEvent} TalentTriggerEvent instance
         */
        TalentTriggerEvent.create = function create(properties) {
            return new TalentTriggerEvent(properties);
        };

        /**
         * Encodes the specified TalentTriggerEvent message. Does not implicitly {@link realtime.TalentTriggerEvent.verify|verify} messages.
         * @function encode
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {realtime.ITalentTriggerEvent} message TalentTriggerEvent message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TalentTriggerEvent.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.talentId != null && Object.hasOwnProperty.call(message, "talentId"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.talentId);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
            if (message.effectType != null && Object.hasOwnProperty.call(message, "effectType"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.effectType);
            if (message.extraDamage != null && Object.hasOwnProperty.call(message, "extraDamage"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.extraDamage);
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.message);
            if (message.partX != null && Object.hasOwnProperty.call(message, "partX"))
                writer.uint32(/* id 6, wireType 0 =*/48).int32(message.partX);
            if (message.partY != null && Object.hasOwnProperty.call(message, "partY"))
                writer.uint32(/* id 7, wireType 0 =*/56).int32(message.partY);
            return writer;
        };

        /**
         * Encodes the specified TalentTriggerEvent message, length delimited. Does not implicitly {@link realtime.TalentTriggerEvent.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {realtime.ITalentTriggerEvent} message TalentTriggerEvent message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TalentTriggerEvent.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a TalentTriggerEvent message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.TalentTriggerEvent} TalentTriggerEvent
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TalentTriggerEvent.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.TalentTriggerEvent();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.talentId = reader.string();
                        break;
                    }
                case 2: {
                        message.name = reader.string();
                        break;
                    }
                case 3: {
                        message.effectType = reader.string();
                        break;
                    }
                case 4: {
                        message.extraDamage = reader.int64();
                        break;
                    }
                case 5: {
                        message.message = reader.string();
                        break;
                    }
                case 6: {
                        message.partX = reader.int32();
                        break;
                    }
                case 7: {
                        message.partY = reader.int32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a TalentTriggerEvent message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.TalentTriggerEvent} TalentTriggerEvent
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TalentTriggerEvent.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a TalentTriggerEvent message.
         * @function verify
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        TalentTriggerEvent.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.talentId != null && message.hasOwnProperty("talentId"))
                if (!$util.isString(message.talentId))
                    return "talentId: string expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.effectType != null && message.hasOwnProperty("effectType"))
                if (!$util.isString(message.effectType))
                    return "effectType: string expected";
            if (message.extraDamage != null && message.hasOwnProperty("extraDamage"))
                if (!$util.isInteger(message.extraDamage) && !(message.extraDamage && $util.isInteger(message.extraDamage.low) && $util.isInteger(message.extraDamage.high)))
                    return "extraDamage: integer|Long expected";
            if (message.message != null && message.hasOwnProperty("message"))
                if (!$util.isString(message.message))
                    return "message: string expected";
            if (message.partX != null && message.hasOwnProperty("partX"))
                if (!$util.isInteger(message.partX))
                    return "partX: integer expected";
            if (message.partY != null && message.hasOwnProperty("partY"))
                if (!$util.isInteger(message.partY))
                    return "partY: integer expected";
            return null;
        };

        /**
         * Creates a TalentTriggerEvent message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.TalentTriggerEvent} TalentTriggerEvent
         */
        TalentTriggerEvent.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.TalentTriggerEvent)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.TalentTriggerEvent();
            if (object.talentId != null)
                message.talentId = String(object.talentId);
            if (object.name != null)
                message.name = String(object.name);
            if (object.effectType != null)
                message.effectType = String(object.effectType);
            if (object.extraDamage != null)
                if ($util.Long)
                    (message.extraDamage = $util.Long.fromValue(object.extraDamage)).unsigned = false;
                else if (typeof object.extraDamage === "string")
                    message.extraDamage = parseInt(object.extraDamage, 10);
                else if (typeof object.extraDamage === "number")
                    message.extraDamage = object.extraDamage;
                else if (typeof object.extraDamage === "object")
                    message.extraDamage = new $util.LongBits(object.extraDamage.low >>> 0, object.extraDamage.high >>> 0).toNumber();
            if (object.message != null)
                message.message = String(object.message);
            if (object.partX != null)
                message.partX = object.partX | 0;
            if (object.partY != null)
                message.partY = object.partY | 0;
            return message;
        };

        /**
         * Creates a plain object from a TalentTriggerEvent message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {realtime.TalentTriggerEvent} message TalentTriggerEvent
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        TalentTriggerEvent.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.talentId = "";
                object.name = "";
                object.effectType = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.extraDamage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.extraDamage = options.longs === String ? "0" : 0;
                object.message = "";
                object.partX = 0;
                object.partY = 0;
            }
            if (message.talentId != null && message.hasOwnProperty("talentId"))
                object.talentId = message.talentId;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.effectType != null && message.hasOwnProperty("effectType"))
                object.effectType = message.effectType;
            if (message.extraDamage != null && message.hasOwnProperty("extraDamage"))
                if (typeof message.extraDamage === "number")
                    object.extraDamage = options.longs === String ? String(message.extraDamage) : message.extraDamage;
                else
                    object.extraDamage = options.longs === String ? $util.Long.prototype.toString.call(message.extraDamage) : options.longs === Number ? new $util.LongBits(message.extraDamage.low >>> 0, message.extraDamage.high >>> 0).toNumber() : message.extraDamage;
            if (message.message != null && message.hasOwnProperty("message"))
                object.message = message.message;
            if (message.partX != null && message.hasOwnProperty("partX"))
                object.partX = message.partX;
            if (message.partY != null && message.hasOwnProperty("partY"))
                object.partY = message.partY;
            return object;
        };

        /**
         * Converts this TalentTriggerEvent to JSON.
         * @function toJSON
         * @memberof realtime.TalentTriggerEvent
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        TalentTriggerEvent.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for TalentTriggerEvent
         * @function getTypeUrl
         * @memberof realtime.TalentTriggerEvent
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TalentTriggerEvent.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.TalentTriggerEvent";
        };

        return TalentTriggerEvent;
    })();

    realtime.BossPartStateDelta = (function() {

        /**
         * Properties of a BossPartStateDelta.
         * @memberof realtime
         * @interface IBossPartStateDelta
         * @property {number|null} [x] BossPartStateDelta x
         * @property {number|null} [y] BossPartStateDelta y
         * @property {number|Long|null} [damage] BossPartStateDelta damage
         * @property {number|Long|null} [beforeHp] BossPartStateDelta beforeHp
         * @property {number|Long|null} [afterHp] BossPartStateDelta afterHp
         * @property {string|null} [partType] BossPartStateDelta partType
         */

        /**
         * Constructs a new BossPartStateDelta.
         * @memberof realtime
         * @classdesc Represents a BossPartStateDelta.
         * @implements IBossPartStateDelta
         * @constructor
         * @param {realtime.IBossPartStateDelta=} [properties] Properties to set
         */
        function BossPartStateDelta(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * BossPartStateDelta x.
         * @member {number} x
         * @memberof realtime.BossPartStateDelta
         * @instance
         */
        BossPartStateDelta.prototype.x = 0;

        /**
         * BossPartStateDelta y.
         * @member {number} y
         * @memberof realtime.BossPartStateDelta
         * @instance
         */
        BossPartStateDelta.prototype.y = 0;

        /**
         * BossPartStateDelta damage.
         * @member {number|Long} damage
         * @memberof realtime.BossPartStateDelta
         * @instance
         */
        BossPartStateDelta.prototype.damage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * BossPartStateDelta beforeHp.
         * @member {number|Long} beforeHp
         * @memberof realtime.BossPartStateDelta
         * @instance
         */
        BossPartStateDelta.prototype.beforeHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * BossPartStateDelta afterHp.
         * @member {number|Long} afterHp
         * @memberof realtime.BossPartStateDelta
         * @instance
         */
        BossPartStateDelta.prototype.afterHp = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * BossPartStateDelta partType.
         * @member {string} partType
         * @memberof realtime.BossPartStateDelta
         * @instance
         */
        BossPartStateDelta.prototype.partType = "";

        /**
         * Creates a new BossPartStateDelta instance using the specified properties.
         * @function create
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {realtime.IBossPartStateDelta=} [properties] Properties to set
         * @returns {realtime.BossPartStateDelta} BossPartStateDelta instance
         */
        BossPartStateDelta.create = function create(properties) {
            return new BossPartStateDelta(properties);
        };

        /**
         * Encodes the specified BossPartStateDelta message. Does not implicitly {@link realtime.BossPartStateDelta.verify|verify} messages.
         * @function encode
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {realtime.IBossPartStateDelta} message BossPartStateDelta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossPartStateDelta.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.x != null && Object.hasOwnProperty.call(message, "x"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.x);
            if (message.y != null && Object.hasOwnProperty.call(message, "y"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.y);
            if (message.damage != null && Object.hasOwnProperty.call(message, "damage"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.damage);
            if (message.beforeHp != null && Object.hasOwnProperty.call(message, "beforeHp"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.beforeHp);
            if (message.afterHp != null && Object.hasOwnProperty.call(message, "afterHp"))
                writer.uint32(/* id 5, wireType 0 =*/40).int64(message.afterHp);
            if (message.partType != null && Object.hasOwnProperty.call(message, "partType"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.partType);
            return writer;
        };

        /**
         * Encodes the specified BossPartStateDelta message, length delimited. Does not implicitly {@link realtime.BossPartStateDelta.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {realtime.IBossPartStateDelta} message BossPartStateDelta message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BossPartStateDelta.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a BossPartStateDelta message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.BossPartStateDelta} BossPartStateDelta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossPartStateDelta.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.BossPartStateDelta();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.x = reader.int32();
                        break;
                    }
                case 2: {
                        message.y = reader.int32();
                        break;
                    }
                case 3: {
                        message.damage = reader.int64();
                        break;
                    }
                case 4: {
                        message.beforeHp = reader.int64();
                        break;
                    }
                case 5: {
                        message.afterHp = reader.int64();
                        break;
                    }
                case 6: {
                        message.partType = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a BossPartStateDelta message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.BossPartStateDelta} BossPartStateDelta
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BossPartStateDelta.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a BossPartStateDelta message.
         * @function verify
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        BossPartStateDelta.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.x != null && message.hasOwnProperty("x"))
                if (!$util.isInteger(message.x))
                    return "x: integer expected";
            if (message.y != null && message.hasOwnProperty("y"))
                if (!$util.isInteger(message.y))
                    return "y: integer expected";
            if (message.damage != null && message.hasOwnProperty("damage"))
                if (!$util.isInteger(message.damage) && !(message.damage && $util.isInteger(message.damage.low) && $util.isInteger(message.damage.high)))
                    return "damage: integer|Long expected";
            if (message.beforeHp != null && message.hasOwnProperty("beforeHp"))
                if (!$util.isInteger(message.beforeHp) && !(message.beforeHp && $util.isInteger(message.beforeHp.low) && $util.isInteger(message.beforeHp.high)))
                    return "beforeHp: integer|Long expected";
            if (message.afterHp != null && message.hasOwnProperty("afterHp"))
                if (!$util.isInteger(message.afterHp) && !(message.afterHp && $util.isInteger(message.afterHp.low) && $util.isInteger(message.afterHp.high)))
                    return "afterHp: integer|Long expected";
            if (message.partType != null && message.hasOwnProperty("partType"))
                if (!$util.isString(message.partType))
                    return "partType: string expected";
            return null;
        };

        /**
         * Creates a BossPartStateDelta message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.BossPartStateDelta} BossPartStateDelta
         */
        BossPartStateDelta.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.BossPartStateDelta)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.BossPartStateDelta();
            if (object.x != null)
                message.x = object.x | 0;
            if (object.y != null)
                message.y = object.y | 0;
            if (object.damage != null)
                if ($util.Long)
                    (message.damage = $util.Long.fromValue(object.damage)).unsigned = false;
                else if (typeof object.damage === "string")
                    message.damage = parseInt(object.damage, 10);
                else if (typeof object.damage === "number")
                    message.damage = object.damage;
                else if (typeof object.damage === "object")
                    message.damage = new $util.LongBits(object.damage.low >>> 0, object.damage.high >>> 0).toNumber();
            if (object.beforeHp != null)
                if ($util.Long)
                    (message.beforeHp = $util.Long.fromValue(object.beforeHp)).unsigned = false;
                else if (typeof object.beforeHp === "string")
                    message.beforeHp = parseInt(object.beforeHp, 10);
                else if (typeof object.beforeHp === "number")
                    message.beforeHp = object.beforeHp;
                else if (typeof object.beforeHp === "object")
                    message.beforeHp = new $util.LongBits(object.beforeHp.low >>> 0, object.beforeHp.high >>> 0).toNumber();
            if (object.afterHp != null)
                if ($util.Long)
                    (message.afterHp = $util.Long.fromValue(object.afterHp)).unsigned = false;
                else if (typeof object.afterHp === "string")
                    message.afterHp = parseInt(object.afterHp, 10);
                else if (typeof object.afterHp === "number")
                    message.afterHp = object.afterHp;
                else if (typeof object.afterHp === "object")
                    message.afterHp = new $util.LongBits(object.afterHp.low >>> 0, object.afterHp.high >>> 0).toNumber();
            if (object.partType != null)
                message.partType = String(object.partType);
            return message;
        };

        /**
         * Creates a plain object from a BossPartStateDelta message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {realtime.BossPartStateDelta} message BossPartStateDelta
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        BossPartStateDelta.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.x = 0;
                object.y = 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.damage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.damage = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.beforeHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.beforeHp = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.afterHp = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.afterHp = options.longs === String ? "0" : 0;
                object.partType = "";
            }
            if (message.x != null && message.hasOwnProperty("x"))
                object.x = message.x;
            if (message.y != null && message.hasOwnProperty("y"))
                object.y = message.y;
            if (message.damage != null && message.hasOwnProperty("damage"))
                if (typeof message.damage === "number")
                    object.damage = options.longs === String ? String(message.damage) : message.damage;
                else
                    object.damage = options.longs === String ? $util.Long.prototype.toString.call(message.damage) : options.longs === Number ? new $util.LongBits(message.damage.low >>> 0, message.damage.high >>> 0).toNumber() : message.damage;
            if (message.beforeHp != null && message.hasOwnProperty("beforeHp"))
                if (typeof message.beforeHp === "number")
                    object.beforeHp = options.longs === String ? String(message.beforeHp) : message.beforeHp;
                else
                    object.beforeHp = options.longs === String ? $util.Long.prototype.toString.call(message.beforeHp) : options.longs === Number ? new $util.LongBits(message.beforeHp.low >>> 0, message.beforeHp.high >>> 0).toNumber() : message.beforeHp;
            if (message.afterHp != null && message.hasOwnProperty("afterHp"))
                if (typeof message.afterHp === "number")
                    object.afterHp = options.longs === String ? String(message.afterHp) : message.afterHp;
                else
                    object.afterHp = options.longs === String ? $util.Long.prototype.toString.call(message.afterHp) : options.longs === Number ? new $util.LongBits(message.afterHp.low >>> 0, message.afterHp.high >>> 0).toNumber() : message.afterHp;
            if (message.partType != null && message.hasOwnProperty("partType"))
                object.partType = message.partType;
            return object;
        };

        /**
         * Converts this BossPartStateDelta to JSON.
         * @function toJSON
         * @memberof realtime.BossPartStateDelta
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        BossPartStateDelta.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for BossPartStateDelta
         * @function getTypeUrl
         * @memberof realtime.BossPartStateDelta
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BossPartStateDelta.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.BossPartStateDelta";
        };

        return BossPartStateDelta;
    })();

    realtime.TalentBleedState = (function() {

        /**
         * Properties of a TalentBleedState.
         * @memberof realtime
         * @interface ITalentBleedState
         * @property {number|Long|null} [startedAtMs] TalentBleedState startedAtMs
         * @property {number|Long|null} [nextTickAtMs] TalentBleedState nextTickAtMs
         * @property {number|Long|null} [endsAtMs] TalentBleedState endsAtMs
         * @property {number|Long|null} [durationMs] TalentBleedState durationMs
         * @property {number|Long|null} [tickIntervalMs] TalentBleedState tickIntervalMs
         * @property {number|Long|null} [totalTicks] TalentBleedState totalTicks
         * @property {number|Long|null} [appliedTicks] TalentBleedState appliedTicks
         * @property {number|Long|null} [totalDamage] TalentBleedState totalDamage
         * @property {number|Long|null} [appliedDamage] TalentBleedState appliedDamage
         */

        /**
         * Constructs a new TalentBleedState.
         * @memberof realtime
         * @classdesc Represents a TalentBleedState.
         * @implements ITalentBleedState
         * @constructor
         * @param {realtime.ITalentBleedState=} [properties] Properties to set
         */
        function TalentBleedState(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TalentBleedState startedAtMs.
         * @member {number|Long} startedAtMs
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.startedAtMs = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState nextTickAtMs.
         * @member {number|Long} nextTickAtMs
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.nextTickAtMs = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState endsAtMs.
         * @member {number|Long} endsAtMs
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.endsAtMs = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState durationMs.
         * @member {number|Long} durationMs
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.durationMs = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState tickIntervalMs.
         * @member {number|Long} tickIntervalMs
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.tickIntervalMs = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState totalTicks.
         * @member {number|Long} totalTicks
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.totalTicks = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState appliedTicks.
         * @member {number|Long} appliedTicks
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.appliedTicks = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState totalDamage.
         * @member {number|Long} totalDamage
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.totalDamage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentBleedState appliedDamage.
         * @member {number|Long} appliedDamage
         * @memberof realtime.TalentBleedState
         * @instance
         */
        TalentBleedState.prototype.appliedDamage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new TalentBleedState instance using the specified properties.
         * @function create
         * @memberof realtime.TalentBleedState
         * @static
         * @param {realtime.ITalentBleedState=} [properties] Properties to set
         * @returns {realtime.TalentBleedState} TalentBleedState instance
         */
        TalentBleedState.create = function create(properties) {
            return new TalentBleedState(properties);
        };

        /**
         * Encodes the specified TalentBleedState message. Does not implicitly {@link realtime.TalentBleedState.verify|verify} messages.
         * @function encode
         * @memberof realtime.TalentBleedState
         * @static
         * @param {realtime.ITalentBleedState} message TalentBleedState message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TalentBleedState.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.startedAtMs != null && Object.hasOwnProperty.call(message, "startedAtMs"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.startedAtMs);
            if (message.nextTickAtMs != null && Object.hasOwnProperty.call(message, "nextTickAtMs"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.nextTickAtMs);
            if (message.endsAtMs != null && Object.hasOwnProperty.call(message, "endsAtMs"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.endsAtMs);
            if (message.durationMs != null && Object.hasOwnProperty.call(message, "durationMs"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.durationMs);
            if (message.tickIntervalMs != null && Object.hasOwnProperty.call(message, "tickIntervalMs"))
                writer.uint32(/* id 5, wireType 0 =*/40).int64(message.tickIntervalMs);
            if (message.totalTicks != null && Object.hasOwnProperty.call(message, "totalTicks"))
                writer.uint32(/* id 6, wireType 0 =*/48).int64(message.totalTicks);
            if (message.appliedTicks != null && Object.hasOwnProperty.call(message, "appliedTicks"))
                writer.uint32(/* id 7, wireType 0 =*/56).int64(message.appliedTicks);
            if (message.totalDamage != null && Object.hasOwnProperty.call(message, "totalDamage"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.totalDamage);
            if (message.appliedDamage != null && Object.hasOwnProperty.call(message, "appliedDamage"))
                writer.uint32(/* id 9, wireType 0 =*/72).int64(message.appliedDamage);
            return writer;
        };

        /**
         * Encodes the specified TalentBleedState message, length delimited. Does not implicitly {@link realtime.TalentBleedState.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.TalentBleedState
         * @static
         * @param {realtime.ITalentBleedState} message TalentBleedState message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TalentBleedState.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a TalentBleedState message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.TalentBleedState
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.TalentBleedState} TalentBleedState
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TalentBleedState.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.TalentBleedState();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.startedAtMs = reader.int64();
                        break;
                    }
                case 2: {
                        message.nextTickAtMs = reader.int64();
                        break;
                    }
                case 3: {
                        message.endsAtMs = reader.int64();
                        break;
                    }
                case 4: {
                        message.durationMs = reader.int64();
                        break;
                    }
                case 5: {
                        message.tickIntervalMs = reader.int64();
                        break;
                    }
                case 6: {
                        message.totalTicks = reader.int64();
                        break;
                    }
                case 7: {
                        message.appliedTicks = reader.int64();
                        break;
                    }
                case 8: {
                        message.totalDamage = reader.int64();
                        break;
                    }
                case 9: {
                        message.appliedDamage = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a TalentBleedState message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.TalentBleedState
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.TalentBleedState} TalentBleedState
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TalentBleedState.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a TalentBleedState message.
         * @function verify
         * @memberof realtime.TalentBleedState
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        TalentBleedState.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.startedAtMs != null && message.hasOwnProperty("startedAtMs"))
                if (!$util.isInteger(message.startedAtMs) && !(message.startedAtMs && $util.isInteger(message.startedAtMs.low) && $util.isInteger(message.startedAtMs.high)))
                    return "startedAtMs: integer|Long expected";
            if (message.nextTickAtMs != null && message.hasOwnProperty("nextTickAtMs"))
                if (!$util.isInteger(message.nextTickAtMs) && !(message.nextTickAtMs && $util.isInteger(message.nextTickAtMs.low) && $util.isInteger(message.nextTickAtMs.high)))
                    return "nextTickAtMs: integer|Long expected";
            if (message.endsAtMs != null && message.hasOwnProperty("endsAtMs"))
                if (!$util.isInteger(message.endsAtMs) && !(message.endsAtMs && $util.isInteger(message.endsAtMs.low) && $util.isInteger(message.endsAtMs.high)))
                    return "endsAtMs: integer|Long expected";
            if (message.durationMs != null && message.hasOwnProperty("durationMs"))
                if (!$util.isInteger(message.durationMs) && !(message.durationMs && $util.isInteger(message.durationMs.low) && $util.isInteger(message.durationMs.high)))
                    return "durationMs: integer|Long expected";
            if (message.tickIntervalMs != null && message.hasOwnProperty("tickIntervalMs"))
                if (!$util.isInteger(message.tickIntervalMs) && !(message.tickIntervalMs && $util.isInteger(message.tickIntervalMs.low) && $util.isInteger(message.tickIntervalMs.high)))
                    return "tickIntervalMs: integer|Long expected";
            if (message.totalTicks != null && message.hasOwnProperty("totalTicks"))
                if (!$util.isInteger(message.totalTicks) && !(message.totalTicks && $util.isInteger(message.totalTicks.low) && $util.isInteger(message.totalTicks.high)))
                    return "totalTicks: integer|Long expected";
            if (message.appliedTicks != null && message.hasOwnProperty("appliedTicks"))
                if (!$util.isInteger(message.appliedTicks) && !(message.appliedTicks && $util.isInteger(message.appliedTicks.low) && $util.isInteger(message.appliedTicks.high)))
                    return "appliedTicks: integer|Long expected";
            if (message.totalDamage != null && message.hasOwnProperty("totalDamage"))
                if (!$util.isInteger(message.totalDamage) && !(message.totalDamage && $util.isInteger(message.totalDamage.low) && $util.isInteger(message.totalDamage.high)))
                    return "totalDamage: integer|Long expected";
            if (message.appliedDamage != null && message.hasOwnProperty("appliedDamage"))
                if (!$util.isInteger(message.appliedDamage) && !(message.appliedDamage && $util.isInteger(message.appliedDamage.low) && $util.isInteger(message.appliedDamage.high)))
                    return "appliedDamage: integer|Long expected";
            return null;
        };

        /**
         * Creates a TalentBleedState message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.TalentBleedState
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.TalentBleedState} TalentBleedState
         */
        TalentBleedState.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.TalentBleedState)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.TalentBleedState();
            if (object.startedAtMs != null)
                if ($util.Long)
                    (message.startedAtMs = $util.Long.fromValue(object.startedAtMs)).unsigned = false;
                else if (typeof object.startedAtMs === "string")
                    message.startedAtMs = parseInt(object.startedAtMs, 10);
                else if (typeof object.startedAtMs === "number")
                    message.startedAtMs = object.startedAtMs;
                else if (typeof object.startedAtMs === "object")
                    message.startedAtMs = new $util.LongBits(object.startedAtMs.low >>> 0, object.startedAtMs.high >>> 0).toNumber();
            if (object.nextTickAtMs != null)
                if ($util.Long)
                    (message.nextTickAtMs = $util.Long.fromValue(object.nextTickAtMs)).unsigned = false;
                else if (typeof object.nextTickAtMs === "string")
                    message.nextTickAtMs = parseInt(object.nextTickAtMs, 10);
                else if (typeof object.nextTickAtMs === "number")
                    message.nextTickAtMs = object.nextTickAtMs;
                else if (typeof object.nextTickAtMs === "object")
                    message.nextTickAtMs = new $util.LongBits(object.nextTickAtMs.low >>> 0, object.nextTickAtMs.high >>> 0).toNumber();
            if (object.endsAtMs != null)
                if ($util.Long)
                    (message.endsAtMs = $util.Long.fromValue(object.endsAtMs)).unsigned = false;
                else if (typeof object.endsAtMs === "string")
                    message.endsAtMs = parseInt(object.endsAtMs, 10);
                else if (typeof object.endsAtMs === "number")
                    message.endsAtMs = object.endsAtMs;
                else if (typeof object.endsAtMs === "object")
                    message.endsAtMs = new $util.LongBits(object.endsAtMs.low >>> 0, object.endsAtMs.high >>> 0).toNumber();
            if (object.durationMs != null)
                if ($util.Long)
                    (message.durationMs = $util.Long.fromValue(object.durationMs)).unsigned = false;
                else if (typeof object.durationMs === "string")
                    message.durationMs = parseInt(object.durationMs, 10);
                else if (typeof object.durationMs === "number")
                    message.durationMs = object.durationMs;
                else if (typeof object.durationMs === "object")
                    message.durationMs = new $util.LongBits(object.durationMs.low >>> 0, object.durationMs.high >>> 0).toNumber();
            if (object.tickIntervalMs != null)
                if ($util.Long)
                    (message.tickIntervalMs = $util.Long.fromValue(object.tickIntervalMs)).unsigned = false;
                else if (typeof object.tickIntervalMs === "string")
                    message.tickIntervalMs = parseInt(object.tickIntervalMs, 10);
                else if (typeof object.tickIntervalMs === "number")
                    message.tickIntervalMs = object.tickIntervalMs;
                else if (typeof object.tickIntervalMs === "object")
                    message.tickIntervalMs = new $util.LongBits(object.tickIntervalMs.low >>> 0, object.tickIntervalMs.high >>> 0).toNumber();
            if (object.totalTicks != null)
                if ($util.Long)
                    (message.totalTicks = $util.Long.fromValue(object.totalTicks)).unsigned = false;
                else if (typeof object.totalTicks === "string")
                    message.totalTicks = parseInt(object.totalTicks, 10);
                else if (typeof object.totalTicks === "number")
                    message.totalTicks = object.totalTicks;
                else if (typeof object.totalTicks === "object")
                    message.totalTicks = new $util.LongBits(object.totalTicks.low >>> 0, object.totalTicks.high >>> 0).toNumber();
            if (object.appliedTicks != null)
                if ($util.Long)
                    (message.appliedTicks = $util.Long.fromValue(object.appliedTicks)).unsigned = false;
                else if (typeof object.appliedTicks === "string")
                    message.appliedTicks = parseInt(object.appliedTicks, 10);
                else if (typeof object.appliedTicks === "number")
                    message.appliedTicks = object.appliedTicks;
                else if (typeof object.appliedTicks === "object")
                    message.appliedTicks = new $util.LongBits(object.appliedTicks.low >>> 0, object.appliedTicks.high >>> 0).toNumber();
            if (object.totalDamage != null)
                if ($util.Long)
                    (message.totalDamage = $util.Long.fromValue(object.totalDamage)).unsigned = false;
                else if (typeof object.totalDamage === "string")
                    message.totalDamage = parseInt(object.totalDamage, 10);
                else if (typeof object.totalDamage === "number")
                    message.totalDamage = object.totalDamage;
                else if (typeof object.totalDamage === "object")
                    message.totalDamage = new $util.LongBits(object.totalDamage.low >>> 0, object.totalDamage.high >>> 0).toNumber();
            if (object.appliedDamage != null)
                if ($util.Long)
                    (message.appliedDamage = $util.Long.fromValue(object.appliedDamage)).unsigned = false;
                else if (typeof object.appliedDamage === "string")
                    message.appliedDamage = parseInt(object.appliedDamage, 10);
                else if (typeof object.appliedDamage === "number")
                    message.appliedDamage = object.appliedDamage;
                else if (typeof object.appliedDamage === "object")
                    message.appliedDamage = new $util.LongBits(object.appliedDamage.low >>> 0, object.appliedDamage.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a TalentBleedState message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.TalentBleedState
         * @static
         * @param {realtime.TalentBleedState} message TalentBleedState
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        TalentBleedState.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.startedAtMs = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.startedAtMs = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.nextTickAtMs = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.nextTickAtMs = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.endsAtMs = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.endsAtMs = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.durationMs = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.durationMs = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.tickIntervalMs = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.tickIntervalMs = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.totalTicks = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.totalTicks = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.appliedTicks = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.appliedTicks = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.totalDamage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.totalDamage = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.appliedDamage = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.appliedDamage = options.longs === String ? "0" : 0;
            }
            if (message.startedAtMs != null && message.hasOwnProperty("startedAtMs"))
                if (typeof message.startedAtMs === "number")
                    object.startedAtMs = options.longs === String ? String(message.startedAtMs) : message.startedAtMs;
                else
                    object.startedAtMs = options.longs === String ? $util.Long.prototype.toString.call(message.startedAtMs) : options.longs === Number ? new $util.LongBits(message.startedAtMs.low >>> 0, message.startedAtMs.high >>> 0).toNumber() : message.startedAtMs;
            if (message.nextTickAtMs != null && message.hasOwnProperty("nextTickAtMs"))
                if (typeof message.nextTickAtMs === "number")
                    object.nextTickAtMs = options.longs === String ? String(message.nextTickAtMs) : message.nextTickAtMs;
                else
                    object.nextTickAtMs = options.longs === String ? $util.Long.prototype.toString.call(message.nextTickAtMs) : options.longs === Number ? new $util.LongBits(message.nextTickAtMs.low >>> 0, message.nextTickAtMs.high >>> 0).toNumber() : message.nextTickAtMs;
            if (message.endsAtMs != null && message.hasOwnProperty("endsAtMs"))
                if (typeof message.endsAtMs === "number")
                    object.endsAtMs = options.longs === String ? String(message.endsAtMs) : message.endsAtMs;
                else
                    object.endsAtMs = options.longs === String ? $util.Long.prototype.toString.call(message.endsAtMs) : options.longs === Number ? new $util.LongBits(message.endsAtMs.low >>> 0, message.endsAtMs.high >>> 0).toNumber() : message.endsAtMs;
            if (message.durationMs != null && message.hasOwnProperty("durationMs"))
                if (typeof message.durationMs === "number")
                    object.durationMs = options.longs === String ? String(message.durationMs) : message.durationMs;
                else
                    object.durationMs = options.longs === String ? $util.Long.prototype.toString.call(message.durationMs) : options.longs === Number ? new $util.LongBits(message.durationMs.low >>> 0, message.durationMs.high >>> 0).toNumber() : message.durationMs;
            if (message.tickIntervalMs != null && message.hasOwnProperty("tickIntervalMs"))
                if (typeof message.tickIntervalMs === "number")
                    object.tickIntervalMs = options.longs === String ? String(message.tickIntervalMs) : message.tickIntervalMs;
                else
                    object.tickIntervalMs = options.longs === String ? $util.Long.prototype.toString.call(message.tickIntervalMs) : options.longs === Number ? new $util.LongBits(message.tickIntervalMs.low >>> 0, message.tickIntervalMs.high >>> 0).toNumber() : message.tickIntervalMs;
            if (message.totalTicks != null && message.hasOwnProperty("totalTicks"))
                if (typeof message.totalTicks === "number")
                    object.totalTicks = options.longs === String ? String(message.totalTicks) : message.totalTicks;
                else
                    object.totalTicks = options.longs === String ? $util.Long.prototype.toString.call(message.totalTicks) : options.longs === Number ? new $util.LongBits(message.totalTicks.low >>> 0, message.totalTicks.high >>> 0).toNumber() : message.totalTicks;
            if (message.appliedTicks != null && message.hasOwnProperty("appliedTicks"))
                if (typeof message.appliedTicks === "number")
                    object.appliedTicks = options.longs === String ? String(message.appliedTicks) : message.appliedTicks;
                else
                    object.appliedTicks = options.longs === String ? $util.Long.prototype.toString.call(message.appliedTicks) : options.longs === Number ? new $util.LongBits(message.appliedTicks.low >>> 0, message.appliedTicks.high >>> 0).toNumber() : message.appliedTicks;
            if (message.totalDamage != null && message.hasOwnProperty("totalDamage"))
                if (typeof message.totalDamage === "number")
                    object.totalDamage = options.longs === String ? String(message.totalDamage) : message.totalDamage;
                else
                    object.totalDamage = options.longs === String ? $util.Long.prototype.toString.call(message.totalDamage) : options.longs === Number ? new $util.LongBits(message.totalDamage.low >>> 0, message.totalDamage.high >>> 0).toNumber() : message.totalDamage;
            if (message.appliedDamage != null && message.hasOwnProperty("appliedDamage"))
                if (typeof message.appliedDamage === "number")
                    object.appliedDamage = options.longs === String ? String(message.appliedDamage) : message.appliedDamage;
                else
                    object.appliedDamage = options.longs === String ? $util.Long.prototype.toString.call(message.appliedDamage) : options.longs === Number ? new $util.LongBits(message.appliedDamage.low >>> 0, message.appliedDamage.high >>> 0).toNumber() : message.appliedDamage;
            return object;
        };

        /**
         * Converts this TalentBleedState to JSON.
         * @function toJSON
         * @memberof realtime.TalentBleedState
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        TalentBleedState.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for TalentBleedState
         * @function getTypeUrl
         * @memberof realtime.TalentBleedState
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TalentBleedState.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.TalentBleedState";
        };

        return TalentBleedState;
    })();

    realtime.TalentCombatState = (function() {

        /**
         * Properties of a TalentCombatState.
         * @memberof realtime
         * @interface ITalentCombatState
         * @property {number|null} [omenStacks] TalentCombatState omenStacks
         * @property {Object.<string,realtime.ITalentBleedState>|null} [bleeds] TalentCombatState bleeds
         * @property {Array.<number>|null} [collapseParts] TalentCombatState collapseParts
         * @property {number|Long|null} [collapseEndsAt] TalentCombatState collapseEndsAt
         * @property {number|Long|null} [collapseDuration] TalentCombatState collapseDuration
         * @property {Array.<number>|null} [doomMarks] TalentCombatState doomMarks
         * @property {boolean|null} [hasTriggeredDoom] TalentCombatState hasTriggeredDoom
         * @property {Object.<string,number|Long>|null} [doomMarkCumDamage] TalentCombatState doomMarkCumDamage
         * @property {number|null} [silverStormRemaining] TalentCombatState silverStormRemaining
         * @property {number|Long|null} [silverStormEndsAt] TalentCombatState silverStormEndsAt
         * @property {boolean|null} [silverStormActive] TalentCombatState silverStormActive
         * @property {string|null} [autoStrikeTargetPart] TalentCombatState autoStrikeTargetPart
         * @property {number|Long|null} [autoStrikeComboCount] TalentCombatState autoStrikeComboCount
         * @property {number|Long|null} [autoStrikeExpiresAt] TalentCombatState autoStrikeExpiresAt
         * @property {number|Long|null} [lastFinalCutAt] TalentCombatState lastFinalCutAt
         * @property {Object.<string,number|Long>|null} [judgmentDayUsed] TalentCombatState judgmentDayUsed
         * @property {number|Long|null} [judgmentDayCooldownSec] TalentCombatState judgmentDayCooldownSec
         * @property {Object.<string,number|Long>|null} [partHeavyClickCount] TalentCombatState partHeavyClickCount
         * @property {Object.<string,number|Long>|null} [partJudgmentDayCount] TalentCombatState partJudgmentDayCount
         * @property {Object.<string,number|Long>|null} [partRetainedClicks] TalentCombatState partRetainedClicks
         * @property {Object.<string,number|Long>|null} [partStormComboCount] TalentCombatState partStormComboCount
         * @property {Object.<string,number|Long>|null} [skinnerParts] TalentCombatState skinnerParts
         * @property {Object.<string,number|Long>|null} [skinnerDurationByPart] TalentCombatState skinnerDurationByPart
         * @property {number|Long|null} [skinnerCooldownEndsAt] TalentCombatState skinnerCooldownEndsAt
         * @property {number|Long|null} [skinnerCooldownDuration] TalentCombatState skinnerCooldownDuration
         * @property {number|Long|null} [normalTriggerCount] TalentCombatState normalTriggerCount
         * @property {number|Long|null} [armorTriggerCount] TalentCombatState armorTriggerCount
         * @property {number|Long|null} [judgmentDayTriggerCount] TalentCombatState judgmentDayTriggerCount
         * @property {number|Long|null} [autoStrikeTriggerCount] TalentCombatState autoStrikeTriggerCount
         * @property {number|Long|null} [autoStrikeWindowSec] TalentCombatState autoStrikeWindowSec
         */

        /**
         * Constructs a new TalentCombatState.
         * @memberof realtime
         * @classdesc Represents a TalentCombatState.
         * @implements ITalentCombatState
         * @constructor
         * @param {realtime.ITalentCombatState=} [properties] Properties to set
         */
        function TalentCombatState(properties) {
            this.bleeds = {};
            this.collapseParts = [];
            this.doomMarks = [];
            this.doomMarkCumDamage = {};
            this.judgmentDayUsed = {};
            this.partHeavyClickCount = {};
            this.partJudgmentDayCount = {};
            this.partRetainedClicks = {};
            this.partStormComboCount = {};
            this.skinnerParts = {};
            this.skinnerDurationByPart = {};
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null && keys[i] !== "__proto__")
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TalentCombatState omenStacks.
         * @member {number} omenStacks
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.omenStacks = 0;

        /**
         * TalentCombatState bleeds.
         * @member {Object.<string,realtime.ITalentBleedState>} bleeds
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.bleeds = $util.emptyObject;

        /**
         * TalentCombatState collapseParts.
         * @member {Array.<number>} collapseParts
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.collapseParts = $util.emptyArray;

        /**
         * TalentCombatState collapseEndsAt.
         * @member {number|Long} collapseEndsAt
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.collapseEndsAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState collapseDuration.
         * @member {number|Long} collapseDuration
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.collapseDuration = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState doomMarks.
         * @member {Array.<number>} doomMarks
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.doomMarks = $util.emptyArray;

        /**
         * TalentCombatState hasTriggeredDoom.
         * @member {boolean} hasTriggeredDoom
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.hasTriggeredDoom = false;

        /**
         * TalentCombatState doomMarkCumDamage.
         * @member {Object.<string,number|Long>} doomMarkCumDamage
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.doomMarkCumDamage = $util.emptyObject;

        /**
         * TalentCombatState silverStormRemaining.
         * @member {number} silverStormRemaining
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.silverStormRemaining = 0;

        /**
         * TalentCombatState silverStormEndsAt.
         * @member {number|Long} silverStormEndsAt
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.silverStormEndsAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState silverStormActive.
         * @member {boolean} silverStormActive
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.silverStormActive = false;

        /**
         * TalentCombatState autoStrikeTargetPart.
         * @member {string} autoStrikeTargetPart
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.autoStrikeTargetPart = "";

        /**
         * TalentCombatState autoStrikeComboCount.
         * @member {number|Long} autoStrikeComboCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.autoStrikeComboCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState autoStrikeExpiresAt.
         * @member {number|Long} autoStrikeExpiresAt
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.autoStrikeExpiresAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState lastFinalCutAt.
         * @member {number|Long} lastFinalCutAt
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.lastFinalCutAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState judgmentDayUsed.
         * @member {Object.<string,number|Long>} judgmentDayUsed
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.judgmentDayUsed = $util.emptyObject;

        /**
         * TalentCombatState judgmentDayCooldownSec.
         * @member {number|Long} judgmentDayCooldownSec
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.judgmentDayCooldownSec = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState partHeavyClickCount.
         * @member {Object.<string,number|Long>} partHeavyClickCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.partHeavyClickCount = $util.emptyObject;

        /**
         * TalentCombatState partJudgmentDayCount.
         * @member {Object.<string,number|Long>} partJudgmentDayCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.partJudgmentDayCount = $util.emptyObject;

        /**
         * TalentCombatState partRetainedClicks.
         * @member {Object.<string,number|Long>} partRetainedClicks
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.partRetainedClicks = $util.emptyObject;

        /**
         * TalentCombatState partStormComboCount.
         * @member {Object.<string,number|Long>} partStormComboCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.partStormComboCount = $util.emptyObject;

        /**
         * TalentCombatState skinnerParts.
         * @member {Object.<string,number|Long>} skinnerParts
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.skinnerParts = $util.emptyObject;

        /**
         * TalentCombatState skinnerDurationByPart.
         * @member {Object.<string,number|Long>} skinnerDurationByPart
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.skinnerDurationByPart = $util.emptyObject;

        /**
         * TalentCombatState skinnerCooldownEndsAt.
         * @member {number|Long} skinnerCooldownEndsAt
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.skinnerCooldownEndsAt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState skinnerCooldownDuration.
         * @member {number|Long} skinnerCooldownDuration
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.skinnerCooldownDuration = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState normalTriggerCount.
         * @member {number|Long} normalTriggerCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.normalTriggerCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState armorTriggerCount.
         * @member {number|Long} armorTriggerCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.armorTriggerCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState judgmentDayTriggerCount.
         * @member {number|Long} judgmentDayTriggerCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.judgmentDayTriggerCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState autoStrikeTriggerCount.
         * @member {number|Long} autoStrikeTriggerCount
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.autoStrikeTriggerCount = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TalentCombatState autoStrikeWindowSec.
         * @member {number|Long} autoStrikeWindowSec
         * @memberof realtime.TalentCombatState
         * @instance
         */
        TalentCombatState.prototype.autoStrikeWindowSec = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new TalentCombatState instance using the specified properties.
         * @function create
         * @memberof realtime.TalentCombatState
         * @static
         * @param {realtime.ITalentCombatState=} [properties] Properties to set
         * @returns {realtime.TalentCombatState} TalentCombatState instance
         */
        TalentCombatState.create = function create(properties) {
            return new TalentCombatState(properties);
        };

        /**
         * Encodes the specified TalentCombatState message. Does not implicitly {@link realtime.TalentCombatState.verify|verify} messages.
         * @function encode
         * @memberof realtime.TalentCombatState
         * @static
         * @param {realtime.ITalentCombatState} message TalentCombatState message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TalentCombatState.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.omenStacks != null && Object.hasOwnProperty.call(message, "omenStacks"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.omenStacks);
            if (message.bleeds != null && Object.hasOwnProperty.call(message, "bleeds"))
                for (let keys = Object.keys(message.bleeds), i = 0; i < keys.length; ++i) {
                    writer.uint32(/* id 2, wireType 2 =*/18).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                    $root.realtime.TalentBleedState.encode(message.bleeds[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                }
            if (message.collapseParts != null && message.collapseParts.length) {
                writer.uint32(/* id 3, wireType 2 =*/26).fork();
                for (let i = 0; i < message.collapseParts.length; ++i)
                    writer.int32(message.collapseParts[i]);
                writer.ldelim();
            }
            if (message.collapseEndsAt != null && Object.hasOwnProperty.call(message, "collapseEndsAt"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.collapseEndsAt);
            if (message.collapseDuration != null && Object.hasOwnProperty.call(message, "collapseDuration"))
                writer.uint32(/* id 5, wireType 0 =*/40).int64(message.collapseDuration);
            if (message.doomMarks != null && message.doomMarks.length) {
                writer.uint32(/* id 6, wireType 2 =*/50).fork();
                for (let i = 0; i < message.doomMarks.length; ++i)
                    writer.int32(message.doomMarks[i]);
                writer.ldelim();
            }
            if (message.hasTriggeredDoom != null && Object.hasOwnProperty.call(message, "hasTriggeredDoom"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.hasTriggeredDoom);
            if (message.doomMarkCumDamage != null && Object.hasOwnProperty.call(message, "doomMarkCumDamage"))
                for (let keys = Object.keys(message.doomMarkCumDamage), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 8, wireType 2 =*/66).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.doomMarkCumDamage[keys[i]]).ldelim();
            if (message.silverStormRemaining != null && Object.hasOwnProperty.call(message, "silverStormRemaining"))
                writer.uint32(/* id 9, wireType 0 =*/72).int32(message.silverStormRemaining);
            if (message.silverStormEndsAt != null && Object.hasOwnProperty.call(message, "silverStormEndsAt"))
                writer.uint32(/* id 10, wireType 0 =*/80).int64(message.silverStormEndsAt);
            if (message.silverStormActive != null && Object.hasOwnProperty.call(message, "silverStormActive"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.silverStormActive);
            if (message.autoStrikeTargetPart != null && Object.hasOwnProperty.call(message, "autoStrikeTargetPart"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.autoStrikeTargetPart);
            if (message.autoStrikeComboCount != null && Object.hasOwnProperty.call(message, "autoStrikeComboCount"))
                writer.uint32(/* id 13, wireType 0 =*/104).int64(message.autoStrikeComboCount);
            if (message.autoStrikeExpiresAt != null && Object.hasOwnProperty.call(message, "autoStrikeExpiresAt"))
                writer.uint32(/* id 14, wireType 0 =*/112).int64(message.autoStrikeExpiresAt);
            if (message.lastFinalCutAt != null && Object.hasOwnProperty.call(message, "lastFinalCutAt"))
                writer.uint32(/* id 15, wireType 0 =*/120).int64(message.lastFinalCutAt);
            if (message.judgmentDayUsed != null && Object.hasOwnProperty.call(message, "judgmentDayUsed"))
                for (let keys = Object.keys(message.judgmentDayUsed), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 16, wireType 2 =*/130).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.judgmentDayUsed[keys[i]]).ldelim();
            if (message.judgmentDayCooldownSec != null && Object.hasOwnProperty.call(message, "judgmentDayCooldownSec"))
                writer.uint32(/* id 17, wireType 0 =*/136).int64(message.judgmentDayCooldownSec);
            if (message.partHeavyClickCount != null && Object.hasOwnProperty.call(message, "partHeavyClickCount"))
                for (let keys = Object.keys(message.partHeavyClickCount), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 18, wireType 2 =*/146).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.partHeavyClickCount[keys[i]]).ldelim();
            if (message.partJudgmentDayCount != null && Object.hasOwnProperty.call(message, "partJudgmentDayCount"))
                for (let keys = Object.keys(message.partJudgmentDayCount), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 19, wireType 2 =*/154).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.partJudgmentDayCount[keys[i]]).ldelim();
            if (message.partRetainedClicks != null && Object.hasOwnProperty.call(message, "partRetainedClicks"))
                for (let keys = Object.keys(message.partRetainedClicks), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 20, wireType 2 =*/162).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.partRetainedClicks[keys[i]]).ldelim();
            if (message.partStormComboCount != null && Object.hasOwnProperty.call(message, "partStormComboCount"))
                for (let keys = Object.keys(message.partStormComboCount), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 21, wireType 2 =*/170).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.partStormComboCount[keys[i]]).ldelim();
            if (message.skinnerParts != null && Object.hasOwnProperty.call(message, "skinnerParts"))
                for (let keys = Object.keys(message.skinnerParts), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 22, wireType 2 =*/178).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.skinnerParts[keys[i]]).ldelim();
            if (message.skinnerDurationByPart != null && Object.hasOwnProperty.call(message, "skinnerDurationByPart"))
                for (let keys = Object.keys(message.skinnerDurationByPart), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 23, wireType 2 =*/186).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 0 =*/16).int64(message.skinnerDurationByPart[keys[i]]).ldelim();
            if (message.skinnerCooldownEndsAt != null && Object.hasOwnProperty.call(message, "skinnerCooldownEndsAt"))
                writer.uint32(/* id 24, wireType 0 =*/192).int64(message.skinnerCooldownEndsAt);
            if (message.skinnerCooldownDuration != null && Object.hasOwnProperty.call(message, "skinnerCooldownDuration"))
                writer.uint32(/* id 25, wireType 0 =*/200).int64(message.skinnerCooldownDuration);
            if (message.normalTriggerCount != null && Object.hasOwnProperty.call(message, "normalTriggerCount"))
                writer.uint32(/* id 26, wireType 0 =*/208).int64(message.normalTriggerCount);
            if (message.armorTriggerCount != null && Object.hasOwnProperty.call(message, "armorTriggerCount"))
                writer.uint32(/* id 27, wireType 0 =*/216).int64(message.armorTriggerCount);
            if (message.judgmentDayTriggerCount != null && Object.hasOwnProperty.call(message, "judgmentDayTriggerCount"))
                writer.uint32(/* id 28, wireType 0 =*/224).int64(message.judgmentDayTriggerCount);
            if (message.autoStrikeTriggerCount != null && Object.hasOwnProperty.call(message, "autoStrikeTriggerCount"))
                writer.uint32(/* id 29, wireType 0 =*/232).int64(message.autoStrikeTriggerCount);
            if (message.autoStrikeWindowSec != null && Object.hasOwnProperty.call(message, "autoStrikeWindowSec"))
                writer.uint32(/* id 30, wireType 0 =*/240).int64(message.autoStrikeWindowSec);
            return writer;
        };

        /**
         * Encodes the specified TalentCombatState message, length delimited. Does not implicitly {@link realtime.TalentCombatState.verify|verify} messages.
         * @function encodeDelimited
         * @memberof realtime.TalentCombatState
         * @static
         * @param {realtime.ITalentCombatState} message TalentCombatState message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TalentCombatState.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a TalentCombatState message from the specified reader or buffer.
         * @function decode
         * @memberof realtime.TalentCombatState
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {realtime.TalentCombatState} TalentCombatState
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TalentCombatState.decode = function decode(reader, length, error, long) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            if (long === undefined)
                long = 0;
            if (long > $Reader.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.realtime.TalentCombatState(), key, value;
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.omenStacks = reader.int32();
                        break;
                    }
                case 2: {
                        if (message.bleeds === $util.emptyObject)
                            message.bleeds = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = null;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = $root.realtime.TalentBleedState.decode(reader, reader.uint32(), undefined, long + 1);
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.bleeds, key);
                        message.bleeds[key] = value;
                        break;
                    }
                case 3: {
                        if (!(message.collapseParts && message.collapseParts.length))
                            message.collapseParts = [];
                        if ((tag & 7) === 2) {
                            let end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.collapseParts.push(reader.int32());
                        } else
                            message.collapseParts.push(reader.int32());
                        break;
                    }
                case 4: {
                        message.collapseEndsAt = reader.int64();
                        break;
                    }
                case 5: {
                        message.collapseDuration = reader.int64();
                        break;
                    }
                case 6: {
                        if (!(message.doomMarks && message.doomMarks.length))
                            message.doomMarks = [];
                        if ((tag & 7) === 2) {
                            let end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.doomMarks.push(reader.int32());
                        } else
                            message.doomMarks.push(reader.int32());
                        break;
                    }
                case 7: {
                        message.hasTriggeredDoom = reader.bool();
                        break;
                    }
                case 8: {
                        if (message.doomMarkCumDamage === $util.emptyObject)
                            message.doomMarkCumDamage = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.doomMarkCumDamage, key);
                        message.doomMarkCumDamage[key] = value;
                        break;
                    }
                case 9: {
                        message.silverStormRemaining = reader.int32();
                        break;
                    }
                case 10: {
                        message.silverStormEndsAt = reader.int64();
                        break;
                    }
                case 11: {
                        message.silverStormActive = reader.bool();
                        break;
                    }
                case 12: {
                        message.autoStrikeTargetPart = reader.string();
                        break;
                    }
                case 13: {
                        message.autoStrikeComboCount = reader.int64();
                        break;
                    }
                case 14: {
                        message.autoStrikeExpiresAt = reader.int64();
                        break;
                    }
                case 15: {
                        message.lastFinalCutAt = reader.int64();
                        break;
                    }
                case 16: {
                        if (message.judgmentDayUsed === $util.emptyObject)
                            message.judgmentDayUsed = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.judgmentDayUsed, key);
                        message.judgmentDayUsed[key] = value;
                        break;
                    }
                case 17: {
                        message.judgmentDayCooldownSec = reader.int64();
                        break;
                    }
                case 18: {
                        if (message.partHeavyClickCount === $util.emptyObject)
                            message.partHeavyClickCount = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.partHeavyClickCount, key);
                        message.partHeavyClickCount[key] = value;
                        break;
                    }
                case 19: {
                        if (message.partJudgmentDayCount === $util.emptyObject)
                            message.partJudgmentDayCount = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.partJudgmentDayCount, key);
                        message.partJudgmentDayCount[key] = value;
                        break;
                    }
                case 20: {
                        if (message.partRetainedClicks === $util.emptyObject)
                            message.partRetainedClicks = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.partRetainedClicks, key);
                        message.partRetainedClicks[key] = value;
                        break;
                    }
                case 21: {
                        if (message.partStormComboCount === $util.emptyObject)
                            message.partStormComboCount = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.partStormComboCount, key);
                        message.partStormComboCount[key] = value;
                        break;
                    }
                case 22: {
                        if (message.skinnerParts === $util.emptyObject)
                            message.skinnerParts = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.skinnerParts, key);
                        message.skinnerParts[key] = value;
                        break;
                    }
                case 23: {
                        if (message.skinnerDurationByPart === $util.emptyObject)
                            message.skinnerDurationByPart = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = 0;
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.int64();
                                break;
                            default:
                                reader.skipType(tag2 & 7, long);
                                break;
                            }
                        }
                        if (key === "__proto__")
                            $util.makeProp(message.skinnerDurationByPart, key);
                        message.skinnerDurationByPart[key] = value;
                        break;
                    }
                case 24: {
                        message.skinnerCooldownEndsAt = reader.int64();
                        break;
                    }
                case 25: {
                        message.skinnerCooldownDuration = reader.int64();
                        break;
                    }
                case 26: {
                        message.normalTriggerCount = reader.int64();
                        break;
                    }
                case 27: {
                        message.armorTriggerCount = reader.int64();
                        break;
                    }
                case 28: {
                        message.judgmentDayTriggerCount = reader.int64();
                        break;
                    }
                case 29: {
                        message.autoStrikeTriggerCount = reader.int64();
                        break;
                    }
                case 30: {
                        message.autoStrikeWindowSec = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7, long);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a TalentCombatState message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof realtime.TalentCombatState
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {realtime.TalentCombatState} TalentCombatState
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TalentCombatState.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a TalentCombatState message.
         * @function verify
         * @memberof realtime.TalentCombatState
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        TalentCombatState.verify = function verify(message, long) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                return "maximum nesting depth exceeded";
            if (message.omenStacks != null && message.hasOwnProperty("omenStacks"))
                if (!$util.isInteger(message.omenStacks))
                    return "omenStacks: integer expected";
            if (message.bleeds != null && message.hasOwnProperty("bleeds")) {
                if (!$util.isObject(message.bleeds))
                    return "bleeds: object expected";
                let key = Object.keys(message.bleeds);
                for (let i = 0; i < key.length; ++i) {
                    let error = $root.realtime.TalentBleedState.verify(message.bleeds[key[i]], long + 1);
                    if (error)
                        return "bleeds." + error;
                }
            }
            if (message.collapseParts != null && message.hasOwnProperty("collapseParts")) {
                if (!Array.isArray(message.collapseParts))
                    return "collapseParts: array expected";
                for (let i = 0; i < message.collapseParts.length; ++i)
                    if (!$util.isInteger(message.collapseParts[i]))
                        return "collapseParts: integer[] expected";
            }
            if (message.collapseEndsAt != null && message.hasOwnProperty("collapseEndsAt"))
                if (!$util.isInteger(message.collapseEndsAt) && !(message.collapseEndsAt && $util.isInteger(message.collapseEndsAt.low) && $util.isInteger(message.collapseEndsAt.high)))
                    return "collapseEndsAt: integer|Long expected";
            if (message.collapseDuration != null && message.hasOwnProperty("collapseDuration"))
                if (!$util.isInteger(message.collapseDuration) && !(message.collapseDuration && $util.isInteger(message.collapseDuration.low) && $util.isInteger(message.collapseDuration.high)))
                    return "collapseDuration: integer|Long expected";
            if (message.doomMarks != null && message.hasOwnProperty("doomMarks")) {
                if (!Array.isArray(message.doomMarks))
                    return "doomMarks: array expected";
                for (let i = 0; i < message.doomMarks.length; ++i)
                    if (!$util.isInteger(message.doomMarks[i]))
                        return "doomMarks: integer[] expected";
            }
            if (message.hasTriggeredDoom != null && message.hasOwnProperty("hasTriggeredDoom"))
                if (typeof message.hasTriggeredDoom !== "boolean")
                    return "hasTriggeredDoom: boolean expected";
            if (message.doomMarkCumDamage != null && message.hasOwnProperty("doomMarkCumDamage")) {
                if (!$util.isObject(message.doomMarkCumDamage))
                    return "doomMarkCumDamage: object expected";
                let key = Object.keys(message.doomMarkCumDamage);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.doomMarkCumDamage[key[i]]) && !(message.doomMarkCumDamage[key[i]] && $util.isInteger(message.doomMarkCumDamage[key[i]].low) && $util.isInteger(message.doomMarkCumDamage[key[i]].high)))
                        return "doomMarkCumDamage: integer|Long{k:string} expected";
            }
            if (message.silverStormRemaining != null && message.hasOwnProperty("silverStormRemaining"))
                if (!$util.isInteger(message.silverStormRemaining))
                    return "silverStormRemaining: integer expected";
            if (message.silverStormEndsAt != null && message.hasOwnProperty("silverStormEndsAt"))
                if (!$util.isInteger(message.silverStormEndsAt) && !(message.silverStormEndsAt && $util.isInteger(message.silverStormEndsAt.low) && $util.isInteger(message.silverStormEndsAt.high)))
                    return "silverStormEndsAt: integer|Long expected";
            if (message.silverStormActive != null && message.hasOwnProperty("silverStormActive"))
                if (typeof message.silverStormActive !== "boolean")
                    return "silverStormActive: boolean expected";
            if (message.autoStrikeTargetPart != null && message.hasOwnProperty("autoStrikeTargetPart"))
                if (!$util.isString(message.autoStrikeTargetPart))
                    return "autoStrikeTargetPart: string expected";
            if (message.autoStrikeComboCount != null && message.hasOwnProperty("autoStrikeComboCount"))
                if (!$util.isInteger(message.autoStrikeComboCount) && !(message.autoStrikeComboCount && $util.isInteger(message.autoStrikeComboCount.low) && $util.isInteger(message.autoStrikeComboCount.high)))
                    return "autoStrikeComboCount: integer|Long expected";
            if (message.autoStrikeExpiresAt != null && message.hasOwnProperty("autoStrikeExpiresAt"))
                if (!$util.isInteger(message.autoStrikeExpiresAt) && !(message.autoStrikeExpiresAt && $util.isInteger(message.autoStrikeExpiresAt.low) && $util.isInteger(message.autoStrikeExpiresAt.high)))
                    return "autoStrikeExpiresAt: integer|Long expected";
            if (message.lastFinalCutAt != null && message.hasOwnProperty("lastFinalCutAt"))
                if (!$util.isInteger(message.lastFinalCutAt) && !(message.lastFinalCutAt && $util.isInteger(message.lastFinalCutAt.low) && $util.isInteger(message.lastFinalCutAt.high)))
                    return "lastFinalCutAt: integer|Long expected";
            if (message.judgmentDayUsed != null && message.hasOwnProperty("judgmentDayUsed")) {
                if (!$util.isObject(message.judgmentDayUsed))
                    return "judgmentDayUsed: object expected";
                let key = Object.keys(message.judgmentDayUsed);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.judgmentDayUsed[key[i]]) && !(message.judgmentDayUsed[key[i]] && $util.isInteger(message.judgmentDayUsed[key[i]].low) && $util.isInteger(message.judgmentDayUsed[key[i]].high)))
                        return "judgmentDayUsed: integer|Long{k:string} expected";
            }
            if (message.judgmentDayCooldownSec != null && message.hasOwnProperty("judgmentDayCooldownSec"))
                if (!$util.isInteger(message.judgmentDayCooldownSec) && !(message.judgmentDayCooldownSec && $util.isInteger(message.judgmentDayCooldownSec.low) && $util.isInteger(message.judgmentDayCooldownSec.high)))
                    return "judgmentDayCooldownSec: integer|Long expected";
            if (message.partHeavyClickCount != null && message.hasOwnProperty("partHeavyClickCount")) {
                if (!$util.isObject(message.partHeavyClickCount))
                    return "partHeavyClickCount: object expected";
                let key = Object.keys(message.partHeavyClickCount);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.partHeavyClickCount[key[i]]) && !(message.partHeavyClickCount[key[i]] && $util.isInteger(message.partHeavyClickCount[key[i]].low) && $util.isInteger(message.partHeavyClickCount[key[i]].high)))
                        return "partHeavyClickCount: integer|Long{k:string} expected";
            }
            if (message.partJudgmentDayCount != null && message.hasOwnProperty("partJudgmentDayCount")) {
                if (!$util.isObject(message.partJudgmentDayCount))
                    return "partJudgmentDayCount: object expected";
                let key = Object.keys(message.partJudgmentDayCount);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.partJudgmentDayCount[key[i]]) && !(message.partJudgmentDayCount[key[i]] && $util.isInteger(message.partJudgmentDayCount[key[i]].low) && $util.isInteger(message.partJudgmentDayCount[key[i]].high)))
                        return "partJudgmentDayCount: integer|Long{k:string} expected";
            }
            if (message.partRetainedClicks != null && message.hasOwnProperty("partRetainedClicks")) {
                if (!$util.isObject(message.partRetainedClicks))
                    return "partRetainedClicks: object expected";
                let key = Object.keys(message.partRetainedClicks);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.partRetainedClicks[key[i]]) && !(message.partRetainedClicks[key[i]] && $util.isInteger(message.partRetainedClicks[key[i]].low) && $util.isInteger(message.partRetainedClicks[key[i]].high)))
                        return "partRetainedClicks: integer|Long{k:string} expected";
            }
            if (message.partStormComboCount != null && message.hasOwnProperty("partStormComboCount")) {
                if (!$util.isObject(message.partStormComboCount))
                    return "partStormComboCount: object expected";
                let key = Object.keys(message.partStormComboCount);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.partStormComboCount[key[i]]) && !(message.partStormComboCount[key[i]] && $util.isInteger(message.partStormComboCount[key[i]].low) && $util.isInteger(message.partStormComboCount[key[i]].high)))
                        return "partStormComboCount: integer|Long{k:string} expected";
            }
            if (message.skinnerParts != null && message.hasOwnProperty("skinnerParts")) {
                if (!$util.isObject(message.skinnerParts))
                    return "skinnerParts: object expected";
                let key = Object.keys(message.skinnerParts);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.skinnerParts[key[i]]) && !(message.skinnerParts[key[i]] && $util.isInteger(message.skinnerParts[key[i]].low) && $util.isInteger(message.skinnerParts[key[i]].high)))
                        return "skinnerParts: integer|Long{k:string} expected";
            }
            if (message.skinnerDurationByPart != null && message.hasOwnProperty("skinnerDurationByPart")) {
                if (!$util.isObject(message.skinnerDurationByPart))
                    return "skinnerDurationByPart: object expected";
                let key = Object.keys(message.skinnerDurationByPart);
                for (let i = 0; i < key.length; ++i)
                    if (!$util.isInteger(message.skinnerDurationByPart[key[i]]) && !(message.skinnerDurationByPart[key[i]] && $util.isInteger(message.skinnerDurationByPart[key[i]].low) && $util.isInteger(message.skinnerDurationByPart[key[i]].high)))
                        return "skinnerDurationByPart: integer|Long{k:string} expected";
            }
            if (message.skinnerCooldownEndsAt != null && message.hasOwnProperty("skinnerCooldownEndsAt"))
                if (!$util.isInteger(message.skinnerCooldownEndsAt) && !(message.skinnerCooldownEndsAt && $util.isInteger(message.skinnerCooldownEndsAt.low) && $util.isInteger(message.skinnerCooldownEndsAt.high)))
                    return "skinnerCooldownEndsAt: integer|Long expected";
            if (message.skinnerCooldownDuration != null && message.hasOwnProperty("skinnerCooldownDuration"))
                if (!$util.isInteger(message.skinnerCooldownDuration) && !(message.skinnerCooldownDuration && $util.isInteger(message.skinnerCooldownDuration.low) && $util.isInteger(message.skinnerCooldownDuration.high)))
                    return "skinnerCooldownDuration: integer|Long expected";
            if (message.normalTriggerCount != null && message.hasOwnProperty("normalTriggerCount"))
                if (!$util.isInteger(message.normalTriggerCount) && !(message.normalTriggerCount && $util.isInteger(message.normalTriggerCount.low) && $util.isInteger(message.normalTriggerCount.high)))
                    return "normalTriggerCount: integer|Long expected";
            if (message.armorTriggerCount != null && message.hasOwnProperty("armorTriggerCount"))
                if (!$util.isInteger(message.armorTriggerCount) && !(message.armorTriggerCount && $util.isInteger(message.armorTriggerCount.low) && $util.isInteger(message.armorTriggerCount.high)))
                    return "armorTriggerCount: integer|Long expected";
            if (message.judgmentDayTriggerCount != null && message.hasOwnProperty("judgmentDayTriggerCount"))
                if (!$util.isInteger(message.judgmentDayTriggerCount) && !(message.judgmentDayTriggerCount && $util.isInteger(message.judgmentDayTriggerCount.low) && $util.isInteger(message.judgmentDayTriggerCount.high)))
                    return "judgmentDayTriggerCount: integer|Long expected";
            if (message.autoStrikeTriggerCount != null && message.hasOwnProperty("autoStrikeTriggerCount"))
                if (!$util.isInteger(message.autoStrikeTriggerCount) && !(message.autoStrikeTriggerCount && $util.isInteger(message.autoStrikeTriggerCount.low) && $util.isInteger(message.autoStrikeTriggerCount.high)))
                    return "autoStrikeTriggerCount: integer|Long expected";
            if (message.autoStrikeWindowSec != null && message.hasOwnProperty("autoStrikeWindowSec"))
                if (!$util.isInteger(message.autoStrikeWindowSec) && !(message.autoStrikeWindowSec && $util.isInteger(message.autoStrikeWindowSec.low) && $util.isInteger(message.autoStrikeWindowSec.high)))
                    return "autoStrikeWindowSec: integer|Long expected";
            return null;
        };

        /**
         * Creates a TalentCombatState message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof realtime.TalentCombatState
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {realtime.TalentCombatState} TalentCombatState
         */
        TalentCombatState.fromObject = function fromObject(object, long) {
            if (object instanceof $root.realtime.TalentCombatState)
                return object;
            if (long === undefined)
                long = 0;
            if (long > $util.recursionLimit)
                throw Error("maximum nesting depth exceeded");
            let message = new $root.realtime.TalentCombatState();
            if (object.omenStacks != null)
                message.omenStacks = object.omenStacks | 0;
            if (object.bleeds) {
                if (typeof object.bleeds !== "object")
                    throw TypeError(".realtime.TalentCombatState.bleeds: object expected");
                message.bleeds = {};
                for (let keys = Object.keys(object.bleeds), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.bleeds, keys[i]);
                    if (typeof object.bleeds[keys[i]] !== "object")
                        throw TypeError(".realtime.TalentCombatState.bleeds: object expected");
                    message.bleeds[keys[i]] = $root.realtime.TalentBleedState.fromObject(object.bleeds[keys[i]], long + 1);
                }
            }
            if (object.collapseParts) {
                if (!Array.isArray(object.collapseParts))
                    throw TypeError(".realtime.TalentCombatState.collapseParts: array expected");
                message.collapseParts = [];
                for (let i = 0; i < object.collapseParts.length; ++i)
                    message.collapseParts[i] = object.collapseParts[i] | 0;
            }
            if (object.collapseEndsAt != null)
                if ($util.Long)
                    (message.collapseEndsAt = $util.Long.fromValue(object.collapseEndsAt)).unsigned = false;
                else if (typeof object.collapseEndsAt === "string")
                    message.collapseEndsAt = parseInt(object.collapseEndsAt, 10);
                else if (typeof object.collapseEndsAt === "number")
                    message.collapseEndsAt = object.collapseEndsAt;
                else if (typeof object.collapseEndsAt === "object")
                    message.collapseEndsAt = new $util.LongBits(object.collapseEndsAt.low >>> 0, object.collapseEndsAt.high >>> 0).toNumber();
            if (object.collapseDuration != null)
                if ($util.Long)
                    (message.collapseDuration = $util.Long.fromValue(object.collapseDuration)).unsigned = false;
                else if (typeof object.collapseDuration === "string")
                    message.collapseDuration = parseInt(object.collapseDuration, 10);
                else if (typeof object.collapseDuration === "number")
                    message.collapseDuration = object.collapseDuration;
                else if (typeof object.collapseDuration === "object")
                    message.collapseDuration = new $util.LongBits(object.collapseDuration.low >>> 0, object.collapseDuration.high >>> 0).toNumber();
            if (object.doomMarks) {
                if (!Array.isArray(object.doomMarks))
                    throw TypeError(".realtime.TalentCombatState.doomMarks: array expected");
                message.doomMarks = [];
                for (let i = 0; i < object.doomMarks.length; ++i)
                    message.doomMarks[i] = object.doomMarks[i] | 0;
            }
            if (object.hasTriggeredDoom != null)
                message.hasTriggeredDoom = Boolean(object.hasTriggeredDoom);
            if (object.doomMarkCumDamage) {
                if (typeof object.doomMarkCumDamage !== "object")
                    throw TypeError(".realtime.TalentCombatState.doomMarkCumDamage: object expected");
                message.doomMarkCumDamage = {};
                for (let keys = Object.keys(object.doomMarkCumDamage), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.doomMarkCumDamage, keys[i]);
                    if ($util.Long)
                        (message.doomMarkCumDamage[keys[i]] = $util.Long.fromValue(object.doomMarkCumDamage[keys[i]])).unsigned = false;
                    else if (typeof object.doomMarkCumDamage[keys[i]] === "string")
                        message.doomMarkCumDamage[keys[i]] = parseInt(object.doomMarkCumDamage[keys[i]], 10);
                    else if (typeof object.doomMarkCumDamage[keys[i]] === "number")
                        message.doomMarkCumDamage[keys[i]] = object.doomMarkCumDamage[keys[i]];
                    else if (typeof object.doomMarkCumDamage[keys[i]] === "object")
                        message.doomMarkCumDamage[keys[i]] = new $util.LongBits(object.doomMarkCumDamage[keys[i]].low >>> 0, object.doomMarkCumDamage[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.silverStormRemaining != null)
                message.silverStormRemaining = object.silverStormRemaining | 0;
            if (object.silverStormEndsAt != null)
                if ($util.Long)
                    (message.silverStormEndsAt = $util.Long.fromValue(object.silverStormEndsAt)).unsigned = false;
                else if (typeof object.silverStormEndsAt === "string")
                    message.silverStormEndsAt = parseInt(object.silverStormEndsAt, 10);
                else if (typeof object.silverStormEndsAt === "number")
                    message.silverStormEndsAt = object.silverStormEndsAt;
                else if (typeof object.silverStormEndsAt === "object")
                    message.silverStormEndsAt = new $util.LongBits(object.silverStormEndsAt.low >>> 0, object.silverStormEndsAt.high >>> 0).toNumber();
            if (object.silverStormActive != null)
                message.silverStormActive = Boolean(object.silverStormActive);
            if (object.autoStrikeTargetPart != null)
                message.autoStrikeTargetPart = String(object.autoStrikeTargetPart);
            if (object.autoStrikeComboCount != null)
                if ($util.Long)
                    (message.autoStrikeComboCount = $util.Long.fromValue(object.autoStrikeComboCount)).unsigned = false;
                else if (typeof object.autoStrikeComboCount === "string")
                    message.autoStrikeComboCount = parseInt(object.autoStrikeComboCount, 10);
                else if (typeof object.autoStrikeComboCount === "number")
                    message.autoStrikeComboCount = object.autoStrikeComboCount;
                else if (typeof object.autoStrikeComboCount === "object")
                    message.autoStrikeComboCount = new $util.LongBits(object.autoStrikeComboCount.low >>> 0, object.autoStrikeComboCount.high >>> 0).toNumber();
            if (object.autoStrikeExpiresAt != null)
                if ($util.Long)
                    (message.autoStrikeExpiresAt = $util.Long.fromValue(object.autoStrikeExpiresAt)).unsigned = false;
                else if (typeof object.autoStrikeExpiresAt === "string")
                    message.autoStrikeExpiresAt = parseInt(object.autoStrikeExpiresAt, 10);
                else if (typeof object.autoStrikeExpiresAt === "number")
                    message.autoStrikeExpiresAt = object.autoStrikeExpiresAt;
                else if (typeof object.autoStrikeExpiresAt === "object")
                    message.autoStrikeExpiresAt = new $util.LongBits(object.autoStrikeExpiresAt.low >>> 0, object.autoStrikeExpiresAt.high >>> 0).toNumber();
            if (object.lastFinalCutAt != null)
                if ($util.Long)
                    (message.lastFinalCutAt = $util.Long.fromValue(object.lastFinalCutAt)).unsigned = false;
                else if (typeof object.lastFinalCutAt === "string")
                    message.lastFinalCutAt = parseInt(object.lastFinalCutAt, 10);
                else if (typeof object.lastFinalCutAt === "number")
                    message.lastFinalCutAt = object.lastFinalCutAt;
                else if (typeof object.lastFinalCutAt === "object")
                    message.lastFinalCutAt = new $util.LongBits(object.lastFinalCutAt.low >>> 0, object.lastFinalCutAt.high >>> 0).toNumber();
            if (object.judgmentDayUsed) {
                if (typeof object.judgmentDayUsed !== "object")
                    throw TypeError(".realtime.TalentCombatState.judgmentDayUsed: object expected");
                message.judgmentDayUsed = {};
                for (let keys = Object.keys(object.judgmentDayUsed), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.judgmentDayUsed, keys[i]);
                    if ($util.Long)
                        (message.judgmentDayUsed[keys[i]] = $util.Long.fromValue(object.judgmentDayUsed[keys[i]])).unsigned = false;
                    else if (typeof object.judgmentDayUsed[keys[i]] === "string")
                        message.judgmentDayUsed[keys[i]] = parseInt(object.judgmentDayUsed[keys[i]], 10);
                    else if (typeof object.judgmentDayUsed[keys[i]] === "number")
                        message.judgmentDayUsed[keys[i]] = object.judgmentDayUsed[keys[i]];
                    else if (typeof object.judgmentDayUsed[keys[i]] === "object")
                        message.judgmentDayUsed[keys[i]] = new $util.LongBits(object.judgmentDayUsed[keys[i]].low >>> 0, object.judgmentDayUsed[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.judgmentDayCooldownSec != null)
                if ($util.Long)
                    (message.judgmentDayCooldownSec = $util.Long.fromValue(object.judgmentDayCooldownSec)).unsigned = false;
                else if (typeof object.judgmentDayCooldownSec === "string")
                    message.judgmentDayCooldownSec = parseInt(object.judgmentDayCooldownSec, 10);
                else if (typeof object.judgmentDayCooldownSec === "number")
                    message.judgmentDayCooldownSec = object.judgmentDayCooldownSec;
                else if (typeof object.judgmentDayCooldownSec === "object")
                    message.judgmentDayCooldownSec = new $util.LongBits(object.judgmentDayCooldownSec.low >>> 0, object.judgmentDayCooldownSec.high >>> 0).toNumber();
            if (object.partHeavyClickCount) {
                if (typeof object.partHeavyClickCount !== "object")
                    throw TypeError(".realtime.TalentCombatState.partHeavyClickCount: object expected");
                message.partHeavyClickCount = {};
                for (let keys = Object.keys(object.partHeavyClickCount), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.partHeavyClickCount, keys[i]);
                    if ($util.Long)
                        (message.partHeavyClickCount[keys[i]] = $util.Long.fromValue(object.partHeavyClickCount[keys[i]])).unsigned = false;
                    else if (typeof object.partHeavyClickCount[keys[i]] === "string")
                        message.partHeavyClickCount[keys[i]] = parseInt(object.partHeavyClickCount[keys[i]], 10);
                    else if (typeof object.partHeavyClickCount[keys[i]] === "number")
                        message.partHeavyClickCount[keys[i]] = object.partHeavyClickCount[keys[i]];
                    else if (typeof object.partHeavyClickCount[keys[i]] === "object")
                        message.partHeavyClickCount[keys[i]] = new $util.LongBits(object.partHeavyClickCount[keys[i]].low >>> 0, object.partHeavyClickCount[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.partJudgmentDayCount) {
                if (typeof object.partJudgmentDayCount !== "object")
                    throw TypeError(".realtime.TalentCombatState.partJudgmentDayCount: object expected");
                message.partJudgmentDayCount = {};
                for (let keys = Object.keys(object.partJudgmentDayCount), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.partJudgmentDayCount, keys[i]);
                    if ($util.Long)
                        (message.partJudgmentDayCount[keys[i]] = $util.Long.fromValue(object.partJudgmentDayCount[keys[i]])).unsigned = false;
                    else if (typeof object.partJudgmentDayCount[keys[i]] === "string")
                        message.partJudgmentDayCount[keys[i]] = parseInt(object.partJudgmentDayCount[keys[i]], 10);
                    else if (typeof object.partJudgmentDayCount[keys[i]] === "number")
                        message.partJudgmentDayCount[keys[i]] = object.partJudgmentDayCount[keys[i]];
                    else if (typeof object.partJudgmentDayCount[keys[i]] === "object")
                        message.partJudgmentDayCount[keys[i]] = new $util.LongBits(object.partJudgmentDayCount[keys[i]].low >>> 0, object.partJudgmentDayCount[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.partRetainedClicks) {
                if (typeof object.partRetainedClicks !== "object")
                    throw TypeError(".realtime.TalentCombatState.partRetainedClicks: object expected");
                message.partRetainedClicks = {};
                for (let keys = Object.keys(object.partRetainedClicks), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.partRetainedClicks, keys[i]);
                    if ($util.Long)
                        (message.partRetainedClicks[keys[i]] = $util.Long.fromValue(object.partRetainedClicks[keys[i]])).unsigned = false;
                    else if (typeof object.partRetainedClicks[keys[i]] === "string")
                        message.partRetainedClicks[keys[i]] = parseInt(object.partRetainedClicks[keys[i]], 10);
                    else if (typeof object.partRetainedClicks[keys[i]] === "number")
                        message.partRetainedClicks[keys[i]] = object.partRetainedClicks[keys[i]];
                    else if (typeof object.partRetainedClicks[keys[i]] === "object")
                        message.partRetainedClicks[keys[i]] = new $util.LongBits(object.partRetainedClicks[keys[i]].low >>> 0, object.partRetainedClicks[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.partStormComboCount) {
                if (typeof object.partStormComboCount !== "object")
                    throw TypeError(".realtime.TalentCombatState.partStormComboCount: object expected");
                message.partStormComboCount = {};
                for (let keys = Object.keys(object.partStormComboCount), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.partStormComboCount, keys[i]);
                    if ($util.Long)
                        (message.partStormComboCount[keys[i]] = $util.Long.fromValue(object.partStormComboCount[keys[i]])).unsigned = false;
                    else if (typeof object.partStormComboCount[keys[i]] === "string")
                        message.partStormComboCount[keys[i]] = parseInt(object.partStormComboCount[keys[i]], 10);
                    else if (typeof object.partStormComboCount[keys[i]] === "number")
                        message.partStormComboCount[keys[i]] = object.partStormComboCount[keys[i]];
                    else if (typeof object.partStormComboCount[keys[i]] === "object")
                        message.partStormComboCount[keys[i]] = new $util.LongBits(object.partStormComboCount[keys[i]].low >>> 0, object.partStormComboCount[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.skinnerParts) {
                if (typeof object.skinnerParts !== "object")
                    throw TypeError(".realtime.TalentCombatState.skinnerParts: object expected");
                message.skinnerParts = {};
                for (let keys = Object.keys(object.skinnerParts), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.skinnerParts, keys[i]);
                    if ($util.Long)
                        (message.skinnerParts[keys[i]] = $util.Long.fromValue(object.skinnerParts[keys[i]])).unsigned = false;
                    else if (typeof object.skinnerParts[keys[i]] === "string")
                        message.skinnerParts[keys[i]] = parseInt(object.skinnerParts[keys[i]], 10);
                    else if (typeof object.skinnerParts[keys[i]] === "number")
                        message.skinnerParts[keys[i]] = object.skinnerParts[keys[i]];
                    else if (typeof object.skinnerParts[keys[i]] === "object")
                        message.skinnerParts[keys[i]] = new $util.LongBits(object.skinnerParts[keys[i]].low >>> 0, object.skinnerParts[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.skinnerDurationByPart) {
                if (typeof object.skinnerDurationByPart !== "object")
                    throw TypeError(".realtime.TalentCombatState.skinnerDurationByPart: object expected");
                message.skinnerDurationByPart = {};
                for (let keys = Object.keys(object.skinnerDurationByPart), i = 0; i < keys.length; ++i) {
                    if (keys[i] === "__proto__")
                        $util.makeProp(message.skinnerDurationByPart, keys[i]);
                    if ($util.Long)
                        (message.skinnerDurationByPart[keys[i]] = $util.Long.fromValue(object.skinnerDurationByPart[keys[i]])).unsigned = false;
                    else if (typeof object.skinnerDurationByPart[keys[i]] === "string")
                        message.skinnerDurationByPart[keys[i]] = parseInt(object.skinnerDurationByPart[keys[i]], 10);
                    else if (typeof object.skinnerDurationByPart[keys[i]] === "number")
                        message.skinnerDurationByPart[keys[i]] = object.skinnerDurationByPart[keys[i]];
                    else if (typeof object.skinnerDurationByPart[keys[i]] === "object")
                        message.skinnerDurationByPart[keys[i]] = new $util.LongBits(object.skinnerDurationByPart[keys[i]].low >>> 0, object.skinnerDurationByPart[keys[i]].high >>> 0).toNumber();
                }
            }
            if (object.skinnerCooldownEndsAt != null)
                if ($util.Long)
                    (message.skinnerCooldownEndsAt = $util.Long.fromValue(object.skinnerCooldownEndsAt)).unsigned = false;
                else if (typeof object.skinnerCooldownEndsAt === "string")
                    message.skinnerCooldownEndsAt = parseInt(object.skinnerCooldownEndsAt, 10);
                else if (typeof object.skinnerCooldownEndsAt === "number")
                    message.skinnerCooldownEndsAt = object.skinnerCooldownEndsAt;
                else if (typeof object.skinnerCooldownEndsAt === "object")
                    message.skinnerCooldownEndsAt = new $util.LongBits(object.skinnerCooldownEndsAt.low >>> 0, object.skinnerCooldownEndsAt.high >>> 0).toNumber();
            if (object.skinnerCooldownDuration != null)
                if ($util.Long)
                    (message.skinnerCooldownDuration = $util.Long.fromValue(object.skinnerCooldownDuration)).unsigned = false;
                else if (typeof object.skinnerCooldownDuration === "string")
                    message.skinnerCooldownDuration = parseInt(object.skinnerCooldownDuration, 10);
                else if (typeof object.skinnerCooldownDuration === "number")
                    message.skinnerCooldownDuration = object.skinnerCooldownDuration;
                else if (typeof object.skinnerCooldownDuration === "object")
                    message.skinnerCooldownDuration = new $util.LongBits(object.skinnerCooldownDuration.low >>> 0, object.skinnerCooldownDuration.high >>> 0).toNumber();
            if (object.normalTriggerCount != null)
                if ($util.Long)
                    (message.normalTriggerCount = $util.Long.fromValue(object.normalTriggerCount)).unsigned = false;
                else if (typeof object.normalTriggerCount === "string")
                    message.normalTriggerCount = parseInt(object.normalTriggerCount, 10);
                else if (typeof object.normalTriggerCount === "number")
                    message.normalTriggerCount = object.normalTriggerCount;
                else if (typeof object.normalTriggerCount === "object")
                    message.normalTriggerCount = new $util.LongBits(object.normalTriggerCount.low >>> 0, object.normalTriggerCount.high >>> 0).toNumber();
            if (object.armorTriggerCount != null)
                if ($util.Long)
                    (message.armorTriggerCount = $util.Long.fromValue(object.armorTriggerCount)).unsigned = false;
                else if (typeof object.armorTriggerCount === "string")
                    message.armorTriggerCount = parseInt(object.armorTriggerCount, 10);
                else if (typeof object.armorTriggerCount === "number")
                    message.armorTriggerCount = object.armorTriggerCount;
                else if (typeof object.armorTriggerCount === "object")
                    message.armorTriggerCount = new $util.LongBits(object.armorTriggerCount.low >>> 0, object.armorTriggerCount.high >>> 0).toNumber();
            if (object.judgmentDayTriggerCount != null)
                if ($util.Long)
                    (message.judgmentDayTriggerCount = $util.Long.fromValue(object.judgmentDayTriggerCount)).unsigned = false;
                else if (typeof object.judgmentDayTriggerCount === "string")
                    message.judgmentDayTriggerCount = parseInt(object.judgmentDayTriggerCount, 10);
                else if (typeof object.judgmentDayTriggerCount === "number")
                    message.judgmentDayTriggerCount = object.judgmentDayTriggerCount;
                else if (typeof object.judgmentDayTriggerCount === "object")
                    message.judgmentDayTriggerCount = new $util.LongBits(object.judgmentDayTriggerCount.low >>> 0, object.judgmentDayTriggerCount.high >>> 0).toNumber();
            if (object.autoStrikeTriggerCount != null)
                if ($util.Long)
                    (message.autoStrikeTriggerCount = $util.Long.fromValue(object.autoStrikeTriggerCount)).unsigned = false;
                else if (typeof object.autoStrikeTriggerCount === "string")
                    message.autoStrikeTriggerCount = parseInt(object.autoStrikeTriggerCount, 10);
                else if (typeof object.autoStrikeTriggerCount === "number")
                    message.autoStrikeTriggerCount = object.autoStrikeTriggerCount;
                else if (typeof object.autoStrikeTriggerCount === "object")
                    message.autoStrikeTriggerCount = new $util.LongBits(object.autoStrikeTriggerCount.low >>> 0, object.autoStrikeTriggerCount.high >>> 0).toNumber();
            if (object.autoStrikeWindowSec != null)
                if ($util.Long)
                    (message.autoStrikeWindowSec = $util.Long.fromValue(object.autoStrikeWindowSec)).unsigned = false;
                else if (typeof object.autoStrikeWindowSec === "string")
                    message.autoStrikeWindowSec = parseInt(object.autoStrikeWindowSec, 10);
                else if (typeof object.autoStrikeWindowSec === "number")
                    message.autoStrikeWindowSec = object.autoStrikeWindowSec;
                else if (typeof object.autoStrikeWindowSec === "object")
                    message.autoStrikeWindowSec = new $util.LongBits(object.autoStrikeWindowSec.low >>> 0, object.autoStrikeWindowSec.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a TalentCombatState message. Also converts values to other types if specified.
         * @function toObject
         * @memberof realtime.TalentCombatState
         * @static
         * @param {realtime.TalentCombatState} message TalentCombatState
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        TalentCombatState.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.collapseParts = [];
                object.doomMarks = [];
            }
            if (options.objects || options.defaults) {
                object.bleeds = {};
                object.doomMarkCumDamage = {};
                object.judgmentDayUsed = {};
                object.partHeavyClickCount = {};
                object.partJudgmentDayCount = {};
                object.partRetainedClicks = {};
                object.partStormComboCount = {};
                object.skinnerParts = {};
                object.skinnerDurationByPart = {};
            }
            if (options.defaults) {
                object.omenStacks = 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.collapseEndsAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.collapseEndsAt = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.collapseDuration = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.collapseDuration = options.longs === String ? "0" : 0;
                object.hasTriggeredDoom = false;
                object.silverStormRemaining = 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.silverStormEndsAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.silverStormEndsAt = options.longs === String ? "0" : 0;
                object.silverStormActive = false;
                object.autoStrikeTargetPart = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.autoStrikeComboCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.autoStrikeComboCount = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.autoStrikeExpiresAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.autoStrikeExpiresAt = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.lastFinalCutAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.lastFinalCutAt = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.judgmentDayCooldownSec = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.judgmentDayCooldownSec = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.skinnerCooldownEndsAt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.skinnerCooldownEndsAt = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.skinnerCooldownDuration = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.skinnerCooldownDuration = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.normalTriggerCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.normalTriggerCount = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.armorTriggerCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.armorTriggerCount = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.judgmentDayTriggerCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.judgmentDayTriggerCount = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.autoStrikeTriggerCount = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.autoStrikeTriggerCount = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.autoStrikeWindowSec = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.autoStrikeWindowSec = options.longs === String ? "0" : 0;
            }
            if (message.omenStacks != null && message.hasOwnProperty("omenStacks"))
                object.omenStacks = message.omenStacks;
            let keys2;
            if (message.bleeds && (keys2 = Object.keys(message.bleeds)).length) {
                object.bleeds = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.bleeds, keys2[j]);
                    object.bleeds[keys2[j]] = $root.realtime.TalentBleedState.toObject(message.bleeds[keys2[j]], options);
                }
            }
            if (message.collapseParts && message.collapseParts.length) {
                object.collapseParts = [];
                for (let j = 0; j < message.collapseParts.length; ++j)
                    object.collapseParts[j] = message.collapseParts[j];
            }
            if (message.collapseEndsAt != null && message.hasOwnProperty("collapseEndsAt"))
                if (typeof message.collapseEndsAt === "number")
                    object.collapseEndsAt = options.longs === String ? String(message.collapseEndsAt) : message.collapseEndsAt;
                else
                    object.collapseEndsAt = options.longs === String ? $util.Long.prototype.toString.call(message.collapseEndsAt) : options.longs === Number ? new $util.LongBits(message.collapseEndsAt.low >>> 0, message.collapseEndsAt.high >>> 0).toNumber() : message.collapseEndsAt;
            if (message.collapseDuration != null && message.hasOwnProperty("collapseDuration"))
                if (typeof message.collapseDuration === "number")
                    object.collapseDuration = options.longs === String ? String(message.collapseDuration) : message.collapseDuration;
                else
                    object.collapseDuration = options.longs === String ? $util.Long.prototype.toString.call(message.collapseDuration) : options.longs === Number ? new $util.LongBits(message.collapseDuration.low >>> 0, message.collapseDuration.high >>> 0).toNumber() : message.collapseDuration;
            if (message.doomMarks && message.doomMarks.length) {
                object.doomMarks = [];
                for (let j = 0; j < message.doomMarks.length; ++j)
                    object.doomMarks[j] = message.doomMarks[j];
            }
            if (message.hasTriggeredDoom != null && message.hasOwnProperty("hasTriggeredDoom"))
                object.hasTriggeredDoom = message.hasTriggeredDoom;
            if (message.doomMarkCumDamage && (keys2 = Object.keys(message.doomMarkCumDamage)).length) {
                object.doomMarkCumDamage = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.doomMarkCumDamage, keys2[j]);
                    if (typeof message.doomMarkCumDamage[keys2[j]] === "number")
                        object.doomMarkCumDamage[keys2[j]] = options.longs === String ? String(message.doomMarkCumDamage[keys2[j]]) : message.doomMarkCumDamage[keys2[j]];
                    else
                        object.doomMarkCumDamage[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.doomMarkCumDamage[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.doomMarkCumDamage[keys2[j]].low >>> 0, message.doomMarkCumDamage[keys2[j]].high >>> 0).toNumber() : message.doomMarkCumDamage[keys2[j]];
                }
            }
            if (message.silverStormRemaining != null && message.hasOwnProperty("silverStormRemaining"))
                object.silverStormRemaining = message.silverStormRemaining;
            if (message.silverStormEndsAt != null && message.hasOwnProperty("silverStormEndsAt"))
                if (typeof message.silverStormEndsAt === "number")
                    object.silverStormEndsAt = options.longs === String ? String(message.silverStormEndsAt) : message.silverStormEndsAt;
                else
                    object.silverStormEndsAt = options.longs === String ? $util.Long.prototype.toString.call(message.silverStormEndsAt) : options.longs === Number ? new $util.LongBits(message.silverStormEndsAt.low >>> 0, message.silverStormEndsAt.high >>> 0).toNumber() : message.silverStormEndsAt;
            if (message.silverStormActive != null && message.hasOwnProperty("silverStormActive"))
                object.silverStormActive = message.silverStormActive;
            if (message.autoStrikeTargetPart != null && message.hasOwnProperty("autoStrikeTargetPart"))
                object.autoStrikeTargetPart = message.autoStrikeTargetPart;
            if (message.autoStrikeComboCount != null && message.hasOwnProperty("autoStrikeComboCount"))
                if (typeof message.autoStrikeComboCount === "number")
                    object.autoStrikeComboCount = options.longs === String ? String(message.autoStrikeComboCount) : message.autoStrikeComboCount;
                else
                    object.autoStrikeComboCount = options.longs === String ? $util.Long.prototype.toString.call(message.autoStrikeComboCount) : options.longs === Number ? new $util.LongBits(message.autoStrikeComboCount.low >>> 0, message.autoStrikeComboCount.high >>> 0).toNumber() : message.autoStrikeComboCount;
            if (message.autoStrikeExpiresAt != null && message.hasOwnProperty("autoStrikeExpiresAt"))
                if (typeof message.autoStrikeExpiresAt === "number")
                    object.autoStrikeExpiresAt = options.longs === String ? String(message.autoStrikeExpiresAt) : message.autoStrikeExpiresAt;
                else
                    object.autoStrikeExpiresAt = options.longs === String ? $util.Long.prototype.toString.call(message.autoStrikeExpiresAt) : options.longs === Number ? new $util.LongBits(message.autoStrikeExpiresAt.low >>> 0, message.autoStrikeExpiresAt.high >>> 0).toNumber() : message.autoStrikeExpiresAt;
            if (message.lastFinalCutAt != null && message.hasOwnProperty("lastFinalCutAt"))
                if (typeof message.lastFinalCutAt === "number")
                    object.lastFinalCutAt = options.longs === String ? String(message.lastFinalCutAt) : message.lastFinalCutAt;
                else
                    object.lastFinalCutAt = options.longs === String ? $util.Long.prototype.toString.call(message.lastFinalCutAt) : options.longs === Number ? new $util.LongBits(message.lastFinalCutAt.low >>> 0, message.lastFinalCutAt.high >>> 0).toNumber() : message.lastFinalCutAt;
            if (message.judgmentDayUsed && (keys2 = Object.keys(message.judgmentDayUsed)).length) {
                object.judgmentDayUsed = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.judgmentDayUsed, keys2[j]);
                    if (typeof message.judgmentDayUsed[keys2[j]] === "number")
                        object.judgmentDayUsed[keys2[j]] = options.longs === String ? String(message.judgmentDayUsed[keys2[j]]) : message.judgmentDayUsed[keys2[j]];
                    else
                        object.judgmentDayUsed[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.judgmentDayUsed[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.judgmentDayUsed[keys2[j]].low >>> 0, message.judgmentDayUsed[keys2[j]].high >>> 0).toNumber() : message.judgmentDayUsed[keys2[j]];
                }
            }
            if (message.judgmentDayCooldownSec != null && message.hasOwnProperty("judgmentDayCooldownSec"))
                if (typeof message.judgmentDayCooldownSec === "number")
                    object.judgmentDayCooldownSec = options.longs === String ? String(message.judgmentDayCooldownSec) : message.judgmentDayCooldownSec;
                else
                    object.judgmentDayCooldownSec = options.longs === String ? $util.Long.prototype.toString.call(message.judgmentDayCooldownSec) : options.longs === Number ? new $util.LongBits(message.judgmentDayCooldownSec.low >>> 0, message.judgmentDayCooldownSec.high >>> 0).toNumber() : message.judgmentDayCooldownSec;
            if (message.partHeavyClickCount && (keys2 = Object.keys(message.partHeavyClickCount)).length) {
                object.partHeavyClickCount = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.partHeavyClickCount, keys2[j]);
                    if (typeof message.partHeavyClickCount[keys2[j]] === "number")
                        object.partHeavyClickCount[keys2[j]] = options.longs === String ? String(message.partHeavyClickCount[keys2[j]]) : message.partHeavyClickCount[keys2[j]];
                    else
                        object.partHeavyClickCount[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.partHeavyClickCount[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.partHeavyClickCount[keys2[j]].low >>> 0, message.partHeavyClickCount[keys2[j]].high >>> 0).toNumber() : message.partHeavyClickCount[keys2[j]];
                }
            }
            if (message.partJudgmentDayCount && (keys2 = Object.keys(message.partJudgmentDayCount)).length) {
                object.partJudgmentDayCount = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.partJudgmentDayCount, keys2[j]);
                    if (typeof message.partJudgmentDayCount[keys2[j]] === "number")
                        object.partJudgmentDayCount[keys2[j]] = options.longs === String ? String(message.partJudgmentDayCount[keys2[j]]) : message.partJudgmentDayCount[keys2[j]];
                    else
                        object.partJudgmentDayCount[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.partJudgmentDayCount[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.partJudgmentDayCount[keys2[j]].low >>> 0, message.partJudgmentDayCount[keys2[j]].high >>> 0).toNumber() : message.partJudgmentDayCount[keys2[j]];
                }
            }
            if (message.partRetainedClicks && (keys2 = Object.keys(message.partRetainedClicks)).length) {
                object.partRetainedClicks = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.partRetainedClicks, keys2[j]);
                    if (typeof message.partRetainedClicks[keys2[j]] === "number")
                        object.partRetainedClicks[keys2[j]] = options.longs === String ? String(message.partRetainedClicks[keys2[j]]) : message.partRetainedClicks[keys2[j]];
                    else
                        object.partRetainedClicks[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.partRetainedClicks[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.partRetainedClicks[keys2[j]].low >>> 0, message.partRetainedClicks[keys2[j]].high >>> 0).toNumber() : message.partRetainedClicks[keys2[j]];
                }
            }
            if (message.partStormComboCount && (keys2 = Object.keys(message.partStormComboCount)).length) {
                object.partStormComboCount = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.partStormComboCount, keys2[j]);
                    if (typeof message.partStormComboCount[keys2[j]] === "number")
                        object.partStormComboCount[keys2[j]] = options.longs === String ? String(message.partStormComboCount[keys2[j]]) : message.partStormComboCount[keys2[j]];
                    else
                        object.partStormComboCount[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.partStormComboCount[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.partStormComboCount[keys2[j]].low >>> 0, message.partStormComboCount[keys2[j]].high >>> 0).toNumber() : message.partStormComboCount[keys2[j]];
                }
            }
            if (message.skinnerParts && (keys2 = Object.keys(message.skinnerParts)).length) {
                object.skinnerParts = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.skinnerParts, keys2[j]);
                    if (typeof message.skinnerParts[keys2[j]] === "number")
                        object.skinnerParts[keys2[j]] = options.longs === String ? String(message.skinnerParts[keys2[j]]) : message.skinnerParts[keys2[j]];
                    else
                        object.skinnerParts[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.skinnerParts[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.skinnerParts[keys2[j]].low >>> 0, message.skinnerParts[keys2[j]].high >>> 0).toNumber() : message.skinnerParts[keys2[j]];
                }
            }
            if (message.skinnerDurationByPart && (keys2 = Object.keys(message.skinnerDurationByPart)).length) {
                object.skinnerDurationByPart = {};
                for (let j = 0; j < keys2.length; ++j) {
                    if (keys2[j] === "__proto__")
                        $util.makeProp(object.skinnerDurationByPart, keys2[j]);
                    if (typeof message.skinnerDurationByPart[keys2[j]] === "number")
                        object.skinnerDurationByPart[keys2[j]] = options.longs === String ? String(message.skinnerDurationByPart[keys2[j]]) : message.skinnerDurationByPart[keys2[j]];
                    else
                        object.skinnerDurationByPart[keys2[j]] = options.longs === String ? $util.Long.prototype.toString.call(message.skinnerDurationByPart[keys2[j]]) : options.longs === Number ? new $util.LongBits(message.skinnerDurationByPart[keys2[j]].low >>> 0, message.skinnerDurationByPart[keys2[j]].high >>> 0).toNumber() : message.skinnerDurationByPart[keys2[j]];
                }
            }
            if (message.skinnerCooldownEndsAt != null && message.hasOwnProperty("skinnerCooldownEndsAt"))
                if (typeof message.skinnerCooldownEndsAt === "number")
                    object.skinnerCooldownEndsAt = options.longs === String ? String(message.skinnerCooldownEndsAt) : message.skinnerCooldownEndsAt;
                else
                    object.skinnerCooldownEndsAt = options.longs === String ? $util.Long.prototype.toString.call(message.skinnerCooldownEndsAt) : options.longs === Number ? new $util.LongBits(message.skinnerCooldownEndsAt.low >>> 0, message.skinnerCooldownEndsAt.high >>> 0).toNumber() : message.skinnerCooldownEndsAt;
            if (message.skinnerCooldownDuration != null && message.hasOwnProperty("skinnerCooldownDuration"))
                if (typeof message.skinnerCooldownDuration === "number")
                    object.skinnerCooldownDuration = options.longs === String ? String(message.skinnerCooldownDuration) : message.skinnerCooldownDuration;
                else
                    object.skinnerCooldownDuration = options.longs === String ? $util.Long.prototype.toString.call(message.skinnerCooldownDuration) : options.longs === Number ? new $util.LongBits(message.skinnerCooldownDuration.low >>> 0, message.skinnerCooldownDuration.high >>> 0).toNumber() : message.skinnerCooldownDuration;
            if (message.normalTriggerCount != null && message.hasOwnProperty("normalTriggerCount"))
                if (typeof message.normalTriggerCount === "number")
                    object.normalTriggerCount = options.longs === String ? String(message.normalTriggerCount) : message.normalTriggerCount;
                else
                    object.normalTriggerCount = options.longs === String ? $util.Long.prototype.toString.call(message.normalTriggerCount) : options.longs === Number ? new $util.LongBits(message.normalTriggerCount.low >>> 0, message.normalTriggerCount.high >>> 0).toNumber() : message.normalTriggerCount;
            if (message.armorTriggerCount != null && message.hasOwnProperty("armorTriggerCount"))
                if (typeof message.armorTriggerCount === "number")
                    object.armorTriggerCount = options.longs === String ? String(message.armorTriggerCount) : message.armorTriggerCount;
                else
                    object.armorTriggerCount = options.longs === String ? $util.Long.prototype.toString.call(message.armorTriggerCount) : options.longs === Number ? new $util.LongBits(message.armorTriggerCount.low >>> 0, message.armorTriggerCount.high >>> 0).toNumber() : message.armorTriggerCount;
            if (message.judgmentDayTriggerCount != null && message.hasOwnProperty("judgmentDayTriggerCount"))
                if (typeof message.judgmentDayTriggerCount === "number")
                    object.judgmentDayTriggerCount = options.longs === String ? String(message.judgmentDayTriggerCount) : message.judgmentDayTriggerCount;
                else
                    object.judgmentDayTriggerCount = options.longs === String ? $util.Long.prototype.toString.call(message.judgmentDayTriggerCount) : options.longs === Number ? new $util.LongBits(message.judgmentDayTriggerCount.low >>> 0, message.judgmentDayTriggerCount.high >>> 0).toNumber() : message.judgmentDayTriggerCount;
            if (message.autoStrikeTriggerCount != null && message.hasOwnProperty("autoStrikeTriggerCount"))
                if (typeof message.autoStrikeTriggerCount === "number")
                    object.autoStrikeTriggerCount = options.longs === String ? String(message.autoStrikeTriggerCount) : message.autoStrikeTriggerCount;
                else
                    object.autoStrikeTriggerCount = options.longs === String ? $util.Long.prototype.toString.call(message.autoStrikeTriggerCount) : options.longs === Number ? new $util.LongBits(message.autoStrikeTriggerCount.low >>> 0, message.autoStrikeTriggerCount.high >>> 0).toNumber() : message.autoStrikeTriggerCount;
            if (message.autoStrikeWindowSec != null && message.hasOwnProperty("autoStrikeWindowSec"))
                if (typeof message.autoStrikeWindowSec === "number")
                    object.autoStrikeWindowSec = options.longs === String ? String(message.autoStrikeWindowSec) : message.autoStrikeWindowSec;
                else
                    object.autoStrikeWindowSec = options.longs === String ? $util.Long.prototype.toString.call(message.autoStrikeWindowSec) : options.longs === Number ? new $util.LongBits(message.autoStrikeWindowSec.low >>> 0, message.autoStrikeWindowSec.high >>> 0).toNumber() : message.autoStrikeWindowSec;
            return object;
        };

        /**
         * Converts this TalentCombatState to JSON.
         * @function toJSON
         * @memberof realtime.TalentCombatState
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        TalentCombatState.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for TalentCombatState
         * @function getTypeUrl
         * @memberof realtime.TalentCombatState
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TalentCombatState.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/realtime.TalentCombatState";
        };

        return TalentCombatState;
    })();

    return realtime;
})();

export { $root as default };
