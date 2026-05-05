import { realtime } from '../proto/realtime.js'

export const realtimeBinaryType = {
  clickRequest: 1,
  clickAck: 2,
  publicDelta: 3,
  userDelta: 4,
  roomState: 5,
}

const decodeOptions = {
  longs: Number,
  arrays: true,
  objects: true,
}

const sparseDecodeOptions = {
  longs: Number,
  arrays: false,
  objects: true,
}

const nestedDecodeOptions = {
  longs: Number,
  defaults: true,
  arrays: true,
  objects: true,
}

function packFrame(messageType, encoded) {
  const body = encoded instanceof Uint8Array ? encoded : new Uint8Array(encoded)
  const frame = new Uint8Array(1 + body.length)
  frame[0] = messageType
  frame.set(body, 1)
  return frame
}

function unpackFrame(frame) {
  const bytes = frame instanceof Uint8Array ? frame : new Uint8Array(frame)
  if (bytes.length < 1) {
    throw new Error('empty realtime binary frame')
  }
  return {
    messageType: bytes[0],
    body: bytes.subarray(1),
  }
}

function toPlain(messageType, message, options = decodeOptions) {
  return messageType.toObject(message, options)
}

function normalizeBossPartsFromMessage(payload, message) {
  if (!payload?.boss || !message?.boss?.parts) {
    return payload
  }
  payload.boss.parts = message.boss.parts.map((part) => (
    realtime.BossPart.toObject(part, nestedDecodeOptions)
  ))
  return payload
}

function normalizePartStateDeltasFromMessage(payload, message) {
  if (!payload || !message?.partStateDeltas) {
    return payload
  }
  payload.partStateDeltas = message.partStateDeltas.map((delta) => (
    realtime.BossPartStateDelta.toObject(delta, nestedDecodeOptions)
  ))
  return payload
}

export function encodeRealtimeClickRequest({ slug, comboCount = 0 }) {
  const encoded = realtime.ClickRequest.encode(realtime.ClickRequest.create({
    slug,
    comboCount,
  })).finish()
  return packFrame(realtimeBinaryType.clickRequest, encoded)
}

export function decodeRealtimeBinaryMessage(frame) {
  const { messageType, body } = unpackFrame(frame)

  switch (messageType) {
    case realtimeBinaryType.clickAck:
      {
        const message = realtime.ClickAck.decode(body)
        return {
          type: 'click_ack',
          payload: normalizePartStateDeltasFromMessage(
            toPlain(realtime.ClickAck, message),
            message,
          ),
        }
      }
    case realtimeBinaryType.publicDelta:
      {
        const message = realtime.PublicDelta.decode(body)
        return {
          type: 'public_delta',
          payload: normalizeBossPartsFromMessage(
            toPlain(realtime.PublicDelta, message, sparseDecodeOptions),
            message,
          ),
        }
      }
    case realtimeBinaryType.userDelta:
      return {
        type: 'user_delta',
        payload: toPlain(realtime.UserDelta, realtime.UserDelta.decode(body)),
      }
    case realtimeBinaryType.roomState:
      return {
        type: 'room_state',
        payload: toPlain(realtime.RoomState, realtime.RoomState.decode(body)),
      }
    default:
      throw new Error('unsupported realtime binary message')
  }
}
