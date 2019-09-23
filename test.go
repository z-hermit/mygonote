package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

type Integer struct {
	Number int
}

func main() {
	time.Ticker{}
}

func byteconvert_test() {
	count := 0
	//s := make([][]byte, 0, 0)
	m := make([]MyStruct, 0, 0)
	for count < 1000000 {
		//s = append(s, MyStructToBytes(&MyStruct{}))
		m = append(m, MyStruct{})
		count++
	}
	//time.Sleep(time.Second)
}

type MyStruct struct {
	A int
	B int
}

var sizeOfMyStruct = int(unsafe.Sizeof(MyStruct{}))

func MyStructToBytes(s *MyStruct) []byte {
	var x reflect.SliceHeader
	x.Len = sizeOfMyStruct
	x.Cap = sizeOfMyStruct
	x.Data = uintptr(unsafe.Pointer(s))
	return *(*[]byte)(unsafe.Pointer(&x))
}

func BytesToMyStruct(b []byte) *MyStruct {
	return (*MyStruct)(unsafe.Pointer(
		(*reflect.SliceHeader)(unsafe.Pointer(&b)).Data))
}

type UpperWriter struct {
	io.Writer
}

// 有没有效果是一样的，区别是p.Writer有没有，p.Writer.Write是一定要动态绑定的
//func (p *UpperWriter) Write(data []byte) (n int, err error) {
//	return p.Writer.Write(bytes.ToUpper(data))
//}

//UpperWriter 虚继承了io.Writer 有write方法，在传入os.Stdout时动态绑定write方法
func virtualfunc_test() {
	fmt.Fprintln(&UpperWriter{os.Stdout}, "hello, world")
}

func check_string_type() {
	s := "hello, world"
	fmt.Println("len(s):", (*reflect.StringHeader)(unsafe.Pointer(&s)).Len)
	fmt.Printf("len(s):%#v", []byte(s))
}

func nil_slice_test() {
	var a []int // nil切片, 和 nil 相等, 一般用来表示一个不存在的切片
	var b = []int{}
	pa := (*reflect.SliceHeader)(unsafe.Pointer(&a))
	pb := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	fmt.Println(pa)
	fmt.Printf("%#v", pb)
}

func slice_slip_test() {
	s := make([]int, 1, 10)
	fmt.Println(s)
	//fmt.Println(s[2])  //index out of range
	s1 := s[2:10]
	//s1 := s[2:11] //slice bounds out of range
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	s1h := (*reflect.SliceHeader)(unsafe.Pointer(&s1))
	fmt.Printf("%#v\n", sh)
	fmt.Printf("%#v\n", s1)
	fmt.Printf("%#v\n", s1h)

}

func net_test() {
	//net.Dial()
}

func interface_test() {
	var b1 bee
	var b2 *bee
	b2 = new(bee)
	b1.print()
	b1.bprint()
	b2.print()
	b2.bprint()
}

type bee int

type b interface {
	print()
}

func (ab *bee) print() {
	fmt.Println("a bee")
}

func (ab bee) bprint() {
	fmt.Println("a bee")
}

//每个case都必须是一个通信
//所有channel表达式都会被求值
//如果有一或多个case都可以运行，Select会随机公平地选出一个执行。其他不会执行。
//否则：如果有default子句，则执行该语句。
//如果没有default字句，select将阻塞，直到某个通信可以运行；Go不会重新对channel或值进行求值。
func select_timeafter_test() {
	ch := make(chan int, 10)
	go select_count(ch)
	for {
		select {
		case i := <-ch:
			fmt.Println(i)
			time.Sleep(time.Second)
		case <-select_timeAfter():
			fmt.Println("time after 3s")
			return
		}

	}

	// final solution
	//ticker := time.NewTimer(2 * time.Second)
	//	defer ticker.Stop()
	//	ch := make(chan int, 10)
	//	go add(ch)
	//	for {
	//		select {
	//		case <- ch:
	//			fmt.Println(ch) // if ch not empty, time.After will nerver exec
	//			fmt.Println("sleep one seconds ...")
	//			time.Sleep(1 * time.Second)
	//			fmt.Println("sleep one seconds end...")
	//		default: // forbid block
	//		}
	//		select {
	//		case <- ticker.C:
	//			fmt.Println("timeout")
	//			return
	//		default: // forbid block
	//		}
	//	}
}

func select_count(ch chan int) {
	for i := 0; i < 10; i++ {
		ch <- i
	}
}

func select_timeAfter() <-chan time.Time {
	fmt.Println("new time after")
	return time.After(time.Second * 3)
}

//主线程停止了，goroutine也就停止了
func goroutin_test() {
	c := make(chan int)
	go func() {
		for {
			fmt.Println("hhh")
		}
	}()
	<-c
}

var (
	count int32
	wg    sync.WaitGroup
	mutex sync.Mutex
)

func lock_test() {
	wg.Add(2)
	go incCount()
	go incCount()
	wg.Wait()
	fmt.Println(count)
}

func incCount() {
	defer wg.Done()
	for i := 0; i < 2; i++ {
		mutex.Lock()
		value := count
		runtime.Gosched()
		value++
		count = value
		mutex.Unlock()
	}
}

func select_test() {
	c := make(chan int)
	o := make(chan bool)
	go func() {
		for {
			select {
			case v := <-c:
				println(v)
			case <-time.After(5 * time.Second):
				println("timeout")
				o <- true
				break
			}
		}
	}()
	<-o
}

func time_test() {
	const base_format = "2006-01-02 15:04:05"
	fmt.Println(time.Now().Format(base_format))
}

func string_test() {
	test := "abbcd"
	fmt.Println(test[:len(test)-1])
}

//slice 传值，但底层为长度，容量，数组指针，所以类似传引用，更改新的旧的也会变，扩展方式类似stl vector
func slice_test() {
	//testSlice := []int64{0}
	//fmt.Println(testSlice)
	//a := testSlice
	//a[0] = 1
	//fmt.Println(testSlice)
	//a = append(a, 2)
	//fmt.Println(testSlice)
	//a[0] = 4
	//fmt.Println(testSlice)
	//b := &testSlice
	//*b = append(*b, 3)
	//fmt.Println(testSlice)

	testSlice1 := make([]int, 0, 2)
	fmt.Println(testSlice1)
	a1 := testSlice1
	a1 = append(a1, 1)
	fmt.Println(a1)
	fmt.Println(testSlice1)
	testSlice1 = append(testSlice1, 2)
	testSlice1 = append(testSlice1, 3)
	fmt.Println(testSlice1)
	fmt.Println(a1)

	//i := []int{}
	//fmt.Println(len(i), cap(i)) //0  0
	//i = append(i, 1)
	//fmt.Println(len(i), cap(i)) //1  1
}

func valueORquote() {
	value := []Integer{{4}, {5}, {6}}
	value1 := value
	value1[0] = Integer{1}
	fmt.Println(value, value1)
}

//struct 传值，string值不可改immutable
func valueORquoteStruct() {
	value := Integer{1}
	value1 := value
	value1.Number = 2
	fmt.Println(value, value1)
}

//map，channel传指针
func valueORquoteMap() {
	value := map[int]Integer{1: {4}, 2: {5}, 3: {6}}
	value1 := value
	value1[0] = Integer{1}
	fmt.Println(value, value1)
}

//for range 踩坑
//for range slice会创建元素副本，而不是直接使用引用，当做引用进行&复制就会出错，最后都得到最后一次循环的值,修改也不会影响原值
func forrang_test() {
	value := []Integer{{4}, {5}, {6}}
	//myMap := make(map[int]Integer)
	for _, v := range value {
		v = Integer{1}
		fmt.Println(v)
	}
	fmt.Println(value)
}

//defer,调用顺序类似栈
//defer 调用的函数参数的值在 defer 定义时就确定了, 而 defer 函数内部所使用的变量的值需要在这个函数运行时才确定
func defer_test1() {
	i := 1
	defer fmt.Println("Deferred print:", i)
	i++
	fmt.Println("Normal print:", i)
}

//defer调用时机: 外层函数设置返回值之后, 并且在即将返回之前
func defer_test2() (r int) {
	defer func() {
		r++
	}()
	defer func() {
		r++
	}()
	return 0
	//equal
	//r = 0
	//return
}

//func (u *User) Update()只能指针调用，func (u User) Update()既能指针调用，也可以直接调用
//指针赋值需要显式赋值，调用可以隐式调用
type User struct {
}

type u interface {
	Update()
}

func fff() {
	var test u = &User{}
	//var test u = User{}  //wrong
	test.Update()
	(&User{}).Update()
}

func (u *User) Update() {
}

//文件init
//同一个package中，不同文件中的init方法的执行按照文件名先后执行各个文件中的init方法
//在同一个文件中的多个init方法，按照在代码中编写的顺序依次执行不同的init方法
//不同package中，根据依赖顺序init
