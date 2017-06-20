package timer

import (
	"runtime"
	"sync"
	"time"
)

//时间轮数量
const wheel_cnt uint8 = 5

var (
	//每个时间轮的槽（元素）数量。在256 + 4 × 64 = 512个槽中，表示的范围是2^32
	element_cnt_per_wheel = [wheel_cnt]uint32{256, 64, 64, 64, 64}

	/*当指针指向当前时间轮的最后一位数，再走一位就需要向上进位。每个时间轮进位时，使用右移的方式，实现最快进位，
	  这里是指每个轮的进位二进制位数*/
	right_shift_per_wheel = [wheel_cnt]uint32{8, 6, 6, 6, 6}

	/*记录每个时间轮指针当前指向的位置*/
	base_per_wheel = [wheel_cnt]uint32{1, 256, 256 * 64, 256 * 64 * 64, 256 * 64 * 64 * 64}

	rwmutex sync.RWMutex

	//每个时间轮当前指针所指向的位置
	newset [wheel_cnt]uint32
	//定义5个时间轮
	timewheels [wheel_cnt][]*Node

	//保存待执行的定时器，方便按链表节点指针地址直接删除定时器
	TimerMap map[string]*Node = make(map[string]*Node)
)

type Timer struct {
	Name        string            //定时器名称
	Inteval     uint32            //时间间隔，即以插入该定时器的时间为起点，Inteval秒后执行回调函数DoSomeThing().
	DoSomething func(interface{}) //回调函数
	Args        interface{}       //回调函数所需参数
}

func init() {
	var bucket_no uint8 = 0
	for bucket_no = 0; bucket_no < wheel_cnt; bucket_no++ {
		var i uint32 = 0
		for ; i < element_cnt_per_wheel[bucket_no]; i++ {
			timewheels[bucket_no] = append(timewheels[bucket_no], new(Node))
		}
	}

	//开启定时器管理系统
	runtime.GOMAXPROCS(runtime.NumCPU())
	go Run()
}

func SetTimer(name string, inteval uint32, handler func(interface{}), args interface{}) {
	if inteval <= 0 {
		return
	}

	var bucket_no uint8 = 0
	var offset uint32 = inteval
	var left uint32 = inteval

	for offset >= element_cnt_per_wheel[bucket_no] { //偏移量大于当前时间轮容量, 则需要向高位进位
		//计算高位的值，偏移量除以低位的进制。比如当前低位是256，则右移8个二进制位，就是除以256，得到的就是高位的值
		offset >>= right_shift_per_wheel[bucket_no]
		var tmp uint32 = 1
		if bucket_no == 0 {
			tmp = 0
		}
		left -= base_per_wheel[bucket_no] * (element_cnt_per_wheel[bucket_no] - newset[bucket_no] - tmp)
		bucket_no++
	}

	if offset < 1 {
		return
	}

	if inteval < base_per_wheel[bucket_no]*offset {
		return
	}
	left -= base_per_wheel[bucket_no] * (offset - 1)
	//通过类似hash的方式，找到在时间轮上插入的位置
	pos := (newset[bucket_no] + offset) % element_cnt_per_wheel[bucket_no]

	var node Node
	node.SetData(Timer{name, left, handler, args})

	rwmutex.RLock()
	TimerMap[name] = timewheels[bucket_no][pos].InsertHead(node) //插入定时器
	rwmutex.RUnlock()
}

func DelTimer(name string) {
	rwmutex.RLock()
	Delete(TimerMap[name])
	rwmutex.RUnlock()
}

func step() {
	rwmutex.RLock()

	var bucket_no uint8 = 0

	//遍历所有的槽
	for bucket_no = 0; bucket_no < wheel_cnt; bucket_no++ {
		//当前指针递增1
		newset[bucket_no] = (newset[bucket_no] + 1) % element_cnt_per_wheel[bucket_no]

		//当前指针指向的槽位置的表头
		var head *Node = timewheels[bucket_no][newset[bucket_no]]
		var firstElement *Node = head.Next()
		for firstElement != nil { //链表不为空
			//如果element中确实存储了Timer类型的值，则ok==true
			if value, ok := firstElement.Data().(Timer); ok {
				inteval := value.Inteval
				doSomething := value.DoSomething
				args := value.Args
				if doSomething != nil {
					if bucket_no == 0 || inteval == 0 {
						go doSomething(args)
					} else {
						SetTimer(value.Name, inteval, doSomething, args) //重新插入计时器
					}
				}
				//删除定时器
				Delete(firstElement)
			}
			//重新定位到链表的第一个元素
			firstElement = head.Next()
		}
		//指针不是0，还未转回到原点， 跳出
		//如果回到原点，则说明转完了一圈，需要向高位进位1，则继续循环入高位步进一步
		if newset[bucket_no] != 0 {
			break
		}
	}

	rwmutex.RUnlock()
}

func Run() {
	var i int = 0
	for {
		go step()
		i++
		time.Sleep(1 * time.Second)
	}
}
