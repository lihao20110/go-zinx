package ziface

//IConnManager 连接管理抽象层
type IConnManager interface {
	Add(conn IConnection)                   //添加连接
	Remove(conn IConnection)                //删除连接
	Get(connID uint32) (IConnection, error) //利用ConnID获取连接
	Len() int                               //获取当前连接数量
	ClearConn()                             //停止并删除全部的连接信息
}
