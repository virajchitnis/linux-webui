linux-webui (beta)
==================
A simple web control panel for Linux servers.

This project was originally developed and tested on Gentoo. Most Linux distros are similar enough that the control panel should work on any of them. If you get it running on a distro not listed below, please open an issue so the OS support list can be updated.

OS support
----------

* Gentoo

Requirements
------------

* Apache 2.4+ with the following modules enabled:
  * `mod_php` (or `php-fpm`)
  * `mod_headers` (required for security headers in `.htaccess`)
  * `mod_rewrite` (required if you enable the HTTPS redirect in `.htaccess`)
* PHP 7.0 or later (uses `random_bytes()`, `password_hash()`, and the array
  form of `session_set_cookie_params()` — none of which are available in PHP 5)
* `sudo` access for the web server user, scoped to specific commands only
  (see the Sudoers section below)

Installation via Git (Recommended)
-----------------------------------

### 1. Install Apache and PHP

Install Apache 2.4+ and PHP 7.0+ using your distro's package manager.

### 2. Clone the repository

Clone into your web server's document root (or a subdirectory of it):

```bash
git clone https://github.com/virajchitnis/linux-webui.git
cd linux-webui
```

### 3. Install the sudoers drop-in

linux-webui needs to run `service`, `pmap`, and `rc-update` as root. A
ready-made drop-in file listing only those specific commands is provided —
**do not** use the old `NOPASSWD: ALL` pattern, which grants unrestricted
root access to the entire web server process.

```bash
# Review the file before installing
cat config/sudoers.example

# Edit it if needed (see notes below), then install
sudo cp config/sudoers.example /etc/sudoers.d/linux-webui
sudo chmod 0440 /etc/sudoers.d/linux-webui
sudo visudo -c          # verify syntax before relying on it
```

Things to check in `config/sudoers.example` before installing:

| Setting | Gentoo | Debian / Ubuntu | CentOS / RHEL |
|---------|--------|-----------------|---------------|
| Web server user | `apache` | `www-data` | `apache` |
| Path to `service` | `/sbin/service` | `/usr/sbin/service` | `/sbin/service` |
| Path to `pmap` | `/usr/bin/pmap` | `/usr/bin/pmap` | `/usr/bin/pmap` |

Verify the paths on your system with `which service` and `which pmap`.

### 4. Set the admin password

The application ships with a default password of `changeme`. **Change it
before the server is reachable by anyone else.**

Generate a bcrypt hash for your chosen password:

```bash
php -r "echo password_hash('your_password_here', PASSWORD_DEFAULT);"
```

Open `config/credentials.php` and replace the value of `AUTH_PASSWORD_HASH`
with the output. You can also change `AUTH_USERNAME` from `admin` if you prefer.

### 5. Set permissions on the data directory

The `data/` directory is used to store runtime files (login rate-limit
counters, the update log). It must be writable by the web server user:

```bash
# Gentoo / CentOS / RHEL
sudo chown apache:apache data/

# Debian / Ubuntu
sudo chown www-data:www-data data/
```

The directory is already protected from direct web access by `.htaccess`.

### 6. Enable required Apache modules

```bash
# Debian / Ubuntu (using a2enmod)
sudo a2enmod headers rewrite
sudo systemctl restart apache2

# Gentoo — ensure the following are in your Apache USE flags or httpd.conf:
#   LoadModule headers_module modules/mod_headers.so
#   LoadModule rewrite_module modules/mod_rewrite.so
```

### 7. Set up HTTPS (recommended)

Running a server control panel over plain HTTP means your session cookie and
all commands are visible on the network. Once you have a TLS certificate in
place, enable HTTPS enforcement by uncommenting two blocks in `.htaccess`:

```apache
# Uncomment to redirect all HTTP traffic to HTTPS
RewriteEngine On
RewriteCond %{HTTPS} off
RewriteRule ^ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]

# Uncomment to send HSTS header (tells browsers to always use HTTPS)
Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
```

> **Note:** Do not enable HSTS until HTTPS is fully working. A misconfigured
> HSTS header can lock browsers out of the site entirely.

### 8. Log in

Navigate to the site in your browser. You will be presented with a login page.
Use the credentials set in step 4.

Future updates to the web application can be installed by clicking the **Update** button on the About page.

---

Installation via tar.gz (Not recommended)
------------------------------------------

1. Install Apache 2.4+ and PHP 7.0+.
2. Download the tar.gz for the latest release into your web server directory:
   ```bash
   wget https://github.com/virajchitnis/linux-webui/archive/v1.1.3.tar.gz
   ```
   (Replace `v1.1.3` with the version you wish to download.)
3. Extract: `tar -zxvf v1.1.3.tar.gz`
4. Delete the archive: `rm v1.1.3.tar.gz`
5. Follow steps 3–7 from the Git installation above.

Future updates must be applied by downloading and extracting a newer tar.gz
over the current directory. Using Git is strongly recommended as it allows
one-click updates from the About page.

---

Security overview
-----------------

The following protections are built into this version of linux-webui:

| Area | Protection |
|------|-----------|
| **Authentication** | Session-based login with bcrypt password hashing. Every page requires a valid session. |
| **Session cookies** | `HttpOnly` and `SameSite=Strict` flags set on every session cookie. `Secure` flag is set automatically when the request arrives over HTTPS. |
| **Session timeout** | Sessions expire after 30 minutes of inactivity. |
| **CSRF** | All state-changing operations (service control, reboot, git update) require a POST request with a per-session CSRF token. |
| **Brute force** | Login is rate-limited by IP address. 5 failed attempts trigger a 15-minute lockout stored server-side — clearing cookies does not reset it. |
| **Command injection** | Service operations use a strict whitelist of allowed service names and operations. No user input is interpolated into shell commands. |
| **File read** | `catfile.php` only serves the update log. All other paths return 403. |
| **XSS** | All shell command output is passed through `htmlspecialchars()` before being written into HTML. |
| **Security headers** | `X-Frame-Options`, `X-Content-Type-Options`, `X-XSS-Protection`, `Content-Security-Policy`, and `Referrer-Policy` are set via `.htaccess`. |
| **sudo scope** | The web server user is granted sudo access only to the specific `service`, `pmap`, `rc-update show`, and `reboot` commands that linux-webui actually uses. |
| **Sensitive directories** | The `config/` and `data/` directories are blocked from direct web access via `.htaccess`. |

### Remaining considerations

* **HTTPS**: Strongly recommended. Session cookies and all control traffic are
  sent in plaintext over HTTP. See step 6 above.
* **Shared NAT**: The IP-based brute-force lockout treats all users behind the
  same NAT address as one. On a shared network, one user's failed attempts can
  lock out others. This is acceptable for a single-admin control panel.
* **sudo scope for reboot**: `/sbin/reboot` is permitted without a password.
  The reboot action is protected by authentication and a CSRF token in the
  application layer, but the underlying command is still available to the web
  server user via sudo.
* **git update over HTTP**: If your git remote is an HTTP (not HTTPS) URL, a
  `git pull` is vulnerable to a man-in-the-middle attack delivering malicious
  code. Ensure your remote URL uses `https://` or SSH.
