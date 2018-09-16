# Discord ping-pong BOT
<small>Because !ping is not enough</small>

## What it does
It works as a ping-pong table... kind-of.

Two users can share a friendly game of ping-pong/table tennis, by sending messages back and forth. The BOT outputs whenever the hit was successful or not.

## Running
1. Download the latest release from GitHub
2. Extract the zip into it's own folder
3. Upload pong_ping.png as a custom emoji on the server you want to run the BOT
4. Do \:pong_ping: in chat and copy what was sent
5. Rename config.example.yaml to config.yaml
6. Set `pong-prefix` in config.yaml to what you copied
7. Create an app in [discordapp.com/developers](https://discordapp.com/developers/applications/), create a BOT account and copy its token
8. Set `token` in config.yaml to what you copied
9. Invite the BOT to your server
10. Run ping-pong.exe to start the BOT

## TODO
- [x] A better mechanic to determine whenever a hit was successful
