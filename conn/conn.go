package conn

type Conn interface {
    Read(b []byte) (int, error)
    Write(b []byte) (int, error)
}