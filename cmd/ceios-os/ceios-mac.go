package main

import (
	"bufio"
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

func MacOSUpdate() {
	runLdBar.Suffix = " Updating OS, please wait a moment ... "
	runLdBar.Start()

	osUpdate := exec.Command("softwareupdate", "--all", "--install", "--force")
	errOSUpdate := osUpdate.Run()
	CheckError(errOSUpdate, "Failed to update Operating System")

	runLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "update OS!\n"
	runLdBar.Stop()
}

func MacSoftware() string {
	macSoft := exec.Command("system_profiler", "SPSoftwareDataType")
	softInfo, err := macSoft.Output()
	CheckError(err, "Failed to get macOS hardware information")
	return string(softInfo)
}

func MacHardware() string {
	macHard := exec.Command("system_profiler", "SPHardwareDataType")
	hardInfo, err := macHard.Output()
	CheckError(err, "Failed to get macOS hardware information")
	return string(hardInfo)
}

func MacInformatrion() (string, string) {
	softInfo := strings.Split(MacSoftware(), "\n")
	hardInfo := strings.Split(MacHardware(), "\n")

	osInfo := strings.Split(strings.Join(softInfo[4:5], ""), ": ")
	kernelInfo := strings.Split(strings.Join(softInfo[5:6], ""), ": ")
	deviceInfo := strings.Split(strings.Join(softInfo[8:9], ""), ": ")
	userInfo := strings.Split(strings.Join(softInfo[9:10], ""), ": ")
	modelInfo := strings.Split(strings.Join(hardInfo[5:6], ""), ": ")
	processorInfo := strings.Split(strings.Join(hardInfo[6:7], ""), ": ")
	clockInfo := strings.Split(strings.Join(hardInfo[7:8], ""), ": ")
	memoryInfo := strings.Split(strings.Join(hardInfo[13:14], ""), ": ")

	system := strings.Join(osInfo[1:2], "")
	kernel := strings.Join(kernelInfo[1:2], "")
	device := strings.Join(deviceInfo[1:2], "")
	user := strings.Join(userInfo[1:2], "")
	userfullname := strings.Split(user, " (")
	fullname := strings.Join(userfullname[0:1], "")
	modelver := strings.Split(strings.Join(modelInfo[1:2], ""), ",")
	model := strings.Join(modelver[0:1], "")
	processor := strings.Join(processorInfo[1:2], "")
	clock := strings.Join(clockInfo[1:2], "")
	memory := strings.Join(memoryInfo[1:2], "")

	macInfo := lstDot + clrGreen + "User name" + clrReset + ": " + user + "\n" +
		lstDot + clrGreen + "Device name" + clrReset + ": " + device + "\n" +
		lstDot + clrGreen + "System" + clrReset + ": " + system + " - " + kernel + "\n" +
		//lstDot + clrGreen + "Kernel" + clrReset + ": " + kernel + "\n" +
		lstDot + clrGreen + "Model" + clrReset + ": " + model + "\n" +
		lstDot + clrGreen + "Processor" + clrReset + ": " + processor + " (" + clock + ")\n" +
		lstDot + clrGreen + "Memory" + clrReset + ": " + memory

	return macInfo, fullname
}

func OpenMacApplication(appName string) {
	runApp := exec.Command("open", "/Applications/"+appName+".app")
	err := runApp.Run()
	CheckCmdError(err, "ASDF-VM failed to add", appName)
}

func ChangeMacApplicationIcon(appName, icnName, adminCode string) {
	srcIcn := WorkingDirectory() + ".ceios-app-icn.icns"
	DownloadFile(srcIcn, "https://raw.githubusercontent.com/leelsey/ConfStore/main/icns/"+icnName, 0755)

	appSrc := strings.Replace(appName, " ", "\\ ", -1)
	appPath := "/Applications/" + appSrc + ".app"
	chicnPath := WorkingDirectory() + ".ceios-chicn.sh"
	cvtIcn := WorkingDirectory() + ".ceios-app-icn.rsrc"
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

func ChangeMacWallpaper() {
	srcWp := WorkingDirectory() + ".ceios-wallpaper.png"
	DownloadFile(srcWp, "https://raw.githubusercontent.com/leelsey/CEIOS/main/pictures/wallpaper/desktop.jpeg", 0755)

	chWpPath := WorkingDirectory() + ".ceios-chwap.sh"
	chWpSrc := "osascript -e 'tell application \"Finder\" to set desktop picture to POSIX file \"" + srcWp + "\"'"
	MakeFile(chWpPath, chWpSrc, 0644)

	chWp := exec.Command(cmdSh, chWpPath)
	if err := chWp.Run(); err != nil {
		RemoveFile(srcWp)
		CheckCmdError(err, "Failed change", "desktop background")
	}
	RemoveFile(srcWp)
	RemoveFile(chWpPath)
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
	insBrewPath := WorkingDirectory() + ".ceios-brew.sh"
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

	dlHopperPath := WorkingDirectory() + ".Hopper.dmg"
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
	macLdBar.Suffix = " Setting system environment... "
	macLdBar.Start()

	if CheckExists(prfPath) == true {
		CopyFile(prfPath, HomeDirectory()+".zprofile.bck")
	}
	if CheckExists(shrcPath) == true {
		CopyFile(shrcPath, HomeDirectory()+".zshrc.bck")
	}

	profileContents := "#    ___________  _____   ____  ______ _____ _      ______ \n" +
		"#   |___  /  __ \\|  __ \\ / __ \\|  ____|_   _| |    |  ____|\n" +
		"#      / /| |__) | |__) | |  | | |__    | | | |    | |__   \n" +
		"#     / / |  ___/|  _  /| |  | |  __|   | | | |    |  __|  \n" +
		"#    / /__| |    | | \\ \\| |__| | |     _| |_| |____| |____ \n" +
		"#   /_____|_|    |_|  \\_\\\\____/|_|    |_____|______|______|\n#\n" +
		"#  " + CurrentUsername() + "’s zsh profile\n\n" +
		"# HOMEBREW\n" +
		"eval \"$(" + macPMS + " shellenv)\"\n\n"
	MakeFile(prfPath, profileContents, 0644)

	shrcContents := "#   ______ _____ _    _ _____   _____\n" +
		"#  |___  // ____| |  | |  __ \\ / ____|\n" +
		"#     / /| (___ | |__| | |__) | |\n" +
		"#    / /  \\___ \\|  __  |  _  /| |\n" +
		"#   / /__ ____) | |  | | | \\ \\| |____\n" +
		"#  /_____|_____/|_|  |_|_|  \\_\\\\_____|\n#\n" +
		"#  " + CurrentUsername() + "’s zsh run commands\n\n"
	MakeFile(shrcPath, shrcContents, 0644)

	MakeDirectory(HomeDirectory() + ".config")
	MakeDirectory(HomeDirectory() + ".cache")

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "setup zsh environment!\n"
	macLdBar.Stop()
}

func macDependency(adminCode string) {
	macLdBar.Suffix = " Installing dependencies... "
	macLdBar.Start()

	MacPMSInstall("pkg-config")
	MacPMSInstall("readline")
	MacPMSInstall("autoconf")
	MacPMSInstall("automake")
	MacPMSInstall("ncurses")
	MacPMSInstall("ca-certificates")
	MacPMSInstall("openssl@3")
	MacPMSInstall("openssl@1.1")
	MacPMSInstall("krb5")
	MacPMSInstall("gmp")
	MacPMSInstall("coreutils")
	MacPMSInstall("gnupg")
	MacPMSInstall("gnu-getopt")

	MacPMSInstall("xz")
	MacPMSInstall("oniguruma")
	MacPMSInstall("wxwidgets")
	MacPMSInstall("swig")
	MacPMSInstall("bison")
	MacPMSInstall("icu4c")
	MacPMSInstall("bzip2")
	MacPMSInstall("re2c")
	MacPMSInstall("fop")
	MacPMSInstall("gd")
	MacPMSInstall("imagemagick")
	MacPMSInstall("glib")
	MacPMSInstall("zlib")
	MacPMSInstall("libgpg-error")
	MacPMSInstall("libgcrypt")
	MacPMSInstall("libsodium")
	MacPMSInstall("libiconv")
	MacPMSInstall("libyaml")
	MacPMSInstall("libxslt")
	MacPMSInstall("libzip")

	MacPMSInstall("sqlite")
	MacPMSInstall("sqlite-analyzer")
	MacPMSInstall("pcre")
	MacPMSInstall("pcre2")
	MacPMSInstall("ccache")
	MacPMSInstall("gawk")
	MacPMSInstall("tcl-tk")
	MacPMSInstall("ruby")
	MacPMSInstall("python@3.10")
	MacPMSInstall("llvm")
	MacPMSInstall("gcc")
	MacPMSInstall("openjdk")
	MacJavaHome("", "", adminCode)
	MacPMSInstall("openjdk@17")
	MacJavaHome("@17", "-17", adminCode)
	MacPMSInstall("openjdk@11")
	MacJavaHome("@11", "-11", adminCode)
	if CheckArchitecture() == "amd64" {
		MacPMSInstall("openjdk@8")
		MacJavaHome("@8", "-8", adminCode)
	}
	MacPMSInstall("php")
	MacPMSInstall("ghc")
	MacPMSInstall("cabal-install")
	MacPMSInstall("haskell-language-server")
	MacPMSInstall("stylish-haskell")

	MacPMSInstall("httpd")
	MacPMSInstall("tomcat")
	MacPMSInstall("mysql")
	MacPMSInstall("redis")

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install dependencies!\n"
	macLdBar.Stop()
}

func macUtility(adminCode string) {
	macLdBar.Suffix = " Installing - applications... "
	macLdBar.Start()

	ConfigAlias4sh()
	MacPMSInstall("bash")
	MacPMSInstall("zsh")
	MacPMSInstall("openssh")
	MacPMSInstall("mosh")
	MacPMSInstall("tmux")
	MacPMSInstall("tmuxinator")
	MacPMSInstall("wget")
	MacPMSInstall("curl")
	MacPMSInstall("inetutils")
	MacPMSInstall("unzip")
	MacPMSInstall("sevenzip")
	MacPMSInstall("p7zip")
	MacPMSInstall("vim")
	MacPMSInstall("neovim")
	MacPMSInstall("zsh-completions")
	MacPMSInstall("zsh-syntax-highlighting")
	MacPMSInstall("zsh-autosuggestions")
	MacPMSInstall("z")
	MakeFile(HomeDirectory()+".z", "", 0644)
	MacPMSInstall("fzf")
	MacPMSInstall("exa")
	MacPMSInstall("bat")
	MacPMSInstall("tree")
	MacPMSInstall("diffutils")
	MacPMSInstall("diffr")
	MacPMSInstall("tldr")
	MacPMSInstall("htop")
	MacPMSInstall("btop")
	MacPMSInstall("iperf3")
	MacPMSInstall("neofetch")
	MacPMSInstall("asciinema")
	MacPMSInstall("transmission-cli")

	MacPMSRepository("romkatv/powerlevel10k")
	MacPMSInstall("romkatv/powerlevel10k/powerlevel10k")
	p10kConfPath := HomeDirectory() + ".config/p10k/"
	p10kCachePath := HomeDirectory() + ".cache/p10k-" + CurrentUsername()
	fontLibPath := HomeDirectory() + "Library/Fonts/"
	MakeDirectory(p10kConfPath)
	MakeDirectory(p10kCachePath)
	DownloadFile(p10kConfPath+"p10k-term.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-minimalism.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-iterm2.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-atelier.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-tmux.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-seeking.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-ops.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-operations.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-etc.zsh", "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-engineering.zsh", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Bold Italic.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Bold%20Italic.ttf", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Bold.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Bold.ttf", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Italic.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Italic.ttf", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Regular.ttf", "https://raw.githubusercontent.com/romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/MesloLGS%20NF%20Regular.ttf", 0644)
	DownloadFile(HomeDirectory()+"Library/Preferences/com.googlecode.iterm2.plist", "https://raw.githubusercontent.com/leelsey/ConfStore/main/iterm2/iTerm2.plist", 0644)

	profileAppend := "# ZSH\n" +
		"export SHELL=zsh\n\n" +
		"# POWERLEVEL10K\n" +
		"source " + brewPrefix + "opt/powerlevel10k/powerlevel10k.zsh-theme\n" +
		"if [[ -r \"${XDG_CACHE_HOME:-" + p10kCachePath + "}/p10k-instant-prompt-${(%):-%n}.zsh\" ]]; then\n" +
		"  source \"${XDG_CACHE_HOME:-" + p10kCachePath + "}/p10k-instant-prompt-${(%):-%n}.zsh\"\n" +
		"fi\n" +
		"if [[ -d /Applications/iTerm.app ]]; then\n" +
		"  if [[ $TERM_PROGRAM = \"Apple_Terminal\" ]]; then\n" +
		"    [[ ! -f " + p10kConfPath + "p10k-term.zsh ]] || source " + p10kConfPath + "p10k-term.zsh\n" +
		"  elif [[ $TERM_PROGRAM = \"iTerm.app\" ]]; then\n" +
		"    [[ ! -f " + p10kConfPath + "p10k-iterm2.zsh ]] || source " + p10kConfPath + "p10k-iterm2.zsh\n" +
		"    echo ''; neofetch --bold off\n" +
		"  elif [[ $TERM_PROGRAM = \"tmux\" ]]; then\n" +
		"    [[ ! -f " + p10kConfPath + "p10k-tmux.zsh ]] || source " + p10kConfPath + "p10k-tmux.zsh\n" +
		"    echo ''; neofetch --bold off\n" +
		"  else\n" +
		"    [[ ! -f " + p10kConfPath + "p10k-etc.zsh ]] || source " + p10kConfPath + "p10k-etc.zsh\n" +
		"  fi\n" +
		"else\n" +
		"  [[ ! -f " + p10kConfPath + "p10k-term.zsh ]] || source " + p10kConfPath + "p10k-term.zsh\n" +
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
		"source " + HomeDirectory() + "/.config/alias4sh/alias4.sh\n\n" +
		"# Edit\n" +
		"export EDITOR=/usr/bin/vi\n" +
		"edit () { $EDITOR \"$@\" }\n" +
		"#vi () { $EDITOR \"$@\" }\n\n"
	AppendFile(prfPath, profileAppend, 0644)

	MacPMSInstallCask("iina", "IINA")
	MacPMSInstallCask("sensei", "Sensei")
	MacPMSInstallCask("rectangle", "Rectangle")
	MacPMSInstallCask("dropbox", "Dropbox")
	MacPMSInstallCask("dropbox-capture", "Dropbox Capture")
	MacPMSInstallCask("keka", "Keka")
	MacPMSInstallCask("transmission", "Transmission")
	ChangeMacApplicationIcon("Transmission", "Transmission.icns", adminCode)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install CLI applications!\n"
	macLdBar.Stop()
}

func macProductivity(adminCode string) {
	macLdBar.Suffix = " Installing - applications... "
	macLdBar.Start()

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

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install CLI applications!\n"
	macLdBar.Stop()
}

func macCreativity(adminCode string) {
	macLdBar.Suffix = " Installing - applications... "
	macLdBar.Start()

	MacPMSInstall("bash")
	MacPMSInstall("zsh")
	MacPMSInstall("openssh")
	MacPMSInstall("mosh")
	MacPMSInstall("wget")
	MacPMSInstall("curl")
	MacPMSInstall("git")
	MacPMSInstall("inetutils")
	MacPMSInstall("openvpn")
	MacPMSInstall("wireguard-go")
	MacPMSInstall("wireguard-tools")
	MacPMSInstall("tor")
	MacPMSInstall("torsocks")

	MacPMSInstallCask("sketch", "Sketch")
	MacPMSInstallCask("zeplin", "Zeplin")
	MacPMSInstallCask("blender", "Blender")
	ChangeMacApplicationIcon("Blender", "Blender.icns", adminCode)
	MacPMSInstallCaskSudo("loopback", "Loopback", "/Applications/Loopback.app", adminCode)
	OpenMacApplication("Loopback")
	MacPMSInstallCask("obs", "OBS")

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install CLI applications!\n"
	macLdBar.Stop()
}

func macDevelopment(adminCode string) {
	macLdBar.Suffix = " Installing - applications... "
	macLdBar.Start()

	MacPMSInstall("make")
	MacPMSInstallQuiet("cmake")
	MacPMSInstall("ninja")
	MacPMSInstall("maven")
	MacPMSInstall("gradle")
	MacPMSInstall("rustup-init")
	MacPMSInstall("opencv")
	MacPMSInstall("git")
	MacPMSInstall("git-lfs")
	MacPMSInstall("gh")
	MacPMSInstall("hub")
	MacPMSInstall("tig")
	MacPMSInstall("qemu")
	MacPMSInstall("curlie")
	MacPMSInstall("jq")
	MacPMSInstall("yq")
	MacPMSInstall("dasel")
	MacPMSInstall("watchman")
	MacPMSInstall("direnv")
	MacPMSInstall("asdf")

	shrcAppend := "# DIRENV\n" +
		"eval \"$(direnv hook zsh)\"\n\n" +
		"# ASDF VM\n" +
		"source " + brewPrefix + "opt/asdf/libexec/asdf.sh\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	asdfrcContents := "#              _____ _____  ______  __      ____  __ \n" +
		"#       /\\    / ____|  __ \\|  ____| \\ \\    / /  \\/  |\n" +
		"#      /  \\  | (___ | |  | | |__ ____\\ \\  / /| \\  / |\n" +
		"#     / /\\ \\  \\___ \\| |  | |  __|_____\\ \\/ / | |\\/| |\n" +
		"#    / ____ \\ ____) | |__| | |         \\  /  | |  | |\n" +
		"#   /_/    \\_\\_____/|_____/|_|          \\/   |_|  |_|\n#\n" +
		"#  " + CurrentUsername() + "’s ASDF-VM run commands\n\n" +
		"legacy_version_file = yes\n" +
		"use_release_candidates = no\n" +
		"always_keep_download = no\n" +
		"plugin_repository_last_check_duration = 0\n" +
		"disable_plugin_short_name_repository = no\n" +
		"java_macos_integration_enable = yes\n"
	MakeFile(HomeDirectory()+".asdfrc", asdfrcContents, 0644)

	ASDFInstall("perl", "latest")
	ASDFInstall("ruby", "latest")
	ASDFInstall("python", "latest")
	ASDFInstall("java", "openjdk-11.0.2")
	ASDFInstall("java", "openjdk-17.0.2")
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

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install CLI applications!\n"
	macLdBar.Stop()
}

func macSecurity(adminCode string) {
	macLdBar.Suffix = " Installing - applications... "
	macLdBar.Start()

	MacPMSInstall("openvpn")
	MacPMSInstall("wireguard-go")
	MacPMSInstall("wireguard-tools")
	MacPMSInstall("tor")
	MacPMSInstall("torsocks")
	MacPMSInstall("nmap")
	MacPMSInstall("radare2")
	MacPMSInstall("sleuthkit")
	MacPMSInstall("autopsy")
	MacPMSInstall("virustotal-cli")

	MacPMSInstallCaskSudo("codeql", "CodeQL", brewPrefix+"Caskroom/Codeql", adminCode)
	MacPMSInstallCask("little-snitch", "Little Snitch")
	ChangeMacApplicationIcon("Little Snitch", "Little Snitch.icns", adminCode)
	MacPMSInstallCask("micro-snitch", "Micro Snitch")
	ChangeMacApplicationIcon("Micro Snitch", "Micro Snitch.icns", adminCode)
	MacPMSInstallCask("imazing", "iMazing")
	ChangeMacApplicationIcon("iMazing", "iMazing.icns", adminCode)
	MacInstallHopper(adminCode)
	MacPMSInstallCask("cutter", "Cutter")
	// Install Ghidra // TODO: will add
	MacPMSInstallCask("apparency", "Apparency")
	MacPMSInstallCask("suspicious-package", "Suspicious Package")
	MacPMSInstallCask("fsmonitor", "FSMonitor")
	ChangeMacApplicationIcon("FSMonitor", "FSMonitor.icns", adminCode)
	MacPMSInstallCaskSudo("wireshark", "Wireshark", "/Applications/Wireshark.app", adminCode)
	ChangeMacApplicationIcon("Wireshark", "Wireshark.icns", adminCode)
	MacPMSInstallCask("burp-suite", "Burp Suite Community Edition")
	MacPMSInstallCask("burp-suite-professional", "Burp Suite Professional")
	MacPMSInstallCask("owasp-zap", "OWASP ZAP")
	ChangeMacApplicationIcon("OWASP ZAP", "OWASP ZAP.icns", adminCode)
	MacPMSInstallCaskSudo("zenmap", "Zenmap", "/Applications/Zenmap.app", adminCode)
	ChangeMacApplicationIcon("Zenmap", "Zenmap.icns", adminCode)

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install CLI applications!\n"
	macLdBar.Stop()
}

func macEnd(userName, userEmail string) {
	macLdBar.Suffix = " Finishing... "
	macLdBar.Start()

	shrcAppend := "\n######## ADD CUSTOM VALUES UNDER HERE ########\n\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	MacPMSUpgrade()
	MacPMSCleanup()
	MacPMSRemoveCache()

	ConfigGit4sh(userName, userEmail)
	ChangeMacWallpaper()

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "clean up homebrew's cache!\n"
	macLdBar.Stop()
}

func CEIOSmacOS(adminCode string) {
	fmt.Println(clrCyan + "Return your information" + clrReset)
	consoleReader := bufio.NewScanner(os.Stdin)
	fmt.Print("User name: ")
	consoleReader.Scan()
	userName := consoleReader.Text()
	fmt.Print("User email: ")
	consoleReader.Scan()
	userEmail := consoleReader.Text()
	ClearLine(3)

	macInfo, fullName := MacInformatrion()

	if fullName != userName {
		var alertAnswer string
		fmt.Print(clrRed + "Warning\n" + clrReset + "Your user name is different from the system.\n" + "If you wish to continue type (Yes) then press return: ")
		_, errG4sOpt := fmt.Scanln(&alertAnswer)
		if errG4sOpt != nil {
			alertAnswer = "Enter"
		}
		if alertAnswer == "Yes" {
			ClearLine(3)
		} else {
			os.Exit(0)
		}
	}
	fmt.Println(clrCyan + "User Information\n" + clrReset +
		lstDot + clrGreen + "User name" + clrReset + ": " + userName + "\n" +
		lstDot + clrGreen + "User email" + clrReset + ": " + userEmail + "\n" +
		clrCyan + "System Information\n" + clrReset + macInfo)

	fmt.Println(clrCyan + "CEIOS OS Installation" + clrReset)
	macBegin(adminCode)
	macEnv()
	macDependency(adminCode)
	macUtility(adminCode)
	macProductivity(adminCode)
	macCreativity(adminCode)
	macDevelopment(adminCode)
	macSecurity(adminCode)
	macEnd(userName, userEmail)

	fmt.Println(" ------------------------------------------------------------\n" +
		clrCyan + "Finished CEIOS OS Installation" + clrReset +
		"\n Please" + clrRed + " RESTART " + clrReset + "your terminal and macOS!\n" +
		lstDot + "Use \"exec -l $SHELL\" or on terminal.\n" +
		lstDot + "Or restart the Terminal application by yourself.\n" +
		lstDot + "Also you need " + clrRed + "RESTART macOS " + clrReset + " to apply " + "the changes.\n" +
		clrCyan + "System Update and Restart OS" + clrReset)
	MacOSUpdate()
}
