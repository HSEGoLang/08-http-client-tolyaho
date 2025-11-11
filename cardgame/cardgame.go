//go:build !solution

package cardgame

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://deckofcardsapi.com/api/deck"

type Client struct {
	baseURL string
	client  *http.Client
	output  io.Writer
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		client:  http.DefaultClient,
	}
}

func (c *Client) PlayGame(userGuess int) (bool, error) {
	resp, err := c.client.Get(c.baseURL + "/new/shuffle/?deck_count=1")
	if err != nil {
		return false, fmt.Errorf("create deck: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("create deck: status %d", resp.StatusCode)
	}

	var deck DeckResponse
	if err := json.NewDecoder(resp.Body).Decode(&deck); err != nil {
		return false, fmt.Errorf("decode deck: %w", err)
	}
	if deck.DeckID == "" {
		return false, errors.New("empty deck id")
	}

	rounds := 0
	for {
		drawResp, err := c.client.Get(fmt.Sprintf("%s/%s/draw/?count=1", c.baseURL, deck.DeckID))
		if err != nil {
			return false, fmt.Errorf("draw card: %w", err)
		}
		if drawResp.StatusCode != http.StatusOK {
			drawResp.Body.Close()
			return false, fmt.Errorf("draw card: status %d", drawResp.StatusCode)
		}

		var draw DrawResponse
		if err := json.NewDecoder(drawResp.Body).Decode(&draw); err != nil {
			drawResp.Body.Close()
			return false, fmt.Errorf("decode draw: %w", err)
		}
		drawResp.Body.Close()

		if len(draw.Cards) == 0 {
			return false, errors.New("no cards returned")
		}

		card := draw.Cards[0]
		rounds++
		c.printf("%s of %s\n", card.Value, card.Suit)

		if card.Value == "QUEEN" {
			break
		}
		if draw.Remaining == 0 {
			return false, errors.New("queen not found")
		}
	}

	if rounds == userGuess {
		c.printf("Вы угадали!\n")
		return true, nil
	}

	c.printf("Вы проиграли! Правильный ответ: %d\n", rounds)
	return false, nil
}

func (c *Client) printf(format string, args ...interface{}) {
	if c.output != nil {
		fmt.Fprintf(c.output, format, args...)
		return
	}
	fmt.Printf(format, args...)
}

func PlayGame(userGuess int) (bool, error) {
	return NewClient().PlayGame(userGuess)
}
