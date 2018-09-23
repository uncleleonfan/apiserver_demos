package service

/**
一般在 handler 中主要做解析参数、返回数据操作，简单的逻辑也可以在 handler 中做，像新增用户、删除用户、更新用户，代码量不大，
所以也可以放在 handler 中。有些代码量很大的逻辑就不适合放在 handler 中，因为这样会导致 handler 逻辑不是很清晰，这时候实际
处理的部分通常放在 service 包
*/
import (
	"fmt"
	"sync"

	"apiserver_demos/demo07/model"
	"apiserver_demos/demo07/util"
)

func ListUser(username string, offset, limit int) ([]*model.UserInfo, uint64, error) {
	//用户信息列表
	infos := make([]*model.UserInfo, 0)
	users, count, err := model.ListUser(username, offset, limit)
	if err != nil {
		return nil, count, err
	}

	//定义用户id列表
	ids := []uint64{}
	//获取所有的id列表，保存原来的顺序
	for _, user := range users {
		ids = append(ids, user.Id)
	}

	wg := sync.WaitGroup{}
	//使用Lock是因为在并发处理中，更新同一个变量为了保证数据一致性，通常需要做锁处理。
	//使用 IdMap 是因为查询的列表通常需要按时间顺序进行排序，一般数据库查询后的列表已经排过序了，
	// 但是为了减少延时，程序中用了并发，这时候会打乱排序，所以通过 IdMap 来记录并发处理前的顺序，处理后再重新复位。
	userList := model.UserList{
		Lock:  new(sync.Mutex),
		IdMap: make(map[uint64]*model.UserInfo, len(users)),
	}

	//创建通道
	errChan := make(chan error, 1)
	finished := make(chan bool, 1)

	// Improve query efficiency in parallel
	//遍历所有的用户数据，对User数据进行并发处理
	for _, u := range users {
		wg.Add(1)
		go func(u *model.UserModel) {
			defer wg.Done()

			shortId, err := util.GenShortId()
			if err != nil {
				//将err写入通道
				errChan <- err
				return
			}

			userList.Lock.Lock()
			defer userList.Lock.Unlock()
			//更新变量IdMap时加锁
			userList.IdMap[u.Id] = &model.UserInfo{
				Id:        u.Id,
				Username:  u.Username,
				SayHello:  fmt.Sprintf("Hello %s", shortId),
				Password:  u.Password,
				CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
		}(u)
	}

	go func() {
		//等待并发结束
		wg.Wait()
		//关闭通道
		close(finished)
	}()

	//select随机执行一个可运行的case。如果没有case可运行，它将阻塞，直到有case可运行。一个默认的子句应该总是可运行的。
	//如果finish或是出现错误则会退出阻塞
	select {
	case <-finished: //通道关闭会结束阻塞
	case err := <-errChan:
		return nil, count, err
	}

	//恢复原来顺序
	for _, id := range ids {
		infos = append(infos, userList.IdMap[id])
	}

	return infos, count, nil
}
