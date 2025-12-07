# Clipboard Manager

A lightweight clipboard history manager for Linux with Fyne GUI.

## Installation

### 1. Install Dependencies

```bash
sudo apt install -y xclip xsel xdotool libgl1-mesa-dev xorg-dev
```

### 2. Build & Install

```bash
git clone https://github.com/phaneendra24/Clipboard-Manager.git
cd Clipboard-Manager
go build -o clipboard-manager .
sudo mv clipboard-manager /usr/local/bin/
```

### 3. Setup Auto-start Daemon

```bash
mkdir -p ~/.config/systemd/user
cp clipboard-manager.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now clipboard-manager.service
```

### 4. Add Keybinding

#### For i3wm

Add to `~/.config/i3/config`:

```
for_window [title="Clipboard Manager"] floating enable, border none, resize set 700 500, move position center
bindsym $mod+v exec --no-startup-id /usr/local/bin/clipboard-manager gui
```

Reload i3: `$mod+Shift+r`

#### For GNOME

```bash
# Add custom keybinding (Super+V)
gsettings set org.gnome.settings-daemon.plugins.media-keys custom-keybindings "['/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/clipboard-manager/']"
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/clipboard-manager/ name 'Clipboard Manager'
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/clipboard-manager/ command '/usr/local/bin/clipboard-manager gui'
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/clipboard-manager/ binding '<Super>v'
```

Or manually: **Settings → Keyboard → Custom Shortcuts → Add**
- Name: `Clipboard Manager`
- Command: `/usr/local/bin/clipboard-manager gui`
- Shortcut: `Super+V`

---

## Usage

Press your keybinding to open. Use **↑/↓** to navigate, **Enter** to paste, **Escape** to close.
