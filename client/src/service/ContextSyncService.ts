import { MessageError } from '../model/Error';
import { ContextItem, Message, MessageKindEnum, StatusCodeEnum } from '../model/Message';
import { StateEnum, StateEnumKeys } from '../model/State';

export interface ContextSyncOptions {
	url: string;
	version: number;
	application: string;
	onConnected: () => void;
	onSubscribed: (rejectReason?: string, statusCode?: number) => void;
	switchContext: (ctx?: ContextItem[]) => void;
	contextSwitchRequest: (ctx?: ContextItem[]) => void;
	contextSwitchRejected: (reason: string) => void;
	onClose: () => void;
	onError: (error: MessageError) => void;
	debugMode?: boolean;
}

export class ContextSyncService {
	private url: URL;
	private port: string;
	private version: number;
	private application: string;
	private debugMode: boolean;
	private onConnectedCallback: () => void;
	private onSubscribedCallback: (rejectReason?: string, statusCode?: number) => void;
	private switchContextCallback: (ctx?: ContextItem[]) => void; // Tell the frontend to switch context.
	private contextSwitchRequestCallback: (ctx?: ContextItem[]) => void;
	private contextSwitchRejectedCallback: ( reason: string, ctx?: ContextItem[]) => void;
	private onCloseCallback: () => void; // Called when the WebSocket closes.
	private onErrorCallback: (error: MessageError) => void; // Called when there is an error. The frontend can decide what it wants to do based on the error.

	private state: StateEnumKeys;
	private socket?: WebSocket;
	private context?: ContextItem[];
	private previousContex?: ContextItem[];
	private newContext?: ContextItem[];
	private connected = false;
	private subscribed = false;

	constructor({
		url,
		version,
		application,
		debugMode,
		onConnected,
		onSubscribed,
		switchContext,
		contextSwitchRequest,
		contextSwitchRejected,
		onClose,
		onError,
	}: ContextSyncOptions) {
		this.url = new URL(url);
		this.port = this.url.port;
		this.version = version;
		this.application = application;
		this.debugMode = debugMode ?? false;
		this.onConnectedCallback = onConnected;
		this.onSubscribedCallback = onSubscribed;
		this.switchContextCallback = switchContext;
		this.contextSwitchRequestCallback = contextSwitchRequest;
		this.contextSwitchRejectedCallback = contextSwitchRejected;
		this.onCloseCallback = onClose;
		this.onErrorCallback = onError;

		this.state = StateEnum.waiting;
	}

	private setState = (state: StateEnumKeys) => {
		if (state === this.state) {
			return;
		}

		this.state = state;
		if (this.debugMode) {
			console.log(`CtxSync: State: '${this.state}'`);
		}
	};

	private send = (message: Message) => {
		if (this.socket === undefined) {
			return;
		}

		if (this.socket.readyState !== WebSocket.OPEN) {
			if (this.debugMode) {
				console.log('CtxSync: WebSocket not open, returning from send.');
			}

			this.onErrorCallback({ message: 'WebSocket not open.', status: StatusCodeEnum.ServerError });
			return;
		}

		const { kind } = message;

		if (!this.subscribed) {
			if (kind !== MessageKindEnum.subscription_request) {
				if (this.debugMode) {
					console.log(`CtxSync: Tried to send ${kind} message while not subscribed, returning from send.`);
				}

				this.onErrorCallback({ message: `Tried to send ${kind} message while not subscribed.`, status: StatusCodeEnum.BadRequest });
				return;
			}
		}

		if (this.state === StateEnum.error) {
			if (this.debugMode) {
				console.log('CtxSync: In error state returning from send.');
			}

			this.onErrorCallback({ message: 'Tried to send message while in error state.', status: StatusCodeEnum.ServerError });
			return;
		}

		if (this.debugMode) {
			console.log(`CtxSync: Sending '${kind}' message`);
		}

		switch (kind) {
			case MessageKindEnum.subscription_request:
				this.setState(StateEnum.subscription_request_sent);
				break;
			case MessageKindEnum.context_change_request:
				this.setState(StateEnum.ctx_change_request_sent);
				break;
			case MessageKindEnum.context_change_reject:
				this.setState(StateEnum.reject_sent);
				break;
			case MessageKindEnum.context_change_accept:
			case MessageKindEnum.out_of_sync_error:
				break;
			case MessageKindEnum.subscription_accept:
			case MessageKindEnum.subscription_rejection:
				if (this.debugMode) {
					console.log('Tried to send subscription accept/reject message. These can only be sent by the host.');
				}

				this.onErrorCallback({
					message: 'Tried to send subscription accept/reject message. These can only be sent by the host.',
					status: StatusCodeEnum.BadRequest
				});
				this.setState(StateEnum.error);
				return;
			default:
				if (this.debugMode) {
					console.log(`CtxSync: Unhandled message type '${kind}' in send.`);
				}
				this.onErrorCallback({ message: `Unhandled message type '${kind}' in send.`, status: StatusCodeEnum.ServerError });
				this.setState(StateEnum.error);
				return;
		}

		const messageStr = JSON.stringify(message);
		this.socket.send(messageStr);
	};

	private handleMessage = async (event: MessageEvent) => {
		if (typeof event.data !== 'string') {
			if (this.debugMode) {
				console.warn("CtxSync: Expected 'string', got ", typeof event.data);
				console.log('CtxSync: Received message raw:', event.data);
			}

			return;
		}

		const message: Message = JSON.parse(event.data);
		const eventMsg = `CtxSync: Received '${message.kind}' message`;
		if (this.debugMode) {
			console.log(eventMsg);
			console.log('CtxSync: Received message', message);
		}

		const { kind } = message;
		switch (kind) {
			case MessageKindEnum.subscription_accept:
				this.context = message.context;
				this.subscribed = true;
				this.onSubscribedCallback();
				if (this.context) {
					this.switchContextCallback(this.context);
				}

				this.setState(StateEnum.waiting);
				break;
			case MessageKindEnum.subscription_rejection:
				this.subscribed = false;
				this.onSubscribedCallback(message.rejection?.reason, message.rejection?.status);
				this.setState(StateEnum.waiting);
				break;
			case MessageKindEnum.context_change_request:
				this.newContext = message.context;
				this.contextSwitchRequestCallback(message.context);
				break;
			case MessageKindEnum.context_change_accept:
				this.previousContex = this.newContext;
				this.context = this.newContext;
				this.newContext = undefined;

				this.switchContextCallback(this.context);
				this.setState(StateEnum.waiting);
				break;
			case MessageKindEnum.context_change_reject:
				if (this.debugMode) {
					console.log('CtxSync: Context change is declined.');
				}

				this.contextSwitchRejectedCallback(message.rejection?.reason ?? 'Unknown reason.', this.newContext);

				this.newContext = undefined;
				this.setState(StateEnum.waiting);
				break;
			case MessageKindEnum.subscription_request:
				if (this.debugMode) {
					console.log(`Received ${kind} message, the host should never send this event.`);
				}

				break;
			default:
				const errorMsg = `Unknown message kind '${kind}'`;
				this.onErrorCallback({ message: errorMsg, status: StatusCodeEnum.BadRequest });
				if (this.debugMode) {
					console.log(errorMsg);
				}

				break;
		}
	};

	private subscribe = () => {
		this.send({
			kind: MessageKindEnum.subscription_request,
			info: {
				version: 1,
				application: this.application,
			},
		});
	};

	readonly accept = () => {
		this.previousContex = this.context;
		this.context = this.newContext;
		this.newContext = undefined;

		this.switchContextCallback(this.context);
		this.send({
			kind: MessageKindEnum.context_change_accept,
			context: this.context,
		});

		this.setState(StateEnum.waiting);
	};

	readonly reject = () => {
		this.send({
			kind: MessageKindEnum.context_change_reject,
			current_context: this.context,
			context: this.newContext,
			rejection: {
				reason: 'rejection reason here',
				status: StatusCodeEnum.Conflict,
			},
		});

		this.newContext = undefined;
		this.setState(StateEnum.waiting);
	};

	readonly requestContextChange = (ctx: ContextItem[]) => {
		if (ctx.length === 0) {
			return;
		}

		this.newContext = ctx;
		this.send({
			kind: MessageKindEnum.context_change_request,
			context: ctx,
		});
	};

	readonly connect = () => {
		if (this.debugMode) {
			console.log('CtxSync: Connect called.');
		}

		if (this.connected) {
			if (this.debugMode) {
				console.log('CtxSync: Already connected.');
			}

			if (!this.subscribed) {
				if (this.debugMode) {
					console.log('CtxSync: Connected but not subscribed, calling subscribe.');
				}

				this.subscribe();
			}

			return;
		}

		this.socket = new WebSocket(this.url);

		this.socket.addEventListener('error', async (event) => {
			if (!this.connected) {
				if (this.debugMode) {
					console.log(`CtxSync: Failed to connect to '${this.url.toString()}'. Port ${this.port} is closed.`);
				}

				return;
			}

			if (this.debugMode) {
				console.error('CtxSync: WebSocket error:', event);
			}

			this.onErrorCallback({
				message: 'Unkown WebSocket error',
				error: event,
			});
		});

		this.socket.addEventListener('open', () => {
			if (this.debugMode) {
				console.log('CtxSync: WebSocket opened.');
			}

			this.connected = true;
			this.onConnectedCallback();
			this.subscribe();
		});

		this.socket.addEventListener('message', (event) => {
			this.handleMessage(event);
		});

		this.socket.addEventListener('close', (event) => {
			if (this.connected) {
				this.onCloseCallback(); // Only call the close callback if we were connected.
			}

			if (this.debugMode) {
				console.log('CtxSync: WebSocket closed:', event);
			}

			this.subscribed = false;
			this.connected = false;
		});
	};

	readonly close = () => {
		if (this.socket !== undefined && this.socket.readyState === WebSocket.OPEN) {
			this.socket.close();
			this.subscribed = false;
			this.connected = false;
			this.onCloseCallback();
		}
	};

	readonly isConnected = () => {
		return this.connected;
	};

	readonly isSubscribed = () => {
		return this.subscribed;
	};

	readonly protocolVersion = () => {
		return this.version;
	};
}
