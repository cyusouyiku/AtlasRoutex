package neo4j

type Driver struct {
DSN string
}

func NewDriver(dsn string) *Driver { return &Driver{DSN: dsn} }

func (d *Driver) Close() error { return nil }
