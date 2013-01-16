go-fsm
======

Simple Finite-State Machines for Go.

View the docs [here](http://godoc.org/github.com/vaughan0/go-fsm).

Example
-------

Here is a simple example that simulates a secure door:

```go
import "github.com/vaughan0/go-fsm"
import "fmt"

type Door struct {
  // The code that was used to lock the door with
  code string
}

// These are the states that a Door can be in:
var locked, unlocked fsm.State

// Initialize the states
locked = fsm.Actions{
  "enter-pin": function(d *Door, pin string) fsm.State {
    if pin == d.code {
      fmt.Println("PIN correct")
      // Change to the unlocked state
      return unlocked
    }
    fmt.Println("Incorrect pin")
    // Stay in the locked state
    return nil
  },
  "turn-handle": function(d *Door) {
    fmt.Println("You can't open a locked door!")
  },
}
unlocked = fsm.Actions{
  "enter-pin": function(d *Door, pin string) fsm.State {
    d.code = pin
    fmt.Println("Locked the door")
    return locked
  },
  "turn-handle": function(d *Door) {
    fmt.Println("Door open")
  },
}

// Create a new FSM to represent the door
door := fsm.New(new(Door), unlocked)

door.Trigger("enter-pin", "1234") // Locked the door
door.Trigger("turn-handle")       // You can't open a locked door!
door.Trigger("enter-pin", "4321") // Incorrect pin
door.Trigger("enter-pin", "1234") // PIN correct
door.Trigger("turn-handle")       // Door open
```
