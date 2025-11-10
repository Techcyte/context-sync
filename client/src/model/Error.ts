import { Message } from './Message';

export interface MessageError {
	message: string;
	status?: number;
	error?: unknown;
	data?: Message;
}