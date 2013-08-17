#!/bin/bash

echo "Downloading rig..."
rig_path=/usr/local/bin/rig
rigd_path=/usr/local/bin/rigd

# TODO: change to this when rig is public:
#wget -O $rig_path  https://github.com/gocardless/rig/releases/download/v0.1/rig
#wget -O $rigd_path https://github.com/gocardless/rig/releases/download/v0.1/rigd

tarball_hash=d778bba35ff753c8ef3c56b53eae3f6cf3ab18ed
if [[ $(shasum /tmp/rig-v0.1.tgz | awk '{print $1}') != $tarball_hash ]]; then
  wget -O /tmp/rig-v0.1.tgz http://cl.ly/1K000I290c2Q/download/rig-v0.1.tgz
fi

tar -C /usr/local/bin -U -xzf /tmp/rig-v0.1.tgz

if [ ! -d $HOME/.rig ]; then
  echo "Creating ~/.rig directory.."
  mkdir $HOME/.rig
fi

echo "Installing LaunchAgent.."
launch_agent_name=com.gocardless.rigd
plist_path="$HOME/Library/LaunchAgents/$launch_agent_name.plist"
cat > $plist_path <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>$launch_agent_name</string>
  <key>ProgramArguments</key>
  <array>
    <string>$rigd_path</string>
    <string>$HOME/.rig</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>StandardErrorPath</key>
  <string>/usr/local/var/log/rigd.log</string>
  <key>StandardOutPath</key>
  <string>/usr/local/var/log/rigd.log</string>
</dict>
</plist>
EOF

if launchctl list | grep "$launch_agent_name" > /dev/null; then
  echo "Unloading old LaunchAgent.."
  launchctl unload $plist_path
fi

echo "Loading LaunchAgent.."
launchctl load $plist_path

if ps -ef | grep $rigd_path | grep -v grep > /dev/null; then
  echo "Running!"
else
  echo "Something went wrong..."
fi

