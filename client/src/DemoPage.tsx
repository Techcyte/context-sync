import { Box, Button, TextField, Typography } from '@mui/material';
import { ContextSyncService } from './service/ContextSyncService';
import { useRef, useState } from 'react';

interface DemoPageType {
	service: React.RefObject<ContextSyncService | undefined>;
	caseNumber: string;
	newCase?: string;
	rejectCallback: () => void;
	connected: boolean;
	subscribed: boolean;
}

const DemoPage = ({ service, caseNumber, newCase, rejectCallback, connected, subscribed }: DemoPageType) => {
	const [userCaseNumber, setUserCaseNumber] = useState<string | undefined>();
	const inputRef = useRef<HTMLInputElement>(null);

	const connect = () => {
		service.current?.connect();
	};

	const requestContextChange = () => {
		if (!userCaseNumber) {
			return;
		}

		console.log(`Requesting context change to case ${userCaseNumber}`);
		inputRef.current?.blur();
		service.current?.requestContextChange([{ key: 'case', value: userCaseNumber }]);
		setUserCaseNumber(undefined);
	};

	const accept = () => {
		service.current?.accept();
	};

	const reject = () => {
		service.current?.reject();
		rejectCallback();
	};

	return (
		<Box m={2}>
			<Box display="flex" flexDirection="row" gap={2}>
				<Typography>Connected: {connected ? 'yes' : 'no'}</Typography>
				<Typography>Subscribed: {subscribed ? 'yes' : 'no'}</Typography>
				<Typography> Protocol Version: {service.current?.protocolVersion()}</Typography>
			</Box>
			{!connected && (
				<Button variant="contained" onClick={connect}>
					Connect
				</Button>
			)}
			<Box>
				<Typography>Current case: {caseNumber ? caseNumber : "{Unset}"}</Typography>
				<Box display="flex" gap={2}>
					<TextField
						label="Case Number"
						inputRef={inputRef}
						disabled={!subscribed}
						value={userCaseNumber ?? ''}
						onChange={(e) => setUserCaseNumber(e.target.value)} 
						onKeyDown={(e) => {
							if (e.key === 'Enter') {
								requestContextChange();
							}
						}}
						
					/>
					<Button variant="contained" onClick={requestContextChange} disabled={!userCaseNumber}>
						Request Change
					</Button>
				</Box>
			</Box>
			{newCase && (
				<Box display="flex" flexDirection="column" sx={{ m: 1 }}>
					<Typography>Change case to {newCase}?</Typography>
					<Box display="flex" sx={{ my: 1, gap: 2 }}>
						<Button variant="contained" onClick={accept} disabled={!newCase}>
							Accept
						</Button>
						<Button variant="contained" onClick={reject} disabled={!newCase}>
							Reject
						</Button>
					</Box>
				</Box>
			)}
		</Box>
	);
};

export default DemoPage;
