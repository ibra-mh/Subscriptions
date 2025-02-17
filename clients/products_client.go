package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Product represents the structure of the product data.
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// Client is a struct that will hold base URL and HTTP client.
type Client struct {
	BaseURL string
}

// NewClient creates a new client to communicate with the products service.
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) GetOffers() ([]Product, error) {
	url := fmt.Sprintf("%s/offers", c.BaseURL)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching offers: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error: Received non-OK response: %s", body)
		return nil, fmt.Errorf("failed to fetch offers: %v", resp.Status)
	}

	var offers []Product
	if err := json.NewDecoder(resp.Body).Decode(&offers); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil, err
	}

	return offers, nil
}

func (c *Client) GetOfferById(id int) (*Product, error) {
	url := fmt.Sprintf("%s/offers/%d", c.BaseURL, id)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching offer by ID: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error: Received non-OK response: %s", body)
		return nil, fmt.Errorf("failed to fetch offer by id: %v", resp.Status)
	}

	var offer Product
	if err := json.NewDecoder(resp.Body).Decode(&offer); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil, err
	}

	return &offer, nil
}
