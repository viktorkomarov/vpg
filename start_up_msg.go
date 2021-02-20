package vpg

const (
	protocolVersion = int32(196608)
)

type startUpMsg struct {
	Version  int32  `pg_order:"1"`
	User     string `pg_order:"2" pg_preffix:"user"`
	Database string `pg_order:"3" pg_preffix:"database"`
	//Replication
}
