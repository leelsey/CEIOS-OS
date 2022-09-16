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
	arm64Path = "/opt/homebrew/"
	amd64Path = "/usr/local/"
	pmsPrefix = MacPMSPrefix()
	macPMS    = MacPMSPath()
	macAlt    = "--cask"
	macGit    = "/usr/bin/git"
	macRepo   = "tap"
	macLdBar  = spinner.New(spinner.CharSets[16], 50*time.Millisecond)
)

func MacPMSPrefix() string {
	if archType == "arm64" {
		return arm64Path
	} else if archType == "amd64" {
		return amd64Path
	}
	return ""
}

func MacPMSPath() string {
	if archType == "arm64" {
		return arm64Path + "bin/brew"
	} else if archType == "amd64" {
		return amd64Path + "bin/brew"
	}
	return ""
}

func MacASDFPath() string {
	asdfPath := "opt/asdf/libexec/bin/asdf"
	if archType == "arm64" {
		return arm64Path + asdfPath
	} else if archType == "amd64" {
		return amd64Path + asdfPath
	}
	return ""
}

func MacDockerPath() string {
	asdfPath := "bin/docker"
	if archType == "arm64" {
		return arm64Path + asdfPath
	} else if archType == "amd64" {
		return amd64Path + asdfPath
	}
	return ""
}

func MacOSUpdate() {
	runLdBar.Suffix = " Updating OS, please wait a moment ... "
	runLdBar.Start()

	osUpdate := exec.Command("softwareupdate", "--all", "--install", "--force")
	errOSUpdate := osUpdate.Run()
	CheckError(errOSUpdate, "Failed to update Operating System")

	runLdBar.FinalMSG = lstDot + fntGreen + "Succeed " + fntReset + "update OS!\n"
	runLdBar.Stop()
}

func MacSoftware() string {
	softInfo, err := exec.Command("system_profiler", "SPSoftwareDataType").Output()
	CheckError(err, "Failed to get macOS hardware information")
	return string(softInfo)
}

func MacHardware() string {
	macInfo, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	CheckError(err, "Failed to get macOS hardware information")
	return string(macInfo)
}

func MacInformation() (string, string, string, string, string, string, string) {
	runLdBar.Suffix = " Checking system... "
	runLdBar.Start()

	softInfo := strings.Split(MacSoftware(), "\n")
	hardInfo := strings.Split(MacHardware(), "\n")

	osVer := strings.Split(strings.Split(softInfo[4], ": ")[1], " ")[1]
	deviceName := strings.Split(softInfo[8], ": ")[1]
	userFullName := strings.Split(strings.Split(softInfo[9], ": ")[1], " (")[0]

	osCode := strings.Split(osVer, ".")[0]
	var osName string
	if osCode == "13" {
		osName = "macOS Ventura"
	} else if osCode == "12" {
		osName = "macOS Monterey"
	} else if osCode == "11" {
		osName = "macOS Big Sur"
	} else {
		runLdBar.Stop()
		MessageError("fatal", "Unsupported", "macOS version")
		return "", "", "", "", "", "", ""
	}

	if archType == "arm64" {
		modelInfo := strings.Split(hardInfo[4], ": ")[1]
		chipInfo := strings.Split(hardInfo[6], ": ")[1]
		memoryInfo := strings.Split(hardInfo[8], ": ")[1]
		runLdBar.Stop()
		return osName, osVer, modelInfo, chipInfo, memoryInfo, deviceName, userFullName
	} else if archType == "amd64" {
		modelInfo := strings.Split(hardInfo[4], ": ")[1]
		processorInfo := strings.Split(hardInfo[6], ": ")[1] + " " + strings.Split(hardInfo[7], ": ")[1]
		memoryInfo := strings.Split(hardInfo[13], ": ")[1]
		runLdBar.Stop()
		return osName, osVer, modelInfo, processorInfo, memoryInfo, deviceName, userFullName
	}
	return "", "", "", "", "", "", ""
}

func StartMacApplication(appName string) {
	runApp := exec.Command("open", "/Applications/"+appName+".app")
	err := runApp.Run()
	CheckCmdError(err, "Failed to start", appName)
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

func ChangeMacWallpaper(srcWp string) bool {
	//srcWp := HomeDirectory() + "Pictures/" + wpPath
	chWpPath := WorkingDirectory() + ".ceios-chwap.sh"
	chWpSrc := "osascript -e 'tell application \"Finder\" to set desktop picture to POSIX file \"" + srcWp + "\"'"
	MakeFile(chWpPath, chWpSrc, 0644)

	chWp := exec.Command(cmdSh, chWpPath)
	if err := chWp.Run(); err != nil {
		RemoveFile(chWpPath)
		return false
	}
	RemoveFile(chWpPath)
	return true
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
	if CheckExists(pmsPrefix+"Homebrew/Library/Taps/"+repoPath) != true {
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
	if CheckExists(pmsPrefix+"Cellar/"+pkg) != true {
		MacPMSUpdate()
		brewIns := exec.Command(macPMS, optIns, pkg)
		brewIns.Stderr = os.Stderr
		err := brewIns.Run()
		CheckCmdError(err, "Brew failed to install", pkg)
	}
}

func MacPMSInstallQuiet(pkg string) {
	if CheckExists(pmsPrefix+"Cellar/"+pkg) != true {
		MacPMSUpdate()
		brewIns := exec.Command(macPMS, optIns, "--quiet", pkg)
		err := brewIns.Run()
		CheckCmdError(err, "Brew failed to install", pkg)
	}
}

func MacPMSInstallCask(pkg, appName string) {
	if CheckExists(pmsPrefix+"Caskroom/"+pkg) != true {
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
	if CheckExists(pmsPrefix+"Caskroom/"+pkg) != true {
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
	if CheckExists(pmsPrefix+"Cellar/openjdk"+srcVer) == true {
		LinkFile(pmsPrefix+"opt/openjdk"+srcVer+" /libexec/openjdk.jdk", "/Library/Java/JavaVirtualMachines/openjdk"+dstVer+".jdk", "symbolic", "root", adminCode)
	}
}

func MacInstallRosetta2() {
	osUpdate := exec.Command("softwareupdate", "--install-rosetta", "--agree-to-license")
	if err := osUpdate.Run(); err != nil {
		CheckCmdError(err, "Failed to install", "Rosetta 2")
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
	CheckError(errMount, "Failed to mount "+fntYellow+"Hopper.dmg"+fntReset)
	RemoveFile(dlHopperPath)

	appName := "Hopper Disassembler v4"
	CopyDirectory("/Volumes/Hopper Disassembler/"+appName+".app", "/Applications/"+appName+".app")

	unmountDmg := exec.Command("hdiutil", "unmount", "/Volumes/Hopper Disassembler")
	errUnmount := unmountDmg.Run()
	CheckError(errUnmount, "Failed to unmount "+fntYellow+"Hopper Disassembler"+fntReset)

	if archType == "arm64" {
		ChangeMacApplicationIcon(appName, "Hopper Disassembler ARM64.icns", adminCode)
	} else if archType == "amd64" {
		ChangeMacApplicationIcon(appName, "Hopper Disassembler AMD64.icns", adminCode)
	}
}

func macBegin(adminCode string) {
	if CheckExists(macPMS) == true {
		macLdBar.Suffix = " Updating homebrew... "
		macLdBar.Start()
		macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "update homebrew!\n"
	} else {
		macLdBar.Suffix = " Installing homebrew... "
		macLdBar.Start()
		macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "install and update homebrew!\n"

		MacInstallBrew(adminCode)
	}
	err := os.Chmod(pmsPrefix+"share", 0755)
	CheckError(err, "Failed to change permissions on "+pmsPrefix+"share to 755")

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

	if CheckArchitecture() == "arm64" {
		MacInstallRosetta2()
	}

	PicturesPath := HomeDirectory() + "Pictures/"
	DownloadFile(PicturesPath+"Cube Glass Light.jpg", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Cube Glass Light.jpg", 0644)
	DownloadFile(PicturesPath+"Cube Glass Dark.jpg", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Cube Glass Dark.jpg", 0644)
	DownloadFile(PicturesPath+"Cube Glass Light and Dark.heic", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Cube Glass Light and Dark.heic", 0644)
	DownloadFile(PicturesPath+"Orb Glass Light.jpg", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Orb Glass Light.jpg", 0644)
	DownloadFile(PicturesPath+"Orb Glass White.jpg", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Orb Glass White.jpg", 0644)
	DownloadFile(PicturesPath+"Orb Glass Green.jpg", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Orb Glass Green.jpg", 0644)
	DownloadFile(PicturesPath+"Orb Glass Blue.jpg", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Orb Glass Blue.jpg", 0644)
	DownloadFile(PicturesPath+"Orb Glass Dynamic.heic", "https://raw.githubusercontent.com/leelsey/ConfStore/main/wallpaper/Orb Glass Dynamic.heic", 0644)

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "setup zsh environment!\n"
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
	if archType == "amd64" {
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

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "install dependencies!\n"
	macLdBar.Stop()
}

func macUtility(adminCode string) {
	macLdBar.Suffix = " Installing - applications... "
	macLdBar.Start()

	Alias4shSet()
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
		"source " + pmsPrefix + "opt/powerlevel10k/powerlevel10k.zsh-theme\n" +
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
		"  FPATH=" + pmsPrefix + "share/zsh-completions:$FPATH\n" +
		"  autoload -Uz compinit\n" +
		"  compinit\n" +
		"fi\n\n" +
		"# ZSH SYNTAX HIGHLIGHTING\n" +
		"source " + pmsPrefix + "share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh\n\n" +
		"# ZSH AUTOSUGGESTIONS\n" +
		"source " + pmsPrefix + "share/zsh-autosuggestions/zsh-autosuggestions.zsh\n\n" +
		"# Z\n" +
		"source " + pmsPrefix + "etc/profile.d/z.sh\n\n" +
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

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "install CLI applications!\n"
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

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "install CLI applications!\n"
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
	StartMacApplication("Loopback")
	MacPMSInstallCask("obs", "OBS")

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "install CLI applications!\n"
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

	shrcAppend := "# DIRENV\n" +
		"eval \"$(direnv hook zsh)\"\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	MacPMSInstallCask("iterm2", "iTerm")
	MacPMSInstallCask("neovide", "Neovide")
	MacPMSInstallCask("visual-studio-code", "Visual Studio Code")
	MacPMSInstallCask("atom", "Atom")
	MacPMSInstallCask("intellij-idea", "IntelliJ IDEA")
	ChangeMacApplicationIcon("IntelliJ IDEA", "IntelliJ IDEA.icns", adminCode)
	ChangeMacApplicationIcon("Neovide", "Neovide.icns", adminCode)
	MacPMSInstallCask("github", "Github")
	MacPMSInstallCask("fork", "Fork")
	MacPMSInstallCask("tableplus", "TablePlus")
	MacPMSInstallCask("proxyman", "Proxyman")
	MacPMSInstallCask("paw", "Paw")
	MacPMSInstallCask("boop", "Boop")
	MacPMSInstallCask("httpie", "HTTPie")
	MacPMSInstallCask("vnc-viewer", "VNC Viewer")
	ChangeMacApplicationIcon("VNC Viewer", "VNC Viewer.icns", adminCode)
	MacPMSInstallCask("forklift", "ForkLift")
	MacPMSInstallCask("drawio", "draw.io")
	MacPMSInstallCask("staruml", "StarUML")
	ChangeMacApplicationIcon("StarUML", "StarUML.icns", adminCode)

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "install CLI applications!\n"
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

	MacPMSInstallCaskSudo("codeql", "CodeQL", pmsPrefix+"Caskroom/Codeql", adminCode)
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

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "install CLI applications!\n"
	macLdBar.Stop()
}

func macVirtualMachines(adminCode string) {
	MacPMSInstall("asdf")
	MacPMSInstallCask("docker", "Docker")
	StartMacApplication("Docker")
	MacPMSInstallCaskSudo("vmware-fusion", "VMware Fusion", "/Applications/VMware Fusion.app", adminCode)
	ChangeMacApplicationIcon("VMware Fusion", "VMware Fusion.icns", adminCode)

	shrcAppend := "# ASDF VM\n" +
		"source " + pmsPrefix + "opt/asdf/libexec/asdf.sh\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	ASDFSet(MacASDFPath())
	DockerSet(MacDockerPath())
}

func macEnd(userName, userEmail string) {
	macLdBar.Suffix = " Finishing... "
	macLdBar.Start()

	shrcAppend := "\n######## ADD CUSTOM VALUES UNDER HERE ########\n\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)

	MacPMSUpgrade()
	MacPMSCleanup()
	MacPMSRemoveCache()
	Git4shSet(userName, userEmail)
	if ChangeMacWallpaper(HomeDirectory()+"Pictures/Orb Glass Dynamic.heic") != true {
		macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "clean up homebrew's cache!\n"
		macLdBar.Stop()
		fmt.Println(fntBold + fntRed + "   Failed " + fntReset + "change desktop wallpaper and cofigure git!")
	}

	macLdBar.FinalMSG = fntBold + fntGreen + "   Succeed " + fntReset + "clean up cache, config git and wallpaper!\n"
	macLdBar.Stop()
}

func CEIOS4macOS(adminCode string) bool {
	TitleLine("User Information")
	if userName, usrName, userEmail, usrEmail, userSts := CheckUserInformation(); userSts == true {
		TitleLine("Check Computer Status")
		if CheckNetworkStatus() != true {
			AlertLine("Network connect failed")
			fmt.Println(errors.New(lstDot + "Please check your internet connection."))
			return false
		}
		osName, osVer, modelInfo, chipInfo, memoryInfo, deviceName, userFullName := MacInformation()
		ClearLine(1)

		var chipArch string
		if archType == "arm64" {
			chipArch = "   Chip "
		} else if archType == "amd64" {
			chipArch = "   Processor "
		}
		macInfo := fntBold + " " + osName + fntReset + "\n" + fntBold + "   Version " + fntReset + osVer + "\n" +
			fntBold + " " + modelInfo + fntReset + "\n" + fntBold + chipArch + fntReset + chipInfo + "\n" +
			fntBold + "   Memory " + fntReset + memoryInfo + "\n" + fntBold + "   Device " + fntReset + deviceName + "\n" +
			fntBold + "   User " + fntReset + userFullName

		if userFullName != userName {
			AlertLine("Warning!")
			fmt.Println(errors.New(lstDot + "Your user username is different from the system."))
			fmt.Print(" If you wish to continue type (Yes) then press return: ")
			var alertAnswer string
			_, errG4sOpt := fmt.Scanln(&alertAnswer)
			if errG4sOpt != nil {
				alertAnswer = "Enter"
			}
			if alertAnswer == "Yes" {
				ClearLine(3)
				TitleLine("System Information")
				fmt.Println(macInfo + " (" + usrName + " - " + usrEmail + ")")
			} else {
				ClearLine(1)
				fmt.Println(errors.New(lstDot + "Installation canceled by user."))
				return false
			}
		} else {
			TitleLine("System Information")
			fmt.Println(macInfo + " (" + usrEmail + ")")
		}

		time.Sleep(time.Millisecond * 300)
		TitleLine("CEIOS OS Installation")
		macBegin(adminCode)
		macEnv()
		macDependency(adminCode)
		macUtility(adminCode)
		macProductivity(adminCode)
		macCreativity(adminCode)
		macDevelopment(adminCode)
		macSecurity(adminCode)
		macVirtualMachines(adminCode)
		macEnd(userName, userEmail)

		fmt.Println(" ------------------------------------------------------------")
		TitleLine("Finished CEIOS OS Installation")
		fmt.Println(" Please" + fntBold + fntRed + " RESTART " + fntReset + "your " + fntPurple + "Terminal and macOS!\n" +
			fntReset + lstDot + "Use \"exec -l $SHELL\" or on terminal.\n" +
			lstDot + "Or restart the Terminal application by yourself.\n" +
			lstDot + "Also you need restart macOS to apply the changes.")
		TitleLine("System Update and Restart OS")
		MacOSUpdate()
	} else {
		return false
	}
	return true
}
