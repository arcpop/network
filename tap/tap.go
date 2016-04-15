package tap

//Tap specifies the necessary methods which are needed to use a tap interface
type Tap interface {
    GetMTU() (n int)
    Read(b []byte) (err error)
    Write(b []byte) (err error)
    Close() (err error)
}



