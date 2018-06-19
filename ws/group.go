package ws

// Group 用來管理一堆Hub
// 新增刪除搜尋hub
type Group struct {
	hubs          map[*Hub]bool
	addHubChan    chan *Hub
	findHubChan   chan int
	deleteHubChan chan int
}

var groupFindHubChan = make(chan *Hub)
var group *Group

// CreateGroup init Group
func CreateGroup() *Group {
	group = newGroup()
	go group.run()

	return group
}

func newGroup() *Group {
	return &Group{
		hubs:          make(map[*Hub]bool),
		addHubChan:    make(chan *Hub),
		findHubChan:   make(chan int),
		deleteHubChan: make(chan int),
	}
}

// 搜尋hub
// 回傳hub or nil
func (g *Group) findHub(ID int) {
	for hub, open := range g.hubs {
		if hub.id == ID && open {
			groupFindHubChan <- hub
			return
		}
	}

	groupFindHubChan <- nil
}

// 刪掉hub
func (g *Group) deleteHub(ID int) {
	for hub, _ := range g.hubs {
		// fmt.Println(open)
		if hub.id == ID {
			hub.destory <- true
			delete(g.hubs, hub)
		}
	}
}

func (g *Group) run() {
	for {
		select {
		case hubID := <-g.findHubChan:
			g.findHub(hubID)
		// 新增hub
		case hub := <-g.addHubChan:
			g.hubs[hub] = true
		// 刪hub
		case hubID := <-g.deleteHubChan:
			g.deleteHub(hubID)
		}
	}
}
