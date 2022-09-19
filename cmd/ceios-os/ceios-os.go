package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"golang.org/x/term"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"
)

var (
	appVer   = "0.1"
	lstDot   = " • "
	archType = CheckArchitecture()
	osType   = CheckOperatingSystem()
	shrcPath = HomeDirectory() + ".zshrc"
	prfPath  = HomeDirectory() + ".zprofile"
	ghRaw    = "https://raw.githubusercontent.com/"
	CfgSto   = ghRaw + "leelsey/ConfStore/main/"
	cmdAdmin = "sudo"
	cmdSh    = "/bin/bash"
	optIns   = "install"
	optReIn  = "reinstall"
	//optUnIn    = "uninstall"
	//optRm      = "remove"
	fntReset  = "\033[0m"
	fntBold   = "\033[1m"
	fntRed    = "\033[31m"
	fntGreen  = "\033[32m"
	fntYellow = "\033[33m"
	fntBlue   = "\033[34m"
	fntPurple = "\033[35m"
	fntCyan   = "\033[36m"
	fntGrey   = "\033[37m"
	runLdBar  = spinner.New(spinner.CharSets[11], 50*time.Millisecond)
	insLdBar  = spinner.New(spinner.CharSets[14], 50*time.Millisecond)
)

func MessageError(handling, msg, code string) {
	errOccurred := fntRed + "\nError occurred " + fntReset + "at "
	errMsgFormat := "\n" + fntRed + "Error >> " + fntReset + msg + " (" + code + ")"
	if handling == "fatal" || handling == "stop" {
		fmt.Print(errors.New("\n" + lstDot + "Fatal error" + errOccurred))
		log.Fatalln(errMsgFormat)
	} else if handling == "print" || handling == "continue" {
		log.Println(errMsgFormat)
	} else if handling == "panic" || handling == "detail" {
		fmt.Print(errors.New("\n" + lstDot + "Panic error" + errOccurred))
		panic(errMsgFormat)
	} else {
		fmt.Print(errors.New("\n" + lstDot + "Unknown error" + errOccurred))
		log.Fatalln(errMsgFormat)
	}
}

func CheckError(err error, msg string) {
	if err != nil {
		MessageError("fatal", " "+msg, err.Error())
	}
}

func CheckCmdError(err error, msg, pkg string) {
	if err != nil {
		MessageError("print", msg+" "+fntYellow+pkg+fntReset, err.Error())
	}
}

func CheckArchitecture() string {
	switch runtime.GOARCH {
	case "arm64":
		return "arm64"
	case "amd64":
		return "amd64"
	default:
		return "unknown"
	}
}

func CheckOperatingSystem() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	default:
		return "unknown"
	}
}

func CheckExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	} else {
		return false
	}
}

func CheckSize(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}

	defer func() {
		errFile := file.Close()
		CheckError(errFile, "Failed to close file")
	}()

	stat, err := file.Stat()
	CheckError(err, "Failed to get file stat")

	return int(stat.Size())
}

func CheckNetworkStatus() bool {
	runLdBar.Suffix = " Checking internet connection... "
	runLdBar.Start()

	getTimeout := 10000 * time.Millisecond
	client := http.Client{
		Timeout: getTimeout,
	}
	_, err := client.Get("https://9.9.9.9")
	if err != nil {
		runLdBar.Stop()
		return false
	}

	runLdBar.Stop()
	return true
}

func CheckUserInformation() (string, string, string, string, bool) {
	var (
		userName  string
		usrName   string
		userEmail string
		usrEmail  string
	)

	consoleReader := bufio.NewScanner(os.Stdin)
	fmt.Print(" User name: ")
	consoleReader.Scan()
	userName = consoleReader.Text()

	if userName == "" {
		usrName = "Unknown User"
	} else {
		usrName = userName
	}

	var tryLoop = 0
	for tryLoop < 3 {
		tryLoop++
		fmt.Print(" User email: ")
		consoleReader.Scan()
		userEmail = consoleReader.Text()
		if userEmail == "" {
			usrEmail = "No Email Address"
			break
		} else if strings.Count(userEmail, "@") == 1 && len(strings.Split(userEmail, "@")[0]) > 0 &&
			strings.Count(userEmail, ".") > 0 && len(strings.Split(userEmail, "@")[0]) > 0 &&
			len(strings.Split(strings.Split(userEmail, "@")[1], ".")[0]) > 1 &&
			len(strings.Split(strings.Split(userEmail, "@")[1], ".")[1]) > 1 {
			usrEmail = userEmail
			break
		} else {
			ClearLine(1)
			AlertLine("Sorry, try again")
			if tryLoop >= 3 {
				fmt.Println(errors.New(lstDot + "3 incorrect email format attempts."))
				return "", "", "", "", false
			} else {
				fmt.Println(errors.New(lstDot + "Invalid email address format."))
			}
		}
	}
	if tryLoop == 1 {
		ClearLine(3)
	} else {
		ClearLine(tryLoop*2 + 1)
	}
	return userName, usrName, userEmail, usrEmail, true
}

func CheckPassword() (string, bool) {
	var tryLoop = 0
	for tryLoop < 3 {
		fmt.Print(" Password:")
		bytePw, _ := term.ReadPassword(0)

		runLdBar.Suffix = " Checking password... "
		runLdBar.Start()

		tryLoop++
		strPw := string(bytePw)
		inputPw := exec.Command("echo", strPw)
		checkPw := exec.Command(cmdAdmin, "-Sv")
		checkPw.Env = os.Environ()
		checkPw.Stdout = os.Stdout
		checkPw.Stdin, _ = inputPw.StdoutPipe()
		_ = checkPw.Start()
		_ = inputPw.Run()
		errSudo := checkPw.Wait()
		if errSudo != nil {
			runLdBar.FinalMSG = fntRed + " Password check failed" + fntReset + "\n"
			runLdBar.Stop()
			if tryLoop >= 3 {
				fmt.Println(errors.New(lstDot + "3 incorrect password attempts."))
			} else {
				fmt.Println(errors.New(lstDot + "Sorry, try again."))
			}
		} else {
			runLdBar.FinalMSG = ""
			runLdBar.Stop()
			if tryLoop == 1 {
				ClearLine(tryLoop - 1)
			} else {
				ClearLine(tryLoop*2 - 2)
			}
			return strPw, true
		}
	}
	return "", false
}

func NeedPermission(strPw string) {
	inputPw := exec.Command("echo", strPw)
	checkPw := exec.Command(cmdAdmin, "-Sv")
	checkPw.Env = os.Environ()
	checkPw.Stdout = os.Stdout

	checkPw.Stdin, _ = inputPw.StdoutPipe()
	_ = checkPw.Start()
	_ = inputPw.Run()
	errSudo := checkPw.Wait()
	CheckError(errSudo, "Failed to run root permission")

	runRoot := exec.Command(cmdAdmin, "whoami")
	runRoot.Env = os.Environ()
	whoAmI, _ := runRoot.Output()

	if string(whoAmI) != "root\n" {
		msg := "Incorrect user, please check permission of sudo.\n" +
			lstDot + "It need sudo command of \"" + fntRed + "root" + fntReset + "\" user's permission.\n" +
			lstDot + "Working username: " + string(whoAmI)
		MessageError("fatal", msg, "User")
	}
}

func RebootOS(adminCode string) {
	runLdBar.Suffix = " Restarting OS, please wait a moment ... "
	runLdBar.Start()

	NeedPermission(adminCode)
	reboot := exec.Command(cmdAdmin, "shutdown", "-r", "now")
	time.Sleep(time.Second * 3)
	if err := reboot.Run(); err != nil {
		runLdBar.FinalMSG = fntRed + "Error: " + fntReset
		runLdBar.Stop()
		fmt.Println(errors.New(lstDot + "Failed to reboot Operating System"))
	}

	runLdBar.FinalMSG = "⣾ Restarting OS, please wait a moment ... "
	runLdBar.Stop()
}

func ClearLine(line int) {
	for clear := 0; clear < line; clear++ {
		fmt.Printf("\033[1A\033[K")
	}
}

func TitleLine(msg string) {
	fmt.Println(fntBold + fntCyan + " " + msg + fntReset)
}

func AlertLine(msg string) {
	fmt.Println(errors.New(fntBold + fntRed + " " + msg + fntReset))
}

func NetHTTP(urlPath string) string {
	resp, err := http.Get(urlPath)
	CheckError(err, "Failed to connect "+urlPath)

	defer func() {
		errBodyClose := resp.Body.Close()
		CheckError(errBodyClose, "Failed to download from "+urlPath)
	}()

	rawFile, err := io.ReadAll(resp.Body)
	CheckError(err, "Failed to read file information from "+urlPath)
	return string(rawFile)
}

//func NetJSON(urlPath, key string) string {
//	resp, err := http.Get(urlPath)
//	CheckError(err, "Failed to connect "+urlPath)
//
//	defer func() {
//		errBodyClose := resp.Body.Close()
//		CheckError(errBodyClose, "Failed to download from "+urlPath)
//	}()
//
//	jsonFile, err := io.ReadAll(resp.Body)
//	CheckError(err, "Failed to read file information from "+urlPath)
//
//	var res map[string]interface{}
//	errMarshal := json.Unmarshal(jsonFile, &res)
//	CheckError(errMarshal, "Failed to parse JSON file from "+urlPath)
//	return res[key].(string)
//}

func HomeDirectory() string {
	homeDirPath, err := os.UserHomeDir()
	CheckError(err, "Failed to get home directory")
	return homeDirPath + "/"
}

func WorkingDirectory() string {
	workingDirPath, err := os.Getwd()
	CheckError(err, "Failed to get working directory")
	return workingDirPath + "/"
}

func CurrentUsername() string {
	workingUser, err := user.Current()
	CheckError(err, "Failed to get current user")
	return workingUser.Username
}

func MakeDirectory(dirPath string) {
	if CheckExists(dirPath) != true {
		err := os.MkdirAll(dirPath, 0755)
		CheckError(err, "Failed to make directory")
	}
}

func CopyDirectory(srcPath, dstPath string) {
	if CheckExists(dstPath) != true {
		cpDir := exec.Command("cp", "-rf", srcPath, dstPath)
		cpDir.Stderr = os.Stderr
		err := cpDir.Run()
		CheckError(err, "Failed to copy directory from \""+srcPath+"\" to \""+dstPath+"\"")
	}
}

func MakeFile(filePath, fileContents string, fileMode int) {
	targetFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(fileMode))
	CheckError(err, "Failed to get file information to make new file from \""+filePath+"\"")

	defer func() {
		err := targetFile.Close()
		CheckError(err, "Failed to finish make file to \""+filePath+"\"")
	}()

	_, err = targetFile.Write([]byte(fileContents))
	CheckError(err, "Failed to fill in information to \""+filePath+"\"")
}

func CopyFile(srcPath, dstPath string) {
	srcFile, err := os.Open(srcPath)
	CheckError(err, "Failed to get file information to copy from \""+srcPath+"\"")
	dstFile, err := os.Create(dstPath)
	CheckError(err, "Failed to get file information to copy to \""+dstPath+"\"")

	defer func() {
		errSrcFileClose := srcFile.Close()
		CheckError(errSrcFileClose, "Failed to finish copy file from \""+srcPath+"\"")
		errDstFileClose := dstFile.Close()
		CheckError(errDstFileClose, "Failed to finish copy file to \""+dstPath+"\"")
	}()

	_, errCopy := io.Copy(dstFile, srcFile)
	CheckError(errCopy, "Failed to copy file from \""+srcPath+"\" to \""+dstPath+"\"")
	errSync := dstFile.Sync()
	CheckError(errSync, "Failed to sync file from \""+srcPath+"\" to \""+dstPath+"\"")
}

func RemoveFile(filePath string) {
	if CheckExists(filePath) == true {
		err := os.Remove(filePath)
		CheckError(err, "Failed to remove file \""+filePath+"\"")
	}
}

func LinkFile(srcPath, dstPath, linkType, permission, adminCode string) {
	if linkType == "hard" {
		if permission == "root" || permission == "sudo" || permission == "admin" {
			NeedPermission(adminCode)
			lnFile := exec.Command(cmdAdmin, "ln", "-sfn", srcPath, dstPath)
			lnFile.Stderr = os.Stderr
			err := lnFile.Run()
			CheckCmdError(err, "Add failed to hard link file", "\""+srcPath+"\"->\""+dstPath+"\"")
		} else {
			if CheckExists(srcPath) == true {
				if CheckExists(dstPath) == true {
					RemoveFile(dstPath)
				}
				errHardlink := os.Link(srcPath, dstPath)
				CheckCmdError(errHardlink, "Add failed to hard link", "\""+srcPath+"\"->\""+dstPath+"\"")
			}
		}
	} else if linkType == "symbolic" {
		if permission == "root" || permission == "sudo" || permission == "admin" {
			NeedPermission(adminCode)
			lnFile := exec.Command(cmdAdmin, "ln", "-sfn", srcPath, dstPath)
			lnFile.Stderr = os.Stderr
			err := lnFile.Run()
			CheckCmdError(err, "Add failed to symbolic link", "\""+srcPath+"\"->\""+dstPath+"\"")
		} else {
			if CheckExists(srcPath) == true {
				if CheckExists(dstPath) == true {
					RemoveFile(dstPath)
				}
				errSymlink := os.Symlink(srcPath, dstPath)
				CheckCmdError(errSymlink, "Add failed to symbolic link\"", srcPath+"\"->\""+dstPath+"\"")
				errLinkOwn := os.Lchown(dstPath, os.Getuid(), os.Getgid())
				CheckError(errLinkOwn, "Failed to change ownership of symlink \""+dstPath+"\"")
			}
		}
	} else {
		MessageError("fatal", "Invalid link type", "Link file")
	}
}

func AppendFile(filePath, fileContents string, fileMode int) {
	targetFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.FileMode(fileMode))
	CheckError(err, "Failed to get file information to append contents from \""+filePath+"\"")

	defer func() {
		err := targetFile.Close()
		CheckError(err, "Failed to finish append contents to \""+filePath+"\"")
	}()

	_, err = targetFile.Write([]byte(fileContents))
	CheckError(err, "Failed to append contents to \""+filePath+"\"")
}

func DownloadFile(filePath, urlPath string, fileMode int) {
	MakeFile(filePath, NetHTTP(urlPath), fileMode)
}

func ASDFReshim(asdfPath string) {
	reshim := exec.Command(asdfPath, "reshim")
	err := reshim.Run()
	CheckCmdError(err, "ASDF failed to", "reshim")
}

func ASDFInstall(asdfPath, asdfPlugin, asdfVersion string) {
	insLdBar.Suffix = " ASDF-VM is installing " + asdfPlugin + " ... "
	insLdBar.Start()

	if CheckExists(HomeDirectory()+".asdf/plugins/"+asdfPlugin) != true {
		asdfPluginAdd := exec.Command(asdfPath, "plugin", "add", asdfPlugin)
		err := asdfPluginAdd.Run()
		CheckCmdError(err, "ASDF-VM failed to add", asdfPlugin)
	}

	ASDFReshim(asdfPath)
	asdfIns := exec.Command(asdfPath, optIns, asdfPlugin, asdfVersion)
	asdfIns.Env = os.Environ()
	errIns := asdfIns.Run()
	CheckCmdError(errIns, "ASDF-VM", asdfPlugin)

	asdfGlobal := exec.Command(asdfPath, "global", asdfPlugin, asdfVersion)
	asdfGlobal.Env = os.Environ()
	errConf := asdfGlobal.Run()
	CheckCmdError(errConf, "ASDF-VM failed to install", asdfPlugin)

	insLdBar.Stop()
}

func ASDFSet(asdfPath string) {
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

	ASDFInstall(asdfPath, "perl", "latest")
	ASDFInstall(asdfPath, "ruby", "latest")
	ASDFInstall(asdfPath, "python", "latest")
	ASDFInstall(asdfPath, "java", "openjdk-11.0.2")
	ASDFInstall(asdfPath, "java", "openjdk-17.0.2")
	ASDFInstall(asdfPath, "rust", "latest")
	ASDFInstall(asdfPath, "golang", "latest")
	ASDFInstall(asdfPath, "lua", "latest")
	ASDFInstall(asdfPath, "nodejs", "latest")
	ASDFInstall(asdfPath, "dart", "latest")
	ASDFInstall(asdfPath, "php", "latest")
	ASDFInstall(asdfPath, "groovy", "latest")
	ASDFInstall(asdfPath, "kotlin", "latest")
	ASDFInstall(asdfPath, "scala", "latest")
	ASDFInstall(asdfPath, "clojure", "latest")
	ASDFInstall(asdfPath, "erlang", "latest")
	ASDFInstall(asdfPath, "elixir", "latest")
	ASDFInstall(asdfPath, "gleam", "latest")
	ASDFInstall(asdfPath, "haskell", "latest")
	ASDFReshim(asdfPath)
}

func DockerStatus(dockerPath string) int {
	if CheckExists(dockerPath) != true {
		return 1
	}

	dockerSts := exec.Command(MacDockerPath(), "version")
	var dockerErr bytes.Buffer
	dockerSts.Stderr = &dockerErr
	if err := dockerSts.Run(); err != nil {
		errStr := dockerErr.String()
		if strings.Contains(errStr, "Cannot connect to the docker daemon") == true {
			return 1
		} else if strings.Contains(errStr, "Error response from daemon: dial unix docker.raw.sock: connect: no such file or director") == true {
			return 2
		}
	}
	return 0
}

func DockerInstall(dockerPath, dockerImage string) {
	insLdBar.Suffix = " Docker is downloading " + dockerImage + " image ... "
	insLdBar.Start()

	startDocker := exec.Command(dockerPath, "pull", dockerImage)
	err := startDocker.Run()
	CheckCmdError(err, "Docker failed to pull", dockerImage)

	insLdBar.Stop()
}

func DockerSet(dockerPath string) bool {
dockerStatus:
	runLdBar.Suffix = " Checking Docker status, please wait a moment ... "
	runLdBar.Start()

	errSts := DockerStatus(dockerPath)
	runLdBar.Stop()
	if errSts != 0 {

		AlertLine("Stopped installation Docker images")

		if errSts == 1 {
			fmt.Println(errors.New(lstDot + "Docker isn't initially started, please start docker and try again."))
		} else if errSts == 2 {
			fmt.Println(errors.New(lstDot + "Docker isn't running, please start docker and try again"))
		} else if errSts == 3 {
			fmt.Println(errors.New(lstDot + "Docker is preparing, please wait a moment and try again."))
		}

		fmt.Print(" If you wish to skip type (Skip) press return, else retry: ")
		var alertAnswer string
		_, errG4sOpt := fmt.Scanln(&alertAnswer)
		if errG4sOpt != nil {
			alertAnswer = "Enter"
		}
		if alertAnswer == "Skip" {
			ClearLine(3)
			AlertLine("Skipped Download docker Images")
			fmt.Println(errors.New(lstDot + "Please manually pull docker images."))
			return false
		}
		ClearLine(3)
		goto dockerStatus
	}

	DockerInstall(dockerPath, "alpine")
	DockerInstall(dockerPath, "ubuntu")
	DockerInstall(dockerPath, "httpd")
	DockerInstall(dockerPath, "nginx")
	DockerInstall(dockerPath, "tomcat")
	DockerInstall(dockerPath, "redis")
	DockerInstall(dockerPath, "mysql")
	DockerInstall(dockerPath, "mariadb")
	DockerInstall(dockerPath, "postgres")
	if osType == "darwin" {
		DockerInstall(dockerPath, "docker")
		DockerInstall(dockerPath, "debian")
		DockerInstall(dockerPath, "ubuntu")
		DockerInstall(dockerPath, "centos")
		DockerInstall(dockerPath, "fedora")
		DockerInstall(dockerPath, "archlinux")
		DockerInstall(dockerPath, "kalilinux/kali-rolling")
		DockerInstall(dockerPath, "wordpress")
		DockerInstall(dockerPath, "drupal")
		DockerInstall(dockerPath, "ghost")
	}
	return true
}

func Alias4shSet() {
	insLdBar.Suffix = " macOS is installing Alias4sh ... "
	insLdBar.Start()

	a4sPath := HomeDirectory() + ".config/alias4sh"
	MakeDirectory(a4sPath)
	MakeFile(a4sPath+"/alias4.sh", "# ALIAS4SH", 0644)

	dlA4sPath := WorkingDirectory() + ".ceios-alias4sh.sh"
	DownloadFile(dlA4sPath, "https://raw.githubusercontent.com/leelsey/Alias4sh/main/install.sh", 0644)

	installA4s := exec.Command("sh", dlA4sPath)
	if err := installA4s.Run(); err != nil {
		RemoveFile(dlA4sPath)
		CheckError(err, "Failed to install Alias4sh")
	}

	RemoveFile(dlA4sPath)
	insLdBar.Stop()
}

func Git4shSet(gitUserName, gitUserEmail string) {
	setGitUserName := exec.Command(macGit, "config", "--global", "user.name", gitUserName)
	errGitUserName := setGitUserName.Run()
	CheckError(errGitUserName, "Failed to set git user name")
	setGitUserEmail := exec.Command(macGit, "config", "--global", "user.email", gitUserEmail)
	errGitUserEmail := setGitUserEmail.Run()
	CheckError(errGitUserEmail, "Failed to set git user email")

	setGitBranch := exec.Command(macGit, "config", "--global", "init.defaultBranch", "main")
	errGitBranch := setGitBranch.Run()
	CheckError(errGitBranch, "Failed to change branch default name (master -> main)")

	setGitColor := exec.Command(macGit, "config", "--global", "color.ui", "true")
	errGitColor := setGitColor.Run()
	CheckError(errGitColor, "Failed to setup colourising")

	setGitEditor := exec.Command(macGit, "config", "--global", "core.editor", "vi")
	errGitEditor := setGitEditor.Run()
	CheckError(errGitEditor, "Failed to setup editor vi (vim)")

	ignoreDirPath := HomeDirectory() + ".config/git/"
	ignorePath := ignoreDirPath + "gitignore_global"
	MakeDirectory(ignoreDirPath)
	DownloadFile(ignorePath, "https://raw.githubusercontent.com/leelsey/Git4set/main/gitignore-sample", 0644)
	setExcludesFile := exec.Command(macGit, "config", "--global", "core.excludesfile", ignorePath)
	errExcludesFile := setExcludesFile.Run()
	CheckError(errExcludesFile, "Failed to set git global ignore file")
}

func main() {
	fmt.Println(fntBlue + "\n\t   ____________________  _____    ____  _____\n" +
		"\t  / ____/ ____/  _/ __ \\/ ___/   / __ \\/ ___/\n" +
		"\t / /   / __/  / // / / /\\__ \\   / / / /\\__ \\ \n" +
		"\t/ /___/ /____/ // /_/ /___/ /  / /_/ /___/ / \n" +
		"\t\\____/_____/___/\\____//____/   \\____//____/\n" + fntReset +
		"\t   C E I O S  O S  -  B Y  L E E L S E Y\n" +
		"     C Y B E R S E C U R I T Y  O P E R A T I O N S  O S\n" +
		fntGrey + "\t\t\t Version " + appVer + "\n" +
		"\t\t    contact@leelsey.com\n" + fntReset +
		" ------------------------------------------------------------")

	if CurrentUsername() == "root" {
		AlertLine("Security Problem")
		fmt.Println(errors.New(lstDot + "Don't run this as root user"))
		return
	}

	if archType != "arm64" && archType != "amd64" {
		AlertLine("Architecture Problem")
		fmt.Println(errors.New(lstDot + "This OS only supports arm64 and amd64"))
		return
	}

	TitleLine("Need Permission")
	if adminCode, adminSts := CheckPassword(); adminSts == true {
		ClearLine(1)
		NeedPermission(adminCode)
		if osType == "darwin" {
			if CEIOS4macOS(adminCode) != true {
				return
			}
		} else if osType == "linux" {
			//if CEIOS4Kali(adminCode) != true {
			//	goto exitPoint
			//}
		} else {
			AlertLine("Operating System Problem")
			fmt.Println(errors.New(lstDot + "Unsupported operating system" + fntReset))
			return
		}
		RebootOS(adminCode)
	} else {
		return
	}
	return
}
