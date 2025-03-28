# Lunaris Installer
# !!Not Ready yet dont use it please hit the readme of HyprLuna to install the dots!!

A TUI (Text User Interface) application for installing Lunaris on Arch Linux.

## Features

- Select an AUR helper (yay or paru)
- Choose packages to install from various categories
- Install Lunaris and selected packages
- Copy configuration files
- Clone the Lunaric repository for Lunaris configuration

## Requirements

- Arch Linux
- Go 1.21 or later
- Git

## Installation

Clone the repository:

```bash
git clone https://github.com/Lunaris-Project/lunaris-installer.git
cd lunaris-installer
```

Build the installer:

```bash
go build -o lunaris-installer cmd/main.go
```

Run the installer:

```bash
./lunaris-installer
```

## Usage

1. Select an AUR helper (yay or paru)
2. Choose packages to install from various categories
3. Start the installation
4. Enter your sudo password when prompted
5. Wait for the installation to complete
6. Log out and select Lunaris from your display manager

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

- `.config/hypr`
- `.config/waybar`
- `.config/rofi`
- `.config/dunst`

## License

MIT 
