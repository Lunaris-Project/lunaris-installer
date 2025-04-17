package config

// AURHelpers is a list of available AUR helpers
var AURHelpers = []string{"yay", "paru"}

// BasePackages is a list of base packages that are always installed
var BasePackages = []string{
	"hyprland",
	"axel",
	"bc",
	"coreutils",
	"cliphist",
	"cmake",
	"curl",
	"rofi-wayland",
	"rsync",
	"wget",
	"ripgrep",
	"jq",
	"npm",
	"meson",
	"typescript",
	"gjs",
	"xdg-user-dirs",
	"brightnessctl",
	"ddcutil",
	"pavucontrol",
	"wireplumber",
	"libdbusmenu-gtk3",
	"playerctl",
	"swww",
	"git",
	"gobject-introspection",
	"glib2-devel",
	"gvfs",
	"glib2",
	"glibc",
	"gtk3",
	"gtk-layer-shell",
	"libpulse",
	"pam",
	"gnome-bluetooth-3.0",
	"gammastep",
	"libsoup3",
	"libnotify",
	"networkmanager",
	"power-profiles-daemon",
	"upower",
	"adw-gtk-theme-git",
	"qt5ct",
	"qt5-wayland",
	"fontconfig",
	"ttf-readex-pro",
	"ttf-jetbrains-mono-nerd",
	"ttf-material-symbols-variable-git",
	"apple-fonts",
	"ttf-space-mono-nerd",
	"ttf-rubik-vf",
	"ttf-gabarito-git",
	"fish",
	"foot",
	"starship",
	"polkit-gnome",
	"gnome-keyring",
	"gnome-control-center",
	"blueberry",
	"webp-pixbuf-loader",
	"gtksourceview3",
	"yad",
	"ydotool",
	"xdg-user-dirs-gtk",
	"tinyxml2",
	"gtkmm3",
	"gtksourceviewmm",
	"cairomm",
	"xdg-desktop-portal",
	"xdg-desktop-portal-gtk",
	"xdg-desktop-portal-hyprland",
	"gradience",
	"python-libsass",
	"python-pywalfox",
	"matugen-bin",
	"python-build",
	"python-pillow",
	"python-pywal",
	"python-setuptools-scm",
	"python-wheel",
	"swappy",
	"wf-recorder",
	"grim",
	"tesseract",
	"tesseract-data-eng",
	"slurp",
	"dart-sass",
	"python-pywayland",
	"python-psutil",
	"hypridle",
	"hyprutils",
	"hyprlock",
	"wlogout",
	"wl-clipboard",
	"hyprpicker",
	"ghostty",
	"ttf-noto-sans-cjk-vf",
	"noto-fonts-emoji",
	"metar",
	"ttf-material-symbols-variable-git",
}

// ConfigRepo is the URL of the repository containing the configuration files
var ConfigRepo = "https://github.com/Lunaris-Project/HyprLuna.git"

// PackageCategory represents a category of packages
type PackageCategory struct {
	Name        string
	Description string
	Options     []PackageOption
	Required    bool
}

// PackageOption represents a package option
type PackageOption struct {
	Name        string
	Description string
	Packages    []string
	Default     bool
}

// PackageCategories is a list of package categories
var PackageCategories = []PackageCategory{
	{
		Name:        "Terminals",
		Description: "Terminal emulators",
		Options: []PackageOption{
			{
				Name:        "Alacritty",
				Description: "A fast, cross-platform, OpenGL terminal emulator",
				Packages:    []string{"alacritty"},
				Default:     true,
			},
			{
				Name:        "Kitty",
				Description: "A modern, hackable, featureful, OpenGL-based terminal emulator",
				Packages:    []string{"kitty"},
				Default:     false,
			},
			{
				Name:        "Foot",
				Description: "A fast, lightweight and minimalistic Wayland terminal emulator",
				Packages:    []string{"foot"},
				Default:     false,
			},
		},
		Required: false,
	},
	{
		Name:        "Shells",
		Description: "Command-line shells",
		Options: []PackageOption{
			{
				Name:        "Zsh",
				Description: "A powerful shell with many features",
				Packages:    []string{"zsh", "zsh-completions", "zsh-syntax-highlighting", "zsh-autosuggestions"},
				Default:     true,
			},
			{
				Name:        "Fish",
				Description: "A smart and user-friendly command line shell",
				Packages:    []string{"fish"},
				Default:     false,
			},
			{
				Name:        "Bash",
				Description: "The default shell for most Linux distributions",
				Packages:    []string{"bash", "bash-completion"},
				Default:     false,
			},
		},
		Required: false,
	},
	{
		Name:        "Browsers",
		Description: "Web browsers",
		Options: []PackageOption{
			{
				Name:        "Firefox",
				Description: "A free and open-source web browser",
				Packages:    []string{"firefox"},
				Default:     true,
			},
			{
				Name:        "Chromium",
				Description: "An open-source browser project that aims to build a safer, faster, and more stable way for all users to experience the web",
				Packages:    []string{"chromium"},
				Default:     false,
			},
			{
				Name:        "Brave",
				Description: "A free and open-source web browser focused on privacy and speed",
				Packages:    []string{"brave-bin"},
				Default:     false,
			},
		},
		Required: false,
	},
	{
		Name:        "File Managers",
		Description: "File managers",
		Options: []PackageOption{
			{
				Name:        "Thunar",
				Description: "A modern file manager for the Xfce Desktop Environment",
				Packages:    []string{"thunar", "thunar-archive-plugin", "thunar-volman", "tumbler"},
				Default:     true,
			},
			{
				Name:        "Dolphin",
				Description: "The default file manager for the KDE Plasma desktop",
				Packages:    []string{"dolphin"},
				Default:     false,
			},
			{
				Name:        "Nautilus",
				Description: "The default file manager for the GNOME desktop",
				Packages:    []string{"nautilus"},
				Default:     false,
			},
		},
		Required: false,
	},
	{
		Name:        "Text Editors",
		Description: "Text editors",
		Options: []PackageOption{
			{
				Name:        "Neovim",
				Description: "Hyperextensible Vim-based text editor",
				Packages:    []string{"neovim"},
				Default:     true,
			},
			{
				Name:        "Visual Studio Code",
				Description: "Code editing. Redefined.",
				Packages:    []string{"visual-studio-code-bin"},
				Default:     false,
			},
			{
				Name:        "Gedit",
				Description: "A text editor for the GNOME desktop environment",
				Packages:    []string{"gedit"},
				Default:     false,
			},
		},
		Required: false,
	},
	{
		Name:        "Media Players",
		Description: "Media players",
		Options: []PackageOption{
			{
				Name:        "VLC",
				Description: "A free and open source cross-platform multimedia player",
				Packages:    []string{"vlc"},
				Default:     true,
			},
			{
				Name:        "MPV",
				Description: "A free, open source, and cross-platform media player",
				Packages:    []string{"mpv"},
				Default:     false,
			},
			{
				Name:        "Celluloid",
				Description: "A simple GTK+ frontend for mpv",
				Packages:    []string{"celluloid"},
				Default:     false,
			},
		},
		Required: false,
	},
}

// ConfigDirs is a list of configuration directories to copy
var ConfigDirs = []string{
	".config",
	".local",
	".fonts",
	".ags",
	"Pictures",
	".cursor",
	".vscode",
}
