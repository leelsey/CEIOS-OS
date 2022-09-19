package main

import (
	"errors"
	"fmt"
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

func MacStartApplication(appName string) {
	insLdBar.Suffix = " macOS is running " + appName + " ... "
	insLdBar.Start()

	runApp := exec.Command("open", "/Applications/"+appName+".app")
	err := runApp.Run()
	CheckCmdError(err, "Failed to start", appName)
	insLdBar.Stop()
}

func ChangeMacIcon(trgPath, icnName, adminCode string) {
	srcIcn := WorkingDirectory() + ".ceios-icn.icns"
	DownloadFile(srcIcn, CfgSto+"icns/"+icnName, 0755)
	if CheckSize(srcIcn) == 0 {
		RemoveFile(srcIcn)
		return
	}

	chicnSh := WorkingDirectory() + ".ceios-chicn.sh"
	cvtIcn := WorkingDirectory() + ".ceios-app-icn.rsrc"
	chIcnSrc := "sudo rm -rf \"" + trgPath + "\"$'/Icon\\r'\n" +
		"sips -i " + srcIcn + " > /dev/null\n" +
		"DeRez -only icns " + srcIcn + " > " + cvtIcn + "\n" +
		"sudo Rez -append " + cvtIcn + " -o " + trgPath + "$'/Icon\\r'\n" +
		"sudo SetFile -a C " + trgPath + "\n" +
		"sudo SetFile -a V " + trgPath + "$'/Icon\\r'"
	MakeFile(chicnSh, chIcnSrc, 0644)

	NeedPermission(adminCode)
	chicn := exec.Command(cmdSh, chicnSh)
	chicn.Env = os.Environ()
	chicn.Stderr = os.Stderr
	errChicn := chicn.Run()
	CheckCmdError(errChicn, "Failed change icon of", "\""+trgPath+"\"")

	RemoveFile(srcIcn)
	RemoveFile(cvtIcn)
	RemoveFile(chicnSh)
}

func ChangeMacApplicationIcon(appName, icnName, adminCode string) {
	insLdBar.Suffix = " macOS is changing application icon ... "
	insLdBar.Start()

	trgPath := "/Applications/" + strings.Replace(appName, " ", "\\ ", -1) + ".app"
	ChangeMacIcon(trgPath, icnName, adminCode)
	insLdBar.Stop()
}

func ChangeMacWallpaper(srcWp string) bool {
	insLdBar.Suffix = " macOS is configuring desktop wallpaper ... "
	insLdBar.Start()

	chWpPath := WorkingDirectory() + ".ceios-chwap.sh"
	chWpSrc := "osascript -e 'tell application \"Finder\" to set desktop picture to POSIX file \"" + srcWp + "\"'"
	MakeFile(chWpPath, chWpSrc, 0644)

	chWp := exec.Command(cmdSh, chWpPath)
	if err := chWp.Run(); err != nil {
		RemoveFile(chWpPath)
		insLdBar.Stop()
		return false
	}
	RemoveFile(chWpPath)
	insLdBar.Stop()
	return true
}

func MacPMSUpdate() {
	updateHomebrew := exec.Command(macPMS, "update", "--auto-update")
	err := updateHomebrew.Run()
	CheckCmdError(err, "Homebrew failed to", "update repositories")
}

func MacPMSUpgrade() {
	MacPMSUpdate()
	upgradeHomebrew := exec.Command(macPMS, "upgrade", "--greedy")
	err := upgradeHomebrew.Run()
	CheckCmdError(err, "Homebrew failed to", "upgrade packages")
}

func MacPMSRepository(repo string) {
	brewRepoName := strings.Split(repo, "/")
	repoPath := strings.Join(brewRepoName[0:1], "") + "/homebrew-" + strings.Join(brewRepoName[1:2], "")
	if CheckExists(pmsPrefix+"Homebrew/Library/Taps/"+repoPath) != true {
		brewRepo := exec.Command(macPMS, macRepo, repo)
		err := brewRepo.Run()
		CheckCmdError(err, "Homebrew failed to add ", repo)
	}
}

func MacPMSCleanup() {
	upgradeHomebrew := exec.Command(macPMS, "cleanup", "--prune=all", "-nsd")
	err := upgradeHomebrew.Run()
	CheckCmdError(err, "Homebrew failed to", "cleanup old packages")
}

func MacPMSRemoveCache() {
	upgradeHomebrew := exec.Command("rm", "-rf", "\"$(brew --cache)\"")
	err := upgradeHomebrew.Run()
	CheckCmdError(err, "Homebrew failed to", "remove cache")
}

func MacPMSInstall(pkg string) {
	insLdBar.Suffix = " Homebrew is installing " + pkg + " ... "
	insLdBar.Start()

	if CheckExists(pmsPrefix+"Cellar/"+pkg) != true {
		MacPMSUpdate()
		brewIns := exec.Command(macPMS, optIns, pkg)
		brewIns.Stderr = os.Stderr
		err := brewIns.Run()
		CheckCmdError(err, "Homebrew failed to install", pkg)
	}
	insLdBar.Stop()
}

func MacPMSInstallQuiet(pkg string) {
	insLdBar.Suffix = " Homebrew is installing " + pkg + " ... "
	insLdBar.Start()

	if CheckExists(pmsPrefix+"Cellar/"+pkg) != true {
		MacPMSUpdate()
		brewIns := exec.Command(macPMS, optIns, "--quiet", pkg)
		err := brewIns.Run()
		CheckCmdError(err, "Homebrew failed to install", pkg)
	}
	insLdBar.Stop()
}

func MacPMSInstallCask(pkg, appName string) {
	insLdBar.Suffix = " Homebrew is installing " + pkg + " ... "
	insLdBar.Start()

	if CheckExists(pmsPrefix+"Caskroom/"+pkg) != true {
		MacPMSUpdate()
		if CheckExists("/Applications/"+appName+".app") != true {
			brewIns := exec.Command(macPMS, optIns, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Homebrew failed to install cask", pkg)
		} else {
			brewIns := exec.Command(macPMS, optReIn, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Homebrew failed to reinstall cask", pkg)
		}
	}
	insLdBar.Stop()
}

func MacPMSInstallCaskSudo(pkg, appName, appPath, adminCode string) {
	insLdBar.Suffix = " Homebrew is installing " + pkg + " ... "
	insLdBar.Start()

	if CheckExists(pmsPrefix+"Caskroom/"+pkg) != true {
		MacPMSUpdate()
		NeedPermission(adminCode)
		if CheckExists(appPath) != true {
			brewIns := exec.Command(macPMS, optIns, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Homebrew failed to install cask", appName)
		} else {
			brewIns := exec.Command(macPMS, optReIn, macAlt, pkg)
			err := brewIns.Run()
			CheckCmdError(err, "Homebrew failed to install cask", appName)
		}
	}
	insLdBar.Stop()
}

func MacJavaHome(srcVer, dstVer, adminCode string) {
	insLdBar.Suffix = " macOS is adding java" + dstVer + " to jvm home ... "
	insLdBar.Start()

	if CheckExists(pmsPrefix+"Cellar/openjdk"+srcVer) == true {
		LinkFile(pmsPrefix+"opt/openjdk"+srcVer+" /libexec/openjdk.jdk", "/Library/Java/JavaVirtualMachines/openjdk"+dstVer+".jdk", "symbolic", "root", adminCode)
	}
	insLdBar.Stop()
}

func MacInstallRosetta2() {
	insLdBar.Suffix = " macOS is installing Rosetta2 ... "
	insLdBar.Start()

	osUpdate := exec.Command("softwareupdate", "--install-rosetta", "--agree-to-license")
	if err := osUpdate.Run(); err != nil {
		CheckCmdError(err, "Failed to install", "Rosetta 2")
	}
	insLdBar.Stop()
}

func MacInstallBrew(adminCode string) {
	insLdBar.Suffix = " macOS is installing homebrew ... "
	insLdBar.Start()

	insBrewPath := WorkingDirectory() + ".ceios-brew.sh"
	DownloadFile(insBrewPath, ghRaw+"Homebrew/install/HEAD/install.sh", 0755)

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
	insLdBar.Stop()
}

func MacInstallHopper(adminCode string) {
	insLdBar.Suffix = " macOS is installing hopper disassembler ... "
	insLdBar.Start()

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
	var macBeginFinalMSG string
	if CheckExists(macPMS) == true {
		fmt.Println("- Update Homebrew")
		macBeginFinalMSG = "update homebrew!"
	} else {
		fmt.Println("- Install and Update Homebrew")
		macBeginFinalMSG = "install and update homebrew!"
		MacInstallBrew(adminCode)
	}

	insLdBar.Suffix = " Homebrew is updating ... "
	insLdBar.Start()

	err := os.Chmod(pmsPrefix+"share", 0755)
	CheckError(err, "Failed to change permissions on "+pmsPrefix+"share to 755")
	MacPMSUpdate()
	MacPMSRepository("homebrew/core")
	MacPMSRepository("homebrew/cask")
	MacPMSRepository("homebrew/cask-versions")
	MacPMSRepository("romkatv/powerlevel10k")
	MacPMSUpgrade()

	insLdBar.Stop()
	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + macBeginFinalMSG)
}

func macEnvironment(userName, userEmail string) {
	fmt.Println("- Environment Configuration")
	insLdBar.Suffix = " Initial setting zsh environment ... "
	insLdBar.Start()

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
	shrcContents := "#   ______ _____ _    _ _____   _____\n" +
		"#  |___  // ____| |  | |  __ \\ / ____|\n" +
		"#     / /| (___ | |__| | |__) | |\n" +
		"#    / /  \\___ \\|  __  |  _  /| |\n" +
		"#   / /__ ____) | |  | | | \\ \\| |____\n" +
		"#  /_____|_____/|_|  |_|_|  \\_\\\\_____|\n#\n" +
		"#  " + CurrentUsername() + "’s zsh run commands\n\n"
	MakeFile(prfPath, profileContents, 0644)
	MakeFile(shrcPath, shrcContents, 0644)

	MakeDirectory(HomeDirectory() + ".config")
	MakeDirectory(HomeDirectory() + ".cache")
	insLdBar.Stop()

	insLdBar.Suffix = " macOS is configuring global git ... "
	insLdBar.Start()

	Git4shSet(userName, userEmail)
	insLdBar.Stop()

	insLdBar.Suffix = " macOS is downloading desktop wallpaper ... "
	insLdBar.Start()

	picturesPath := HomeDirectory() + "Pictures/"
	DownloadFile(picturesPath+"Cube Glass.heic", CfgSto+"wallpaper/Cube Glass.heic", 0644)
	DownloadFile(picturesPath+"Orb Glass.heic", CfgSto+"wallpaper/Orb Glass.heic", 0644)
	DownloadFile(picturesPath+"Oval Wave.heic", CfgSto+"wallpaper/Oval Wave.heic", 0644)
	DownloadFile(picturesPath+"Silk Wave.heic", CfgSto+"wallpaper/Silk Wave.heic", 0644)
	DownloadFile(picturesPath+"Stone Wave.heic", CfgSto+"wallpaper/Stone Wave.heic", 0644)
	DownloadFile(picturesPath+"Blue Wave.heic", CfgSto+"wallpaper/Blue Wave.jpg", 0644)
	DownloadFile(picturesPath+"Purple Wave.heic", CfgSto+"wallpaper/Purple Wave.jpg", 0644)
	DownloadFile(picturesPath+"Rain Wave.heic", CfgSto+"wallpaper/Rain Glass.jpg", 0644)
	//DownloadFile(picturesPath+"CEIOS OS.heic", CfgSto+"wallpaper/CEIOS OS.heic", 0644)               // TODO: Add CEIOS OS wallpaper
	//DownloadFile(picturesPath+"CEIOS Ops.heic", CfgSto+"wallpaper/CEIOS Ops.heic", 0644)             // TODO: Add CEIOS OS wallpaper
	//DownloadFile(picturesPath+"CEIOS Red Team.heic", CfgSto+"wallpaper/CEIOS Red Team.heic", 0644)   // TODO: Add CEIOS OS wallpaper
	//DownloadFile(picturesPath+"CEIOS Blue Team.heic", CfgSto+"wallpaper/CEIOS Blue Team.heic", 0644) // TODO: Add CEIOS OS wallpaper
	insLdBar.Stop()

	if CheckArchitecture() == "arm64" {
		MacInstallRosetta2()
	}

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup environment!")
}

func macDependency(adminCode string) {
	fmt.Println("- Dependency Installation")

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

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup dependencies!")
}

func macUtility(adminCode string) {
	fmt.Println("- Utility Installation")

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
	// Install P        // TODO: will add
	// Install JCtl     // TODO: will add
	// Install ShCtl    // TODO: will add
	// Install Verifier // TODO: will add
	MacPMSInstall("diffutils")
	MacPMSInstall("diffr")
	MacPMSInstall("tldr")
	MacPMSInstall("htop")
	MacPMSInstall("btop")
	MacPMSInstall("iperf3")
	MacPMSInstall("neofetch")
	MacPMSInstall("transmission-cli")
	MacPMSInstall("romkatv/powerlevel10k/powerlevel10k")

	insLdBar.Suffix = " macOS is configuring zsh environment ... "
	insLdBar.Start()

	p10kConfPath := HomeDirectory() + ".config/p10k/"
	p10kCachePath := HomeDirectory() + ".cache/p10k-" + CurrentUsername()
	fontLibPath := HomeDirectory() + "Library/Fonts/"
	nerdfontDlPath := ghRaw + "romkatv/dotfiles-public/master/.local/share/fonts/NerdFonts/"
	MakeDirectory(p10kConfPath)
	MakeDirectory(p10kCachePath)
	DownloadFile(p10kConfPath+"p10k-term.zsh", CfgSto+"p10k/p10k-minimalism.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-iterm2.zsh", CfgSto+"p10k/p10k-atelier.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-tmux.zsh", CfgSto+"p10k/p10k-seeking.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-ops.zsh", CfgSto+"p10k/p10k-operations.zsh", 0644)
	DownloadFile(p10kConfPath+"p10k-etc.zsh", CfgSto+"p10k/p10k-engineering.zsh", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Bold Italic.ttf", nerdfontDlPath+"MesloLGS NF Bold Italic.ttf", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Bold.ttf", nerdfontDlPath+"MesloLGS NF Bold.ttf", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Italic.ttf", nerdfontDlPath+"MesloLGS NF Italic.ttf", 0644)
	DownloadFile(fontLibPath+"MesloLGS NF Regular.ttf", nerdfontDlPath+"MesloLGS NF Regular.ttf", 0644)
	DownloadFile(HomeDirectory()+"Library/Preferences/com.googlecode.iterm2.plist", CfgSto+"iterm2/iTerm2.plist", 0644)

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
	insLdBar.Stop()

	MacPMSInstallCask("iina", "IINA")
	MacPMSInstallCask("sensei", "Sensei")
	MacPMSInstallCask("rectangle", "Rectangle")
	MacPMSInstallCask("dropbox", "Dropbox")
	MacPMSInstallCask("dropbox-capture", "Dropbox Capture")
	MacStartApplication("Dropbox")
	MacPMSInstallCask("keka", "Keka")
	MacPMSInstallCask("transmission", "Transmission")
	ChangeMacApplicationIcon("Transmission", "Transmission.icns", adminCode)

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup utilities!")
}

func macProductivity(adminCode string) {
	fmt.Println("- Productivity Installation")

	MacPMSInstallCask("google-chrome", "Google Chrome")
	MacStartApplication("Google Chrome")
	MacPMSInstallCask("firefox", "Firefox")
	ChangeMacApplicationIcon("Firefox", "Firefox.icns", adminCode)
	MacPMSInstallCask("tor-browser", "Tor Browser")
	ChangeMacApplicationIcon("Tor Browser", "Tor Browser.icns", adminCode)
	//MacPMSInstallCask("chromium", "Chromium") TODO: Will add Grimoire (LE Chromium)
	MacPMSInstallCask("spotify", "Spotify")
	ChangeMacApplicationIcon("Spotify", "Spotify.icns", adminCode)
	MacPMSInstallCask("signal", "Signal")
	MacPMSInstallCask("discord", "Discord")
	MacPMSInstallCask("jetbrains-space", "JetBrains Space")
	ChangeMacApplicationIcon("JetBrains Space", "JetBrains Space.icns", adminCode)
	MacPMSInstallCask("notion", "Notion")
	ChangeMacApplicationIcon("Notion", "Notion.icns", adminCode)

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup productivity!")
}

func macCreativity(adminCode string) {
	fmt.Println("- Creativity Installation")

	MacPMSInstall("asciinema")

	MacPMSInstallCask("sketch", "Sketch")
	MacPMSInstallCask("zeplin", "Zeplin")
	MacPMSInstallCask("blender", "Blender")
	ChangeMacApplicationIcon("Blender", "Blender.icns", adminCode)
	MacPMSInstallCask("obs", "OBS")
	MacPMSInstallCaskSudo("loopback", "Loopback", "/Applications/Loopback.app", adminCode)
	MacStartApplication("Loopback")

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup creativity!")
}

func macDevelopment(adminCode string) {
	fmt.Println("- Developer Tool Installation")

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

	insLdBar.Suffix = " macOS is configuring zsh for direnv ... "
	insLdBar.Start()

	shrcAppend := "# DIRENV\n" +
		"eval \"$(direnv hook zsh)\"\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)
	insLdBar.Stop()

	MacPMSInstallCask("iterm2", "iTerm")
	MacStartApplication("iTerm")
	MacPMSInstallCask("intellij-idea", "IntelliJ IDEA")
	ChangeMacApplicationIcon("IntelliJ IDEA", "IntelliJ IDEA.icns", adminCode)
	MacPMSInstallCask("visual-studio-code", "Visual Studio Code")
	MacPMSInstallCask("neovide", "Neovide")
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

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup developer tools!")
}

func macSecurity(adminCode string) {
	fmt.Println("- Security Tool Installation")

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

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup security tools!")
}

func macVirtualMachine(vmSts bool, adminCode string) {
	fmt.Println("- virtualisation Tool Installation")

	MacPMSInstall("asdf")

	insLdBar.Suffix = " macOS is configuring zsh for ASDF-VM ... "
	insLdBar.Start()

	shrcAppend := "# ASDF-VM\n" +
		"source " + pmsPrefix + "opt/asdf/libexec/asdf.sh\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)
	insLdBar.Stop()

	if vmSts != true {
		MacPMSInstallCask("docker", "Docker")
		MacStartApplication("Docker")
		MacPMSInstallCaskSudo("vmware-fusion", "VMware Fusion", "/Applications/VMware Fusion.app", adminCode)
		ChangeMacApplicationIcon("VMware Fusion", "VMware Fusion.icns", adminCode)
		MacStartApplication("VMware Fusion")
	}

	ASDFSet(MacASDFPath())
	if vmSts != true {
		DockerSet(MacDockerPath())
	}

	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "install and setup virtualisation tools!")
}

func macEnd() {
	fmt.Println("- Final Organisation")
	insLdBar.Suffix = " macOS is configuring zsh for finishing ... "
	insLdBar.Start()

	shrcAppend := "\n######## ADD CUSTOM VALUES UNDER HERE ########\n\n\n"
	AppendFile(shrcPath, shrcAppend, 0644)
	insLdBar.Stop()

	insLdBar.Suffix = " macOS is configuring workspace directory ... "
	insLdBar.Start()

	keyPath := HomeDirectory() + "Key/"
	if CheckExists(keyPath) == true {
		if CheckExists(keyPath+"AWS") == true {
			err := os.Chmod(HomeDirectory()+"Dropbox/Key/AWS", 0600)
			CheckError(err, "Failed to change permissions on AWS Key on Dropbox to 600")
		}
		if CheckExists(keyPath+"GCP") == true {
			err := os.Chmod(HomeDirectory()+"Dropbox/Key/GCP", 0600)
			CheckError(err, "Failed to change permissions on GCP Key on Dropbox to 600")
		}
		if CheckExists(keyPath+"SSL") == true {
			err := os.Chmod(HomeDirectory()+"Dropbox/Key/SSL", 0600)
			CheckError(err, "Failed to change permissions on SSL Key on Dropbox to 600")
		}
	}

	RemoveFile(HomeDirectory() + "Applications")
	RemoveFile(HomeDirectory() + "Virtual Machines")
	MakeDirectory(HomeDirectory() + "Public/OS Images")
	MakeDirectory(HomeDirectory() + "Public/Share Box")
	MakeDirectory(HomeDirectory() + "Public/Virtual Machines")
	insLdBar.Stop()

	insLdBar.Suffix = " Clearing homebrew caches ... "
	insLdBar.Start()

	MacPMSUpgrade()
	MacPMSCleanup()
	MacPMSRemoveCache()

	insLdBar.Stop()
	ClearLine(1)
	fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "clean up cache, configure git!")
}

func macExtended() {
	fmt.Println("- Additional Configuration")
	if ChangeMacWallpaper(HomeDirectory()+"Pictures/Orb Glass Dynamic.heic") != true {
		ClearLine(1)
		fmt.Println(errors.New(fntBold + fntRed + "   Failed " + fntReset + "change desktop wallpaper."))
	} else {
		ClearLine(1)
		fmt.Println(fntBold + fntGreen + "   Succeed " + fntReset + "change desktop wallpaper!")
	}
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
		var vmSts bool
		if strings.Contains(chipInfo, "Unknown") == true {
			chipInfo = "Virtual Machine"
			vmSts = true
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
		macEnvironment(userName, userEmail)
		macDependency(adminCode)
		macUtility(adminCode)
		macProductivity(adminCode)
		macCreativity(adminCode)
		macDevelopment(adminCode)
		macSecurity(adminCode)
		macVirtualMachine(vmSts, adminCode)
		macEnd()
		macExtended()

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
