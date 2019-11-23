-   基于 go 1.13.3

*   runtime/netpoll_epoll.go

netpollinit 创建 epoll 对象（整个进程只有一个实例） 使用的 epoll_create，epoll_create 等价于 glibc 的 epoll_create1 和 epoll_create 函数

netpollopen 将 fd 添加到 epoll (每个 fd 一开始都关注了读写事件，并且采用的是边缘触发，除此之外还关注了一个不常见的新事件 EPOLLRDHUP，这个事件是在较新的内核版本添加的，目的是解决对端 socket 关闭，epoll 本身并不能直接感知到这个关闭动作的问题。注意任何一个 fd 在添加到 epoll 中的时候就关注了 EPOLLOUT 事件的话，就立马产生一次写事件，这次事件可能是多余浪费的。)

netpollclose 将 fd 从 epoll 删除

netpoll(block) 得到所有发生事件的 fd, if block , timeout = max, else timeout = 0, 将每个 fd 对应的 goroutine(用户态线程)通过链表返回

-   runtime/netpoll.go

poll_runtime_pollWait 函数：park 当前读写的 goroutine

核心数据结构

```
type pollDesc struct {
	link *pollDesc // in pollcache, protected by pollcache.lock

	// The lock protects pollOpen, pollSetDeadline, pollUnblock and deadlineimpl operations.
	// This fully covers seq, rt and wt variables. fd is constant throughout the PollDesc lifetime.
	// pollReset, pollWait, pollWaitCanceled and runtime·netpollready (IO readiness notification)
	// proceed w/o taking the lock. So closing, everr, rg, rd, wg and wd are manipulated
	// in a lock-free way by all operations.
	// NOTE(dvyukov): the following code uses uintptr to store *g (rg/wg),
	// that will blow up when GC starts moving objects.
	lock    mutex // protects the following fields
	fd      uintptr
	closing bool
	everr   bool    // marks event scanning error happened
	user    uint32  // user settable cookie
	rseq    uintptr // protects from stale read timers
	rg      uintptr // pdReady, pdWait, G waiting for read or nil
	rt      timer   // read deadline timer (set if rt.f != nil)
	rd      int64   // read deadline
	wseq    uintptr // protects from stale write timers
	wg      uintptr // pdReady, pdWait, G waiting for write or nil
	wt      timer   // write deadline timer
	wd      int64   // write deadline
}
```

对 fd 进行了封装，并且绑定了对应的 goroutine

-   poll/fd_poll_runtime.go

runtime 中的 epoll 事件驱动抽象层其实在进入 net 库后，又被封装了一次

```
type pollDesc struct {
	runtimeCtx uintptr
}
```

runtimeCtx 就是 runtime 中的 pollDesc

```
var serverInit sync.Once

func (pd *pollDesc) init(fd *FD) error {
	serverInit.Do(runtime_pollServerInit)
	ctx, errno := runtime_pollOpen(uintptr(fd.Sysfd))
	if errno != 0 {
		if ctx != 0 {
			runtime_pollUnblock(ctx)
			runtime_pollClose(ctx)
		}
		return errnoErr(syscall.Errno(errno))
	}
	pd.runtimeCtx = ctx
	return nil
}
```

通过 sync.Once, 只会调用一次 epoll 的初始化

-   poll/fd_unix.go

Accept 方法节选

```
for {
		s, rsa, errcall, err := accept(fd.Sysfd)
		if err == nil {
			return s, rsa, "", err
		}
		switch err {
		case syscall.EAGAIN:
			if fd.pd.pollable() {
				if err = fd.pd.waitRead(fd.isFile); err == nil {
					continue
				}
			}
		case syscall.ECONNABORTED:
			// This means that a socket on the listen
			// queue was closed before we Accept()ed it;
			// it's a silly error, so try again.
			continue
		}
		return -1, nil, errcall, err
	}
```

```
// Wrapper around the accept system call that marks the returned file
// descriptor as nonblocking and close-on-exec.
func accept(s int) (int, syscall.Sockaddr, string, error) {
	ns, sa, err := Accept4Func(s, syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC)
	// On Linux the accept4 system call was introduced in 2.6.28
	// kernel and on FreeBSD it was introduced in 10 kernel. If we
	// get an ENOSYS error on both Linux and FreeBSD, or EINVAL
	// error on Linux, fall back to using accept.
	switch err {
	case nil:
		return ns, sa, "", nil
	default: // errors other than the ones listed
		return -1, sa, "accept4", err
	case syscall.ENOSYS: // syscall missing
	case syscall.EINVAL: // some Linux use this instead of ENOSYS
	case syscall.EACCES: // some Linux use this instead of ENOSYS
	case syscall.EFAULT: // some Linux use this instead of ENOSYS
	}

	// See ../syscall/exec_unix.go for description of ForkLock.
	// It is probably okay to hold the lock across syscall.Accept
	// because we have put fd.sysfd into non-blocking mode.
	// However, a call to the File method will put it back into
	// blocking mode. We can't take that risk, so no use of ForkLock here.
	ns, sa, err = AcceptFunc(s)
	if err == nil {
		syscall.CloseOnExec(ns)
	}
	if err != nil {
		return -1, nil, "accept", err
	}
	if err = syscall.SetNonblock(ns, true); err != nil {
		CloseFunc(ns)
		return -1, nil, "setnonblock", err
	}
	return ns, sa, "", nil
}
```

可以看到 accept 方法返回的 fd 是 nonblocking and close-on-exec.

> 关于 close-on-exec：我们经常会碰到需要 fork 子进程的情况，而且子进程很可能会继续 exec 新的程序。这就不得不提到子进程中无用文件描述符的问题！
> fork 函数的使用本不是这里讨论的话题，但必须提一下的是：子进程以写时复制（COW，Copy-On-Write）方式获得父进程的数据空间、堆和栈副本，这其中也包括文件描述符。刚刚 fork 成功时，父子进程中相同的文件描述符指向系统文件表中的同一项（这也意味着他们共享同一文件偏移量）。 1.子进程继承的时候这些 fd 就是打开的，本身就有越权的行为可能存在，操作它不该操作的东西。 2.如果该 socketfd 被子进程继承并占用，或者未关闭，就会导致新的父进程重新启动时不能正常使用这些网络端口

Read 方法分析

```
for {
		n, err := syscall.Read(fd.Sysfd, p)
		if err != nil {
			n = 0
			if err == syscall.EAGAIN && fd.pd.pollable() {
				if err = fd.pd.waitRead(fd.isFile); err == nil {
					continue
				}
			}

			// On MacOS we can see EINTR here if the user
			// pressed ^Z.  See issue #22838.
			if runtime.GOOS == "darwin" && err == syscall.EINTR {
				continue
			}
		}
}
```

> EAGAIN:非阻塞(non-blocking)操作(对文件或 socket)的时候。例如，以 O_NONBLOCK 的标志打开文件/socket/FIFO，如果你连续做 read 操作而没有数据可读。此时程序不会阻塞起来等待数据准备就绪返回，read 函数会返回一个错误 EAGAIN，提示你的应用程序现在没有数据可读请稍后再试。

read 方法调用 syscall.Read 不会阻塞，而是在 EAGAIN 后 park goroutine，使得 M 不用阻塞，只是挂起 goroutine，这是高并发的关键

-   在上层会进入 net 包，由 netFD 在做一层封装

那 park 的 goroutine 是怎么被唤醒的呢？
通过上文中的 runtime.netpoll 进行唤醒，而 runtime.netpoll 所有的调用都在 runtime.proc 中

-   runtime.proc

The scheduler's job is to distribute ready-to-run goroutines over worker threads.

G, P, M 的设计文档可以这里查看https://golang.org/s/go11sched
