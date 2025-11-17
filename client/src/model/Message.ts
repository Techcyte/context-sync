import { ValuesOf } from '../util/Types';

export const MessageKindEnum = {
	subscription_request: 'sub-request', // The client is requesting to subscribe.
	subscription_accept: 'sub-accept', // The host accepts the subscription request.
	subscription_rejection: 'sub-reject', // The host already has an active connection.
	context_change_request: 'ctx-change-request', // Request to change the current context.
	context_change_accept: 'ctx-change-accept', // The requestee accepts the context change request.
	context_change_reject: 'ctx-change-reject', // One of the parties rejects the change request with a given reason.
	empty_context: 'ctx-null', // Sent when the user navigates away from the active context to a page without an active context.
	out_of_sync_error: 'sync-error', // Sent if there is an error that results in a desync.
} as const;

export type MessageKindEnumKeys = ValuesOf<typeof MessageKindEnum>;

export const ContextItemKeyEnum = {
	case_number: 'case',
} as const;

export type ContextItemKeys = ValuesOf<typeof ContextItemKeyEnum>;

export const StatusCodeEnum = {
	OK: 200,
	BadRequest: 400,
	MethodNotAllowed: 405,
	RequestTimeout: 408,
	Conflict: 409,
	ConflictWithRetry: 419,
	UpgradeRequired: 426,
	TooManyRequests: 429,
	ServerError: 500,
} as const;

export type StatusCodeEnumKeys = ValuesOf<typeof StatusCodeEnum>;

export interface ContextItem {
	key: ContextItemKeys;
	value: string;
}

export interface MessageRejection {
	reason: string;
	status: StatusCodeEnumKeys;
}

export interface MessageError {
	message: string;
	status: number;
}

export interface ConnectionInfo {
	version: number;
	application: string;
	timeout?: number;
	replace_exiting_client?: boolean;
}

export interface Message {
	kind: MessageKindEnumKeys;
	info?: ConnectionInfo;
	context?: ContextItem[];
	current_context?: ContextItem[];
	rejection?: MessageRejection;
	error?: MessageError;
}
