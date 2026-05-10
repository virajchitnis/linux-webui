#!/usr/bin/env bash
# linux-admin installer
# Usage: sudo bash install.sh
#        sudo bash install.sh --uninstall

set -euo pipefail

BINARY_NAME="linux-admin"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/linux-admin"
TLS_DIR="$CONFIG_DIR/tls"
SYSTEM_USER="linux-admin"
SERVICE_NAME="linux-admin"
SUDOERS_FILE="/etc/sudoers.d/linux-admin"
POLKIT_RULE="/etc/polkit-1/rules.d/50-linux-admin.rules"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info() { echo -e "${GREEN}[info]${NC} $*"; }
warn() { echo -e "${YELLOW}[warn]${NC} $*"; }
die()  { echo -e "${RED}[error]${NC} $*" >&2; exit 1; }

[[ $EUID -ne 0 ]] && die "Must be run as root"

if [[ "${1:-}" == "--uninstall" ]]; then
    info "Uninstalling linux-admin…"
    systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    systemctl daemon-reload
    rm -f "$INSTALL_DIR/$BINARY_NAME"
    rm -f "$SUDOERS_FILE"
    rm -f "$POLKIT_RULE"
    if id "$SYSTEM_USER" &>/dev/null; then
        userdel -r "$SYSTEM_USER" 2>/dev/null || true
    fi
    warn "Config and database at $CONFIG_DIR were NOT removed — delete manually if needed"
    info "Uninstall complete"
    exit 0
fi

# ── 1. Binary ─────────────────────────────────────────────────────────────────
BINARY_SRC="${BINARY_NAME}"
[[ -f "$BINARY_SRC" ]] || die "Binary '$BINARY_SRC' not found in current directory. Build it first: make build"
info "Installing binary to $INSTALL_DIR/$BINARY_NAME"
install -m 0755 "$BINARY_SRC" "$INSTALL_DIR/$BINARY_NAME"

# ── 2. System user ────────────────────────────────────────────────────────────
if ! id "$SYSTEM_USER" &>/dev/null; then
    info "Creating system user '$SYSTEM_USER'"
    useradd --system --no-create-home --shell /usr/sbin/nologin "$SYSTEM_USER"
fi

# ── 3. Directories ────────────────────────────────────────────────────────────
info "Creating config directory $CONFIG_DIR"
mkdir -p "$TLS_DIR"
chown -R "$SYSTEM_USER:$SYSTEM_USER" "$CONFIG_DIR"
chmod 750 "$CONFIG_DIR"
chmod 700 "$TLS_DIR"

# ── 4. Config file ────────────────────────────────────────────────────────────
CONFIG_FILE="$CONFIG_DIR/config.toml"
if [[ ! -f "$CONFIG_FILE" ]]; then
    info "Writing default config to $CONFIG_FILE"
    cat > "$CONFIG_FILE" <<'EOF'
[server]
listen      = "0.0.0.0:8443"
http_listen = "0.0.0.0:8080"
tls_cert    = "/etc/linux-admin/tls/cert.pem"
tls_key     = "/etc/linux-admin/tls/key.pem"

[auth]
session_timeout_minutes = 30
max_login_attempts      = 5
lockout_minutes         = 15
bcrypt_cost             = 12

[features]
terminal_enabled     = true
file_browser_enabled = true
file_browser_roots   = ["/etc", "/var/log", "/home"]
max_terminal_sessions = 3

[ollama]
enabled  = false
endpoint = "http://127.0.0.1:11434"
model    = "llama3.2"

[monitoring]
prometheus = false

[audit]
retention_days = 90
EOF
    chown "$SYSTEM_USER:$SYSTEM_USER" "$CONFIG_FILE"
    chmod 640 "$CONFIG_FILE"
else
    warn "Config file $CONFIG_FILE already exists — not overwriting"
fi

# ── 5. sudoers ────────────────────────────────────────────────────────────────
info "Installing sudoers rules"
cat > "$SUDOERS_FILE" <<EOF
# linux-admin sudoers — managed by install.sh
linux-admin ALL=(root) NOPASSWD: /sbin/reboot
linux-admin ALL=(root) NOPASSWD: /usr/bin/apt-get update
linux-admin ALL=(root) NOPASSWD: /usr/bin/apt-get upgrade -y
linux-admin ALL=(root) NOPASSWD: /usr/bin/apt-get install --
linux-admin ALL=(root) NOPASSWD: /usr/bin/apt-get remove --
linux-admin ALL=(root) NOPASSWD: /usr/sbin/ufw *
linux-admin ALL=(root) NOPASSWD: /usr/sbin/useradd *, /usr/sbin/userdel *, /usr/bin/passwd *
linux-admin ALL=(root) NOPASSWD: /usr/bin/wg set *, /usr/bin/wg-quick up *, /usr/bin/wg-quick down *
linux-admin ALL=(root) NOPASSWD: /usr/bin/tailscale up, /usr/bin/tailscale down, /usr/bin/tailscale set *
EOF
chmod 440 "$SUDOERS_FILE"
visudo -cf "$SUDOERS_FILE" || die "Generated sudoers file is invalid"

# ── 6. Polkit rule ────────────────────────────────────────────────────────────
info "Installing polkit rule"
cat > "$POLKIT_RULE" <<'EOF'
// Allow linux-admin to manage systemd units without a password
polkit.addRule(function(action, subject) {
    if (action.id === "org.freedesktop.systemd1.manage-units" &&
        subject.user === "linux-admin") {
        return polkit.Result.YES;
    }
});
EOF

# ── 7. journal group ──────────────────────────────────────────────────────────
if getent group systemd-journal &>/dev/null; then
    info "Adding $SYSTEM_USER to systemd-journal group"
    usermod -aG systemd-journal "$SYSTEM_USER" 2>/dev/null || true
fi

# Add to docker group if Docker is installed
if getent group docker &>/dev/null; then
    info "Docker detected — adding $SYSTEM_USER to docker group (see README security notes)"
    usermod -aG docker "$SYSTEM_USER" 2>/dev/null || true
fi

# ── 8. systemd unit ───────────────────────────────────────────────────────────
info "Installing systemd unit"
cat > "/etc/systemd/system/$SERVICE_NAME.service" <<EOF
[Unit]
Description=linux-admin server administration panel
After=network.target

[Service]
Type=simple
User=$SYSTEM_USER
Group=$SYSTEM_USER
ExecStart=$INSTALL_DIR/$BINARY_NAME --config $CONFIG_FILE
Restart=on-failure
RestartSec=5
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=$CONFIG_DIR
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

# ── 9. Done ───────────────────────────────────────────────────────────────────
echo ""
info "Installation complete!"
echo ""
echo "  Start: sudo systemctl start $SERVICE_NAME"
echo "  Open:  https://$(hostname -I | awk '{print $1}'):8443"
echo "         (accept the self-signed certificate)"
echo ""
warn "On first visit you will be prompted to create an admin account."
