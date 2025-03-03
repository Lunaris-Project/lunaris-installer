# Hyprland Installer

A TUI (Terminal User Interface) installer for Hyprland with the Lunaris configuration.

## Features

- Install an AUR helper of your choice (yay or paru)
- Install required packages for Hyprland
- Select optional packages by category
- Copy configuration files to your home directory

## Requirements

- Arch Linux or Arch-based distribution
- Go 1.18 or later
- Base development tools (git, base-devel)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/Lunaris-Project/Lunaric.git
cd Lunaric
```

2. Clone the installer repository:

```bash
git clone https://github.com/nixev/hyprland-installer.git
cd hyprland-installer
```

3. Build the installer:

```bash
go build -o hyprland-installer cmd/installer/main.go
```

4. Run the installer:

```bash
./hyprland-installer
```

## Usage

The installer provides a simple TUI interface that guides you through the installation process:

1. **Welcome Screen**: Introduction to the installer
2. **AUR Helper Selection**: Choose between yay and paru
3. **Package Selection**: Select which packages to install by category
4. **Installation**: Install the selected packages and copy configuration files
5. **Completion**: Installation complete

### Navigation

- Use the arrow keys or `h`, `j`, `k`, `l` to navigate
- Use `Tab` to switch between categories and options
- Use `Enter` or `Space` to select an option
- Use `Esc` to go back
- Use `Ctrl+C` or `q` to quit
- Use `?` to show help

## What Gets Installed

The installer will:

1. Install the selected AUR helper (yay or paru)
2. Install required packages for Hyprland
3. Install optional packages based on your selection
4. Copy configuration files to your home directory

## Configuration Files

The following directories will be copied to your home directory:

- `.config`: Configuration files for various applications
- `.local`: Local files and scripts
- `.fonts`: Custom fonts

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- [Lunaris Project](https://github.com/Lunaris-Project) for the Hyprland configuration
- [Charm](https://charm.sh) for the Bubble Tea TUI framework 