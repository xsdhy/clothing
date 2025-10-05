package entity

// demo表
type DbDemo struct {
	ID        uint  `gorm:"primarykey" json:"id"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`

	UUID string `gorm:"column:uuid;type:varchar(255);uniqueIndex" json:"uuid"`

	Phone             string `gorm:"column:phone;type:varchar(255);index" json:"phone"`
	Nickname          string `gorm:"column:nickname;type:varchar(255)" json:"nickname"`
	LastHeartbeatTime int64  `gorm:"column:last_heartbeat_time;type:int(11);default:0" json:"last_heartbeat_time,omitempty"` //上一次心跳时间，时间戳
}

// TableName 指定表名
func (DbDemo) TableName() string {
	return "demo"
}
