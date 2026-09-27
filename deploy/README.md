# DeskOrder deploy

This directory provides one PowerShell entrypoint to build and deploy all three targets:

- backend (`backend`)
- admin web (`admin-web`)
- merchant H5 (`merchant-h5`)

## Files

- `deploy.ps1`: build, upload, deploy, and optional nginx reload
- `config.example.psd1`: copy to `config.psd1` and edit values
- `config.dev.psd1` / `config.trial.psd1` / `config.prod.psd1`: optional local profiles copied from `config.example.psd1`; these files are ignored by Git
- `nginx.subdomain.http.example.conf`: HTTP example for first certificate issuance
- `nginx.subdomain.https.example.conf`: HTTPS example for post-certificate deployment
- `nginx.single-host.multi-env.http.example.conf`: one-host dev/trial/prod HTTP example
- `nginx.single-host.multi-env.https.example.conf`: one-host dev/trial/prod HTTPS example

## Usage

1. Copy `config.example.psd1` to `config.psd1`, or create one ignored profile per environment
2. Edit SSH settings, remote directories, `Backend.ConfigDir`, and `Backend.RuntimeAssetsDir`
3. Run:

```powershell
pwsh -File .\deploy\deploy.ps1 -ConfigPath .\deploy\config.psd1
```

For one host with multiple environments, do not reuse a single config file. Keep one deploy config per environment, for example:

```powershell
pwsh -File .\deploy\deploy.ps1 -ConfigPath .\deploy\config.dev.psd1
pwsh -File .\deploy\deploy.ps1 -ConfigPath .\deploy\config.trial.psd1
pwsh -File .\deploy\deploy.ps1 -ConfigPath .\deploy\config.prod.psd1
```

## Backend production config

The backend always starts with `configs/config.yaml` inside the deployed package.

The deploy script decides what gets packed into that `configs/` directory by reading `Backend.ConfigDir` from `config.psd1`.

Recommended practice:

- copy `backend/configs/config.example.yaml` to the ignored `backend/configs/config.yaml` for local development
- keep every deploy environment's real config in a separate directory outside the repository
- point `Backend.ConfigDir` to the corresponding external directory when deploying

Example:

```powershell
Backend = @{
	ProjectDir = 'backend'
	Entry = './cmd/backend'
	ConfigDir = 'D:\path\to\desk-order-secrets\backend\prod'
	RuntimeAssetsDir = 'D:\path\to\desk-order-secrets\runtime-assets'
}
```

That directory should contain at least:

```text
config.yaml
```

If you need to deploy runtime-sensitive files together with the backend, for example `apiclient_cert.p12`, put them under `Backend.RuntimeAssetsDir`.

The deploy script will merge all files from that directory into the backend package root, while automatically skipping `Backend.ConfigDir` when it is a child directory.

For example, when the external runtime assets directory contains `cert/apiclient_cert.p12`:

- local certificate path: `D:\path\to\desk-order-secrets\runtime-assets\cert\apiclient_cert.p12`
- remote certificate path after deploy: `/srv/deskorder/backend/cert/apiclient_cert.p12`

So the backend production config can point `wechat.refund_cert_p12_path` to:

```yaml
wechat:
  refund_cert_p12_path: "/srv/deskorder/backend/cert/apiclient_cert.p12"
```

On each deployment, the script will copy that directory into the artifact as `configs/`, so the remote backend always starts with the production config and you do not need to manually replace files on the server.

Optional flags:

- `-SkipBackend`
- `-SkipAdmin`
- `-SkipMerchantH5`
- `-ForceNginxReload`
- `-DryRun`

## What the script does

### Backend

- builds `backend` with configurable `BuildOS` and `BuildArch`
- packs the binary plus the directory selected by `Backend.ConfigDir`
- optionally merges `Backend.RuntimeAssetsDir` into the backend package root
- uploads archive by `scp`
- extracts into remote `RemoteDir`
- kills old process by `PidFile` and `pkill -f`
- starts new process with `nohup`

Important: automated remote restart is implemented for Linux hosts because it uses `ssh` + `nohup`. If you change backend build target to `windows`, the script will stop and ask you to switch back to `linux` for remote deploy.

### Admin web and merchant H5

- runs `npm run build`
- supports optional `BuildEnv` values from `config.psd1`, so each environment can inject its own `VITE_*` variables without editing repo `.env.production`
- packs `dist`
- uploads archive by `scp`
- clears target directory except `.well-known`
- extracts the new build into the configured remote directory

Example:

```powershell
AdminWeb = @{
	ProjectDir = 'admin-web'
	RemoteDir = '/srv/www/deskorder/dev/admin-web'
	BuildEnv = @{
		VITE_API_ORIGIN = 'https://api-dev.xxx.com'
	}
}
```

The same pattern works for `MerchantH5`.

## One host, multiple environments

If you need `develop` / `trial` / `release` on the same server, the best layout is not one shared deployment with switches. The safer layout is one host, three isolated runtime stacks.

Recommended isolation boundaries:

- separate backend directories, for example `/srv/deskorder/dev/backend`, `/srv/deskorder/trial/backend`, `/srv/deskorder/prod/backend`
- separate frontend publish directories, for example `/srv/www/deskorder/dev/admin-web` and `/srv/www/deskorder/dev/merchant-h5`
- separate domains or subdomains, for example `api-dev.xxx.com`, `api-trial.xxx.com`, `api.xxx.com`
- separate MySQL databases, for example `deskorder_dev`, `deskorder_trial`, `deskorder`
- separate Redis DB indexes or separate Redis instances
- separate backend config directories on your local machine, one per environment

Why this keeps data from interfering:

- backend uploads are stored under each backend runtime directory as `./uploads`, so separate `Backend.RemoteDir` also separates uploaded files and generated QR assets
- backend `configs/config.yaml` already carries `server.port`, `mysql.dbname`, and `redis.db`; those are the core knobs for process and data isolation
- frontends can now build against different API origins by using `BuildEnv`, so dev/trial/prod no longer need to share one compiled API target

Recommended same-host topology:

```text
dev:
	api-dev.xxx.com        -> 127.0.0.1:18080 -> /srv/deskorder/dev/backend
	admin-dev.xxx.com      -> /srv/www/deskorder/dev/admin-web
	merchant-h5-dev.xxx.com -> /srv/www/deskorder/dev/merchant-h5

trial:
	api-trial.xxx.com      -> 127.0.0.1:28080 -> /srv/deskorder/trial/backend
	admin-trial.xxx.com    -> /srv/www/deskorder/trial/admin-web
	merchant-h5-trial.xxx.com -> /srv/www/deskorder/trial/merchant-h5

prod:
	api.xxx.com            -> 127.0.0.1:8080  -> /srv/deskorder/prod/backend
	admin.xxx.com          -> /srv/www/deskorder/prod/admin-web
	merchant-h5.xxx.com    -> /srv/www/deskorder/prod/merchant-h5
```

Operationally, the least risky way is:

- keep one nginx server block set per environment
- keep one deploy config file per environment
- keep one backend config directory per environment
- never let two environments point to the same MySQL database unless you explicitly want shared data
- never let two environments point to the same `RemoteDir`, `LogFile`, or `PidFile`

If you want stable production and cheap dev/test on one host, the usual priority order is:

- production gets its own database and Redis DB/index first
- trial gets its own database next
- development can share the same MySQL server process and Redis server process, but still must use its own database name / Redis DB and its own runtime directories

## Should nginx reload?

Usually no, if all of these stay true:

- backend upstream address does not change
- nginx config files do not change
- admin-web and merchant-h5 still publish into the same directories

Reload nginx only when you changed nginx config, certificate files, upstream targets, domains, or routing rules.

The script exposes nginx reload as an option instead of forcing it every time.

## Recommended hostname layout

Prefer subdomains over path-based routing for these three apps:

- `api.xxx.com`
- `admin.xxx.com`
- `merchant-h5.xxx.com`

Why this is better:

- cleaner nginx config
- cleaner cookies and auth boundaries
- fewer SPA base-path and history fallback issues
- easier CDN and TLS management
- clearer separation between API and two independent frontends

Path-based routing such as `api.xxx.com/admin/login` and `api.xxx.com/merchant-h5/login` is possible, but it creates more coupling and more rewrite/base-path maintenance.

## HTTPS with acme.sh

For this project, `acme.sh` is recommended over `certbot` because it is lightweight, works well with multiple subdomains, and is easier to combine with DNS verification later if needed.

### Target domains

Current recommended HTTPS domains:

- `api.example.com`
- `admin.example.com`
- `merchant-h5.example.com`

### Install acme.sh

```bash
curl https://get.acme.sh | sh
source ~/.bashrc
~/.acme.sh/acme.sh --version
```

### Recommended nginx switching strategy

For first certificate issuance, the least confusing approach is to keep two nginx files:

- `/etc/nginx/sites-available/deskorder.http.conf`
- `/etc/nginx/sites-available/deskorder.https.conf`

Recommended usage:

- `deskorder.http.conf`: HTTP only, used for first issuance
- `deskorder.https.conf`: HTTPS enabled, used after certificate install

Important:

- both files can exist in `sites-available` at the same time
- do not enable both at the same time in `sites-enabled` if they contain the same `listen` + `server_name`
- keep one stable symlink such as `/etc/nginx/sites-enabled/deskorder.conf` and switch that symlink between the two files

Example switch commands:

```bash
ln -sfn /etc/nginx/sites-available/deskorder.http.conf /etc/nginx/sites-enabled/deskorder.conf
nginx -t && systemctl reload nginx
```

After certificate install:

```bash
ln -sfn /etc/nginx/sites-available/deskorder.https.conf /etc/nginx/sites-enabled/deskorder.conf
nginx -t && systemctl reload nginx
```

### First issuance: prepare HTTP validation

Create the challenge webroot:

```bash
mkdir -p /var/www/_letsencrypt/.well-known/acme-challenge
```

For the very first issuance, save the following full config as `/etc/nginx/sites-available/deskorder.http.conf`.

This version keeps all three sites on HTTP temporarily so `acme.sh --webroot` can complete without certificate dependencies.

```nginx
upstream deskorder_backend {
	server 127.0.0.1:8080;
	keepalive 32;
}

server {
	listen 80;
	server_name api.example.com;

	client_max_body_size 10m;

	location ^~ /.well-known/acme-challenge/ {
		root /var/www/_letsencrypt;
		default_type "text/plain";
		try_files $uri =404;
		allow all;
	}

	location /uploads/ {
		proxy_http_version 1.1;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_pass http://deskorder_backend;
	}

	location / {
		proxy_http_version 1.1;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_pass http://deskorder_backend;
	}
}

server {
	listen 80;
	server_name admin.example.com;

	root /srv/www/deskorder/admin-web;
	index index.html;

	location ^~ /.well-known/acme-challenge/ {
		root /var/www/_letsencrypt;
		default_type "text/plain";
		try_files $uri =404;
		allow all;
	}

	location / {
		try_files $uri $uri/ /index.html;
	}
}

server {
	listen 80;
	server_name merchant-h5.example.com;

	root /srv/www/deskorder/merchant-h5;
	index index.html;

	location ^~ /.well-known/acme-challenge/ {
		root /var/www/_letsencrypt;
		default_type "text/plain";
		try_files $uri =404;
		allow all;
	}

	location / {
		try_files $uri $uri/ /index.html;
	}
}
```

Enable it:

```bash
ln -sfn /etc/nginx/sites-available/deskorder.http.conf /etc/nginx/sites-enabled/deskorder.conf
nginx -t && systemctl reload nginx
```

Before issuing the certificate, it is worth testing one challenge file manually:

```bash
echo ok > /var/www/_letsencrypt/.well-known/acme-challenge/test.txt
curl http://api.example.com/.well-known/acme-challenge/test.txt
curl http://admin.example.com/.well-known/acme-challenge/test.txt
curl http://merchant-h5.example.com/.well-known/acme-challenge/test.txt
```

### Issue one certificate for all three domains

This issues one SAN certificate that covers all three domains:

```bash
~/.acme.sh/acme.sh --issue \
	-d api.example.com \
	-d admin.example.com \
	-d merchant-h5.example.com \
	--webroot /var/www/_letsencrypt
```

Important:

- the first `-d` is treated as the primary domain
- later install and renew commands should continue to use `api.example.com`
- the domains must already resolve to the server and port 80 must be reachable
- with `--webroot`, nginx keeps full control and renewals are easier to reason about than `--nginx` auto-edit mode

### Install certificate into fixed nginx paths

Create a stable certificate directory:

```bash
mkdir -p /etc/nginx/ssl/deskorder
```

Install the certificate and configure automatic nginx reload after renew:

```bash
~/.acme.sh/acme.sh --install-cert -d api.example.com \
	--key-file /etc/nginx/ssl/deskorder/key.pem \
	--fullchain-file /etc/nginx/ssl/deskorder/fullchain.pem \
	--reloadcmd "systemctl reload nginx"
```

Why only `-d api.example.com` is used here:

- `api.example.com` was the primary domain when the SAN certificate was issued
- `acme.sh` stores that one certificate set under the primary domain entry
- the installed `/etc/nginx/ssl/deskorder/fullchain.pem` and `/etc/nginx/ssl/deskorder/key.pem` are the shared certificate files for all three domains in that SAN certificate
- nginx can point `api.example.com`, `admin.example.com`, and `merchant-h5.example.com` to the same two files

So yes, in this setup all three domains share one certificate.

You only need separate `--install-cert` commands if you issued separate certificates per domain.

After install, nginx should use only these two files:

- `/etc/nginx/ssl/deskorder/key.pem`
- `/etc/nginx/ssl/deskorder/fullchain.pem`

Do not point nginx directly at acme.sh internal storage paths.

### Switch to HTTPS config

After the certificate files exist, switch the enabled nginx file to `deskorder.https.conf`:

Save the following full config as `/etc/nginx/sites-available/deskorder.https.conf`:

```nginx
upstream deskorder_backend {
	server 127.0.0.1:8080;
	keepalive 32;
}

server {
	listen 80;
	server_name api.example.com;

	location ^~ /.well-known/acme-challenge/ {
		root /var/www/_letsencrypt;
		default_type "text/plain";
		try_files $uri =404;
		allow all;
	}

	location / {
		return 301 https://$host$request_uri;
	}
}

server {
	listen 443 ssl http2;
	server_name api.example.com;

	ssl_certificate     /etc/nginx/ssl/deskorder/fullchain.pem;
	ssl_certificate_key /etc/nginx/ssl/deskorder/key.pem;
	ssl_session_timeout 1d;
	ssl_session_cache shared:SSL:10m;
	ssl_protocols TLSv1.2 TLSv1.3;

	client_max_body_size 10m;

	location = /api/v1/pos/ws {
		proxy_http_version 1.1;
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_read_timeout 3600s;
		proxy_send_timeout 3600s;
		proxy_pass http://deskorder_backend;
	}

	location = /api/v1/merchant/ws {
		proxy_http_version 1.1;
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_read_timeout 3600s;
		proxy_send_timeout 3600s;
		proxy_pass http://deskorder_backend;
	}

	location = /api/v1/pos/ws {
		proxy_http_version 1.1;
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_read_timeout 3600s;
		proxy_send_timeout 3600s;
		proxy_pass http://deskorder_backend;
	}

	location = /api/v1/merchant/ws {
		proxy_http_version 1.1;
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_read_timeout 3600s;
		proxy_send_timeout 3600s;
		proxy_pass http://deskorder_backend;
	}

	location /uploads/ {
		proxy_http_version 1.1;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_pass http://deskorder_backend;
	}

	location / {
		proxy_http_version 1.1;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_pass http://deskorder_backend;
	}
}

server {
	listen 80;
	server_name admin.example.com;

	location ^~ /.well-known/acme-challenge/ {
		root /var/www/_letsencrypt;
		default_type "text/plain";
		try_files $uri =404;
		allow all;
	}

	location / {
		return 301 https://$host$request_uri;
	}
}

server {
	listen 443 ssl http2;
	server_name admin.example.com;

	ssl_certificate     /etc/nginx/ssl/deskorder/fullchain.pem;
	ssl_certificate_key /etc/nginx/ssl/deskorder/key.pem;
	ssl_session_timeout 1d;
	ssl_session_cache shared:SSL:10m;
	ssl_protocols TLSv1.2 TLSv1.3;

	root /srv/www/deskorder/admin-web;
	index index.html;

	location / {
		try_files $uri $uri/ /index.html;
	}
}

server {
	listen 80;
	server_name merchant-h5.example.com;

	location ^~ /.well-known/acme-challenge/ {
		root /var/www/_letsencrypt;
		default_type "text/plain";
		try_files $uri =404;
		allow all;
	}

	location / {
		return 301 https://$host$request_uri;
	}
}

server {
	listen 443 ssl http2;
	server_name merchant-h5.example.com;

	ssl_certificate     /etc/nginx/ssl/deskorder/fullchain.pem;
	ssl_certificate_key /etc/nginx/ssl/deskorder/key.pem;
	ssl_session_timeout 1d;
	ssl_session_cache shared:SSL:10m;
	ssl_protocols TLSv1.2 TLSv1.3;

	root /srv/www/deskorder/merchant-h5;
	index index.html;

	location / {
		try_files $uri $uri/ /index.html;
	}
}
```

Then enable it:

```bash
ln -sfn /etc/nginx/sites-available/deskorder.https.conf /etc/nginx/sites-enabled/deskorder.conf
nginx -t
systemctl reload nginx
```

### Nginx example notes

`deploy/nginx.subdomain.http.example.conf` and `deploy/nginx.subdomain.https.example.conf` match the two-file workflow documented above.

If you use the two-file strategy described above:

- make sure your real port 80 config always keeps `/.well-known/acme-challenge/`
- do not keep a top-level `return 301 ...` that swallows challenge requests

The shared certificate files remain:

- `/etc/nginx/ssl/deskorder/fullchain.pem`
- `/etc/nginx/ssl/deskorder/key.pem`

### Renewal

`acme.sh` normally installs a cron job automatically. Check it with:

```bash
crontab -l
```

Renewal does not require a separate nginx command if `--reloadcmd` was set during `--install-cert`.

If you continue to use HTTP-01 with `--webroot`, keep `/.well-known/acme-challenge/` permanently available on port 80. Renewal still needs it.

To manually force a renewal test:

```bash
~/.acme.sh/acme.sh --renew -d api.example.com --force
```

That command renews the same SAN certificate that contains:

- `api.example.com`
- `admin.example.com`
- `merchant-h5.example.com`

### Verification

Check certificate files:

```bash
ls -l /etc/nginx/ssl/deskorder
```

Check nginx config:

```bash
nginx -t
```

Check HTTPS responses:

```bash
curl -I https://api.example.com
curl -I https://admin.example.com
curl -I https://merchant-h5.example.com
```

Check SAN entries in the certificate:

```bash
openssl x509 -in /etc/nginx/ssl/deskorder/fullchain.pem -noout -text | grep -A1 "Subject Alternative Name"
```

### Notes

- If HTTP validation is unreliable in your environment, consider switching to DNS verification later.
- Frontend production env should use `https://api.example.com` as `VITE_API_ORIGIN`.
- HTTPS is required for reliable mobile browser geolocation in merchant H5.
