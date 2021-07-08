package main

import (
	"bufio"
	"github.com/briandowns/spinner"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
)

func (app *application) parseFile(filename string) []Product {
	s := spinner.New(spinner.CharSets[25], 100*time.Millisecond)
	s.Suffix = "parsing files"
	s.Start()

	data, err := os.OpenFile(filename, os.O_RDONLY, 0444)
	if err != nil {
		app.logger.Fatal(err)
	}

	decoder := charmap.Windows1251.NewDecoder()
	reader := decoder.Reader(data)

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var productLines []string

	for scanner.Scan() {
		productLines = append(productLines, scanner.Text())
	}

	var products []Product

	for _, line := range productLines {
		p := createProduct(line)
		products = append(products, p)
	}

	app.logger.Println("file successfully parsed")
	s.Stop()
	return app.mergeProductsById(&products)
}

func (app *application) mergeProductsById(products *[]Product) []Product {
	s := spinner.New(spinner.CharSets[25], 100*time.Millisecond)
	s.Suffix = "merging products"
	s.Start()
	mergedProducts := []Product{}
	for _, product := range *products {
		index := findProductInProduct(&mergedProducts, product.ID)
		if index != -1 {
			mergedProducts[index].Barcodes = append(mergedProducts[index].Barcodes, product.Barcodes...)
		} else {
			mergedProducts = append(mergedProducts, product)
		}

	}
	s.Stop()
	return mergedProducts
}

func (app *application) copyFile(original, destination string) {
	data, err := ioutil.ReadFile(original)
	if err != nil {
		app.logger.Fatal(err)
	}

	err = ioutil.WriteFile(destination, data, 0666)
	if err != nil {
		app.logger.Fatal(err)
	}
}

func findProductInProduct(products *[]Product, id string) int {
	for index, product := range *products {
		if product.ID == id {
			return index
		}
	}
	return -1
}

func createProduct(productLine string) Product {
	line := strings.Split(productLine, ",")
	product := Product{}

	product.ID = line[0]
	product.Barcodes = []string{line[1]}
	product.Title = line[3]
	product.Plu = line[7]
	product.Cash = line[2]

	if line[7] == "0" {
		product.IsWeight = false
	} else {
		product.IsWeight = true
	}

	return product
}

func (app *application) filterNewProducts(newProducts *[]Product, oldProducts *[]Product) []Product {
	s := spinner.New(spinner.CharSets[25], 100*time.Millisecond)
	s.Suffix = "filtering products"
	s.Start()
	var filteredProducts []Product
	for _, product := range *newProducts {
		if !checkIfProductContains(oldProducts, product) {
			filteredProducts = append(filteredProducts, product)
		}
	}
	s.Stop()
	app.logger.Println("products successfully filtered")
	return filteredProducts

}

func checkIfProductContains(products *[]Product, product Product) bool {
	for _, p := range *products {
		if p.ID == product.ID && reflect.DeepEqual(p, product) {
			return true
		}
	}
	return false
}

func (app *application) createProductsToUpload(config *config) []Product {
	newProducts := app.parseFile(config.newProductsFile)
	var oldProducts []Product
	if _, err := os.Stat(config.oldProductsFile); os.IsNotExist(err) {
		oldProducts = []Product{}
	} else {
		oldProducts = app.parseFile(config.oldProductsFile)
	}

	productsToUpload := app.filterNewProducts(&newProducts, &oldProducts)
	app.copyFile(config.newProductsFile, config.oldProductsFile)
	return productsToUpload
}
