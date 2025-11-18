import { ValuesOf } from '../util/Types';

export const StateEnum = {
	error: 'error',
	waiting: 'waiting',
	sync_request_sent: 'sync_request_sent',
	ctx_change_request_sent: 'ctx_change_request_sent',
	accept_sent: 'accept_sent',
	reject_sent: 'reject_sent',
} as const;

export type StateEnumKeys = ValuesOf<typeof StateEnum>;
