# ABOUT

This is a simple telegram bot designed to work with [Valetudo](https://valetudo.cloud/) custom firmware for vacuum robot cleaners. It allows to:

 - Start complete and zoned cleaning
 - Pause cleaning, send robot to the dock
 - View basic robot status

# BUILDING

Change values for `GOARCH` and `GOARM` according to your robot's target architecture

```
GOOS=linux GOARCH=arm GOARM=7 go build
```

# CONFGIRATION

Configuration is specified via `-config` property (defaults to `/etc/vacuum_bot.json`). Configuration file should contain JSON object with following fields:

`apiUrl` - API URL of your robot. Usually something like `http://127.0.0.1`, if binary is executed on robot itself

`botToken` - Telegram bot token

`authorizedUsers` - Users authorized to use this bot (user name array)

# INSTALLATION

Copy binary and config file to your robot's filesystem. If your robot uses initd, there's a sample init script in `scripts` folder.
