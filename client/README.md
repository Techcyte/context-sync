# Client

The demo client is an `npm` project so npm must be installed to run it. It can be installed from [here](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm).

To run the demo client first run `npm install`. After that finishes run `npm run start` to run the demo client.
The demo client should open up in the default browser.

The UI in the demo client is very simple. For the best understanding of how the LIS protocol works it is recommended
to have the console open. There will be messages printed to the console that show the flow of the LIS protocol.

## Connecting to the server (secure WebSockets)

The client connects to the server over a **secure WebSocket** at `wss://localhost:4002/cm`. For the browser to allow
this connection, the server's self-signed certificate authority must be trusted on your machine first. Start the server
at least once (it generates and installs the certificate automatically) and follow the trust instructions in the
`server` folder's `README.md`. If the certificate isn't trusted the browser will silently fail to connect, there is no
"proceed anyway" prompt for WebSocket connections.
