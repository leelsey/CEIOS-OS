package main

import (
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	arm64Path  = "/opt/homebrew/"
	amd64Path  = "/usr/local/"
	brewPrefix = MacPMSPrefix()
	macPMS     = MacPMSPath()
	macGit     = "/usr/bin/git"
	macASDF    = MacASDFPath()
	macAlt     = "--cask"
	macRepo    = "tap"
	fontPath   = HomeDir() + "Library/Fonts/"
	p10kPath   = HomeDir() + ".config/p10k/"
	p10kCache  = HomeDir() + ".cache/p10k-" + WorkingUser()
	macLdBar   = spinner.New(spinner.CharSets[16], 50*time.Millisecond)
)

func MacPMSPrefix() string {
	if CheckArchitecture() == "arm64" {
		return arm64Path
	} else if CheckArchitecture() == "amd64" {
		return amd64Path
	} else {
		MessageError("fatal", "Error:", "Unknown architecture")
		return ""
	}
}

func MacPMSPath() string {
	if CheckArchitecture() == "arm64" {
		return arm64Path + "bin/brew"
	} else if CheckArchitecture() == "amd64" {
		return amd64Path + "bin/brew"
	} else {
		MessageError("fatal", "Error:", "Unknown architecture")
		return ""
	}
}

func MacASDFPath() string {
	asdfPath := "opt/asdf/libexec/bin/asdf"
	if CheckArchitecture() == "arm64" {
		return arm64Path + asdfPath
	} else if CheckArchitecture() == "amd64" {
		return amd64Path + asdfPath
	} else {
		MessageError("fatal", "Error:", "Unknown architecture")
		return ""
	}
}

func MacUpdate() {
	runLdBar.Suffix = " Updating OS, please wait a moment ... "
	runLdBar.Start()

	osUpdate := exec.Command("softwareupdate", "--all", "--install", "--force")
	errOSUpdate := osUpdate.Run()
	CheckError(errOSUpdate, "Failed to update Operating System")

	runLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "update OS!\n"
	runLdBar.Stop()
}

func MacReboot(adminCode string) {
	runLdBar.Suffix = " Restarting OS, please wait a moment ... "
	runLdBar.Start()

	NeedPermission(adminCode)
	reboot := exec.Command(cmdAdmin, "shutdown", "-r", "now")
	time.Sleep(time.Second * 3)
	if err := reboot.Run(); err != nil {
		runLdBar.FinalMSG = clrRed + "Error: " + clrReset
		runLdBar.Stop()
		fmt.Println(errors.New("failed to reboot Operating System"))
	}

	runLdBar.FinalMSG = "⣾ Restarting OS, please wait a moment ... "
	runLdBar.Stop()
}

func OpenMacApplication(appName string) {
	runApp := exec.Command("open", "/Applications/"+appName+".app")
	err := runApp.Run()
	CheckCmdError(err, "ASDF-VM failed to add", appName)
}

func ChangeMacApplicationIcon(appName, icnName, adminCode string) {
	srcIcn := WorkingDir() + ".dev4mac-app-icn.icns"
	DownloadFile(srcIcn, "https://raw.githubusercontent.com/leelsey/ConfStore/main/icns/"+icnName, 0755)

	appSrc := strings.Replace(appName, " ", "\\ ", -1)
	appPath := "/Applications/" + appSrc + ".app"
	chicnPath := WorkingDir() + ".dev4mac-chicn.sh"
	cvtIcn := WorkingDir() + ".dev4mac-app-icn.rsrc"
	chIcnSrc := "sudo rm -rf \"" + appPath + "\"$'/Icon\\r'\n" +
		"sips -i " + srcIcn + " > /dev/null\n" +
		"DeRez -only icns " + srcIcn + " > " + cvtIcn + "\n" +
		"sudo Rez -append " + cvtIcn + " -o " + appPath + "$'/Icon\\r'\n" +
		"sudo SetFile -a C " + appPath + "\n" +
		"sudo SetFile -a V " + appPath + "$'/Icon\\r'"
	MakeFile(chicnPath, chIcnSrc, 0644)

	NeedPermission(adminCode)
	chicn := exec.Command(cmdSh, chicnPath)
	chicn.Env = os.Environ()
	chicn.Stderr = os.Stderr
	err := chicn.Run()
	CheckCmdError(err, "Failed change icon of", appName+".app")

	RemoveFile(srcIcn)
	RemoveFile(cvtIcn)
	RemoveFile(chicnPath)
}

func MacPMSUpdate() {
	updateHomebrew := exec.Command(macPMS, "update", "--auto-update")
	err := updateHomebrew.Run()
	CheckCmdError(err, "Brew failed to", "update repositories")
}

func MacPMSUpgrade() {
	MacPMSUpdate()
	upgradeHomebrew := exec.Command(macPMS, "upgrade", "--greedy")
	err := upgradeHomebrew.Run()
	CheckCmdError(err, "Brew failed to", "upgrade packages")
}

func MacPMSRepository(repo string) {
	brewRepo := strings.Split(repo, "/")
	repoPath := strings.Join(brewRepo[0:1], "") + "/homebrew-" + strings.Join(brewRepo[1:2], "")
	if CheckExists(brewPrefix+"Homebrew/Library/Taps/"+repoPath) != true {
		brewRepo := exec.Command(macPMS, macRepo, repo)
		err := brewRepo.Run()
		CheckCmdError(err, "Brew failed to add ", repo)
	}
}

func MacPMSCleanup() {
	upgradeHomebrew := exec.Command(macPMS, "cleanup", "--prune=all", "-nsd")
	err := upgradeHomebrew.Run()
	CheckCmdError(err, "Brew failed to", "cleanup old packages")
}

func MacPMSRemoveCache() {
	upgradeHomebrew := exec.Command("rm", "-rf", "\"$(brew --cache)\"")
	err := upgradeHomebrew.Run()
	CheckCmdError(err, "Brew failed to", "remove cache")
}

func MacPMSInstall(pkg string) {
	if CheckExists(brewPrefix+"Cellar/"+pkg) != true {
		MacPMSUpdate()
		brewIns := exec.Command(macPMS, optIns, pkg)
		brewIns.Stderr = os.Stderr
		err := brewIns.Run()
		CheckCmdError(err, "Brew failed to install", pkg)
	}
}

func MacPMSInstallQuiet(pkg string) {
	if CheckExists(brewPrefix+"Cellar/"+pkg) != true {
		MacPMSUpdate()
		brewIns := exec.Command(macPMS, optIns, "--quiet", pkg)
		err := brewIns.Run()
		CheckCmdError(err, "Brew failed to install", pkg)
	}
}

func MacPMSInstallCask(pkg, appName string) {
	if CheckExists(brewPrefix+"Caskroom/"+pkg) != true {
		MacPMSUpdate()
		if CheckExists("/Applications/"+appName+".app") != true {
			brewIns := exec.Command(macPMS, optIns, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Brew failed to install cask", pkg)
		} else {
			brewIns := exec.Command(macPMS, optReIn, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Brew failed to reinstall cask", pkg)
		}
	}
}

func MacPMSInstallCaskSudo(pkg, appName, appPath, adminCode string) {
	if CheckExists(brewPrefix+"Caskroom/"+pkg) != true {
		MacPMSUpdate()
		NeedPermission(adminCode)
		if CheckExists(appPath) != true {
			brewIns := exec.Command(macPMS, optIns, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Brew failed to install cask", appName)
		} else {
			brewIns := exec.Command(macPMS, optReIn, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Brew failed to install cask", appName)
		}
	}
}

func MacJavaHome(srcVer, dstVer, adminCode string) {
	if CheckExists(brewPrefix+"Cellar/openjdk"+srcVer) == true {
		LinkFile(brewPrefix+"opt/openjdk"+srcVer+" /libexec/openjdk.jdk", "/Library/Java/JavaVirtualMachines/openjdk"+dstVer+".jdk", "symbolic", "root", adminCode)
	}
}

func MacInstallBrew(adminCode string) {
	insBrewPath := WorkingDir() + ".dev4mac-brew.sh"
	DownloadFile(insBrewPath, "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh", 0755)

	NeedPermission(adminCode)
	installHomebrew := exec.Command(cmdSh, "-c", insBrewPath)
	installHomebrew.Env = append(os.Environ(), "NONINTERACTIVE=1")
	if err := installHomebrew.Run(); err != nil {
		RemoveFile(insBrewPath)
		CheckError(err, "Failed to install homebrew")
	}
	RemoveFile(insBrewPath)

	if CheckExists(macPMS) == false {
		MessageError("fatal", "Installed brew failed, please check your system", "Can't find homebrew")
	}
}

func MacInstallHopper(adminCode string) {
	hopperRSS := strings.Split(NetHTTP("https://www.hopperapp.com/rss/html_changelog.php"), " ")
	hopperVer := strings.Join(hopperRSS[1:2], "")

	dlHopperPath := WorkingDir() + ".Hopper.dmg"
	DownloadFile(dlHopperPath, "https://d2ap6ypl1xbe4k.cloudfront.net/Hopper-"+hopperVer+"-demo.dmg", 0755)

	mountHopper := exec.Command("hdiutil", "attach", dlHopperPath)
	errMount := mountHopper.Run()
	CheckError(errMount, "Failed to mount "+clrYellow+"Hopper.dmg"+clrReset)
	RemoveFile(dlHopperPath)

	appName := "Hopper Disassembler v4"
	CopyDirectory("/Volumes/Hopper Disassembler/"+appName+".app", "/Applications/"+appName+".app")

	unmountDmg := exec.Command("hdiutil", "unmount", "/Volumes/Hopper Disassembler")
	errUnmount := unmountDmg.Run()
	CheckError(errUnmount, "Failed to unmount "+clrYellow+"Hopper Disassembler"+clrReset)

	if CheckArchitecture() == "arm64" {
		ChangeMacApplicationIcon(appName, "Hopper Disassembler ARM64.icns", adminCode)
	} else if CheckArchitecture() == "amd64" {
		ChangeMacApplicationIcon(appName, "Hopper Disassembler AMD64.icns", adminCode)
	}
}

func macBegin(adminCode string) {
	if CheckExists(macPMS) == true {
		macLdBar.Suffix = " Updating homebrew... "
		macLdBar.Start()
		macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "update homebrew!\n"
	} else {
		macLdBar.Suffix = " Installing homebrew... "
		macLdBar.Start()
		macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install and update homebrew!\n"

		MacInstallBrew(adminCode)
	}
	err := os.Chmod(brewPrefix+"share", 0755)
	CheckError(err, "Failed to change permissions on "+brewPrefix+"share to 755")

	MacPMSUpdate()
	MacPMSRepository("homebrew/core")
	MacPMSRepository("homebrew/cask")
	MacPMSRepository("homebrew/cask-versions")
	MacPMSUpgrade()

	macLdBar.Stop()
}

func macEnv() {
	macLdBar.Suffix = " Setting basic environment... "
	macLdBar.Start()

	if CheckExists(prfPath) == true {
		CopyFile(prfPath, HomeDir()+".zprofile.bck")
	}
	if CheckExists(shrcPath) == true {
		CopyFile(shrcPath, HomeDir()+".zshrc.bck")
	}

	profileContents := "#    ___________  _____   ____  ______ _____ _      ______ \n" +
		"#   |___  /  __ \\|  __ \\ / __ \\|  ____|_   _| |    |  ____|\n" +
		"#      / /| |__) | |__) | |  | | |__    | | | |    | |__   \n" +
		"#     / / |  ___/|  _  /| |  | |  __|   | | | |    |  __|  \n" +
		"#    / /__| |    | | \\ \\| |__| | |     _| |_| |____| |____ \n" +
		"#   /_____|_|    |_|  \\_\\\\____/|_|    |_____|______|______|\n#\n" +
		"#  " + WorkingUser() + "’s zsh profile\n\n" +
		"# HOMEBREW\n" +
		"eval \"$(" + macPMS + " shellenv)\"\n\n"
	MakeFile(prfPath, profileContents, 0644)

	shrcContents := "#   ______ _____ _    _ _____   _____\n" +
		"#  |___  // ____| |  | |  __ \\ / ____|\n" +
		"#     / /| (___ | |__| | |__) | |\n" +
		"#    / /  \\___ \\|  __  |  _  /| |\n" +
		"#   / /__ ____) | |  | | | \\ \\| |____\n" +
		"#  /_____|_____/|_|  |_|_|  \\_\\\\_____|\n#\n" +
		"#  " + WorkingUser() + "’s zsh run commands\n\n"
	MakeFile(shrcPath, shrcContents, 0644)

	MakeDirectory(HomeDir() + ".config")
	MakeDirectory(HomeDir() + ".cache")

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "setup zsh environment!\n"
	macLdBar.Stop()
}

func macDependency() {
	macLdBar.Suffix = " Installing dependencies... "
	macLdBar.Start()

	MacPMSInstall("pkg-config")
	MacPMSInstall("ca-certificates")
	MacPMSInstall("ncurses")
	MacPMSInstall("openssl@3")
	MacPMSInstall("openssl@1.1")
	MacPMSInstall("readline")
	MacPMSInstall("autoconf")
	MacPMSInstall("automake")
	MacPMSInstall("mpdecimal")
	MacPMSInstall("utf8proc")
	MacPMSInstall("m4")
	MacPMSInstall("gmp")
	MacPMSInstall("mpfr")
	MacPMSInstall("gettext")
	MacPMSInstall("jpeg-turbo")
	MacPMSInstall("libtool")
	MacPMSInstall("libevent")
	MacPMSInstall("libffi")
	MacPMSInstall("libtiff")
	MacPMSInstall("libvmaf")
	MacPMSInstall("libpng")
	MacPMSInstall("libyaml")
	MacPMSInstall("giflib")
	MacPMSInstall("xz")
	MacPMSInstall("gdbm")
	MacPMSInstall("sqlite")
	MacPMSInstall("lz4")
	MacPMSInstall("zstd")
	MacPMSInstall("hiredis")
	MacPMSInstall("berkeley-db")
	MacPMSInstall("asciidoctor")
	MacPMSInstall("freetype")
	MacPMSInstall("fontconfig")

	MacPMSInstall("pcre")
	MacPMSInstall("pcre2")
	MacPMSInstall("ccache")
	MacPMSInstall("gawk")
	MacPMSInstall("tcl-tk")
	MacPMSInstall("perl")
	MacPMSInstall("ruby")
	MacPMSInstall("python@3.10")
	MacPMSInstall("openjdk")

	MacPMSInstall("ghc")
	MacPMSInstall("krb5")
	MacPMSInstall("libsodium")
	MacPMSInstall("nettle")
	MacPMSInstall("coreutils")
	MacPMSInstall("gnu-getopt")
	MacPMSInstall("ldns")
	MacPMSInstall("isl")
	MacPMSInstall("npth")
	MacPMSInstall("gzip")
	MacPMSInstall("bzip2")
	MacPMSInstall("fop")
	MacPMSInstall("little-cms2")
	MacPMSInstall("imath")
	MacPMSInstall("openldap")
	MacPMSInstall("openexr")
	MacPMSInstall("openjpeg")
	MacPMSInstall("jpeg-xl")
	MacPMSInstall("webp")
	MacPMSInstall("rtmpdump")
	MacPMSInstall("aom")
	MacPMSInstall("screenresolution")
	MacPMSInstall("brotli")
	MacPMSInstall("bison")
	MacPMSInstall("swig")
	MacPMSInstall("re2c")
	MacPMSInstall("icu4c")
	MacPMSInstall("bdw-gc")
	MacPMSInstall("guile")
	MacPMSInstall("wxwidgets")
	MacPMSInstall("sphinx-doc")
	MacPMSInstall("docbook")
	MacPMSInstall("docbook2x")
	MacPMSInstall("docbook-xsl")
	MacPMSInstall("xmlto")
	MacPMSInstall("html-xml-utils")
	MacPMSInstall("shared-mime-info")
	MacPMSInstall("x265")
	MacPMSInstall("oniguruma")
	MacPMSInstall("libgpg-error")
	MacPMSInstall("libgcrypt")
	MacPMSInstall("libunistring")
	MacPMSInstall("libatomic_ops")
	MacPMSInstall("libiconv")
	MacPMSInstall("libmpc")
	MacPMSInstall("libidn")
	MacPMSInstall("libidn2")
	MacPMSInstall("libssh2")
	MacPMSInstall("libnghttp2")
	MacPMSInstall("libxml2")
	MacPMSInstall("libtasn1")
	MacPMSInstall("libxslt")
	MacPMSInstall("libavif")
	MacPMSInstall("libzip")
	MacPMSInstall("libde265")
	MacPMSInstall("libheif")
	MacPMSInstall("libksba")
	MacPMSInstall("libusb")
	MacPMSInstall("liblqr")
	MacPMSInstall("libomp")
	MacPMSInstall("libassuan")
	MacPMSInstall("p11-kit")
	MacPMSInstall("gnutls")
	MacPMSInstall("gd")
	MacPMSInstall("ghostscript")
	MacPMSInstall("imagemagick")
	MacPMSInstall("pinentry")
	MacPMSInstall("gnupg")
	MacPMSInstall("curl")
	MacPMSInstall("wget")
	MacPMSInstall("glib")
	MacPMSInstall("zlib")

	shrcAppend := "# NCURSES\n" +
		"export PATH=\"" + brewPrefix + "opt/ncurses/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/ncurses/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/ncurses/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/ncurses/lib/pkgconfig\"\n\n" +
		"# OPENSSL-3\n" +
		"export PATH=\"" + brewPrefix + "opt/openssl@3/bin:$PATH\"\n" +
		"export LDFLAGS=\"-L" + brewPrefix + "opt/openssl@3/lib\"\n" +
		"export CPPFLAGS=\"-I" + brewPrefix + "opt/openssl@3/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/openssl@3/lib/pkgconfig\"\n\n" +
		"# OPENSSL-1.1\n" +
		"export PATH=\"" + brewPrefix + "opt/openssl@1.1/bin:$PATH\"\n" +
		"export LDFLAGS=\"-L" + brewPrefix + "opt/openssl@1.1/lib\"\n" +
		"export CPPFLAGS=\"-I" + brewPrefix + "opt/openssl@1.1/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/openssl@1.1/lib/pkgconfig\"\n\n" +
		"# KRB5\n" +
		"export PATH=\"" + brewPrefix + "opt/krb5/bin:$PATH\"\n" +
		"export PATH=\"" + brewPrefix + "opt/krb5/sbin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/krb5/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/krb5/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/krb5/lib/pkgconfig\"\n\n" +
		"# COREUTILS\n" +
		"#export PATH=\"" + brewPrefix + "opt/coreutils/libexec/gnubin:$PATH\"\n\n" +
		"# GNU GETOPT\n" +
		"export PATH=\"" + brewPrefix + "opt/gnu-getopt/bin:$PATH\"\n\n" +
		"# TCL-TK\n" +
		"export PATH=\"" + brewPrefix + "opt/tcl-tk/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/tcl-tk/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/tcl-tk/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/tcl-tk/lib/pkgconfig\"\n\n" +
		"# BZIP2\n" +
		"export PATH=\"" + brewPrefix + "opt/bzip2/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/bzip2/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/bzip2/include\"\n\n" +
		"# BISON\n" +
		"export PATH=\"" + brewPrefix + "opt/bison/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/bison/lib\"\n\n" +
		"# ICU4C\n" +
		"export PATH=\"" + brewPrefix + "opt/icu4c/bin:$PATH\"\n" +
		"export PATH=\"" + brewPrefix + "opt/icu4c/sbin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/icu4c/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/icu4c/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/icu4c/lib/pkgconfig\"\n\n" +
		"# DOCBOOK\n" +
		"export XML_CATALOG_FILES=\"" + brewPrefix + "etc/xml/catalog\"\n\n" +
		"# LIBICONV\n" +
		"export PATH=\"" + brewPrefix + "opt/libiconv/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/libiconv/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/libiconv/include\"\n\n" +
		"# LIBXML2\n" +
		"export PATH=\"" + brewPrefix + "opt/libxml2/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/libxml2/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/libxml2/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/libxml2/lib/pkgconfig\"\n\n" +
		"# LIBXSLT\n" +
		"export PATH=\"" + brewPrefix + "opt/libxslt/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/libxslt/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/libxslt/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/libxslt/lib/pkgconfig\"\n\n" +
		"# CURL\n" +
		"export PATH=\"" + brewPrefix + "opt/curl/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/curl/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/curl/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/curl/lib/pkgconfig\"\n\n" +
		"# ZLIB\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/zlib/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/zlib/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/zlib/lib/pkgconfig\"\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install dependencies!\n"
	macLdBar.Stop()
}

func macTerminal() {
	macLdBar.Suffix = " Installing zsh with useful tools... "
	macLdBar.Start()

	ConfigAlias4sh()
	MacPMSInstall("zsh-completions")
	MacPMSInstall("zsh-syntax-highlighting")
	MacPMSInstall("zsh-autosuggestions")
	MacPMSInstall("z")
	MakeFile(HomeDir()+".z", "", 0644)
	MacPMSInstall("tree")
	MacPMSInstall("fzf")
	MacPMSInstall("tmux")
	MacPMSInstall("tmuxinator")
	MacPMSInstall("neofetch")
	dliTerm2Conf := HomeDir() + "Library/Preferences/com.googlecode.iterm2.plist"
	DownloadFile(dliTerm2Conf, "https://raw.githubusercontent.com/leelsey/ConfStore/main/iterm2/iTerm2.plist", 0644)

	MacPMSRepository("romkatv/powerlevel10k")
	MacPMSInstall("romkatv/powerlevel10k/powerlevel10k")
	MakeDirectory(p10kPath)
	MakeDirectory(p10kCache)
	DownloadFile(p10kPath+"p10k-term.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-minimalism.zsh", 0644)

	DownloadFile(p10kPath+"p10k-iterm2.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-atelier.zsh", 0644)
	DownloadFile(p10kPath+"p10k-tmux.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-seeking.zsh", 0644)
	DownloadFile(p10kPath+"p10k-ops.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-operations.zsh", 0644)
	DownloadFile(p10kPath+"p10k-etc.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-engineering.zsh", 0644)
	DownloadFile(fontPath+"MesloLGS NF Bold Italic.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Bold%20Italic.ttf", 0644)
	DownloadFile(fontPath+"MesloLGS NF Bold.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Bold.ttf", 0644)
	DownloadFile(fontPath+"MesloLGS NF Italic.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Italic.ttf", 0644)
	DownloadFile(fontPath+"MesloLGS NF Regular.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Regular.ttf", 0644)

	profileAppend := "# ZSH\n" +
		"export SHELL=zsh\n\n" +
		"# POWERLEVEL10K\n" +
		"source " + brewPrefix + "opt/powerlevel10k/powerlevel10k.zsh-theme\n" +
		"if [[ -r \"${XDG_CACHE_HOME:-" + p10kCache + "}/p10k-instant-prompt-${(%):-%n}.zsh\" ]]; then\n" +
		"  source \"${XDG_CACHE_HOME:-" + p10kCache + "}/p10k-instant-prompt-${(%):-%n}.zsh\"\n" +
		"fi\n" +
		"if [[ -d /Applications/iTerm.app ]]; then\n" +
		"  if [[ $TERM_PROGRAM = \"Apple_Terminal\" ]]; then\n" +
		"    [[ ! -f " + p10kPath + "p10k-term.zsh ]] || source " + p10kPath + "p10k-term.zsh\n" +
		"  elif [[ $TERM_PROGRAM = \"iTerm.app\" ]]; then\n" +
		"    [[ ! -f " + p10kPath + "p10k-iterm2.zsh ]] || source " + p10kPath + "p10k-iterm2.zsh\n" +
		"    echo ''; neofetch --bold off\n" +
		"  elif [[ $TERM_PROGRAM = \"tmux\" ]]; then\n" +
		"    [[ ! -f " + p10kPath + "p10k-tmux.zsh ]] || source " + p10kPath + "p10k-tmux.zsh\n" +
		"    echo ''; neofetch --bold off\n" +
		"  else\n" +
		"    [[ ! -f " + p10kPath + "p10k-etc.zsh ]] || source " + p10kPath + "p10k-etc.zsh\n" +
		"  fi\n" +
		"else\n" +
		"  [[ ! -f " + p10kPath + "p10k-term.zsh ]] || source " + p10kPath + "p10k-term.zsh\n" +
		"fi\n\n" +
		"# ZSH-COMPLETIONS\n" +
		"if type brew &>/dev/null; then\n" +
		"  FPATH=" + brewPrefix + "share/zsh-completions:$FPATH\n" +
		"  autoload -Uz compinit\n" +
		"  compinit\n" +
		"fi\n\n" +
		"# ZSH SYNTAX HIGHLIGHTING\n" +
		"source " + brewPrefix + "share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh\n\n" +
		"# ZSH AUTOSUGGESTIONS\n" +
		"source " + brewPrefix + "share/zsh-autosuggestions/zsh-autosuggestions.zsh\n\n" +
		"# Z\n" +
		"source " + brewPrefix + "etc/profile.d/z.sh\n\n" +
		"# ALIAS4SH\n" +
		"source " + HomeDir() + "/.config/alias4sh/alias4.sh\n\n" +
		"# Edit\n" +
		"export EDITOR=/usr/bin/vi\n" +
		"edit () { $EDITOR \"$@\" }\n" +
		"#vi () { $EDITOR \"$@\" }\n\n"
	AppendFile(prfPath, profileAppend, 0644)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install and configure for terminal!\n"
	macLdBar.Stop()
}

func macLanguage(adminCode string) {
	macLdBar.Suffix = " Installing computer programming language... "
	macLdBar.Start()

	MacPMSInstall("llvm")
	MacPMSInstall("gcc") // fortran
	MacJavaHome("", "", adminCode)
	MacPMSInstall("openjdk@17")
	MacJavaHome("@17", "-17", adminCode)
	MacPMSInstall("openjdk@11")
	MacJavaHome("@11", "-11", adminCode)
	if CheckArchitecture() == "amd64" {
		MacPMSInstall("openjdk@8")
		MacJavaHome("@8", "-8", adminCode)
	}
	//MacPMSInstall("rust")
	//MacPMSInstall("go")
	//MacPMSInstall("lua")
	//MacPMSInstall("php")
	//MacPMSInstall("node")
	//MacPMSInstall("typescript")
	//MacPMSRepository("dart-lang/dart")
	//MacPMSInstall("dart")
	//MacPMSInstall("groovy")
	//MacPMSInstall("kotlin")
	//MacPMSInstall("scala")
	//MacPMSInstall("clojure")
	//MacPMSInstall("erlang")
	//MacPMSInstall("elixir")
	MacPMSInstall("cabal-install")
	MacPMSInstall("haskell-language-server")
	MacPMSInstall("stylish-haskell")

	shrcAppend := "# CCACHE\n" +
		"export PATH=\"" + brewPrefix + "opt/ccache/libexec:$PATH\"\n\n" +
		"# RUBY\n" +
		"export PATH=\"" + brewPrefix + "opt/ruby/bin:$PATH\"\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install languages!\n"
	macLdBar.Stop()
}

func macServer() {
	macLdBar.Suffix = " Installing developing tools for server and database... "
	macLdBar.Start()

	MacPMSInstall("httpd")
	MacPMSInstall("tomcat")
	MacPMSInstall("sqlite-analyzer")
	MacPMSInstall("mysql")
	MacPMSInstall("redis")

	shrcAppend := "# SQLITE3\n" +
		"export PATH=\"" + brewPrefix + "opt/sqlite/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/sqlite/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/sqlite/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/sqlite/lib/pkgconfig\"\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install servers and databases!\n"
	macLdBar.Stop()
}

func macDevVM() {
	macLdBar.Suffix = " Installing developer tools version management tool with plugin... "
	macLdBar.Start()

	MacPMSInstall("asdf")

	shrcAppend := "# ASDF VM\n" +
		"source " + brewPrefix + "opt/asdf/libexec/asdf.sh\n" +
		"export RUBY_CONFIGURE_OPTS=\"--with-openssl-dir=$(brew --prefix openssl@1.1)\"\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	asdfrcContents := "#              _____ _____  ______  __      ____  __ \n" +
		"#       /\\    / ____|  __ \\|  ____| \\ \\    / /  \\/  |\n" +
		"#      /  \\  | (___ | |  | | |__ ____\\ \\  / /| \\  / |\n" +
		"#     / /\\ \\  \\___ \\| |  | |  __|_____\\ \\/ / | |\\/| |\n" +
		"#    / ____ \\ ____) | |__| | |         \\  /  | |  | |\n" +
		"#   /_/    \\_\\_____/|_____/|_|          \\/   |_|  |_|\n#\n" +
		"#  " + WorkingUser() + "’s ASDF-VM run commands\n\n" +
		"legacy_version_file = yes\n" +
		"use_release_candidates = no\n" +
		"always_keep_download = no\n" +
		"plugin_repository_last_check_duration = 0\n" +
		"disable_plugin_short_name_repository = no\n" +
		"java_macos_integration_enable = yes\n"
	MakeFile(HomeDir()+".asdfrc", asdfrcContents, 0644)

	ASDFInstall("perl", "latest")
	ASDFInstall("ruby", "latest")
	ASDFInstall("python", "latest")
	ASDFInstall("java", "openjdk-11.0.2") // JDK LTS 11
	ASDFInstall("java", "openjdk-17.0.2") // JDK LTS 17
	ASDFInstall("rust", "latest")
	ASDFInstall("golang", "latest")
	ASDFInstall("lua", "latest")
	ASDFInstall("nodejs", "latest")
	ASDFInstall("dart", "latest")
	ASDFInstall("php", "latest")
	ASDFInstall("groovy", "latest")
	ASDFInstall("kotlin", "latest")
	ASDFInstall("scala", "latest")
	ASDFInstall("clojure", "latest")
	ASDFInstall("erlang", "latest")
	ASDFInstall("elixir", "latest")
	ASDFInstall("gleam", "latest")
	ASDFInstall("haskell", "latest")
	ASDFReshim()

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install ASDF-VM with languages!\n"
	macLdBar.Stop()
}

func macCLIApp() {
	macLdBar.Suffix = " Installing CLI applications... "
	macLdBar.Start()

	MacPMSInstall("bash")
	MacPMSInstall("zsh")
	MacPMSInstall("openssh")
	MacPMSInstall("mosh")
	MacPMSInstall("git")
	MacPMSInstall("inetutils")
	MacPMSInstall("openvpn")
	MacPMSInstall("wireguard-go")
	MacPMSInstall("wireguard-tools")
	MacPMSInstall("tor")
	MacPMSInstall("torsocks")

	MacPMSInstall("asciinema")
	MacPMSInstall("unzip")
	MacPMSInstall("diffutils")
	MacPMSInstall("transmission-cli")
	MacPMSInstall("exa")
	MacPMSInstall("bat")
	MacPMSInstall("diffr")
	MacPMSInstall("tldr")

	MacPMSInstall("git-lfs")
	MacPMSInstall("gh")
	MacPMSInstall("hub")
	MacPMSInstall("tig")
	MacPMSInstall("direnv")
	MacPMSInstall("watchman")

	MacPMSInstall("make")
	MacPMSInstallQuiet("cmake")
	MacPMSInstall("ninja")
	MacPMSInstall("maven")
	MacPMSInstall("gradle")
	MacPMSInstall("rustup-init")
	MacPMSInstall("htop")
	MacPMSInstall("qemu")
	MacPMSInstall("vim")
	MacPMSInstall("neovim")
	MacPMSInstall("curlie")
	MacPMSInstall("jq")
	MacPMSInstall("yq")
	MacPMSInstall("dasel")

	MacPMSInstall("nmap")
	MacPMSInstall("radare2")
	MacPMSInstall("sleuthkit")
	MacPMSInstall("autopsy")
	MacPMSInstall("virustotal-cli")

	shrcAppend := "# DIRENV\n" +
		"eval \"$(direnv hook zsh)\"\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install CLI applications!\n"
	macLdBar.Stop()
}

func macGUIApp(adminCode string) {
	macLdBar.Suffix = " Installing GUI applications... "
	macLdBar.Start()

	MacPMSInstallCask("sensei", "Sensei")
	MacPMSInstallCask("keka", "Keka")
	MacPMSInstallCask("iina", "IINA")
	MacPMSInstallCask("transmission", "Transmission")
	ChangeMacApplicationIcon("Transmission", "Transmission.icns", adminCode)
	MacPMSInstallCask("rectangle", "Rectangle")
	MacPMSInstallCask("google-chrome", "Google Chrome")
	MacPMSInstallCask("firefox", "Firefox")
	ChangeMacApplicationIcon("Firefox", "Firefox.icns", adminCode)
	MacPMSInstallCask("tor-browser", "Tor Browser")
	ChangeMacApplicationIcon("Tor Browser", "Tor Browser.icns", adminCode)
	//MacPMSInstallCask("chromium", "Chromium") TODO: Will add Grimoire (LE Chromium)
	MacPMSInstallCask("spotify", "Spotify")
	ChangeMacApplicationIcon("Spotify", "Spotify.icns", adminCode)
	MacPMSInstallCask("signal", "Signal")
	MacPMSInstallCask("discord", "Discord")
	MacPMSInstallCask("slack", "Slack")
	MacPMSInstallCask("jetbrains-space", "JetBrains Space")
	ChangeMacApplicationIcon("JetBrains Space", "JetBrains Space.icns", adminCode)

	MacPMSInstallCask("dropbox", "Dropbox")
	MacPMSInstallCask("dropbox-capture", "Dropbox Capture")
	MacPMSInstallCask("sketch", "Sketch")
	MacPMSInstallCask("zeplin", "Zeplin")
	MacPMSInstallCask("blender", "Blender")
	ChangeMacApplicationIcon("Blender", "Blender.icns", adminCode)
	MacPMSInstallCask("obs", "OBS")
	MacPMSInstallCaskSudo("loopback", "Loopback", "/Applications/Loopback.app", adminCode)
	OpenMacApplication("Loopback")

	MacPMSInstallCask("iterm2", "iTerm")
	MacPMSInstallCask("intellij-idea", "IntelliJ IDEA")
	ChangeMacApplicationIcon("IntelliJ IDEA", "IntelliJ IDEA.icns", adminCode)
	MacPMSInstallCask("visual-studio-code", "Visual Studio Code")
	MacPMSInstallCask("atom", "Atom")
	MacPMSInstallCask("neovide", "Neovide")
	ChangeMacApplicationIcon("Neovide", "Neovide.icns", adminCode)
	MacPMSInstallCaskSudo("vmware-fusion", "VMware Fusion", "/Applications/VMware Fusion.app", adminCode)
	ChangeMacApplicationIcon("VMware Fusion", "VMware Fusion.icns", adminCode)
	MacPMSInstallCask("docker", "Docker")
	MacPMSInstallCask("github", "Github")
	MacPMSInstallCask("fork", "Fork")
	MacPMSInstallCask("tableplus", "TablePlus")
	MacPMSInstallCask("proxyman", "Proxyman")
	MacPMSInstallCask("postman", "Postman")
	MacPMSInstallCask("paw", "Paw")
	MacPMSInstallCask("boop", "Boop")
	MacPMSInstallCask("httpie", "HTTPie")
	MacPMSInstallCask("vnc-viewer", "VNC Viewer")
	ChangeMacApplicationIcon("VNC Viewer", "VNC Viewer.icns", adminCode)
	MacPMSInstallCask("forklift", "ForkLift")
	MacPMSInstallCask("drawio", "draw.io")
	MacPMSInstallCask("staruml", "StarUML")
	ChangeMacApplicationIcon("StarUML", "StarUML.icns", adminCode)

	MacPMSInstallCaskSudo("codeql", "CodeQL", brewPrefix+"Caskroom/Codeql", adminCode)
	MacPMSInstallCask("little-snitch", "Little Snitch")
	ChangeMacApplicationIcon("Little Snitch", "Little Snitch.icns", adminCode)
	MacPMSInstallCask("micro-snitch", "Micro Snitch")
	ChangeMacApplicationIcon("Micro Snitch", "Micro Snitch.icns", adminCode)
	MacPMSInstallCask("burp-suite", "Burp Suite Community Edition")
	MacPMSInstallCask("burp-suite-professional", "Burp Suite Professional")
	MacPMSInstallCask("owasp-zap", "OWASP ZAP")
	ChangeMacApplicationIcon("OWASP ZAP", "OWASP ZAP.icns", adminCode)
	MacPMSInstallCaskSudo("wireshark", "Wireshark", "/Applications/Wireshark.app", adminCode)
	ChangeMacApplicationIcon("Wireshark", "Wireshark.icns", adminCode)
	MacPMSInstallCaskSudo("zenmap", "Zenmap", "/Applications/Zenmap.app", adminCode)
	ChangeMacApplicationIcon("Zenmap", "Zenmap.icns", adminCode)
	MacInstallHopper(adminCode)
	MacPMSInstallCask("cutter", "Cutter")
	// Install Ghidra // TODO: will add
	MacPMSInstallCask("imazing", "iMazing")
	ChangeMacApplicationIcon("iMazing", "iMazing.icns", adminCode)
	MacPMSInstallCask("apparency", "Apparency")
	MacPMSInstallCask("suspicious-package", "Suspicious Package")
	MacPMSInstallCask("fsmonitor", "FSMonitor")
	ChangeMacApplicationIcon("FSMonitor", "FSMonitor.icns", adminCode)

	shrcAppend := "# ANDROID STUDIO\n" +
		"export ANDROID_HOME=$HOME/Library/Android/sdk\n" +
		"export PATH=$PATH:$ANDROID_HOME/emulator\n" +
		"export PATH=$PATH:$ANDROID_HOME/tools\n" +
		"export PATH=$PATH:$ANDROID_HOME/tools/bin\n" +
		"export PATH=$PATH:$ANDROID_HOME/platform-tools\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install GUI applications!\n"
	macLdBar.Stop()
}

func macEnd() {
	macLdBar.Suffix = " Finishing... "
	macLdBar.Start()

	shrcAppend := "\n######## ADD CUSTOM VALUES UNDER HERE ########\n\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	MacPMSUpgrade()
	MacPMSCleanup()
	MacPMSRemoveCache()

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "clean up homebrew's cache!\n"
	macLdBar.Stop()
}

func CEIOSmacOS(adminCode string) {
	var brewSts string

	if CheckExists(macPMS) == true {
		brewSts = "Update"
	} else {
		brewSts = "Install"
	}

	fmt.Println(clrCyan + "CEIOS The Development tools of Essential and Various for macOS\n" + clrReset)

	runEgMsg := lstDot + "Run " + clrPurple + "" + clrReset + " installation\n" + lstDot + brewSts + " homebrew with configure shell"
	fmt.Println(runEgMsg + ", then install Dependencies, Languages, Server, Database, management DevTools and Terminal/CLI/GUI applications with set basic preferences.")

	alMsg := lstDot + "Use root permission to install "
	if brewSts == "Install" {
		alMsg = alMsg + "homebrew " + clrReset + "and " + clrPurple + "applications" + clrReset + ": "
	} else if brewSts == "Update" {
		alMsg = alMsg + "applications" + clrReset + ": "
	}
	fmt.Println(alMsg + "Java, Loopback, VMware Fusion, Wireshark and Zenmap")

	macBegin(adminCode)
	macEnv()
	macDependency()
	macTerminal()
	macLanguage(adminCode)
	macServer()
	macDevVM()
	macCLIApp()
	macGUIApp(adminCode)
	macEnd()

	var g4sOpt string
	fmt.Print(clrCyan + "\nConfigure git global easily\n" + clrReset + "To continue we setup git global configuration.\nIf you wish to continue type (Y) then press return: ")
	_, errG4sOpt := fmt.Scanln(&g4sOpt)
	if errG4sOpt != nil {
		g4sOpt = "Enter"
	}
	if g4sOpt == "y" || g4sOpt == "Y" || g4sOpt == "yes" || g4sOpt == "Yes" || g4sOpt == "YES" {
		ClearLine(3)
		ConfigGit4sh()
	} else {
		ClearLine(4)
	}

	fmt.Print(clrCyan + "\nFinishing CEIOS-OS installation\n" + clrReset)
	MacUpdate()
	MacReboot(adminCode)
}
