import React, { useEffect, useRef, useState } from 'react';
import CssBaseline from '@mui/material/CssBaseline';
import { createTheme, Snackbar, ThemeProvider, Typography } from '@mui/material';
import { indigo, pink } from '@mui/material/colors';
import DemoPage from './DemoPage';
import { ContextSyncOptions, ContextSyncService } from './service/ContextSyncService';
import { ContextItem } from './model/Message';
import { MessageError } from './model/Error';

function App() {
	const service = useRef<ContextSyncService | undefined>(undefined);
	const [caseNumber, setCaseNumber] = useState('');
	const [rejectCaseNumber, setRejectCaseNumber] = useState<string | undefined>();
	const [rejectReason, setRejectReason] = useState<string | undefined>();
	const [newCaseNumber, setNewCaseNumber] = useState<string | undefined>();
	const [connected, setConnected] = useState(false);
	const [subscribed, setSubscribed] = useState(false);
	const [showRejectAlert, setShowRejectAlert] = useState(false);

	const onConnected = () => {
		setConnected(true);
	};

	const onSubscribed = (rejectReason?: string, statusCode?: number) => {
		setSubscribed(!rejectReason);

		if (!rejectReason) {
			console.log('Subscribed!');
		} else {
			console.log(`Subscription request rejected. Reason: ${rejectReason}`);
		}
	};

	const switchContext = (ctx?: ContextItem[]) => {
		ctx?.forEach((item) => {
			if (item.key === 'case') {
				setCaseNumber(item.value);
				return;
			}
		});

		setNewCaseNumber(undefined);
	};

	const contextSwitchRequest = (ctx?: ContextItem[]) => {
		ctx?.forEach((item) => {
			if (item.key === 'case') {
				setNewCaseNumber(item.value);
				return;
			}
		});
	};

	const contextSwitchRejected = (reason: string, ctx?: ContextItem[]) => {
		console.log(`Context switch rejected. Reason: ${reason}`);
		setRejectReason(reason);
		ctx?.forEach((item) => {
			if (item.key === 'case') {
				setRejectCaseNumber(item.value);
				return;
			}
		});

		setShowRejectAlert(true);
	}

	const onClose = () => {
		setConnected(false);
		setSubscribed(false);
		console.log('Connection closed.');
	}

	const onError = (error: MessageError) => {
		console.log('Demo error:', error);
	};

	const onReject = () => {
		setNewCaseNumber(undefined);
	};

	const darkTheme = createTheme({
		palette: {
			mode: 'dark',
			primary: {
				main: indigo[800],
			},
			secondary: pink,
		},
	});

	useEffect(() => {
		if (service.current !== undefined) {
			return;
		}

		const options: ContextSyncOptions = {
			url: 'ws://localhost:4002/cm',
			version: 1,
			application: 'Demo',
			debugMode: true,
			onConnected,
			onSubscribed,
			switchContext,
			contextSwitchRequest,
			contextSwitchRejected,
			onClose,
			onError,
		};

		service.current = new ContextSyncService(options);
		service.current.connect();
	// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	return (
		<ThemeProvider theme={darkTheme}>
			<CssBaseline />
			<DemoPage service={service} caseNumber={caseNumber} newCase={newCaseNumber} rejectCallback={onReject} connected={connected} subscribed={subscribed} />
			<Snackbar
				open={showRejectAlert}
				autoHideDuration={6000}
				onClose={() => {
					setShowRejectAlert(false);
					setRejectReason(undefined);
				}}
				anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
			>
				<Typography>
					Context change to case {rejectCaseNumber} rejected. Reason: {rejectReason}
				</Typography>
			</Snackbar>
		</ThemeProvider>
	);
}

export default App;
