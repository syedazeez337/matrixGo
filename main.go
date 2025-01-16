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
func (t *TerminalState) UpdateGrid() {
	for row := t.Rows - 1; row >= 0; row-- {
		for col := 0; col < t.Cols; col++ {
			index := row*t.Cols + col

			// Handle falling characters
			if row < t.Rows-1 && t.Grid[index] != ' ' {
				below := (row+1)*t.Cols + col
				t.Grid[below] = t.Grid[index]
				t.Grid[index] = ' '
			}

			// Generate new characters at the top row
			if row == 0 && rand.Intn(10) < 2 { // 20% chance of new char
				t.Grid[index] = getRandomChar()
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
func getRandomChar() rune {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return rune(chars[rand.Intn(len(chars))])
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
	rand.Seed(time.Now().UnixNano())

	// Initialize terminal state
	state, err := InitializeTerminal()
	if err != nil {
		fmt.Println("Error initializing terminal:", err)
		return
	}
	defer HandleExit()

	// Hide cursor and clear screen
	fmt.Print("\033[?25l")
	ClearScreen()

	// Signal handling for graceful exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		HandleExit()
	}()

	// Main loop
	ticker := time.NewTicker(66 * time.Millisecond) // ~15 FPS
	defer ticker.Stop()

	for range ticker.C {
		// Resize terminal if necessary
		if err := state.Resize(); err != nil {
			fmt.Println("Error resizing terminal:", err)
			break
		}

		// Update and render the grid
		state.UpdateGrid()
		state.RenderGrid()
	}
}
