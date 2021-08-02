package main

type Product struct {
	ID       string   `json:"id"`
	Barcodes []string `json:"barcodes"`
	Title    string   `json:"title"`
	Plu      int      `json:"plu"`
	Cash     int      `json:"cash"`
	IsWeight bool     `json:"isWeight"`
}

type Store struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
