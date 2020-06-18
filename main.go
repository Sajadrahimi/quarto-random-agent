package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Gimulator/client-go"
	uuid "github.com/satori/go.uuid"
)

var (
	name string = "rando-agent" + uuid.NewV4().String()[0:5]
)

func main() {
	rand.Seed(time.Now().UnixNano())

	a, err := newAgent()
	if err != nil {
		panic(err)
	}
	a.listen()
}

type agent struct {
	*client.Client

	ch chan client.Object
}

func newAgent() (*agent, error) {
	ch := make(chan client.Object)

	cli, err := client.NewClient(ch)
	if err != nil {
		return nil, err
	}

	if err := cli.Set(client.Key{
		Namespace: "quarto",
		Name:      name,
		Type:      "register",
	}, ""); err != nil {
		return nil, err
	}

	if err := cli.Watch(client.Key{
		Namespace: "quarto",
		Name:      "board",
		Type:      "verdict",
	}); err != nil {
		return nil, err
	}

	return &agent{
		ch:     ch,
		Client: cli,
	}, nil
}

func (a *agent) listen() {
	for {
		obj := <-a.ch

		data, ok := obj.Value.(string)
		if !ok {
			fmt.Println("could not cast value to []byte")
			continue
		}

		board := Board{}
		err := json.Unmarshal([]byte(data), &board)
		if err != nil {
			fmt.Println("could not unmarshal data to board struct:", err.Error())
			continue
		}

		if board.Turn != name {
			continue
		}

		if err := a.action(board); err != nil {
			fmt.Println("could not execute action:", err.Error())
			continue
		}
	}
}

func (a *agent) action(board Board) error {
	avPieces := make([]int, 0)
	avPositions := make([]Position, 0)
	fmt.Println("starting making an action...")

	for id := range board.Pieces {
		isUsed := false
		for _, pos := range board.Positions {
			if pos.Piece == id {
				isUsed = true
			}
		}
		if !isUsed && id != board.Picked {
			avPieces = append(avPieces, id)
		}
	}

	for _, pos := range board.Positions {
		if pos.Piece == 0 {
			avPositions = append(avPositions, pos)
		}
	}

	n := rand.Intn(len(avPieces))
	fmt.Printf("%d hase been chosen\n", n)
	if n == 0 {
		fmt.Println("no available positions, exiting.")
		os.Exit(0)
	}


	ac := Action{
		Picked: avPieces[n],
		X:      avPositions[n].X,
		Y:      avPositions[n].Y,
	}

	b, err := json.Marshal(ac)
	if err != nil {
		fmt.Println("couldn't marshall the action", err.Error())
		return err
	}

	if err := a.Set(client.Key{
		Namespace: "quarto",
		Name:      name,
		Type:      "action",
	}, string(b)); err != nil {
		fmt.Println("couldn't set the action", err.Error())
		return err
	}

	return nil
}

//********************************* types
type Board struct {
	Pieces    map[int]Piece
	Positions []Position
	Turn      string
	Picked    int
}

type Position struct {
	X     int
	Y     int
	Piece int
}

type Piece struct {
	Length string
	Shape  string
	Color  string
	Hole   string
}

type Action struct {
	Picked int
	X      int
	Y      int
}
