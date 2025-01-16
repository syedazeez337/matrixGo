package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
)

// TerminalState holds the terminal size and the map to render
type TerminalState struct {
	Rows int
	Cols int
	Grid []rune
}

// InitializeTerminal gets the terminal size and initializes the grid
func InitializeTerminal() (*TerminalState, error) {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return nil, err
	}

	grid := make([]rune, rows*cols)
	for i := range grid {
		grid[i] = ' '
	}

	return &TerminalState{Rows: rows, Cols: cols, Grid: grid}, nil
}

// Resize adjusts the grid when the terminal is resized
func (t *TerminalState) Resize() error {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	if rows != t.Rows || cols != t.Cols {
		t.Rows = rows
		t.Cols = cols
		t.Grid = make([]rune, rows*cols)
		for i := range t.Grid {
			t.Grid[i] = ' '
		}
	}
	return nil
}

// UpdateGrid updates the grid with falling characters
func (t *TerminalState) UpdateGrid(rng *rand.Rand) {
    for row := t.Rows - 1; row >= 0; row-- {
        for col := 0; col < t.Cols; col++ {
            index := row*t.Cols + col

            if row < t.Rows-1 && t.Grid[index] != ' ' {
                below := (row+1)*t.Cols + col
                t.Grid[below] = t.Grid[index]
                t.Grid[index] = ' '
            }

            if row == 0 && rng.Intn(10) < 2 {
                t.Grid[index] = getRandomChar(rng)
            }
        }
    }
}

// RenderGrid draws the grid to the terminal
func (t *TerminalState) RenderGrid() {
	// Move cursor to top-left
	fmt.Print("\033[H")
	for i, cell := range t.Grid {
		if i > 0 && i%t.Cols == 0 {
			fmt.Print("\n")
		}
		if cell != ' ' {
			fmt.Printf("\033[92m%c\033[0m", cell) // Green characters
		} else {
			fmt.Print(" ")
		}
	}
}

// getRandomChar generates a random character
func getRandomChar(rng *rand.Rand) rune {
    chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    return rune(chars[rng.Intn(len(chars))])
}

// ClearScreen clears the terminal
func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

// HandleExit cleans up on exit
func HandleExit() {
	fmt.Print("\033[?25h") // Show cursor
	ClearScreen()
	fmt.Println("Exiting...")
	os.Exit(0)
}

func main() {
    rng := rand.New(rand.NewSource(time.Now().UnixNano())) // Local random generator

    state, err := InitializeTerminal()
    if err != nil {
        fmt.Println("Error initializing terminal:", err)
        return
    }
    defer HandleExit()

    fmt.Print("\033[?25l")
    ClearScreen()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigs
        HandleExit()
    }()

    ticker := time.NewTicker(66 * time.Millisecond)
    defer ticker.Stop()

    for range ticker.C {
        if err := state.Resize(); err != nil {
            fmt.Println("Error resizing terminal:", err)
            break
        }

        state.UpdateGrid(rng)
        state.RenderGrid()
    }
}
