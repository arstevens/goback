build:
	go get -u github.com/arstevens/goback/...
	go get -u github.com/fsnotify/fsnotify
	go build -o /usr/local/bin/gobackd daemon/*.go
	echo "[Unit]" > /etc/systemd/system/gobackd.service
	echo "Description=Goback backup daemon" >> /etc/systemd/system/gobackd.service
	echo "After=network.target" >> /etc/systemd/system/gobackd.service
	echo "StartLimitIntervalSec=0" >> /etc/systemd/system/gobackd.service
	echo "[Service]" >> /etc/systemd/system/gobackd.service
	echo "Type=simple" >> /etc/systemd/system/gobackd.service
	echo "Restart=always" >> /etc/systemd/system/gobackd.service
	echo "RestartSec=1" >> /etc/systemd/system/gobackd.service
	echo "User=root" >> /etc/systemd/system/gobackd.service
	echo "ExecStart=/usr/local/bin/gobackd" >> /etc/systemd/system/gobackd.service
	echo "[Install]" >> /etc/systemd/system/gobackd.service
	echo "WantedBy=multi-user.target" >> /etc/systemd/system/gobackd.service
	systemctl start gobackd
	systemctl enable gobackd
	go build -o /usr/local/bin/goback cli/main.go
