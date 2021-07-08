package main

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"flag"
	"github.com/briandowns/spinner"
	"github.com/cheggaaa/pb/v3"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
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
	opt             option.ClientOption
}

func main() {

	config := &config{}

	flag.StringVar(&config.newProductsFile, "products", "itemfull.txt", "File with products")
	fbkey := flag.String("key", "fb_key.json", "Firebase json key")

	flag.Parse()

	config.oldProductsFile = "itemfull_old.txt"
	config.opt = option.WithCredentialsFile(*fbkey)

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

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
			_, err = app.firestore.Collection("products").Doc(product.ID).Set(context.Background(), product)
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
