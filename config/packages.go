package config

// AURHelpers defines the available AUR helpers
var AURHelpers = []string{"yay", "paru"}

// PackageCategory represents a category of packages with options
type PackageCategory struct {
	Name        string
	Description string
	Options     []PackageOption
	Required    bool
}

// PackageOption represents a selectable package option within a category
type PackageOption struct {
	Name        string
	Description string
	Packages    []string
	Default     bool
}

// AURPackages is the list of AUR packages to install
var AURPackages = []string{
	"axel", "bc", "coreutils", "cliphist", "cmake", "curl", "rofi-wayland",
	"rsync", "wget", "ripgrep", "jq", "npm", "meson", "typescript", "gjs",
	"xdg-user-dirs", "brightnessctl", "ddcutil", "pavucontrol", "wireplumber",
	"libdbusmenu-gtk3", "playerctl", "swww", "git", "gobject-introspection",
	"glib2-devel", "gvfs", "glib2", "glibc", "gtk3", "gtk-layer-shell",
	"libpulse", "pam", "gnome-bluetooth-3.0", "gammastep", "libsoup3",
	"libnotify", "networkmanager", "power-profiles-daemon", "upower",
	"adw-gtk-theme-git", "qt5ct", "qt5-wayland", "fontconfig",
	"ttf-readex-pro", "ttf-jetbrains-mono-nerd", "ttf-material-symbols-variable-git",
	"ttf-space-mono-nerd", "ttf-rubik-vf", "ttf-gabarito-git", "fish", "foot",
	"starship", "polkit-gnome", "gnome-keyring", "gnome-control-center",
	"blueberry", "webp-pixbuf-loader", "gtksourceview3", "yad", "ydotool",
	"xdg-user-dirs-gtk", "tinyxml2", "gtkmm3", "gtksourceviewmm", "cairomm",
	"xdg-desktop-portal", "xdg-desktop-portal-gtk", "xdg-desktop-portal-hyprland",
	"gradience", "python-libsass", "python-pywalfox", "matugen-bin",
	"python-build", "python-pillow", "python-pywal", "python-setuptools-scm",
	"python-wheel", "swappy", "wf-recorder", "grim", "tesseract",
	"tesseract-data-eng", "slurp", "dart-sass", "python-pywayland",
	"python-psutil", "hypridle", "hyprutils", "hyprlock", "wlogout",
	"wl-clipboard", "hyprpicker",
}

// PackageCategories defines the categories of packages with options
var PackageCategories = []PackageCategory{
	{
		Name:        "Terminals",
		Description: "Terminal emulators",
		Required:    false,
		Options: []PackageOption{
			{
				Name:        "Foot",
				Description: "A fast, lightweight and minimalistic Wayland terminal emulator",
				Packages:    []string{"foot"},
				Default:     true,
			},
			{
				Name:        "Kitty",
				Description: "A fast, feature-rich, GPU based terminal emulator",
				Packages:    []string{"kitty"},
				Default:     false,
			},
			{
				Name:        "Alacritty",
				Description: "A cross-platform, GPU-accelerated terminal emulator",
				Packages:    []string{"alacritty"},
				Default:     false,
			},
		},
	},
	{
		Name:        "Shells",
		Description: "Command line shells",
		Required:    false,
		Options: []PackageOption{
			{
				Name:        "Fish",
				Description: "The friendly interactive shell",
				Packages:    []string{"fish"},
				Default:     true,
			},
			{
				Name:        "Zsh",
				Description: "A shell designed for interactive use",
				Packages:    []string{"zsh", "zsh-completions", "zsh-syntax-highlighting", "zsh-autosuggestions"},
				Default:     false,
			},
			{
				Name:        "Bash",
				Description: "The Bourne Again SHell",
				Packages:    []string{"bash", "bash-completion"},
				Default:     false,
			},
		},
	},
	{
		Name:        "Browsers",
		Description: "Web browsers",
		Required:    false,
		Options: []PackageOption{
			{
				Name:        "Firefox",
				Description: "Mozilla Firefox web browser",
				Packages:    []string{"firefox"},
				Default:     true,
			},
			{
				Name:        "Chromium",
				Description: "Open-source web browser from Google",
				Packages:    []string{"chromium"},
				Default:     false,
			},
			{
				Name:        "Brave",
				Description: "Privacy-focused web browser",
				Packages:    []string{"brave-bin"},
				Default:     false,
			},
		},
	},
	{
		Name:        "File Managers",
		Description: "File managers",
		Required:    false,
		Options: []PackageOption{
			{
				Name:        "Thunar",
				Description: "Modern file manager for the Xfce Desktop",
				Packages:    []string{"thunar", "thunar-archive-plugin", "thunar-volman"},
				Default:     true,
			},
			{
				Name:        "Nautilus",
				Description: "GNOME file manager",
				Packages:    []string{"nautilus"},
				Default:     false,
			},
			{
				Name:        "PCManFM",
				Description: "Extremely fast and lightweight file manager",
				Packages:    []string{"pcmanfm-gtk3"},
				Default:     false,
			},
		},
	},
	{
		Name:        "Text Editors",
		Description: "Text editors",
		Required:    false,
		Options: []PackageOption{
			{
				Name:        "Visual Studio Code",
				Description: "Code editing. Redefined.",
				Packages:    []string{"visual-studio-code-bin"},
				Default:     true,
			},
			{
				Name:        "Neovim",
				Description: "Hyperextensible Vim-based text editor",
				Packages:    []string{"neovim", "python-pynvim"},
				Default:     false,
			},
		},
	},
	{
		Name:        "Media Players",
		Description: "Media players",
		Required:    false,
		Options: []PackageOption{
			{
				Name:        "VLC",
				Description: "Multi-platform multimedia player",
				Packages:    []string{"vlc"},
				Default:     true,
			},
			{
				Name:        "MPV",
				Description: "Free, open source, and cross-platform media player",
				Packages:    []string{"mpv"},
				Default:     false,
			},
		},
	},
}

// ConfigDirs defines the directories to copy from the repository to the user's home directory
var ConfigDirs = []string{
	".config",
	".local",
	".fonts",
}
