package container

import (
	"os/exec"
	"syscall"
	"os"
	log "github.com/sirupsen/logrus"
	"fmt"
)

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, err
	}
	return false, err
}

func NewParentProcess(tty bool,rootUrl string,MntUrl string)(*exec.Cmd,*os.File) {

	readPipe,writePipe,err:=os.Pipe()
	if(err!=nil){
    	log.Errorf("new Pipe error %v",err)
		return nil,nil
	}

    //os.Args[0]
    //path, err := os.Executable("/proc/self/exe")
    //path, err := os.Readlink("/proc/self/exe")
	cmd := exec.Command("/proc/self/exe","init")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty{
		cmd.Stdin=os.Stdin
		cmd.Stdout=os.Stdout
		cmd.Stderr=os.Stderr
	}

	cmd.ExtraFiles=[]*os.File{readPipe}

	if err:=newWorkSpace(rootUrl,MntUrl);err!=nil{
		log.Errorf("newWorkSpace error %v",err)
		return nil,nil
	}
	cmd.Dir=MntUrl

	return cmd,writePipe
}

func newWorkSpace(rootUrl string,MntUrl string) error{

    if err:=createReadOnlyLayer(rootUrl);err!=nil{
        return err
    }
	if err:=createWriteLayer(rootUrl);err!=nil{
	    return err
	}
	if err:=createMountPoint(rootUrl,MntUrl);err!=nil{
	    return err
	}
	return nil
}

func createReadOnlyLayer(rootUrl string)error{
    readlayer:=rootUrl+"busybox/"
	readlayerpath,err:=pathExist(readlayer)
	if err!=nil{
	    return err
	}
	if !readlayerpath{
	    return fmt.Errorf("readlayer(busybox) dir don't exist: %s", readlayer)
	}
	return nil
}

func createWriteLayer(rootUrl string)error{
    writelayer:=rootUrl+"writelayer/"
	if err:=os.Mkdir(writelayer,0777);err!=nil{
	    return fmt.Errorf("create write layer failed: %v", err)
	}
	return nil
}


func createMountPoint(rootUrl string,MntUrl string)error{

	log.Errorf("createMountPoint is running--------------------------")
	

	if err := os.Mkdir(MntUrl, 0777); err != nil {
		log.Errorf("Mkdir mntUrl failed: %v", err)
		return fmt.Errorf("mkdir faild: %v", err)
	}
	
	aPath:=rootUrl+"writelayer/"
	bPath:=rootUrl+"busybox/"
	cPath:=MntUrl

	log.Errorf("MntUrl is %s",MntUrl)


	upperDir, _ := exec.Command("mktemp", "-d").Output()
	workDir, _ := exec.Command("mktemp", "-d").Output()
	// 去除末尾的换行符
	upperDirPath := string(upperDir[:len(upperDir)-1])
	workDirPath := string(workDir[:len(workDir)-1])
	// 构建挂载命令
	cmd := exec.Command("sudo", "mount", "-t", "overlay",
		"overlay",
		"-o", fmt.Sprintf("lowerdir=%s:%s,upperdir=%s,workdir=%s", aPath, bPath, upperDirPath, workDirPath),
		cPath)

	// 构建命令
	//cmd := exec.Command("sudo", "mount", "-t", "overlay", "overlay",
	//"-o", fmt.Sprintf("lowerdir=%s:%s,upperdir=$(mktemp -d),workdir=$(mktemp -d),%s", aPath, bPath, MntUrl))

    //lowerDirs := "lowerdir=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
    //cmd := exec.Command("sudo","mount", "-t", "overlay", "-o", lowerDirs, "overlay", MntUrl)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err:=cmd.Run();err!=nil{
	     return fmt.Errorf("mmt dir err: %v", err)
	} 

	log.Errorf("createMountPoint is end---------------------")
	return nil
}

