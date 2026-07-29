# SplitNow

Bill-splitting chatroom app.

# Development

`sudo wscat -c ws://localhost:8080/ws?room=test` - to open a websocket in the room test
-> needs wscat through npm (`sudo npm install -g wscat`)

`fuser -k 8080/tcp` -> to stop any process listening to port :8080 


