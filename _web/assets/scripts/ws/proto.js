// TODO: maybe payload will need another number of bytes.
// Encode and Decode functions should use and allocate 0 bytes of memory
// while dealing with real data. So benchmark that.
export const MaxMessageLength = 1 + 100 // MessageKind + payload.

export const MessageKind = {
	// KindPing used to maintain a TCP connection (keepalive) and measure network latency.
	// Payload is a byte sequence (1 or more bytes) that represents the latency in milliseconds.
	KindPing: 0,
	// KindPong is sent in response to [Ping]. Payload is nil.
	KindPong: 1,
	// KindJoin used to register a player in matchmaking pool. Payload is nil.
	KindJoin: 2,
	// KindLeave is used to unregister a player from matchmaking pool. Payload is nil.
	KindLeave: 3,
	// KindCounter is used to notify a player about number of players in matchmaking queue.
	// Payload is an integer encoded in a byte sequence.
	KindCounter: 4,
	// KindAdd is used to notify Coordinator about new topic. Payload is nil.
	KindAdd: 5,
	// KindRemove is used to notify Coordinator about removed topic. Payload is nil.
	KindRemove: 6,
	// KindMatch is emmited when a new game is being created.
	// Payload is [Player1ID_Player2ID_GameId]. Note that '_' is used only for readability
	// and is ommited in real message.
	KindMatch: 7,
}

/**
 * @parameter {number} x
 * @return {Uint8Array}
 */
export function encodeInt(x) {
	const encoded = new Uint8Array(3)
	// Write 6 bits at a time.
	let i = 0
	for (; ; i++) {
		const chunk = (x & 0x3F)
		encoded[i] = chunk
		x >>= 6
		if (x == 0) {
			break
		}
		encoded[i] |= 0x40 // Set continuation bit.
	}
	return encoded.slice(0, i+1)
}

/**
 * @parameter {Uint8Array} encoded
 * @return {number}
 */
export function decodeInt(encoded) {
	let decoded = 0
	let coeff = 0
	for (let i = 0; i < encoded.length; i++) {
		decoded |= (encoded[i]&0x3F) << coeff
		coeff += 6
	}
	return decoded
}

export class Message {
	/**
	 * @type {Uint8Array}
	 */
	payload
	/**
	 * @type {MessageKind}
	 */
	kind

	constructor(payload, id, kind) {
		this.payload = payload
		this.id = id
		this.kind = kind
	}
}
