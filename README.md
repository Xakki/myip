# Описание проекта и функционал:

- При запуске используется фаил .env в котором указывается: 
    - WEB=host:port (обязятелен)
    - REDIS=host:port (обязателен)
    - REDIS_USER и REDIS_PASS (не обязательны)
    - RDAP_API=https://rdap.db.ripe.net/ip/{REMOTE_IP} (не обязателен)
    - LOG_TYPE=console/syslog/gelf/system
    - LOG_ADDR=
- Используем простую архитектуру с применением шаблонов для html. При компиляции все фаилы сохраняеются в бинарник, кроме .env
- Endpoint  только корневой /
- Получаем информацию об IP по запросу из RDAP_API и кешируем его в редис (храним неделю, но обновляем через сутки). Если ошибка в запросе то не падаем и в ответе просто будет пустой результат.
- При каждом запросе в редис сохраняем счетчик обращений по этому IP (count_call)
- если Content-type = json, то ответ выдаем в json формате. Базовая информация выдается:  
      ip,
      count_call,
      country, (этот и далее данные берутся из RDAP_API)
      handle, 
      ipVersion, 
      name, 
      type, 
      events,
- Если это обычный запрос из браузера, собираем дополнительно информацию из браузера на одной странице:
    - выводим информацию по IP полученную на бэке.
    - Всю доступную информацию браузера.
    - вывести информацию как в "Check for Proxy Detection" по аналогии как в https://www.whatismyip.com/proxy-check/
    - вывести информацию как в https://browserleaks.com/canvas
    - также вывести https://browserleaks.com/webrtc
    - https://browserleaks.com/javascript
    - https://browserleaks.com/fonts
    - https://browserleaks.com/canvas
 
# Примеры

- https://myip.xakki.pro/
- https://myip.xakki.pro/api
- https://api.myip.xakki.pro/
- https://myip.xakki.pro/?ip=127.0.0.1
- https://myip.xakki.pro/api?ip=127.0.0.1


# Запуск проекта как сервис

1. Скачать релизный бинарник и распаковать (например в /var/www/myip)
```
wget -O myip.tar.gz https://github.com/Xakki/myip/releases/download/v0.1/myip_0.1_linux_amd64.tar.gz
mkdir myip
tar -xvzf myip.tar.gz -C ./myip
cd myip
```

3. `nano .env` и отредактировать под свои нужды

4. Используем свой редис или запускаем `docker run -d --name keydb -p 6378:6379 -v myip-keydb:/data eqalpha/keydb`

5. Делаем автозапуск.
   создать фаил `nano /etc/systemd/system/myip.service`
```
[Unit]
Description=MyIp simple service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/myip
ExecStart=/var/www/myip/myip
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

6. Выполнить
```
systemctl daemon-reload
systemctl enable myip
systemctl start myip
systemctl status myip
```

Если статус красный, то смотрим логи `journalctl -u myip -f`

