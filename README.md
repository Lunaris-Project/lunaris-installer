# HyprLuna Installer

A TUI (Text User Interface) application for installing HyprLuna on Arch Linux.

## Features

- Select an AUR helper (yay or paru)
- Install base-devel and the selected AUR helper
- Install HyprLuna packages with the chosen AUR helper
- Option to install dotfiles with backup functionality
- Clone the HyprLuna repository for configuration
- Make scripts executable and set up the environment

## Requirements

- Arch Linux
- Go 1.21 or later
- Git

## Installation

Clone the repository:

```bash
git clone https://github.com/Lunaris-Project/lunaris-installer.git
cd hyprland-installer
```

Build the installer:

```bash
go build -o hyprland-installer cmd/main.go
```

Run the installer:

```bash
./hyprland-installer
```

## Usage

1. Select an AUR helper (yay or paru)
2. Choose packages to install from various categories
3. Start the installation
4. Enter your sudo password when prompted
5. Choose whether to install dotfiles
6. If installing dotfiles, choose whether to backup existing configuration
7. Wait for the installation to complete
8. Log out and select HyprLuna from your display manager

## Package Categories

The installer includes the following package categories:

- Terminals (Alacritty, Kitty, Foot)
- Shells (Zsh, Fish, Bash)
- Browsers (Firefox, Chromium, Brave)
- File Managers (Thunar, Dolphin, Nautilus)
- Text Editors (Neovim, Visual Studio Code, Gedit)
- Media Players (VLC, MPV, Celluloid)

## Configuration

The installer copies configuration files to the following directories:

- `.config` - Contains all configuration files for HyprLuna
- `.local` - Contains local binaries and other files
- `.fonts` - Contains fonts used by HyprLuna
- `.ags` - Contains AGS configuration
- `.cursor` - Contains cursor themes
- `.vscode` - Contains VSCode configuration
- `Pictures` - Contains wallpapers and other images

## License

MIT