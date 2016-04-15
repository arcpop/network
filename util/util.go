//Package util contains utility functions needed in the whole project.
package util



//ChannelClosed returns true if the stop channel is closed and false otherwise.
func ChannelClosed(stop chan bool) bool {
    select {
        case _, ok := <- stop:
            return ok
    }
}
