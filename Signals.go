package main

import (
	"syscall"
)

const (
	//SIGHUP terminal line hangup
	SIGHUP int = iota + 1
	//SIGINT interrupt program
	SIGINT
	//SIGQUIT quit program
	SIGQUIT
	//SIGILL illegal instruction
	SIGILL
	//SIGTRAP trace trap
	SIGTRAP
	//SIGABRT abort program (formerly SIGIOT)
	SIGABRT
	//SIGEMT emulate instruction executed
	SIGEMT
	//SIGFPE floating-point exception
	SIGFPE
	//SIGKILL kill program
	SIGKILL
	//SIGBUS bus error
	SIGBUS
	//SIGSEGV segmentation violation
	SIGSEGV
	//SIGSYS non-existent system call invoked
	SIGSYS
	//SIGPIPE write on a pipe with no reader
	SIGPIPE
	//SIGALRM real-time timer expired
	SIGALRM
	//SIGTERM software termination signal
	SIGTERM
	//SIGURG urgent condition present on socket
	SIGURG
	//SIGSTOP stop (cannot be caught or ignored)
	SIGSTOP
	//SIGTSTP stop signal generated from keyboard
	SIGTSTP
	//SIGCONT continue after stop
	SIGCONT
	//SIGCHLD child status has changed
	SIGCHLD
	//SIGTTIN background read attempted from control terminal
	SIGTTIN
	//SIGTTOU background write attempted to control terminal
	SIGTTOU
	//SIGIO I/O is possible on a descriptor (see fcntl(2))
	SIGIO
	//SIGXCPU cpu time limit exceeded (see setrlimit(2))
	SIGXCPU
	//SIGXFSZ file size limit exceeded (see setrlimit(2))
	SIGXFSZ
	//SIGVTALRM virtual time alarm (see setitimer(2))
	SIGVTALRM
	//SIGPROF profiling timer alarm (see setitimer(2))
	SIGPROF
	//SIGWINCH Window size change
	SIGWINCH
	//SIGINFO status request from keyboard
	SIGINFO
	//SIGUSR1 User defined signal 1
	SIGUSR1
	//SIGUSR2 User defined signal 2
	SIGUSR2
)

var Signals = map[string]syscall.Signal{
	"SIGHUP":    syscall.Signal(SIGHUP),
	"SIGINT":    syscall.Signal(SIGINT),
	"SIGQUIT":   syscall.Signal(SIGQUIT),
	"SIGILL":    syscall.Signal(SIGILL),
	"SIGTRAP":   syscall.Signal(SIGTRAP),
	"SIGABRT":   syscall.Signal(SIGABRT),
	"SIGEMT":    syscall.Signal(SIGEMT),
	"SIGFPE":    syscall.Signal(SIGFPE),
	"SIGKILL":   syscall.Signal(SIGKILL),
	"SIGBUS":    syscall.Signal(SIGBUS),
	"SIGSEGV":   syscall.Signal(SIGSEGV),
	"SIGSYS":    syscall.Signal(SIGSYS),
	"SIGPIPE":   syscall.Signal(SIGPIPE),
	"SIGALRM":   syscall.Signal(SIGALRM),
	"SIGTERM":   syscall.Signal(SIGTERM),
	"SIGURG":    syscall.Signal(SIGURG),
	"SIGSTOP":   syscall.Signal(SIGSTOP),
	"SIGTSTP":   syscall.Signal(SIGTSTP),
	"SIGCONT":   syscall.Signal(SIGCONT),
	"SIGCHLD":   syscall.Signal(SIGCHLD),
	"SIGTTIN":   syscall.Signal(SIGTTIN),
	"SIGTTOU":   syscall.Signal(SIGTTOU),
	"SIGIO":     syscall.Signal(SIGIO),
	"SIGXCPU":   syscall.Signal(SIGXCPU),
	"SIGXFSZ":   syscall.Signal(SIGXFSZ),
	"SIGVTALRM": syscall.Signal(SIGVTALRM),
	"SIGPROF":   syscall.Signal(SIGPROF),
	"SIGWINCH":  syscall.Signal(SIGWINCH),
	"SIGINFO":   syscall.Signal(SIGINFO),
	"SIGUSR1":   syscall.Signal(SIGUSR1),
	"SIGUSR2":   syscall.Signal(SIGUSR2),
}
