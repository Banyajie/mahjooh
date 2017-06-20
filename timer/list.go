package timer

type Node struct {
	data interface{}
	prev *Node
	next *Node
}

type LinkedList struct {
	head   *Node
	last   *Node
	length uint
}

func NewLinkedList() *LinkedList {
	var list *LinkedList = new(LinkedList)
	list.head = nil
	list.last = nil
	list.length = 0

	return list
}

/*Get the list head*/
func (this LinkedList) GetHead() *Node {
	return this.head
}

/*Get the list last*/
func (this LinkedList) GetLast() *Node {
	return this.last
}

func (this LinkedList) Length() uint {
	return this.length
}

/*Insert one node*/
func (this *LinkedList) PushBack(node Node) *Node {
	node.next = nil

	if nil == this.head { //空表
		this.head = &node
		this.head.prev = nil
		this.last = this.head
	} else {
		node.prev = this.last
		this.last.next = &node
		this.last = this.last.next
	}
	//fmt.Println("insert %d %d\n", this.length, this.last.data)
	this.length++

	return this.last
}

/*Erase one node from the listedlist*/
func (this *LinkedList) Erase(node *Node) {
	if nil == node {
		return
	} else if nil == node.next && nil == node.prev {
		return
	}

	if node == this.head && node == this.last {
		this.head = nil
		this.last = nil
		this.length = 0
	} else {
		if node == this.head {
			this.head = this.head.next
			if nil != this.head {
				this.head.prev = nil
			}
		} else if node == this.last {
			node.prev.next = nil
			this.last = node.prev
		} else {
			node.prev.next = node.next
			node.next.prev = node.prev
		}
	}

	this.length--
}

func Delete(node *Node) {
	if nil == node {
		return
	} else if nil == node.prev { //该元素处于表头，不删除，默认表头不存元素
		return
	} else if nil == node.next { //该元素处于表尾
		node.prev.next = nil
		node.prev = nil
	} else {
		node.next.prev = node.prev
		node.prev.next = node.next
		node.prev = nil
		node.next = nil
	}
}

/*Insert the node from the list head*/
func (head *Node) InsertHead(node Node) *Node {
	if nil == head || nil != head.prev { //表头为空或者不是表头
		return nil
	} else {
		if nil != head.prev {
			head.next.prev = &node
			node.next = head.next
		}
		head.next = &node
		node.prev = head
	}

	return &node
}

func (this *Node) Next() *Node {
	return this.next
}

func (this *Node) Prev() *Node {
	return this.prev
}

func (this *Node) Data() (data interface{}) {
	return this.data
}

func (this *Node) SetData(data interface{}) {
	this.data = data
}
