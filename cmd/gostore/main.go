package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/briandowns/spinner"
	"github.com/cheggaaa/pb/v3"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

type application struct {
	logger    *log.Logger
	spinner   *spinner.Spinner
	fb        *firebase.App
	firestore *firestore.Client
	products  []Product
}

type config struct {
	newProductsFile string
	oldProductsFile string
	storeName       string `yaml:"storeName"`
	opt             option.ClientOption
}

func main() {

	config := &config{}
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	flag.StringVar(&config.newProductsFile, "products", "itemfull.txt", "File with products")
	fbkey := flag.String("key", "fb_key.json", "Firebase json key")

	f, err := ioutil.ReadFile("config.yml")
	if err != nil {
		logger.Fatal("Cant find config.yml")
	}
	err = yaml.Unmarshal(f, &config.storeName)
	if err != nil {
		logger.Fatal("Cant parse config")
	}

	flag.Parse()

	config.oldProductsFile = "itemfull_old.txt"
	config.opt = option.WithCredentialsFile(*fbkey)

	fbApp, err := firebase.NewApp(context.Background(), nil, config.opt)
	if err != nil {
		logger.Fatal(err)
	}

	firestoreClient, err := fbApp.Firestore(context.Background())
	if err != nil {
		logger.Fatal(err)
	}
	defer firestoreClient.Close()

	app := &application{
		logger:    logger,
		spinner:   spinner.New(spinner.CharSets[11], 100*time.Millisecond),
		fb:        fbApp,
		firestore: firestoreClient,
	}

	products := app.createProductsToUpload(config)

	if len(products) > 0 {
		bar := pb.StartNew(len(products))

		for _, product := range products {
			_, err = app.firestore.Collection("stores").Doc(config.storeName).Collection("products").Doc(product.ID).Set(context.Background(), product)
			if err != nil {
				app.logger.Println(err)
			}
			bar.Increment()
		}
		bar.Finish()

	} else {
		logger.Println("no new products to upload")
	}

}
